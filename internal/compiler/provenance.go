package compiler

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const ProvenanceResource = ".nutshell-provenance.bin"
const provenanceMagic = "\x00NUTSHELL-PROVENANCE-v1\x00"
const maxProvenanceBytes = 8 * 1024 * 1024

var ErrNoProvenance = errors.New("no valid Nutshell provenance found; recompile this program with Nutshell 0.9.0 or newer")

// Provenance records compiler inputs, not proof of behavior or publisher identity.
type Provenance struct {
	Version         int      `json:"version"`
	Entry           string   `json:"entry"`
	Sources         []Source `json:"sources"`
	SourceSHA256    string   `json:"source_sha256"`
	NutshellVersion string   `json:"nutshell_version"`
	Compiler        string   `json:"compiler"`
	Language        string   `json:"language"`
	Assumptions     []string `json:"assumptions"`
}

func SourceHash(p Program) string {
	sources := append([]Source(nil), p.Sources...)
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	data, _ := json.Marshal(struct {
		Entry   string   `json:"entry"`
		Sources []Source `json:"sources"`
	}{p.Entry, sources})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func provenanceFor(p Program, opts Options, manifest Manifest) Provenance {
	v := opts.NutshellVersion
	if v == "" {
		v = "development"
	}
	return Provenance{1, p.Entry, p.Sources, SourceHash(p), v, opts.Interpreter, manifest.Language, append([]string{}, manifest.Assumptions...)}
}

func validSourcePath(name string) bool {
	return fs.ValidPath(name) && path.Ext(name) == ".nut" && !strings.ContainsAny(name, "\\\x00")
}

func (p Provenance) validate() error {
	if p.Version != 1 || !validSourcePath(p.Entry) || len(p.Sources) == 0 || len(p.Sources) > 128 || p.NutshellVersion == "" || p.Compiler == "" {
		return fmt.Errorf("invalid provenance schema or source count")
	}
	seen := map[string]bool{}
	total := 0
	for _, source := range p.Sources {
		total += len(source.Text)
		if !validSourcePath(source.Path) || seen[source.Path] || !utf8.ValidString(source.Text) || strings.TrimSpace(source.Text) == "" || total > maxSourceBytes {
			return fmt.Errorf("invalid provenance source bundle")
		}
		seen[source.Path] = true
	}
	if !seen[p.Entry] || p.SourceSHA256 != SourceHash(Program{Entry: p.Entry, Sources: p.Sources}) {
		return fmt.Errorf("provenance source hash or entry mismatch")
	}
	return nil
}

// EncodeProvenance creates the opaque resource embedded by the native toolchain.
func EncodeProvenance(p Provenance) ([]byte, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	if len(data) > maxProvenanceBytes {
		return nil, fmt.Errorf("provenance exceeds 8 MiB")
	}
	frame := []byte(provenanceMagic)
	frame = binary.LittleEndian.AppendUint64(frame, uint64(len(data)))
	frame = append(frame, data...)
	sum := sha256.Sum256(data)
	return append(frame, sum[:]...), nil
}

func writeProvenance(dir string, p Provenance) error {
	data, err := EncodeProvenance(p)
	if err != nil {
		return err
	}
	file, err := localPath(dir, filepath.Join("src", ProvenanceResource))
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0600)
}

// Inspect reads a regular file as data. It never executes or loads its code.
func Inspect(filename string) (Provenance, error) {
	pathInfo, err := os.Stat(filename)
	if err != nil {
		return Provenance{}, err
	}
	if !pathInfo.Mode().IsRegular() {
		return Provenance{}, fmt.Errorf("inspection requires a regular file")
	}
	f, err := os.Open(filename)
	if err != nil {
		return Provenance{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return Provenance{}, err
	}
	if !info.Mode().IsRegular() {
		return Provenance{}, fmt.Errorf("inspection requires a regular file")
	}
	var found *Provenance
	var foundData []byte
	const chunk = 64 * 1024
	buf := make([]byte, chunk+len(provenanceMagic)-1)
	var offset int64
	for offset < info.Size() {
		n, err := f.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			return Provenance{}, err
		}
		window := buf[:n]
		for start := 0; start < len(window); {
			index := bytes.Index(window[start:], []byte(provenanceMagic))
			if index < 0 {
				break
			}
			index += start
			p, data, err := readProvenanceAt(f, offset+int64(index), info.Size())
			if err == nil {
				if found != nil && !bytes.Equal(foundData, data) {
					return Provenance{}, fmt.Errorf("binary contains conflicting Nutshell provenance records")
				}
				found, foundData = &p, data
			}
			start = index + len(provenanceMagic)
		}
		offset += chunk
	}
	if found == nil {
		return Provenance{}, ErrNoProvenance
	}
	return *found, nil
}

func readProvenanceAt(f *os.File, offset, size int64) (Provenance, []byte, error) {
	var p Provenance
	pos := offset + int64(len(provenanceMagic))
	var length [8]byte
	if _, err := f.ReadAt(length[:], pos); err != nil {
		return p, nil, err
	}
	n := binary.LittleEndian.Uint64(length[:])
	pos += 8
	if n == 0 || n > maxProvenanceBytes || int64(n)+sha256.Size > size-pos {
		return p, nil, fmt.Errorf("invalid provenance length")
	}
	data := make([]byte, int(n)+sha256.Size)
	if _, err := f.ReadAt(data, pos); err != nil {
		return p, nil, err
	}
	sum := sha256.Sum256(data[:n])
	if !bytes.Equal(sum[:], data[n:]) {
		return p, nil, fmt.Errorf("provenance checksum mismatch")
	}
	decoder := json.NewDecoder(bytes.NewReader(data[:n]))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&p); err != nil {
		return p, nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return p, nil, fmt.Errorf("invalid trailing provenance data")
	}
	if err := p.validate(); err != nil {
		return p, nil, err
	}
	return p, data[:n], nil
}

func verifyProvenance(file string, expected Provenance) error {
	got, err := Inspect(file)
	if err != nil {
		return fmt.Errorf("generated executable must embed src/%s as a retained binary resource: %w", ProvenanceResource, err)
	}
	a, _ := json.Marshal(got)
	b, _ := json.Marshal(expected)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("generated executable contains stale or mismatched source provenance")
	}
	return nil
}
