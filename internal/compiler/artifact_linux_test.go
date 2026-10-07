package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRejectWrongNativeArchitecture(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	// EM_NONE in either byte order: a valid ELF container for no host architecture.
	data[18], data[19] = 0, 0
	path := filepath.Join(t.TempDir(), "wrong-architecture")
	if err := os.WriteFile(path, data, 0700); err != nil {
		t.Fatal(err)
	}
	if err := nativeArtifact(path); err == nil {
		t.Fatal("wrong architecture accepted")
	}
}
