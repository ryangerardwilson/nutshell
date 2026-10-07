//go:build linux || darwin

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func maintenanceFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0700); err != nil {
		t.Fatal(err)
	}
}

func TestLocalInstallerSelectsSourceAndPreservesFailedBuild(t *testing.T) {
	installer, err := filepath.Abs("install.sh")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	source := filepath.Join(root, "selected source")
	bin := filepath.Join(root, "installed bin")
	maintenanceFile(t, filepath.Join(source, "go.mod"), "module github.com/ryangerardwilson/nutshell\n\ngo 1.26.0\n")
	maintenanceFile(t, filepath.Join(source, "VERSION"), "9.8.7\n")
	maintenanceFile(t, filepath.Join(source, "main.go"), "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"nutshell 9.8.7\")}\n")
	run := func(args ...string) ([]byte, error) {
		cmd := exec.Command("bash", append([]string{installer}, args...)...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "NUTSHELL_INSTALL_DIR="+bin)
		return cmd.CombinedOutput()
	}
	if out, err := run("from", source); err != nil || !bytes.Contains(out, []byte("nutshell 9.8.7")) {
		t.Fatalf("install selected checkout: %v: %s", err, out)
	}
	installed := filepath.Join(bin, "nutshell")
	previous, err := os.ReadFile(installed)
	if err != nil {
		t.Fatal(err)
	}
	maintenanceFile(t, filepath.Join(source, "main.go"), "broken Go source")
	if out, err := run("from", source); err == nil {
		t.Fatalf("accepted broken build: %s", out)
	}
	for _, args := range [][]string{{"from"}, {"wrong", source}, {"from", source, "extra"}, {"from", root}} {
		if out, err := run(args...); err == nil {
			t.Fatalf("accepted %v: %s", args, out)
		}
	}
	after, _ := os.ReadFile(installed)
	if !bytes.Equal(previous, after) {
		t.Fatal("failed install replaced existing executable")
	}
	leftovers, _ := filepath.Glob(filepath.Join(bin, ".nutshell-install-*"))
	if len(leftovers) != 0 {
		t.Fatal("failed install left staged binaries")
	}
}

func TestReleaseMetadataValidation(t *testing.T) {
	script, err := os.ReadFile("scripts/check-release.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, version, changelog, tag string
		valid                         bool
	}{
		{"valid", "1.2.3\n", "## [1.2.3] - 2026-10-07\n\nA release.\n\n## [1.2.2] - 2026-10-06\nOld notes.\n", "v1.2.3", true},
		{"mismatch", "1.2.3\n", "## [1.2.3] - 2026-10-07\nNotes.\n", "v1.2.4", false},
		{"leading zero", "01.2.3\n", "## [01.2.3] - 2026-10-07\nNotes.\n", "v01.2.3", false},
		{"missing notes", "1.2.3\n", "## [1.2.3] - 2026-10-07\n\n## [1.2.2] - 2026-10-06\nOld.\n", "v1.2.3", false},
		{"missing date", "1.2.3\n", "## [1.2.3]\nNotes.\n", "v1.2.3", false},
		{"duplicate", "1.2.3\n", "## [1.2.3] - 2026-10-07\nNotes.\n## [1.2.3] - 2026-10-07\nAgain.\n", "v1.2.3", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			maintenanceFile(t, filepath.Join(root, "scripts/check-release.sh"), string(script))
			maintenanceFile(t, filepath.Join(root, "VERSION"), tc.version)
			maintenanceFile(t, filepath.Join(root, "CHANGELOG.md"), tc.changelog)
			out, err := exec.Command("bash", filepath.Join(root, "scripts/check-release.sh"), tc.tag).CombinedOutput()
			if (err == nil) != tc.valid {
				t.Fatalf("validation: %v: %s", err, out)
			}
			if tc.valid && (!strings.Contains(string(out), "A release.") || strings.Contains(string(out), "Old notes")) {
				t.Fatalf("wrong release notes: %s", out)
			}
		})
	}
}
