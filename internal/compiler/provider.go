package compiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Adapter commands are argv, never implicitly passed through a shell.
type Adapter struct {
	Command    []string `json:"command"`
	UnsafeArgs []string `json:"unsafe_args"`
	Prompt     string   `json:"prompt"`
}

func Resolve(name string) (Adapter, error) {
	var a Adapter
	switch name {
	case "codex":
		a = Adapter{[]string{"codex", "exec", "--skip-git-repo-check", "--ephemeral", "--color", "never", "-"}, []string{"--dangerously-bypass-approvals-and-sandbox"}, "stdin"}
	case "grok":
		a = Adapter{[]string{"grok", "--no-plan", "--prompt-file", "{prompt_file}"}, []string{"--always-approve", "--sandbox", "off"}, "file"}
	case "claude":
		a = Adapter{[]string{"claude", "--print", "--output-format", "text", "--no-session-persistence"}, []string{"--dangerously-skip-permissions"}, "stdin"}
	default:
		dir, err := os.UserConfigDir()
		if err != nil {
			return a, err
		}
		path := filepath.Join(dir, "nutshell", "interpreters.json")
		var config struct {
			Interpreters map[string]Adapter `json:"interpreters"`
		}
		if err := readJSON(path, &config); err != nil {
			return a, fmt.Errorf("unknown compilation agent %q; configure it in %s: %w", name, path, err)
		}
		var ok bool
		a, ok = config.Interpreters[name]
		if !ok {
			return a, fmt.Errorf("compilation agent %q is not configured in %s", name, path)
		}
	}
	if err := a.validate(); err != nil {
		return a, fmt.Errorf("compilation agent %s: %w", name, err)
	}
	resolved, err := exec.LookPath(a.Command[0])
	if err != nil {
		return a, fmt.Errorf("compilation agent %q is not installed or not on PATH: %w", a.Command[0], err)
	}
	a.Command[0] = resolved
	return a, nil
}

func (a Adapter) validate() error {
	if len(a.Command) == 0 || a.Command[0] == "" {
		return fmt.Errorf("command must be a nonempty argv array")
	}
	if len(a.UnsafeArgs) == 0 {
		return fmt.Errorf("unsafe_args must explicitly specify the tool's unattended full-permission flags")
	}
	switch a.Prompt {
	case "stdin":
	case "file", "argument":
		placeholder := "{prompt_file}"
		if a.Prompt == "argument" {
			placeholder = "{prompt}"
		}
		found := false
		for _, arg := range a.Command {
			if strings.Contains(arg, placeholder) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("%s prompt requires %s in command", a.Prompt, placeholder)
		}
	default:
		return fmt.Errorf("prompt must be stdin, file, or argument")
	}
	return nil
}

func (a Adapter) invocation(prompt, promptFile string) ([]string, io.Reader) {
	argv := []string{a.Command[0]}
	positioned := false
	for _, arg := range a.Command[1:] {
		if arg == "{unsafe_args}" {
			positioned = true
		}
	}
	if !positioned {
		argv = append(argv, a.UnsafeArgs...)
	}
	for _, arg := range a.Command[1:] {
		if arg == "{unsafe_args}" {
			argv = append(argv, a.UnsafeArgs...)
		} else {
			argv = append(argv, arg)
		}
	}
	for i := 1; i < len(argv); i++ {
		argv[i] = strings.ReplaceAll(argv[i], "{prompt_file}", promptFile)
		if a.Prompt == "argument" {
			argv[i] = strings.ReplaceAll(argv[i], "{prompt}", prompt)
		}
	}
	if a.Prompt == "stdin" {
		return argv, strings.NewReader(prompt)
	}
	return argv, nil
}

func readJSON(path string, dst any) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("expected a regular JSON file: %s", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxSourceBytes+1))
	if err != nil {
		return err
	}
	if len(data) > maxSourceBytes {
		return fmt.Errorf("JSON exceeds 1 MiB")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("expected one JSON object")
	}
	return nil
}
