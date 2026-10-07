package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ryangerardwilson/nutshell/internal/compiler"
)

func TestProgressDisplay(t *testing.T) {
	for _, terminal := range []bool{true, false} {
		var out bytes.Buffer
		p := newProgressDisplay(&out, terminal)
		p.update(compiler.Progress{Stage: "implementing", Message: "Writing code\x1b\n", Step: 2})
		p.update(compiler.Progress{Stage: "understanding", Message: "REGRESSION", Step: 1})
		time.Sleep(150 * time.Millisecond)
		p.finish(nil)
		got := out.String()
		if strings.Contains(got, "REGRESSION") || !strings.Contains(got, "Done") {
			t.Fatal(got)
		}
		if terminal && !strings.Contains(got, "[================]") {
			t.Fatal(got)
		}
		if !terminal && strings.ContainsAny(got, "\r\x1b") {
			t.Fatal(got)
		}
	}
	var out bytes.Buffer
	p := newProgressDisplay(&out, true)
	p.update(compiler.Progress{Stage: "test", Message: "Testing", Step: 6})
	p.finish(errors.New("failed"))
	if strings.Contains(out.String(), "[================]") || !strings.Contains(out.String(), "Build stopped") {
		t.Fatal(out.String())
	}
}

func TestSameStageSummariesAndStaleRecovery(t *testing.T) {
	var out bytes.Buffer
	p := newProgressDisplay(&out, false)
	first := compiler.Progress{Stage: "implementing", Message: "Writing command line argument parser", Step: 2}
	second := compiler.Progress{Stage: "implementing", Message: "Adding empty input validation tests", Step: 2}
	p.update(first)
	p.update(second)
	p.update(second)
	p.mu.Lock()
	p.updated = time.Now().Add(-31 * time.Second)
	if got := p.activity(time.Now(), false); got != "Waiting for AI progress update" {
		t.Fatal(got)
	}
	p.mu.Unlock()
	p.update(second)
	p.mu.Lock()
	if got := p.activity(time.Now(), false); got != second.Message {
		t.Fatal(got)
	}
	p.mu.Unlock()
	p.finish(nil)
	got := out.String()
	if strings.Count(got, first.Message) != 1 || strings.Count(got, second.Message) != 1 {
		t.Fatal(got)
	}
	long := "Preparing executable for final verification"
	if cleanMessage(long) != long {
		t.Fatal("truncated valid activity")
	}
}
