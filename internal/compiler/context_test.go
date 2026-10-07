package compiler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplicitSourceSelection(t *testing.T) {
	root, outputRoot := t.TempDir(), t.TempDir()
	t.Chdir(root)
	put(t, filepath.Join(root, "src", "ignored.go"), "implicit context must be ignored")
	selected := filepath.Join(root, "custom implementation")
	put(t, filepath.Join(selected, "existing.go"), "selected implementation")
	output := filepath.Join(outputRoot, "program")
	c, err := captureContext("custom implementation", output)
	if err != nil {
		t.Fatal(err)
	}
	if c.Input.Path != selected || c.Destination.Path != selected {
		t.Fatalf("%+v", c)
	}
	work := t.TempDir()
	if err := c.seed(work); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(work, "src", "existing.go"))
	if string(data) != "selected implementation" {
		t.Fatal(string(data))
	}
	if _, err := os.Stat(filepath.Join(work, "src", "ignored.go")); !os.IsNotExist(err) {
		t.Fatal("implicit source was included")
	}
	fresh := filepath.Join(root, "new implementation")
	c, err = captureContext(fresh, output)
	if err != nil {
		t.Fatal(err)
	}
	if c.Input.Exists || c.Input.Path != fresh {
		t.Fatalf("fell back to src: %+v", c)
	}
	if _, err := os.Stat(fresh); !os.IsNotExist(err) {
		t.Fatal("created source during preflight")
	}
	if _, err := captureContext("", output); err == nil {
		t.Fatal("accepted implicit source")
	}
	if _, err := captureContext(filepath.Join(root, "missing parent", "code"), output); err == nil {
		t.Fatal("accepted missing parent")
	}
	if _, err := captureContext(selected, filepath.Join(selected, "program")); err == nil {
		t.Fatal("accepted output inside source")
	}
	if _, err := captureContext(selected, selected); err == nil {
		t.Fatal("accepted output equal to source")
	}
}

func TestSourceSnapshotChanges(t *testing.T) {
	for _, change := range []string{"edit", "add", "delete", "directory", "permissions", "symlink"} {
		t.Run(change, func(t *testing.T) {
			root := t.TempDir()
			file := filepath.Join(root, "src", "main.go")
			put(t, file, "original")
			c := exportContext(t, filepath.Join(root, "main"))
			switch change {
			case "edit":
				put(t, file, "edited")
			case "add":
				put(t, filepath.Join(root, "src", "new.go"), "new")
			case "delete":
				if err := os.Remove(file); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(filepath.Join(root, "src", "empty"), 0700); err != nil {
					t.Fatal(err)
				}
			case "permissions":
				if err := os.Chmod(file, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Symlink(file, filepath.Join(root, "src", "link.go")); err != nil {
					t.Fatal(err)
				}
			}
			if err := c.unchanged(); err == nil {
				t.Fatalf("missed %s", change)
			}
		})
	}
}

func TestLegacyReceiptIsNotRequiredOrCopied(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "src", "main.go"), "user edits")
	put(t, filepath.Join(root, "src", sourceReceipt), "outdated receipt")
	c := exportContext(t, filepath.Join(root, "main"))
	work := t.TempDir()
	if err := c.seed(work); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(work, "src", sourceReceipt)); !os.IsNotExist(err) {
		t.Fatal("copied obsolete receipt")
	}
}

func TestSourceChangedWhileSeeding(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "src", "main.go")
	put(t, file, "original")
	c := exportContext(t, filepath.Join(root, "main"))
	put(t, file, "changed")
	if err := c.seed(t.TempDir()); err == nil {
		t.Fatal("accepted changed context")
	}
}

func TestNutInsidePublishedSource(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "src", "rules.nut"), "Keep this source.")
	p := Program{Root: root, Sources: []Source{{Path: "src/rules.nut", Text: "Keep this source."}}}
	working := t.TempDir()
	put(t, filepath.Join(working, "rules.nut"), "Keep this source.")
	if err := protectNutInputs(p, working, filepath.Join(root, "src")); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(working, "rules.nut"), "AI rewrote source")
	if err := protectNutInputs(p, working, filepath.Join(root, "src")); err == nil {
		t.Fatal("accepted .nut overwrite")
	}
}

func TestIndependentSourceAndBinaryDestinations(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	entryRoot, outputRoot := t.TempDir(), t.TempDir()
	put(t, filepath.Join(entryRoot, "src", "main.go"), "existing")
	out := filepath.Join(outputRoot, "program")
	c, err := captureContext(filepath.Join(entryRoot, "src"), out)
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	if err := c.seed(work); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(work, "src", "main.go"), "adapted")
	if err := publishOutputs(context.Background(), self, filepath.Join(work, "src"), out, c, nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(entryRoot, "src", "main.go"))
	if string(data) != "adapted" {
		t.Fatal("selected source was not updated")
	}
	if _, err := os.Stat(filepath.Join(outputRoot, "src")); !os.IsNotExist(err) {
		t.Fatal("created implicit source beside binary")
	}
}

func TestContextPrompt(t *testing.T) {
	message := promptFor(Program{Entry: "main.nut"})
	if !strings.Contains(message, "Inspect those\nfiles FIRST") || !strings.Contains(message, "rather than deleting it and starting from scratch") {
		t.Fatal("context instructions missing")
	}
}
