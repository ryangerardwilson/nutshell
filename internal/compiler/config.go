package compiler

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// This is an initialization template, never a runtime fallback for missing names.
//
//go:embed default-compilers.json
var defaultCompilers []byte

type compilerConfig struct {
	Version   int                `json:"version"`
	Compilers map[string]Adapter `json:"compilers"`
}

// ConfigPath follows XDG even on hosts where os.UserConfigDir uses another layout.
func ConfigPath() (string, error) {
	root := os.Getenv("XDG_CONFIG_HOME")
	if !filepath.IsAbs(root) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".config")
	}
	return filepath.Join(root, "nutshell", "compilers.json"), nil
}

func readCompilerConfig(path string) (compilerConfig, error) {
	var config compilerConfig
	if err := readJSON(path, &config); err != nil {
		return config, fmt.Errorf("read compiler config %s: %w", path, err)
	}
	if config.Version != 1 || config.Compilers == nil {
		return config, fmt.Errorf("%s requires version 1 and a compilers object", path)
	}
	for name, adapter := range config.Compilers {
		if strings.TrimSpace(name) == "" {
			return config, fmt.Errorf("%s contains an empty compiler name", path)
		}
		if err := adapter.validate(); err != nil {
			return config, fmt.Errorf("compiler %q in %s: %w", name, path, err)
		}
	}
	return config, nil
}

// InitConfig preserves an existing config, including one created concurrently.
func InitConfig() (string, error) {
	path, err := ConfigPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(path); err == nil {
		_, err = readCompilerConfig(path)
		return path, err
	} else if !os.IsNotExist(err) {
		return path, err
	}
	var config compilerConfig
	if err := json.Unmarshal(defaultCompilers, &config); err != nil {
		return path, err
	}
	legacyPath := filepath.Join(filepath.Dir(path), "interpreters.json")
	if _, err := os.Lstat(legacyPath); err == nil {
		var legacy struct {
			Interpreters map[string]Adapter `json:"interpreters"`
		}
		if err := readJSON(legacyPath, &legacy); err != nil {
			return path, fmt.Errorf("migrate %s: %w", legacyPath, err)
		}
		if legacy.Interpreters == nil {
			return path, fmt.Errorf("%s requires an interpreters object for migration", legacyPath)
		}
		for name, adapter := range legacy.Interpreters {
			if strings.TrimSpace(name) == "" {
				return path, fmt.Errorf("%s contains an empty compiler name", legacyPath)
			}
			if err := adapter.validate(); err != nil {
				return path, fmt.Errorf("migrate compiler %q from %s: %w", name, legacyPath, err)
			}
			config.Compilers[name] = adapter
		}
	} else if !os.IsNotExist(err) {
		return path, err
	}
	if err := createConfig(path, config); err != nil {
		return path, fmt.Errorf("initialize %s: %w", path, err)
	}
	_, err = readCompilerConfig(path)
	return path, err
}

func createConfig(path string, config compilerConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".compilers-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if _, err := temp.Write(append(data, '\n')); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	// Hard-link a completed file into place: readers never see a partial document,
	// and a racing initializer/editor's file is never replaced.
	if err := os.Link(temp.Name(), path); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
