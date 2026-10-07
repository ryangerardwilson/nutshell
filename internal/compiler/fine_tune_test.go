//go:build linux || darwin

package compiler

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFineTunePrompt(t *testing.T) {
	p := Program{Entry: "main.nut", Sources: []Source{{"main.nut", "Print the original greeting."}, {"feature.nut", "Unrelated feature details."}}}
	request := "replace x with \"y\"; preserve 100%\nand all other behavior"
	prompt := promptFor(p, request, false)
	for _, want := range []string{"FINE-TUNE MODE", `replace x with \"y\"; preserve 100%\nand all other behavior`, "source/main.nut", "source/feature.nut", "takes precedence", "Preserve all unrelated behavior", "rgw-ast --root src status --json", "build.json"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, unwanted := range []string{p.Sources[0].Text, p.Sources[1].Text, "SOURCE BUNDLE", "ns inspect", "previous-program", legacyProvenanceResource} {
		if strings.Contains(prompt, unwanted) {
			t.Errorf("unexpected %q", unwanted)
		}
	}
	full := promptFor(p, "", false)
	if !strings.Contains(full, p.Sources[0].Text) || !strings.Contains(full, "FULL COMPILATION") {
		t.Fatal("full compilation lost requirements")
	}
	migration := promptFor(p, "change greeting", true)
	if !strings.Contains(migration, "Remove that resource AND") {
		t.Fatal("missing legacy migration")
	}
}

func TestFineTuneRequiresExistingSourceBeforeProvider(t *testing.T) {
	for _, kind := range []string{"absent", "empty", "resource-only", "nut-only", "blank-request"} {
		t.Run(kind, func(t *testing.T) {
			dir, entry := fake(t, "exit 99")
			src := filepath.Join(dir, "src")
			request := "replace x with y"
			switch kind {
			case "empty":
				if err := os.Mkdir(src, 0700); err != nil {
					t.Fatal(err)
				}
			case "resource-only":
				put(t, filepath.Join(src, legacyProvenanceResource), "old metadata")
			case "nut-only":
				put(t, filepath.Join(src, "context.nut"), "requirements only")
			case "blank-request":
				request = " \n\t"
				put(t, filepath.Join(src, "main.go"), "package main")
			}
			r, err := Compile(context.Background(), Options{Entry: entry, Interpreter: "fake", SourceDir: src, FineTune: request})
			if err == nil || !strings.Contains(err.Error(), "fine tuning requires") || r.Directory != "" {
				t.Fatalf("%+v %v", r, err)
			}
		})
	}
}

func TestFineTuneBuildsAndMigratesExistingImplementation(t *testing.T) {
	fixture := t.TempDir()
	newCode := "package main\nimport \"fmt\"\nfunc greeting() string { return \"y\" }; func main(){fmt.Println(greeting())}\n"
	put(t, filepath.Join(fixture, "main.go"), newCode)
	put(t, filepath.Join(fixture, "main_test.go"), "package main\nimport \"testing\"\nfunc TestGreeting(t *testing.T){if greeting()!=\"y\" {t.Fatal(\"wrong greeting\")}}\n")
	manifest := Manifest{Version: 1, Language: "Go", Summary: "Change greeting", Artifact: "program", Build: [][]string{{"go", "-C", "src", "build", "-o", "../program", "."}}, Test: [][]string{{"go", "-C", "src", "test", "./..."}}}
	if err := writeJSON(filepath.Join(fixture, "build.json"), manifest); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NUTSHELL_FINE_TUNE_FIXTURE", fixture)
	dir, entry := fake(t, `grep -q 'FINE-TUNE MODE' received-prompt.txt
grep -q 'replace x with y' received-prompt.txt
grep -q 'Remove that resource AND' received-prompt.txt
test -f src/.nutshell-provenance.bin
test -f original-src/provenance.go
! test -e previous-program
! test -e previous-context.json
rm src/provenance.go src/.nutshell-provenance.bin
cp "$NUTSHELL_FINE_TUNE_FIXTURE/main.go" src/main.go
cp "$NUTSHELL_FINE_TUNE_FIXTURE/main_test.go" src/main_test.go
cp "$NUTSHELL_FINE_TUNE_FIXTURE/build.json" build.json`)
	put(t, entry, "Print x followed by a newline.\n")
	src := filepath.Join(dir, "src")
	put(t, filepath.Join(src, "go.mod"), "module fine\n\ngo 1.26.0\n")
	put(t, filepath.Join(src, "main.go"), strings.ReplaceAll(newCode, `"y"`, `"x"`))
	put(t, filepath.Join(src, "manual.txt"), "preserve this user asset")
	put(t, filepath.Join(src, legacyProvenanceResource), "old source text")
	put(t, filepath.Join(src, "provenance.go"), "package main\nimport _ \"embed\"\n//go:embed .nutshell-provenance.bin\nvar oldSource []byte\nfunc init(){ if len(oldSource)==0 { panic(\"missing source\") } }\n")
	output := filepath.Join(dir, "app")
	build := exec.Command("go", "-C", src, "build", "-o", output, ".")
	build.Env = append(os.Environ(), "GOWORK=off")
	if data, err := build.CombinedOutput(); err != nil {
		t.Fatalf("legacy build: %s %v", data, err)
	}
	r, err := Compile(context.Background(), Options{Entry: entry, Interpreter: "fake", SourceDir: src, Output: output, FineTune: "replace x with y"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := exec.Command(r.Output).Output()
	if err != nil || string(data) != "y\n" {
		t.Fatalf("%q %v", data, err)
	}
	for _, name := range []string{legacyProvenanceResource, "provenance.go"} {
		if _, err := os.Stat(filepath.Join(src, name)); !os.IsNotExist(err) {
			t.Fatalf("legacy file retained: %s", name)
		}
	}
	for path, want := range map[string]string{entry: "Print x followed by a newline.\n", filepath.Join(src, "manual.txt"): "preserve this user asset"} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != want {
			t.Fatalf("unexpected change to %s: %q %v", path, data, err)
		}
	}
}

func TestLegacyResourceBlocksPublication(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	src, output := filepath.Join(root, "src"), filepath.Join(root, "app")
	put(t, filepath.Join(src, "manual.txt"), "existing source")
	oldBinary, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, oldBinary, 0700); err != nil {
		t.Fatal(err)
	}
	baseline, err := captureContext(src, output)
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	put(t, filepath.Join(work, "main.go"), "new implementation")
	put(t, filepath.Join(work, legacyProvenanceResource), "stale metadata")
	err = publishOutputs(context.Background(), self, work, output, baseline)
	if err == nil || !strings.Contains(err.Error(), "embedding hooks") {
		t.Fatalf("%v", err)
	}
	if err := baseline.unchanged(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil || string(data) != string(oldBinary) {
		t.Fatal("previous output replaced")
	}
}
