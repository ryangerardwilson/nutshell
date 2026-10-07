package compiler

import (
	"fmt"
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

func TestCompositionIsCompilerDefined(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.nut")
	text := "from feature1.nut import abc\nUse the behavior described in feature2.nut.\nInclude \"missing.nut\".\n"
	put(t, entry, text)
	put(t, filepath.Join(dir, "feature1.nut"), "abc greets the user.\n")
	put(t, filepath.Join(dir, "nested", "feature2.nut"), "Ask main.nut for the behavior.\n")
	put(t, filepath.Join(dir, "nested", "main.nut"), "Another available file, not the selected entry.\n")
	for _, excluded := range []string{".git", ".hidden", "node_modules", "vendor", "target", "dist", "build"} {
		put(t, filepath.Join(dir, excluded, "ignored.nut"), "Not compiler input.")
	}
	p, err := Load(entry)
	if err != nil || p.Entry != "main.nut" || len(p.Sources) != 4 || p.Sources[0].Text != text {
		t.Fatalf("%+v %v", p, err)
	}
	// Newly available input changes the bundle even without a parsed directive.
	put(t, filepath.Join(dir, "extra.nut"), "Another available feature.")
	if p.unchanged() == nil {
		t.Fatal("failed to detect added source")
	}
	put(t, filepath.Join(dir, "feature1.nut"), "\xff")
	if _, err := Load(entry); err == nil {
		t.Fatal("accepted invalid discovered input")
	}
}

func TestDiscoveryBoundaries(t *testing.T) {
	dir, outside := t.TempDir(), t.TempDir()
	entry := filepath.Join(dir, "main.nut")
	put(t, entry, "Use available features.")
	put(t, filepath.Join(outside, "hidden.nut"), "Must not follow directory links.")
	if err := os.Symlink(outside, filepath.Join(dir, "linked")); err != nil {
		t.Fatal(err)
	}
	p, err := Load(entry)
	if err != nil || len(p.Sources) != 1 {
		t.Fatalf("followed directory symlink: %+v %v", p, err)
	}
	if err := os.Symlink(filepath.Join(outside, "hidden.nut"), filepath.Join(dir, "linked.nut")); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(entry); err == nil {
		t.Fatal("accepted source symlink")
	}
	if err := os.Remove(filepath.Join(dir, "linked.nut")); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 128; i++ {
		put(t, filepath.Join(dir, fmt.Sprintf("file%03d.nut", i)), "A feature.")
	}
	if _, err := Load(entry); err == nil || !strings.Contains(err.Error(), "maximum 128") {
		t.Fatalf("file limit: %v", err)
	}
}
