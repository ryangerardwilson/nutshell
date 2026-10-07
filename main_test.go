package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
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
	for _, flags := range [][]string{{"-ft", request}, {"--fine-tune", request}, {"-ft=" + request}, {"--fine-tune=" + request}} {
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
	for _, flags := range [][]string{{"-ft"}, {"-ft", ""}, {"--fine-tune="}, {"-ft", " \n\t"}} {
		args := append([]string{"-c", "grok", "-s", "./src"}, flags...)
		if _, err := parse(args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
