//go:build linux || darwin

package compiler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func fake(t *testing.T, body string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	// Compiler fixtures do not depend on a workstation installation of rgw-ast.
	toolDir := t.TempDir()
	put(t, filepath.Join(toolDir, "rgw-ast"), "#!/bin/sh\nexit 0\n")
	if err := os.Chmod(filepath.Join(toolDir, "rgw-ast"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", toolDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	entry := filepath.Join(dir, "main.nut")
	put(t, entry, "Write hello followed by a newline.\n")
	tool := filepath.Join(dir, "fake-ai")
	put(t, tool, "#!/bin/sh\nset -eu\ncat > received-prompt.txt\n"+body+"\n")
	if err := os.Chmod(tool, 0700); err != nil {
		t.Fatal(err)
	}
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)
	data, _ := json.Marshal(map[string]any{"interpreters": map[string]Adapter{"fake": {Command: []string{tool}, UnsafeArgs: []string{"--unsafe"}, Prompt: "stdin"}}})
	put(t, filepath.Join(config, "nutshell", "interpreters.json"), string(data))
	return dir, entry
}

func TestCompileRequiresRGWAstBeforeLaunchingAgent(t *testing.T) {
	for _, nonExecutable := range []bool{false, true} {
		t.Run(fmt.Sprint(nonExecutable), func(t *testing.T) {
			dir, entry := fake(t, "exit 99")
			path := t.TempDir()
			if nonExecutable {
				put(t, filepath.Join(path, "rgw-ast"), "not executable")
			}
			t.Setenv("PATH", path)
			output := filepath.Join(dir, "main")
			self, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			previous, err := os.ReadFile(self)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(output, previous, 0700); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(dir, "implementation")
			r, err := Compile(context.Background(), Options{Entry: entry, SourceDir: source, Interpreter: "fake"})
			if err == nil || !strings.Contains(err.Error(), "requires rgw-ast") || !strings.Contains(err.Error(), "PATH") {
				t.Fatalf("expected dependency guidance, got %v", err)
			}
			if r.Directory != "" {
				t.Fatal("created build attempt before dependency preflight")
			}
			data, _ := os.ReadFile(output)
			if !bytes.Equal(data, previous) {
				t.Fatal("changed existing output")
			}
			if _, err := os.Stat(source); !os.IsNotExist(err) {
				t.Fatal("created source destination")
			}
		})
	}
}

func TestAgentReceivesRGWAstWorkflow(t *testing.T) {
	dir, entry := fake(t, `test -d src
grep -q 'rgw-ast --root src status --json' received-prompt.txt
grep -q 'even when status reports enforced=false' received-prompt.txt
grep -q 'patch with --expect-hash' received-prompt.txt
printf 'workflow received' > workflow-received
exit 9`)
	r, err := Compile(context.Background(), Options{Entry: entry, SourceDir: filepath.Join(dir, "implementation"), Interpreter: "fake"})
	if err == nil {
		t.Fatal("expected fixture exit")
	}
	data, readErr := os.ReadFile(filepath.Join(r.Directory, "workflow-received"))
	if readErr != nil || string(data) != "workflow received" {
		t.Fatalf("agent did not receive workflow or source root: %v (%v)", readErr, err)
	}
}

func TestCompileNativeProgram(t *testing.T) {
	fixture := t.TempDir()
	put(t, filepath.Join(fixture, "src", "go.mod"), "module generated\n\ngo 1.26.0\n")
	put(t, filepath.Join(fixture, "src", "main.go"), "package main\nimport \"fmt\"\nfunc greeting() string { return \"hello\" }; func main(){fmt.Println(greeting())}\n")
	put(t, filepath.Join(fixture, "src", "main_test.go"), "package main\nimport \"testing\"\nfunc TestGreeting(t *testing.T){if greeting()!=\"hello\" { t.Fatal(\"wrong greeting\") }}\n")
	m := Manifest{Version: 1, Language: "Go", Summary: "Greeting", Artifact: "program", Build: [][]string{{"go", "-C", "src", "build", "-o", "../program", "."}}, Test: [][]string{{"go", "-C", "src", "test", "./..."}}}
	if err := writeJSON(filepath.Join(fixture, "build.json"), m); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(fixture, "diagnostics.json"), `{"anything":"obsolete clarification"}`)
	t.Setenv("NUTSHELL_TEST_FIXTURE", fixture)
	dir, entry := fake(t, `printf '%s' '{"version":1,"stage":"implementing","message":"Generating greeting program code"}' > progress.json
echo RAW_PROVIDER_OUTPUT
sleep 0.3
cp -R "$NUTSHELL_TEST_FIXTURE/." .`)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	var updates []Progress
	var log bytes.Buffer
	result, err := Compile(ctx, Options{Entry: entry, SourceDir: filepath.Join(dir, "src"), Interpreter: "fake", Log: &log, OnProgress: func(p Progress) { updates = append(updates, p) }})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "succeeded" || result.Output != filepath.Join(dir, "main") {
		t.Fatalf("%+v", result)
	}
	if !strings.HasPrefix(result.Directory, "/tmp/nutshell-build-") {
		t.Fatal(result.Directory)
	}
	if _, err := os.Stat(filepath.Join(dir, ".nutshell")); !os.IsNotExist(err) {
		t.Fatal("created metadata in source directory")
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "main.go")); err != nil {
		t.Fatal(err)
	}
	if log.Len() != 0 {
		t.Fatalf("raw output leaked: %s", log.String())
	}
	found := false
	for _, u := range updates {
		if u.Stage == "implementing" {
			found = true
		}
	}
	if !found {
		t.Fatalf("no live AI progress: %+v", updates)
	}
	got, err := exec.Command(result.Output).Output()
	if err != nil || string(got) != "hello\n" {
		t.Fatalf("%q %v", got, err)
	}
	for _, name := range []string{"source/main.nut", "source.json", "prompt.txt", "received-prompt.txt", "interpreter.log", "verification.log", "result.json"} {
		if _, err := os.Stat(filepath.Join(result.Directory, name)); err != nil {
			t.Fatal(err)
		}
	}
	var report Result
	if err := readJSON(filepath.Join(result.Directory, "result.json"), &report); err != nil || report.Status != "succeeded" {
		t.Fatalf("%+v %v", report, err)
	}
}

func TestFailedCompilePreservesOutput(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	base := `{"version":1,"language":"fixture","summary":"test","artifact":"program","build":[["true"]],"test":[["true"]]}`
	for _, tc := range []struct{ name, body, want string }{
		{"provider", "exit 3", "compilation agent failed"},
		{"missing", "true", "build.json"},
		{"badjson", "echo nope > build.json", "build.json"},
		{"tests", "cp \"$NUTSHELL_TEST_SELF\" program\nprintf '%s' '" + strings.Replace(base, `"test":[["true"]]`, `"test":[["false"]]`, 1) + "' > build.json", "test failed"},
		{"script", "printf '#!/bin/sh\\nexit 0\\n' > program\nchmod +x program\nprintf '%s' '" + base + "' > build.json", "expected native"},
		{"artifact-escape", "printf '%s' '" + strings.Replace(base, `"program"`, `"../program"`, 1) + "' > build.json", "relative file"},
		{"artifact-symlink", "ln -s \"$NUTSHELL_TEST_SELF\" program\nprintf '%s' '" + base + "' > build.json", "symlinks"},
		{"source-drift", "cp \"$NUTSHELL_TEST_SELF\" program\necho changed > \"$NUTSHELL_TEST_ENTRY\"\nprintf '%s' '" + base + "' > build.json", "source changed"},
		{"no-tests", "printf '%s' '" + strings.Replace(base, `"test":[["true"]]`, `"test":[]`, 1) + "' > build.json", "test must contain"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, entry := fake(t, tc.body)
			t.Setenv("NUTSHELL_TEST_SELF", self)
			t.Setenv("NUTSHELL_TEST_ENTRY", entry)
			previous, err := os.ReadFile(self)
			if err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(dir, "main")
			if err := os.WriteFile(output, previous, 0700); err != nil {
				t.Fatal(err)
			}
			r, err := Compile(context.Background(), Options{Entry: entry, SourceDir: filepath.Join(dir, "src"), Interpreter: "fake"})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("wanted %s: %v", tc.want, err)
			}
			after, _ := os.ReadFile(output)
			if !bytes.Equal(previous, after) {
				t.Fatal("previous output changed")
			}
			var report Result
			if err := readJSON(filepath.Join(r.Directory, "result.json"), &report); err != nil || report.Status != "failed" || report.Error == "" {
				t.Fatalf("bad report %+v %v", report, err)
			}
		})
	}
}

func TestOutputProtection(t *testing.T) {
	dir, entry := fake(t, "exit 99")
	for _, name := range []string{"main.nut", "notes.txt", "linked"} {
		output := filepath.Join(dir, name)
		if name == "notes.txt" {
			put(t, output, "precious text")
		}
		if name == "linked" {
			if err := os.Symlink(entry, output); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := Compile(context.Background(), Options{Entry: entry, SourceDir: filepath.Join(dir, "src"), Output: output, Interpreter: "fake"}); err == nil {
			t.Fatalf("accepted %s", name)
		}
	}
	data, _ := os.ReadFile(entry)
	if !strings.Contains(string(data), "Write hello") {
		t.Fatal("source overwritten")
	}
}

func TestCancellation(t *testing.T) {
	dir, entry := fake(t, "sleep 30 &\necho $! > child.pid\nwait")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	r, err := Compile(ctx, Options{Entry: entry, SourceDir: filepath.Join(dir, "src"), Interpreter: "fake", Log: io.Discard})
	if err == nil || time.Since(start) > 5*time.Second || r.Directory == "" {
		t.Fatalf("cancellation: %+v %v", r, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "main")); !os.IsNotExist(err) {
		t.Fatal("published cancelled build")
	}
	if runtime.GOOS == "linux" {
		pid, err := os.ReadFile(filepath.Join(r.Directory, "child.pid"))
		if err != nil {
			t.Fatal(err)
		}
		status, err := os.ReadFile(filepath.Join("/proc", strings.TrimSpace(string(pid)), "status"))
		if err == nil && !strings.Contains(string(status), "Z (zombie)") && !strings.Contains(string(status), "X (dead)") {
			t.Fatalf("child survived cancellation: %s", status)
		}
	}
}

func TestBuiltInAdapters(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"codex", "grok", "claude"} {
		put(t, filepath.Join(dir, name), "#!/bin/sh\nexit 0\n")
		if err := os.Chmod(filepath.Join(dir, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
	for name, flag := range map[string]string{"codex": "--dangerously-bypass-approvals-and-sandbox", "grok": "--always-approve", "claude": "--dangerously-skip-permissions"} {
		a, err := Resolve(name)
		if err != nil {
			t.Fatal(err)
		}
		argv, in := a.invocation("the program", "/tmp/prompt.txt")
		joined := strings.Join(argv, " ")
		if !strings.Contains(joined, flag) || strings.Contains(joined, "--model") {
			t.Fatal(joined)
		}
		if name == "grok" && (!strings.Contains(joined, "--sandbox off") || !strings.Contains(joined, "--prompt-file /tmp/prompt.txt")) {
			t.Fatal(joined)
		}
		if name != "grok" {
			data, _ := io.ReadAll(in)
			if string(data) != "the program" {
				t.Fatal(string(data))
			}
		}
	}
}

func TestCompileEvolvesExistingImplementation(t *testing.T) {
	fixture := t.TempDir()
	put(t, filepath.Join(fixture, "main.go"), "package main\nimport \"fmt\"\nfunc main(){fmt.Println(existingGreeting())}\n")
	put(t, filepath.Join(fixture, "main_test.go"), "package main\nimport \"testing\"\nfunc TestExistingGreeting(t *testing.T){if existingGreeting()!=\"hello from existing code\"{t.Fatal(\"existing behavior lost\")}}\n")
	m := Manifest{Version: 1, Language: "Go", Summary: "Adapt existing code", Artifact: "program", Build: [][]string{{"go", "-C", "src", "build", "-o", "../program", "."}}, Test: [][]string{{"go", "-C", "src", "test", "./..."}}}
	if err := writeJSON(filepath.Join(fixture, "build.json"), m); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NUTSHELL_INCREMENTAL_FIXTURE", fixture)
	dir, entry := fake(t, `test -f src/existing.go
 test -f original-src/existing.go
 cp "$NUTSHELL_INCREMENTAL_FIXTURE/main.go" src/main.go
 cp "$NUTSHELL_INCREMENTAL_FIXTURE/main_test.go" src/main_test.go
 cp "$NUTSHELL_INCREMENTAL_FIXTURE/build.json" build.json`)
	selected := filepath.Join(dir, "custom implementation")
	output := filepath.Join(t.TempDir(), "greet")
	put(t, filepath.Join(dir, "src", "untouched.txt"), "ignore this implicit directory")
	existing := "package main\n// Keep this user-written implementation.\nfunc existingGreeting()string{return \"hello from existing code\"}\n"
	put(t, filepath.Join(selected, "existing.go"), existing)
	put(t, filepath.Join(selected, "go.mod"), "module incremental\n\ngo 1.26.0\n")
	r, err := Compile(context.Background(), Options{Entry: entry, SourceDir: selected, Output: output, Interpreter: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(r.Output).Output()
	if err != nil || string(out) != "hello from existing code\n" {
		t.Fatalf("%q %v", out, err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(output), "src")); !os.IsNotExist(err) {
		t.Fatal("created implicit output-side src")
	}
	ignored, _ := os.ReadFile(filepath.Join(dir, "src", "untouched.txt"))
	if string(ignored) != "ignore this implicit directory" {
		t.Fatal("modified unrelated src")
	}
	data, _ := os.ReadFile(filepath.Join(selected, "existing.go"))
	if string(data) != existing {
		t.Fatal("existing implementation changed")
	}
}

func TestFailedIncrementalCompilePreservesSource(t *testing.T) {
	dir, entry := fake(t, `printf 'temporary AI edit' > src/manual.txt
 exit 4`)
	put(t, filepath.Join(dir, "src", "manual.txt"), "original user code")
	if _, err := Compile(context.Background(), Options{Entry: entry, SourceDir: filepath.Join(dir, "src"), Interpreter: "fake"}); err == nil {
		t.Fatal("expected failure")
	}
	data, _ := os.ReadFile(filepath.Join(dir, "src", "manual.txt"))
	if string(data) != "original user code" {
		t.Fatal("failed generation changed source")
	}
}
