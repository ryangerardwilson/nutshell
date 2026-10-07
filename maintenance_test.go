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
	alias := filepath.Join(bin, "ns")
	if target, err := os.Readlink(alias); err != nil || target != "nutshell" {
		t.Fatalf("alias: %q %v", target, err)
	}
	if out, err := exec.Command(alias, "--version").CombinedOutput(); err != nil || string(out) != "nutshell 9.8.7\n" {
		t.Fatalf("run alias: %v: %s", err, out)
	}
	// Reinstall a changed source and ensure the alias follows the new executable.
	maintenanceFile(t, filepath.Join(source, "main.go"), "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"nutshell 9.8.8\")}\n")
	if out, err := run("from", source); err != nil {
		t.Fatalf("reinstall: %v: %s", err, out)
	}
	if out, err := exec.Command(alias, "--version").CombinedOutput(); err != nil || string(out) != "nutshell 9.8.8\n" {
		t.Fatalf("alias after upgrade: %v: %s", err, out)
	}
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
	if out, err := exec.Command(alias, "--version").CombinedOutput(); err != nil || string(out) != "nutshell 9.8.8\n" {
		t.Fatalf("alias after failed upgrade: %v: %s", err, out)
	}
	leftovers, _ := filepath.Glob(filepath.Join(bin, ".nutshell-install-*"))
	if len(leftovers) != 0 {
		t.Fatal("failed install left staged binaries")
	}
}

func TestInstallerPreservesUnrelatedNS(t *testing.T) {
	installer, err := filepath.Abs("install.sh")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"file", "directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			bin := t.TempDir()
			installed, alias := filepath.Join(bin, "nutshell"), filepath.Join(bin, "ns")
			maintenanceFile(t, installed, "previous installation")
			switch kind {
			case "file":
				maintenanceFile(t, alias, "another tool")
			case "directory":
				maintenanceFile(t, filepath.Join(alias, "keep"), "another tool")
			case "symlink":
				if err := os.Symlink("unrelated-target", alias); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("bash", installer, "from", source)
			cmd.Env = append(os.Environ(), "NUTSHELL_INSTALL_DIR="+bin)
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), "unrelated command") {
				t.Fatalf("conflict not rejected: %v: %s", err, out)
			}
			previous, _ := os.ReadFile(installed)
			if string(previous) != "previous installation" {
				t.Fatal("changed installed binary despite alias conflict")
			}
			if kind == "symlink" {
				if target, err := os.Readlink(alias); err != nil || target != "unrelated-target" {
					t.Fatal("changed unrelated symlink")
				}
			} else {
				if kind == "directory" {
					alias = filepath.Join(alias, "keep")
				}
				data, _ := os.ReadFile(alias)
				if string(data) != "another tool" {
					t.Fatal("changed unrelated command")
				}
			}
		})
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
