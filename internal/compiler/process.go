package compiler

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func run(ctx context.Context, dir string, argv []string, input io.Reader, output io.Writer) error {
	if len(argv) == 0 || argv[0] == "" {
		return fmt.Errorf("empty command")
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir, cmd.Stdin = dir, input
	cmd.Stdout, cmd.Stderr = output, output
	overrides := map[string]string{
		"GOWORK":     "off",
		"TMPDIR":     filepath.Join(dir, "tmp"),
		"GOCACHE":    filepath.Join(dir, "cache", "go"),
		"GOMODCACHE": filepath.Join(dir, "cache", "modules"),
	}
	for key, value := range overrides {
		if key != "GOWORK" {
			if err := os.MkdirAll(value, 0700); err != nil {
				return err
			}
		}
	}
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		if _, overridden := overrides[key]; !overridden {
			cmd.Env = append(cmd.Env, e)
		}
	}
	for key, value := range overrides {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	cmd.WaitDelay = 2 * time.Second
	configureProcess(cmd)
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%s: %w", argv[0], err)
	}
	return nil
}
