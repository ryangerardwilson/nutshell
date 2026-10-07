package compiler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

const maxSourceBytes = 1024 * 1024

var includeLine = regexp.MustCompile(`(?i)^\s*Include\s+"([^"]+)"\.?\s*$`)

type Source struct {
	Path string `json:"path"`
	Text string `json:"text"`
}

type Program struct {
	Root    string
	Entry   string
	Sources []Source
}

// localPath rejects escapes and symlink components, including dangling links.
func localPath(root, name string) (string, error) {
	if filepath.IsAbs(name) || !filepath.IsLocal(name) {
		return "", fmt.Errorf("path must stay inside the program directory: %q", name)
	}
	p := root
	for _, part := range strings.Split(filepath.Clean(name), string(filepath.Separator)) {
		p = filepath.Join(p, part)
		info, err := os.Lstat(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlinks are not allowed: %s", p)
		}
	}
	return p, nil
}

func Load(entry string) (Program, error) {
	abs, err := filepath.Abs(entry)
	if err != nil {
		return Program{}, err
	}
	root, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if err != nil {
		return Program{}, err
	}
	p := Program{Root: root, Entry: filepath.Base(abs)}
	states := map[string]int{}
	total := 0
	var visit func(string) error
	visit = func(name string) error {
		name = filepath.Clean(name)
		if filepath.Ext(name) != ".nut" {
			return fmt.Errorf("source must have a .nut extension: %s", name)
		}
		if states[name] == 1 {
			return fmt.Errorf("include cycle at %s", name)
		}
		if states[name] == 2 {
			return nil
		}
		if len(states) >= 128 {
			return fmt.Errorf("too many source files (maximum 128)")
		}
		path, err := localPath(root, name)
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if !info.Mode().IsRegular() || info.Size() > maxSourceBytes {
			return fmt.Errorf("%s: expected regular source under 1 MiB", name)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		total += len(data)
		if total > maxSourceBytes {
			return fmt.Errorf("combined source exceeds 1 MiB")
		}
		if !utf8.Valid(data) || len(bytes.TrimSpace(data)) == 0 {
			return fmt.Errorf("%s: source must be nonempty UTF-8", name)
		}
		states[name] = 1
		p.Sources = append(p.Sources, Source{filepath.ToSlash(name), string(data)})
		for n, line := range strings.Split(string(data), "\n") {
			if m := includeLine.FindStringSubmatch(line); m != nil {
				if !filepath.IsLocal(m[1]) {
					return fmt.Errorf("%s:%d: include must be a local relative path", name, n+1)
				}
				if err := visit(filepath.Join(filepath.Dir(name), m[1])); err != nil {
					return fmt.Errorf("%s:%d: %w", name, n+1, err)
				}
			}
		}
		states[name] = 2
		return nil
	}
	if err := visit(p.Entry); err != nil {
		return Program{}, err
	}
	return p, nil
}

func (p Program) unchanged() error {
	for _, s := range p.Sources {
		path, err := localPath(p.Root, filepath.FromSlash(s.Path))
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != s.Text {
			return fmt.Errorf("source changed during compilation: %s", s.Path)
		}
	}
	return nil
}
