package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	for _, args := range [][]string{{"main.nut", "-c", "codex", "-s", "my source"}, {"-s", "my source", "-c", "codex", "main.nut"}, {"--compiler=codex", "--source=my source"}, {"-c=codex", "-s=my source"}} {
		o, err := parse(args)
		if err != nil || o.Entry != "main.nut" || o.Interpreter != "codex" || o.SourceDir != "my source" {
			t.Fatalf("%v: %+v %v", args, o, err)
		}
	}
	for _, args := range [][]string{{}, {"-c"}, {"-c", "codex", "a.nut", "b.nut"}, {"-c", "codex", "--timeout=0"}, {"--wat"}, {"main.nut", "-c", "codex"}, {"-c", "codex", "-s"}, {"-c", "codex", "-s="}} {
		if _, err := parse(args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestInformationalCommandsDoNotInvokeProvider(t *testing.T) {
	t.Setenv("PATH", "")
	for _, arg := range []string{"--help", "help", "--version", "version"} {
		var out, errs bytes.Buffer
		if code := execute(context.Background(), []string{arg}, &out, &errs); code != 0 || out.Len() == 0 {
			t.Fatalf("%s: %d %s", arg, code, errs.String())
		}
	}
}

func TestMissingSourceFailsBeforeCompilation(t *testing.T) {
	t.Setenv("PATH", "")
	var out, errs bytes.Buffer
	code := execute(context.Background(), []string{"main.nut", "-c", "codex"}, &out, &errs)
	if code != 2 || !strings.Contains(errs.String(), "-s <path>") || out.Len() != 0 {
		t.Fatalf("%d %s %s", code, out.String(), errs.String())
	}
}

func TestOldCompilerFlagsGiveMigrationGuidance(t *testing.T) {
	t.Setenv("PATH", "")
	for _, flag := range []string{"-i", "--interpreter", "--interpreter=codex"} {
		var out, errs bytes.Buffer
		code := execute(context.Background(), []string{flag, "codex", "-s", "src"}, &out, &errs)
		if code != 2 || !strings.Contains(errs.String(), "replaced by -c (or --compiler)") || out.Len() != 0 {
			t.Fatalf("%s: %d %s", flag, code, errs.String())
		}
	}
}

func TestFineTuneArguments(t *testing.T) {
	request := "replace x with y\nkeep 100% of other behavior"
	for _, flags := range [][]string{{"-f", request}, {"--fine-tune", request}, {"-f=" + request}, {"--fine-tune=" + request}} {
		for _, before := range []bool{true, false} {
			args := []string{"main.nut", "-c", "grok", "-o", "app", "-s", "./src"}
			if before {
				args = append(append([]string{}, flags...), args...)
			} else {
				args = append(args, flags...)
			}
			got, err := parse(args)
			if err != nil || got.FineTune != request || got.Output != "app" {
				t.Fatalf("%v: %+v %v", args, got, err)
			}
		}
	}
	for _, flags := range [][]string{{"-f"}, {"-f", ""}, {"--fine-tune="}, {"-f", " \n\t"}} {
		args := append([]string{"-c", "grok", "-s", "./src"}, flags...)
		if _, err := parse(args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestRemovedTimeLimitsFailBeforeCompilation(t *testing.T) {
	t.Setenv("PATH", "")
	for _, flags := range [][]string{{"-l", "5"}, {"--limit", "5"}, {"-l=5"}, {"--limit=5"}, {"-l"}, {"--limit="}} {
		for _, before := range []bool{true, false} {
			args := []string{"main.nut", "-c", "grok", "-s", "src", "-f", "fix spacing"}
			if before {
				args = append(append([]string{}, flags...), args...)
			} else {
				args = append(args, flags...)
			}
			var out, errs bytes.Buffer
			code := execute(context.Background(), args, &out, &errs)
			if code != 2 || out.Len() != 0 || !strings.Contains(errs.String(), "was removed; omit the time limit") {
				t.Fatalf("%v: %d %s %s", args, code, out.String(), errs.String())
			}
		}
	}
}

func TestTimeoutRemainsIndependent(t *testing.T) {
	for _, fineTune := range [][]string{nil, {"-f", "fix spacing"}} {
		args := append([]string{"main.nut", "-c", "grok", "-s", "src"}, fineTune...)
		o, err := parse(args)
		if err != nil || o.Timeout != 30*time.Minute {
			t.Fatalf("default timeout: %+v %v", o, err)
		}
		o, err = parse(append(args, "--timeout", "10m"))
		if err != nil || o.Timeout != 10*time.Minute {
			t.Fatalf("explicit timeout: %+v %v", o, err)
		}
	}
}

func TestOldFineTuneFlagGivesMigrationGuidance(t *testing.T) {
	for _, flag := range []string{"-ft", "-ft=fix"} {
		var out, errs bytes.Buffer
		code := execute(context.Background(), []string{"-c", "grok", "-s", "src", flag}, &out, &errs)
		if code != 2 || !strings.Contains(errs.String(), "replaced by -f") {
			t.Fatalf("%d %s", code, errs.String())
		}
	}
}

func TestConfigCommandsDoNotInvokeProvider(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("PATH", "")
	path := filepath.Join(root, "nutshell", "compilers.json")
	for _, args := range [][]string{{"--help"}, {"--version"}, {"config", "path"}} {
		var out, errs bytes.Buffer
		if code := execute(context.Background(), args, &out, &errs); code != 0 {
			t.Fatalf("%v: %d %s", args, code, errs.String())
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%v created config", args)
		}
	}
	var out, errs bytes.Buffer
	if code := execute(context.Background(), []string{"config", "init"}, &out, &errs); code != 0 || strings.TrimSpace(out.String()) != path {
		t.Fatalf("%d %s %s", code, out.String(), errs.String())
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"config"}, {"config", "reset"}, {"config", "init", "extra"}} {
		if code := execute(context.Background(), args, &out, &errs); code != 2 {
			t.Fatalf("accepted %v", args)
		}
	}
}
