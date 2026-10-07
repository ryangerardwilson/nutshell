package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ryangerardwilson/nutshell/internal/compiler"
)

func TestInspectAndDiffNeverExecuteTarget(t *testing.T) {
	t.Setenv("PATH", "") // No provider or helper executable is needed.
	root := t.TempDir()
	entry := filepath.Join(root, "main.nut")
	feature := filepath.Join(root, "feature.nut")
	for path, text := range map[string]string{entry: "Use the abc feature in feature.nut.\n", feature: "abc prints hello.\n"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := compiler.Load(entry)
	if err != nil {
		t.Fatal(err)
	}
	record := compiler.Provenance{Version: 1, Entry: p.Entry, Sources: p.Sources, SourceSHA256: compiler.SourceHash(p), NutshellVersion: version, Compiler: "fake", Language: "Go", Assumptions: []string{"A greeting"}}
	frame, err := compiler.EncodeProvenance(record)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "app")
	marker := filepath.Join(root, "must-not-exist")
	data := append([]byte("#!/bin/sh\nprintf executed > '"+marker+"'\nexit 0\n"), frame...)
	if err := os.WriteFile(target, data, 0700); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) (int, string, string) {
		var out, errs bytes.Buffer
		code := execute(context.Background(), args, &out, &errs)
		return code, out.String(), errs.String()
	}
	code, output, errs := run("inspect", target, "--json")
	var recovered compiler.Provenance
	if code != 0 || json.Unmarshal([]byte(output), &recovered) != nil || len(recovered.Sources) != 2 || recovered.SourceSHA256 != record.SourceSHA256 {
		t.Fatalf("%d %s %s", code, output, errs)
	}
	if code, _, errs := run("diff", target, entry); code != 0 {
		t.Fatalf("equal diff: %d %s", code, errs)
	}
	if err := os.WriteFile(feature, []byte("abc prints goodbye.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	code, output, errs = run("diff", "--json", target, entry)
	var diff compiler.SourceDiff
	if code != 1 || json.Unmarshal([]byte(output), &diff) != nil || len(diff.Changes) != 1 || diff.Changes[0].Path != "feature.nut" {
		t.Fatalf("%d %s %s", code, output, errs)
	}
	for _, name := range []string{entry, feature} {
		if err := os.Remove(name); err != nil {
			t.Fatal(err)
		}
	}
	code, output, errs = run("inspect", target, "--source")
	if code != 0 || !strings.Contains(output, "abc prints hello.") || !strings.Contains(output, "--- feature.nut ---") {
		t.Fatalf("source recovery: %d %s %s", code, output, errs)
	}
	if code, _, _ := run("diff", target, entry); code != 2 {
		t.Fatal("missing source should be an error, not a diff")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("inspection executed the target")
	}
	for _, args := range [][]string{{"inspect"}, {"inspect", target, "--json", "--source"}, {"diff", target, "--source"}, {"inspect", target, entry}, {"inspect", target, "--wat"}} {
		if code, _, _ := run(args...); code != 2 {
			t.Fatalf("invalid arguments accepted: %v", args)
		}
	}
}
