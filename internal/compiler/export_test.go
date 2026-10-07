package compiler

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSourcePublicationAndRollback(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	source := filepath.Join(t.TempDir(), "src")
	put(t, filepath.Join(source, "main.go"), "original code")
	out := filepath.Join(root, "main")
	if err := publishOutputs(context.Background(), self, source, out, exportContext(t, out)); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(source, "main.go"), "replacement code")
	if err := publishOutputs(context.Background(), "/missing-artifact", source, out, exportContext(t, out)); err == nil {
		t.Fatal("accepted failed publication")
	}
	data, _ := os.ReadFile(filepath.Join(root, "src", "main.go"))
	if string(data) != "original code" {
		t.Fatalf("failed to roll back: %s", data)
	}
	if err := publishOutputs(context.Background(), self, source, out, exportContext(t, out)); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(filepath.Join(root, "src", "main.go"))
	if string(data) != "replacement code" {
		t.Fatal(string(data))
	}
	baseline := exportContext(t, out)
	put(t, filepath.Join(root, "src", "main.go"), "user edits")
	if err := publishOutputs(context.Background(), self, source, out, baseline); err == nil {
		t.Fatal("accepted changes made during compilation")
	}
	data, _ = os.ReadFile(filepath.Join(root, "src", "main.go"))
	if string(data) != "user edits" {
		t.Fatal(string(data))
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 2 {
		t.Fatalf("publication debris: %v", entries)
	}
}

func exportContext(t *testing.T, output string) sourceContext {
	t.Helper()
	c, err := captureContext(filepath.Join(filepath.Dir(output), "src"), output)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestUnmanagedSourceIsContext(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "src", "important.txt"), "keep")
	c := exportContext(t, filepath.Join(root, "main"))
	work := t.TempDir()
	if err := c.seed(work); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"src", "original-src"} {
		data, _ := os.ReadFile(filepath.Join(work, dir, "important.txt"))
		if string(data) != "keep" {
			t.Fatal(string(data))
		}
	}
	put(t, filepath.Join(work, "src", "important.txt"), "working copy change")
	data, _ := os.ReadFile(filepath.Join(root, "src", "important.txt"))
	if string(data) != "keep" {
		t.Fatal("source edited before publication")
	}
}

func TestRollbackAcrossPublicationDirectories(t *testing.T) {
	sourceRoot, outputRoot := t.TempDir(), t.TempDir()
	selected := filepath.Join(sourceRoot, "implementation")
	put(t, filepath.Join(selected, "main.go"), "original source")
	output := filepath.Join(outputRoot, "program")
	put(t, output, "previous executable")
	baseline, err := captureContext(selected, output)
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	if err := baseline.seed(work); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(work, "src", "main.go"), "temporary replacement")
	if err := publishOutputs(context.Background(), "/missing-artifact", filepath.Join(work, "src"), output, baseline); err == nil {
		t.Fatal("accepted failed binary publication")
	}
	data, _ := os.ReadFile(filepath.Join(selected, "main.go"))
	if string(data) != "original source" {
		t.Fatal("source rollback failed")
	}
	data, _ = os.ReadFile(output)
	if string(data) != "previous executable" {
		t.Fatal("binary changed")
	}
	for _, parent := range []string{sourceRoot, outputRoot} {
		entries, err := os.ReadDir(parent)
		if err != nil || len(entries) != 1 {
			t.Fatalf("publication debris: %v %v", entries, err)
		}
	}
}
