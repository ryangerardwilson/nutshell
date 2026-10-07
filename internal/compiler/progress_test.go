package compiler

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProgressProtocol(t *testing.T) {
	dir := t.TempDir()
	updates := make(chan Progress, 10)
	stop := watchProgress(context.Background(), dir, func(p Progress) { updates <- p })
	defer stop()
	path := filepath.Join(dir, "progress.json")
	put(t, path, `{"version":1,"stage":"implementing","message":"Writing command line argument parser"}`)
	select {
	case p := <-updates:
		if p.Step != 2 {
			t.Fatal(p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("progress not delivered")
	}
	put(t, path, `{"version":1,"stage":"implementing","message":"Adding empty input validation tests"}`)
	select {
	case p := <-updates:
		if p.Step != 2 || p.Message != "Adding empty input validation tests" {
			t.Fatal(p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("same-stage activity did not change")
	}
	for _, bad := range []string{`{"version":1,"stage":"implementing","message":"Only two"}`, `{"version":1,"stage":"implementing","message":"This message has far too many words here"}`, `{`, `{"version":1,"stage":"understanding","message":"old"}`, `{"version":1,"stage":"done","message":"premature completion"}`, `{"version":2,"stage":"ready"}`} {
		put(t, path, bad)
		select {
		case p := <-updates:
			t.Fatalf("invalid update delivered: %+v", p)
		case <-time.After(250 * time.Millisecond):
		}
	}
	put(t, path, `{"version":1,"stage":"checking","message":"Running generated tests"}`)
	select {
	case p := <-updates:
		if p.Step != 3 {
			t.Fatal(p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not recover")
	}
	put(t, path, `{"version":1,"stage":"implementing","message":"Fixing duplicate entry handling logic"}`)
	select {
	case p := <-updates:
		if p.Step != 3 || p.Message != "Fixing duplicate entry handling logic" {
			t.Fatalf("regressed milestone or lost activity: %+v", p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("fix activity not delivered")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func TestActivitySummary(t *testing.T) {
	for _, input := range []string{"", "Writing code", "This activity message contains more than seven words"} {
		if _, ok := activitySummary(input); ok {
			t.Fatalf("accepted %q", input)
		}
	}
	got, ok := activitySummary("  Adding\tUTF-8 filename\nvalidation tests ")
	if !ok || got != "Adding UTF-8 filename validation tests" {
		t.Fatalf("%q %v", got, ok)
	}
	if _, ok := activitySummary("Writing a simple command line greeting program"); !ok {
		t.Fatal("rejected seven words")
	}
}
