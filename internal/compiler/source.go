package compiler

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const maxSourceBytes = 1024 * 1024

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
	if filepath.Ext(p.Entry) != ".nut" {
		return Program{}, fmt.Errorf("source must have a .nut extension: %s", p.Entry)
	}
	paths := []string{p.Entry}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") || excludedSourceDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(entry.Name()) != ".nut" {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel != p.Entry {
			paths = append(paths, rel)
		}
		if len(paths) > 128 {
			return fmt.Errorf("too many source files (maximum 128)")
		}
		return nil
	})
	if err != nil {
		return Program{}, err
	}
	sort.Strings(paths[1:])
	total := 0
	for _, name := range paths {
		path, err := localPath(root, name)
		if err != nil {
			return Program{}, err
		}
		info, err := os.Stat(path)
		if err != nil {
			return Program{}, fmt.Errorf("%s: %w", name, err)
		}
		if !info.Mode().IsRegular() || info.Size() > maxSourceBytes {
			return Program{}, fmt.Errorf("%s: expected regular source under 1 MiB", name)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return Program{}, err
		}
		total += len(data)
		if total > maxSourceBytes {
			return Program{}, fmt.Errorf("combined source exceeds 1 MiB")
		}
		if !utf8.Valid(data) || len(bytes.TrimSpace(data)) == 0 {
			return Program{}, fmt.Errorf("%s: source must be nonempty UTF-8", name)
		}
		p.Sources = append(p.Sources, Source{filepath.ToSlash(name), string(data)})
	}
	return p, nil
}

func excludedSourceDirectory(name string) bool {
	switch name {
	case "node_modules", "vendor", "target", "dist", "build":
		return true
	}
	return false
}

func (p Program) unchanged() error {
	current, err := Load(filepath.Join(p.Root, p.Entry))
	if err != nil || SourceHash(current) != SourceHash(p) {
		return fmt.Errorf("source changed during compilation: %s", p.Root)
	}
	return nil
}
