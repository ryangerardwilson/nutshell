package compiler

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type previousContext struct {
	Exists     bool        `json:"exists"`
	Binary     string      `json:"binary,omitempty"`
	Provenance *Provenance `json:"provenance,omitempty"`
	Diff       *SourceDiff `json:"diff,omitempty"`
	Reason     string      `json:"reason,omitempty"`
}

func stagePrevious(dir, output string, current Program) error {
	c := previousContext{}
	f, err := os.Open(output)
	if errors.Is(err, os.ErrNotExist) {
		c.Reason = "No previous output exists."
		return writeJSON(filepath.Join(dir, "previous-context.json"), c)
	}
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("previous output must be a regular file")
	}
	c.Exists, c.Binary = true, "previous-program"
	copyPath := filepath.Join(dir, c.Binary)
	copyFile, err := os.OpenFile(copyPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0400)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(copyFile, f)
	closeErr := copyFile.Close()
	if err := errors.Join(copyErr, closeErr); err != nil {
		return err
	}
	p, err := Inspect(copyPath)
	if err != nil {
		c.Reason = err.Error()
	} else {
		d := DiffSources(p, current)
		c.Provenance, c.Diff = &p, &d
	}
	return writeJSON(filepath.Join(dir, "previous-context.json"), c)
}
