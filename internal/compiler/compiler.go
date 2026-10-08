package compiler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Options struct {
	FineTune    string
	Entry       string
	Interpreter string
	SourceDir   string
	Output      string
	Log         io.Writer
	OnProgress  func(Progress)
}

type Manifest struct {
	Version     int        `json:"version"`
	Language    string     `json:"language"`
	Summary     string     `json:"summary"`
	Assumptions []string   `json:"assumptions"`
	Artifact    string     `json:"artifact"`
	Build       [][]string `json:"build"`
	Test        [][]string `json:"test"`
}

type Result struct {
	Version     int       `json:"version"`
	Status      string    `json:"status"`
	Entry       string    `json:"entry"`
	Interpreter string    `json:"interpreter"`
	Directory   string    `json:"directory"`
	Output      string    `json:"output,omitempty"`
	Started     time.Time `json:"started"`
	Finished    time.Time `json:"finished"`
	Manifest    *Manifest `json:"manifest,omitempty"`
	Error       string    `json:"error,omitempty"`
}

type lockedWriter struct {
	sync.Mutex
	w io.Writer
}

func (w *lockedWriter) Write(p []byte) (int, error) { w.Lock(); defer w.Unlock(); return w.w.Write(p) }

func Compile(ctx context.Context, opts Options) (result Result, err error) {
	if opts.Log == nil {
		opts.Log = io.Discard
	}
	opts.progress("preparing", "Reading your program", 0)
	p, err := Load(opts.Entry)
	if err != nil {
		return result, err
	}
	output, err := outputPath(p, opts.Output)
	if err != nil {
		return result, err
	}
	baseline, err := captureContext(opts.SourceDir, output)
	if err != nil {
		return result, err
	}
	if opts.FineTune != "" {
		if strings.TrimSpace(opts.FineTune) == "" {
			return result, fmt.Errorf("fine tuning requires a nonblank change request")
		}
		if !hasImplementation(baseline.Input.Files) {
			return result, fmt.Errorf("fine tuning requires existing implementation files in -s; compile once without -f first")
		}
	}
	a, err := Resolve(opts.Interpreter)
	if err != nil {
		return result, err
	}
	if _, err := exec.LookPath("rgw-ast"); err != nil {
		return result, fmt.Errorf("compilation requires rgw-ast: make an executable rgw-ast available on PATH before compiling: %w", err)
	}
	dir, err := os.MkdirTemp("/tmp", "nutshell-build-")
	if err != nil {
		return result, err
	}
	result = Result{Version: 1, Status: "failed", Entry: filepath.Join(p.Root, p.Entry), Interpreter: opts.Interpreter, Directory: dir, Started: time.Now().UTC()}
	defer func() {
		result.Finished = time.Now().UTC()
		if err != nil {
			result.Error = err.Error()
		}
		if writeErr := writeJSON(filepath.Join(dir, "result.json"), result); writeErr != nil && err == nil {
			err = fmt.Errorf("write result report: %w", writeErr)
		}
	}()
	opts.progress("understanding", "Starting AI compilation session", 1)
	for _, s := range p.Sources {
		dest := filepath.Join(dir, "source", filepath.FromSlash(s.Path))
		if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			return result, err
		}
		if err = os.WriteFile(dest, []byte(s.Text), 0600); err != nil {
			return result, err
		}
	}
	if err = writeJSON(filepath.Join(dir, "source.json"), p.Sources); err != nil {
		return result, err
	}
	if baseline.Input.Exists {
		opts.progress("understanding", "Loading existing implementation source", 1)
	}
	if err = baseline.seed(dir); err != nil {
		return result, err
	}
	if err = os.MkdirAll(filepath.Join(dir, "src"), 0700); err != nil {
		return result, err
	}
	_, legacyProvenance := baseline.Input.Files[legacyProvenanceResource]
	prompt := promptFor(p, opts.FineTune, legacyProvenance)
	promptFile := filepath.Join(dir, "prompt.txt")
	if err = os.WriteFile(promptFile, []byte(prompt), 0600); err != nil {
		return result, err
	}
	log, err := os.OpenFile(filepath.Join(dir, "interpreter.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return result, err
	}
	argv, stdin := a.invocation(prompt, promptFile)
	stopProgress := watchProgress(ctx, dir, opts.OnProgress)
	runErr := run(ctx, dir, argv, stdin, &lockedWriter{w: log})
	stopProgress()
	closeErr := log.Close()
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if opts.FineTune != "" {
		if err = checkFineTuneReview(dir, p); err != nil {
			// Preserve provider failure details without losing source conflict diagnostics.
			if runErr != nil {
				err = errors.Join(err, fmt.Errorf("compilation agent failed: %w", runErr))
			}
			return result, err
		}
	}
	if runErr != nil {
		return result, fmt.Errorf("compilation agent failed: %w", runErr)
	}
	if closeErr != nil {
		return result, closeErr
	}
	var manifest Manifest
	manifestPath, err := localPath(dir, "build.json")
	if err != nil {
		return result, err
	}
	if err = readJSON(manifestPath, &manifest); err != nil {
		return result, fmt.Errorf("compilation agent must produce build.json: %w", err)
	}
	if err = manifest.validate(); err != nil {
		return result, err
	}
	result.Manifest = &manifest
	verification, err := os.OpenFile(filepath.Join(dir, "verification.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return result, err
	}
	defer verification.Close()
	writer := &lockedWriter{w: verification}
	for _, stage := range []struct {
		name     string
		commands [][]string
	}{{"build", manifest.Build}, {"test", manifest.Test}} {
		step := 5
		label := "Building the native executable"
		if stage.name == "test" {
			step, label = 6, "Running generated program tests"
		}
		opts.progress(stage.name, label, step)
		for _, command := range stage.commands {
			encoded, _ := json.Marshal(command)
			fmt.Fprintf(writer, "\n%s: %s\n", stage.name, encoded)
			if err = run(ctx, dir, command, nil, writer); err != nil {
				return result, fmt.Errorf("%s failed: %w", stage.name, err)
			}
		}
	}
	opts.progress("publishing", "Verifying the compiled executable", 7)
	artifact, err := localPath(dir, manifest.Artifact)
	if err != nil {
		return result, err
	}
	if err = nativeArtifact(artifact); err != nil {
		return result, err
	}
	if err = p.unchanged(); err != nil {
		return result, err
	}
	// Recheck the destination after the external interpreter has finished.
	if _, err = outputPath(p, output); err != nil {
		return result, err
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if err = protectNutInputs(p, filepath.Join(dir, "src"), baseline.Destination.Path); err != nil {
		return result, err
	}
	if err = publishOutputs(ctx, artifact, filepath.Join(dir, "src"), output, baseline); err != nil {
		return result, fmt.Errorf("publish: %w", err)
	}
	result.Status, result.Output = "succeeded", output
	return result, nil
}

func (m Manifest) validate() error {
	if m.Version != 1 {
		return fmt.Errorf("build.json version must be 1")
	}
	if strings.TrimSpace(m.Language) == "" || strings.TrimSpace(m.Summary) == "" {
		return fmt.Errorf("build.json requires language and summary")
	}
	if !filepath.IsLocal(m.Artifact) || m.Artifact == "." {
		return fmt.Errorf("artifact must be a relative file inside the build directory")
	}
	for name, commands := range map[string][][]string{"build": m.Build, "test": m.Test} {
		if len(commands) == 0 || len(commands) > 128 {
			return fmt.Errorf("%s must contain 1–128 commands", name)
		}
		for _, cmd := range commands {
			if len(cmd) == 0 || strings.TrimSpace(cmd[0]) == "" {
				return fmt.Errorf("%s contains an empty command", name)
			}
			for _, arg := range cmd {
				if strings.ContainsRune(arg, 0) {
					return fmt.Errorf("command contains NUL")
				}
			}
		}
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}
