package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func put(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestSourceBundle(t *testing.T) {
	dir := t.TempDir()
	put(t, filepath.Join(dir, "main.nut"), "Say hello.\nInclude \"rules.nut\".\ninclude \"rules.nut\"\n")
	put(t, filepath.Join(dir, "rules.nut"), "Exit successfully.\n")
	p, err := Load(filepath.Join(dir, "main.nut"))
	if err != nil || len(p.Sources) != 2 || p.Sources[1].Path != "rules.nut" || p.Sources[1].Text != "Exit successfully.\n" {
		t.Fatalf("%+v %v", p, err)
	}
	put(t, filepath.Join(dir, "rules.nut"), "Changed.")
	if err := p.unchanged(); err == nil {
		t.Fatal("source drift accepted")
	}
}

func TestInvalidSources(t *testing.T) {
	for _, tc := range []struct{ name, source, extra string }{
		{"empty", " \n", ""}, {"utf8", "\xff", ""},
		{"escape", "Include \"../outside.nut\".", ""},
		{"absolute", "Include \"/tmp/outside.nut\".", ""},
		{"missing", "Include \"missing.nut\".", ""},
		{"extension", "Include \"rules.txt\".", ""},
		{"cycle", "Include \"rules.nut\".", "Include \"main.nut\"."},
		{"large", strings.Repeat("x", maxSourceBytes+1), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			entry := filepath.Join(dir, "main.nut")
			put(t, entry, tc.source)
			if tc.extra != "" {
				put(t, filepath.Join(dir, "rules.nut"), tc.extra)
			}
			if _, err := Load(entry); err == nil {
				t.Fatal("invalid source accepted")
			}
		})
	}
	dir := t.TempDir()
	put(t, filepath.Join(dir, "actual.nut"), "hello")
	if err := os.Symlink("actual.nut", filepath.Join(dir, "main.nut")); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(filepath.Join(dir, "main.nut")); err == nil {
		t.Fatal("symlink accepted")
	}
}
