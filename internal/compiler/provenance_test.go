package compiler

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const embeddedFixture = `package main
import (
 _ "embed"
 "crypto/sha256"
)
//go:embed .nutshell-provenance.bin
var provenance []byte
func init() { if sha256.Sum256(provenance) == ([32]byte{}) { panic("invalid provenance") } }
`

func testProvenance() Provenance {
	p := Program{Entry: "main.nut", Sources: []Source{{"main.nut", "Use feature.nut however you see fit.\n"}, {"feature.nut", "Print hello.\n"}}}
	return provenanceFor(p, Options{Interpreter: "fake", NutshellVersion: "0.9.0"}, Manifest{Language: "Go", Assumptions: []string{"A greeting is sufficient"}})
}

func TestInspectEmbeddedResource(t *testing.T) {
	p := testProvenance()
	frame, err := EncodeProvenance(p)
	if err != nil {
		t.Fatal(err)
	}
	// Start the resource across a scan chunk boundary, with a harmless false marker.
	data := append([]byte(provenanceMagic), make([]byte, 64*1024-len(provenanceMagic)-5)...)
	data = append(data, frame...)
	data = append(data, []byte("remaining native file bytes")...)
	data = append(data, frame...) // identical duplicated resources are valid
	file := filepath.Join(t.TempDir(), "program")
	put(t, file, string(data))
	got, err := Inspect(file)
	if err != nil || got.SourceSHA256 != p.SourceSHA256 || got.Sources[1].Text != p.Sources[1].Text {
		t.Fatalf("%+v %v", got, err)
	}
	if err := verifyProvenance(file, p); err != nil {
		t.Fatal(err)
	}
	p.Assumptions = []string{"Changed assumption"}
	if err := verifyProvenance(file, p); err == nil {
		t.Fatal("accepted different assumptions")
	}
	other, _ := EncodeProvenance(p)
	put(t, file, string(append(data, other...)))
	if _, err := Inspect(file); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("ambiguous provenance accepted: %v", err)
	}
}

func TestRejectDamagedProvenance(t *testing.T) {
	frame, _ := EncodeProvenance(testProvenance())
	for _, kind := range []string{"checksum", "truncated", "oversized", "legacy", "schema", "path", "hash"} {
		t.Run(kind, func(t *testing.T) {
			data := append([]byte{}, frame...)
			switch kind {
			case "checksum":
				data[len(data)-1] ^= 1
			case "truncated":
				data = data[:len(data)-12]
			case "oversized":
				binary.LittleEndian.PutUint64(data[len(provenanceMagic):], ^uint64(0))
			case "legacy":
				data = []byte("ordinary binary without metadata")
			default:
				p := testProvenance()
				if kind == "schema" {
					p.Version = 99
				} else if kind == "path" {
					p.Sources[1].Path = "../outside.nut"
				} else {
					p.Sources[1].Text = "Changed without updating the source hash"
				}
				payload, _ := json.Marshal(p)
				data = binary.LittleEndian.AppendUint64([]byte(provenanceMagic), uint64(len(payload)))
				data = append(data, payload...)
				sum := sha256.Sum256(payload)
				data = append(data, sum[:]...)
			}
			file := filepath.Join(t.TempDir(), "program")
			put(t, file, string(data))
			if _, err := Inspect(file); err == nil {
				t.Fatal("accepted invalid provenance")
			}
		})
	}
}

func TestSourceDiff(t *testing.T) {
	p := testProvenance()
	current := Program{Entry: p.Entry, Sources: append([]Source{}, p.Sources...)}
	current.Sources[0], current.Sources[1] = current.Sources[1], current.Sources[0]
	if DiffSources(p, current).Changed {
		t.Fatal("discovery order changed canonical hash")
	}
	current.Sources = []Source{{"main.nut", "New behavior without trailing newline"}, {"new.nut", "New feature.\n"}}
	d := DiffSources(p, current)
	if !d.Changed || len(d.Changes) != 3 || d.Changes[0].Kind != "removed" || d.Changes[1].Kind != "modified" || d.Changes[2].Kind != "added" {
		t.Fatalf("%+v", d)
	}
	var output bytes.Buffer
	d.WriteText(&output)
	if !strings.Contains(output.String(), "+New feature.") || !strings.Contains(output.String(), "-Print hello.") || !strings.Contains(output.String(), "No newline at end of file") {
		t.Fatal(output.String())
	}
	current = Program{Entry: "feature.nut", Sources: p.Sources}
	d = DiffSources(p, current)
	if !d.Changed || len(d.Changes) != 0 {
		t.Fatal("entry change was lost")
	}
}

func TestPreviousContextLegacyBinary(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(t.TempDir(), "previous")
	put(t, output, "a legacy binary")
	if err := stagePrevious(dir, output, Program{Entry: "main.nut"}); err != nil {
		t.Fatal(err)
	}
	var c previousContext
	if err := readJSON(filepath.Join(dir, "previous-context.json"), &c); err != nil || !c.Exists || c.Provenance != nil || c.Reason == "" {
		t.Fatalf("%+v %v", c, err)
	}
	info, err := os.Stat(filepath.Join(dir, "previous-program"))
	if err != nil || info.Mode().Perm()&0333 != 0 {
		t.Fatal("previous copy is not read-only")
	}
}
