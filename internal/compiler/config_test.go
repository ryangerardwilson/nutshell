package compiler

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestXDGConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	for _, xdg := range []string{"", "relative/config", t.TempDir()} {
		t.Setenv("XDG_CONFIG_HOME", xdg)
		want := filepath.Join(home, ".config", "nutshell", "compilers.json")
		if filepath.IsAbs(xdg) {
			want = filepath.Join(xdg, "nutshell", "compilers.json")
		}
		got, err := ConfigPath()
		if err != nil || got != want {
			t.Fatalf("%q %v want %q", got, err, want)
		}
		if _, err := os.Stat(filepath.Dir(got)); !os.IsNotExist(err) {
			t.Fatal("path lookup created config")
		}
	}
}

func TestConfigInitializesOnceAndPreservesOverrides(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := InitConfig()
	if err != nil {
		t.Fatal(err)
	}
	config, err := readCompilerConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Compilers) != 3 {
		t.Fatalf("unexpected starter entries: %+v", config.Compilers)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	config.Compilers = map[string]Adapter{"codex": {Command: []string{self, "custom", "--prompt={prompt}"}, Prompt: "argument"}, "wrapper": {Command: []string{self, "--profile", "custom-profile", "exec", "-"}, Prompt: "stdin"}}
	if err := writeJSON(path, config); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := InitConfig(); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("initialization changed existing configuration")
	}
	a, err := Resolve("codex")
	if err != nil {
		t.Fatal(err)
	}
	prompt := "Say `hello` and $(do nothing)."
	argv, _ := a.invocation(prompt, "unused")
	if !reflect.DeepEqual(argv, []string{self, "custom", "--prompt=" + prompt}) {
		t.Fatalf("override ignored: %q", argv)
	}
	wrapper, err := Resolve("wrapper")
	if err != nil {
		t.Fatal(err)
	}
	argv, stdin := wrapper.invocation("source input", "unused")
	data, _ := io.ReadAll(stdin)
	if argv[2] != "custom-profile" || string(data) != "source input" {
		t.Fatalf("wrapper altered: %q %q", argv, data)
	}
	if _, err := Resolve("grok"); err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("removed name fell back: %v", err)
	}
}

func TestConfigMigratesLegacyWithoutChangingOriginal(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, _ := ConfigPath()
	legacy := filepath.Join(filepath.Dir(path), "interpreters.json")
	self, _ := os.Executable()
	adapter := Adapter{Command: []string{self, "{unsafe_args}", "generate"}, UnsafeArgs: []string{"--custom-yolo"}, Prompt: "stdin"}
	data, _ := json.Marshal(map[string]any{"interpreters": map[string]Adapter{"codex": adapter, "private-wrapper": adapter}})
	put(t, legacy, string(data))
	if _, err := InitConfig(); err != nil {
		t.Fatal(err)
	}
	config, err := readCompilerConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Compilers) != 4 || !reflect.DeepEqual(config.Compilers["codex"], adapter) {
		t.Fatalf("migration lost override: %+v", config)
	}
	original, _ := os.ReadFile(legacy)
	if string(original) != string(data) {
		t.Fatal("legacy file changed")
	}
	put(t, legacy, "invalid now; new file must remain authoritative")
	if _, err := Resolve("private-wrapper"); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidConfigFailsWithoutReset(t *testing.T) {
	for _, payload := range []string{"{", `{"version":2,"compilers":{}}`, `{"version":1}`, `{"version":1,"compilers":{"x":{"command":[],"prompt":"stdin"}}}`, `{"version":1,"compilers":{"x":{"command":["tool"],"prompt":"file"}}}`, `{"version":1,"compilers":{"x":{"command":["tool"],"prompt":"unknown"}}}`} {
		t.Run(payload, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			path, _ := ConfigPath()
			put(t, path, payload)
			if _, err := Resolve("codex"); err == nil || !strings.Contains(err.Error(), path) {
				t.Fatalf("%v", err)
			}
			current, _ := os.ReadFile(path)
			if string(current) != payload {
				t.Fatal("reset invalid user config")
			}
		})
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, _ := ConfigPath()
	put(t, filepath.Join(filepath.Dir(path), "interpreters.json"), "broken")
	if _, err := InitConfig(); err == nil {
		t.Fatal("ignored broken legacy file")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("created config despite migration failure")
	}
}

func TestConcurrentConfigInitialization(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := InitConfig(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	path, _ := ConfigPath()
	config, err := readCompilerConfig(path)
	if err != nil || len(config.Compilers) != 3 {
		t.Fatalf("%+v %v", config, err)
	}
	matches, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".compilers-*"))
	if len(matches) != 0 {
		t.Fatalf("left temporary files: %v", matches)
	}
}
