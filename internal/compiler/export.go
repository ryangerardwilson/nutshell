package compiler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

const sourceReceipt = ".nutshell-generated.json"

func sourceFiles(root, copyTo string) (map[string]string, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("implementation source must be a regular directory: %s", root)
	}
	files := map[string]string{".": fmt.Sprintf("dir:%o", info.Mode().Perm())}
	if copyTo != "" {
		if err := os.Chmod(copyTo, info.Mode().Perm()); err != nil {
			return nil, err
		}
	}
	var total int64
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if len(files) >= 4096 {
			return fmt.Errorf("src exceeds 4096 entries")
		}
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			files[filepath.ToSlash(rel)+"/"] = fmt.Sprintf("dir:%o", info.Mode().Perm())
			if copyTo != "" {
				dest := filepath.Join(copyTo, rel)
				if err := os.MkdirAll(dest, 0700); err != nil {
					return err
				}
				return os.Chmod(dest, info.Mode().Perm())
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("generated src contains a non-regular file: %s", rel)
		}
		if rel == sourceReceipt {
			return nil
		}
		total += info.Size()
		if len(files) >= 4096 || total > 100*1024*1024 {
			return fmt.Errorf("generated src exceeds 4096 files or 100 MiB")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		files[filepath.ToSlash(rel)] = fmt.Sprintf("file:%o:%s", info.Mode().Perm(), hex.EncodeToString(sum[:]))
		if copyTo != "" {
			dest := filepath.Join(copyTo, rel)
			if err := os.WriteFile(dest, data, info.Mode().Perm()); err != nil {
				return err
			}
			return os.Chmod(dest, info.Mode().Perm())
		}
		return nil
	})
	return files, err
}

// Stage source on the destination filesystem, then roll it back if binary
// publication fails. Lock both parent directories when source and binary are
// published separately, so builds sharing either destination cannot collide.
func publishOutputs(ctx context.Context, artifact, source, output string, baseline sourceContext) error {
	dest := baseline.Destination.Path
	parent := filepath.Dir(dest)
	parents := []string{parent}
	if filepath.Dir(output) != parent {
		parents = append(parents, filepath.Dir(output))
	}
	sort.Strings(parents)
	for _, dir := range parents {
		lock := filepath.Join(dir, ".nutshell-publish.lock")
		if err := os.Mkdir(lock, 0700); err != nil {
			return fmt.Errorf("cannot acquire publication lock %s: %w", lock, err)
		}
		defer os.Remove(lock)
	}
	if err := baseline.unchanged(); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(parent, ".nutshell-src-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	files, err := sourceFiles(source, stage)
	if err != nil {
		return fmt.Errorf("generated src: %w", err)
	}
	if _, exists := files[legacyProvenanceResource]; exists {
		return fmt.Errorf("remove legacy %s and its Nutshell-only embedding hooks, then rebuild", legacyProvenanceResource)
	}
	if !hasImplementation(files) {
		return fmt.Errorf("compilation agent must generate implementation files in src/")
	}
	// Copying can take time; check again immediately before replacing source.
	if err := baseline.unchanged(); err != nil {
		return err
	}
	backup, err := os.MkdirTemp(parent, ".nutshell-previous-src-")
	if err != nil {
		return err
	}
	if err := os.Remove(backup); err != nil {
		return err
	}
	hadSource := false
	if _, err := os.Lstat(dest); err == nil {
		if err := os.Rename(dest, backup); err != nil {
			return err
		}
		hadSource = true
	} else if !os.IsNotExist(err) {
		return err
	}
	installed := false
	rollback := func(cause error) error {
		if installed {
			if err := os.RemoveAll(dest); err != nil {
				return errors.Join(cause, fmt.Errorf("source rollback failed; previous source at %s: %w", backup, err))
			}
		}
		if hadSource {
			if err := os.Rename(backup, dest); err != nil {
				return errors.Join(cause, fmt.Errorf("restore source from %s: %w", backup, err))
			}
		}
		return cause
	}
	if err := os.Rename(stage, dest); err != nil {
		return rollback(err)
	}
	installed = true
	if err := ctx.Err(); err != nil {
		return rollback(err)
	}
	if err := publish(artifact, output); err != nil {
		return rollback(err)
	}
	if hadSource {
		_ = os.RemoveAll(backup)
	}
	return nil
}
