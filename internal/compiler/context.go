package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Version 0.9 generated this resource; new builds remove it and its embedding hooks.
const legacyProvenanceResource = ".nutshell-provenance.bin"

type sourceSnapshot struct {
	Path   string            `json:"path"`
	Exists bool              `json:"exists"`
	Files  map[string]string `json:"files,omitempty"`
}

func snapshotSource(path string) (sourceSnapshot, error) {
	s := sourceSnapshot{Path: path}
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return s, nil
	} else if err != nil {
		return s, err
	}
	s.Exists = true
	files, err := sourceFiles(path, "")
	s.Files = files
	return s, err
}

func (s sourceSnapshot) unchanged() error {
	current, err := snapshotSource(s.Path)
	if err != nil {
		return fmt.Errorf("source changed during compilation at %s: %w", s.Path, err)
	}
	if !reflect.DeepEqual(s, current) {
		return fmt.Errorf("source changed during compilation: %s", s.Path)
	}
	return nil
}

type sourceContext struct {
	Input       sourceSnapshot `json:"input"`
	Destination sourceSnapshot `json:"destination"`
}

func captureContext(selected, output string) (sourceContext, error) {
	if strings.TrimSpace(selected) == "" {
		return sourceContext{}, fmt.Errorf("choose an implementation source directory with -s <path>")
	}
	abs, err := filepath.Abs(selected)
	if err != nil {
		return sourceContext{}, err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if err != nil {
		return sourceContext{}, fmt.Errorf("parent of -s directory must exist: %w", err)
	}
	abs = filepath.Join(parent, filepath.Base(abs))
	if rel, err := filepath.Rel(abs, output); err == nil && filepath.IsLocal(rel) {
		return sourceContext{}, fmt.Errorf("executable output must be outside the -s source directory")
	}
	dest, err := snapshotSource(abs)
	if err != nil {
		return sourceContext{}, err
	}
	return sourceContext{Input: dest, Destination: dest}, nil
}

func (c sourceContext) seed(dir string) error {
	if c.Input.Exists {
		// Keep an untouched original, then copy the working tree from it.
		original := filepath.Join(dir, "original-src")
		if err := os.Mkdir(original, 0700); err != nil {
			return err
		}
		files, err := sourceFiles(c.Input.Path, original)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(files, c.Input.Files) {
			return fmt.Errorf("source changed while preparing context: %s", c.Input.Path)
		}
		working := filepath.Join(dir, "src")
		if err := os.Mkdir(working, 0700); err != nil {
			return err
		}
		if _, err := sourceFiles(original, working); err != nil {
			return err
		}
	}
	if err := c.unchanged(); err != nil {
		return err
	}
	return writeJSON(filepath.Join(dir, "source-context.json"), c)
}

func (c sourceContext) unchanged() error {
	if err := c.Input.unchanged(); err != nil {
		return err
	}
	if c.Input.Path != c.Destination.Path {
		return c.Destination.unchanged()
	}
	return nil
}

func hasImplementation(files map[string]string) bool {
	for name, fingerprint := range files {
		if name == legacyProvenanceResource || name == sourceReceipt || filepath.Ext(name) == ".nut" {
			continue
		}
		if strings.HasPrefix(fingerprint, "file:") {
			return true
		}
	}
	return false
}

// A .nut input can itself be inside src/. Publishing implementation must not
// accidentally replace that authoritative source with an AI-edited copy.
func protectNutInputs(p Program, source, dest string) error {
	for _, input := range p.Sources {
		rel, err := filepath.Rel(dest, filepath.Join(p.Root, filepath.FromSlash(input.Path)))
		if err != nil || !filepath.IsLocal(rel) {
			continue
		}
		path, err := localPath(source, rel)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != input.Text {
			return fmt.Errorf("generated src would replace .nut input: %s", input.Path)
		}
	}
	return nil
}
