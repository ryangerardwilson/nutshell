package compiler

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPublicationRechecksProvenanceAndRollsBack(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	out := filepath.Join(root, "program")
	put(t, filepath.Join(root, "src", "main.go"), "previous implementation")
	original, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, original, 0700); err != nil {
		t.Fatal(err)
	}
	baseline := exportContext(t, out)
	work := t.TempDir()
	put(t, filepath.Join(work, "main.go"), "changed implementation")
	expected := testProvenance()
	// This native artifact lacks the expected resource. The staged copy must be
	// rejected even if generation's earlier verification were bypassed or stale.
	if err := publishOutputs(context.Background(), self, work, out, baseline, &expected); err == nil {
		t.Fatal("published artifact without provenance")
	}
	data, _ := os.ReadFile(out)
	if !bytes.Equal(data, original) {
		t.Fatal("replaced previous executable")
	}
	data, _ = os.ReadFile(filepath.Join(root, "src", "main.go"))
	if string(data) != "previous implementation" {
		t.Fatal("failed to roll back implementation")
	}
}
