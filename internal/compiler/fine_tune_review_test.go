package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func reviewProgram(t *testing.T) Program {
	t.Helper()
	root := t.TempDir()
	put(t, filepath.Join(root, "main.nut"), "A greeting program.\n\nPrint hello world.\n")
	put(t, filepath.Join(root, "features", "greeting.nut"), "Keep these words:\r\nhello world\r\nunchanged.\r\n")
	p, err := Load(filepath.Join(root, "main.nut"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func conflictReview() fineTuneReview {
	return fineTuneReview{Version: 1, Status: "conflict", ReviewedSources: []string{"main.nut", "features/greeting.nut"}, Conflicts: []fineTuneConflict{
		{Path: "main.nut", StartLine: 3, EndLine: 3, Quote: "Print hello world.", Reason: "The requested replacement changes the required greeting."},
		{Path: "features/greeting.nut", StartLine: 1, EndLine: 3, Quote: "Keep these words:\nhello world\nunchanged.", Reason: "The requested replacement changes words explicitly required to remain unchanged."},
	}}
}

func TestConflictReviewLocations(t *testing.T) {
	p := reviewProgram(t)
	dir := t.TempDir()
	if err := writeJSON(filepath.Join(dir, "fine-tune.json"), conflictReview()); err != nil {
		t.Fatal(err)
	}
	err := checkFineTuneReview(dir, p)
	if err == nil {
		t.Fatal("accepted conflicting request")
	}
	for _, want := range []string{filepath.Join(p.Root, "main.nut") + ":3:", filepath.Join(p.Root, "features/greeting.nut") + ":1-3:", "3 | Print hello world.", "2 | hello world", "Change the fine-tune request"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in %s", want, err)
		}
	}
	put(t, filepath.Join(p.Root, "main.nut"), "Changed requirements.\n")
	if err := checkFineTuneReview(dir, p); err == nil || !strings.Contains(err.Error(), "source changed") {
		t.Fatalf("stale diagnostics: %v", err)
	}
}

func TestInvalidFineTuneReviews(t *testing.T) {
	p := reviewProgram(t)
	for _, tc := range []struct {
		name   string
		change func(*fineTuneReview)
	}{
		{"version", func(r *fineTuneReview) { r.Version = 2 }},
		{"status", func(r *fineTuneReview) { r.Status = "approved" }},
		{"missing source", func(r *fineTuneReview) { r.ReviewedSources = r.ReviewedSources[:1] }},
		{"duplicate source", func(r *fineTuneReview) { r.ReviewedSources[1] = "main.nut" }},
		{"unknown reviewed source", func(r *fineTuneReview) { r.ReviewedSources[1] = "outside.nut" }},
		{"compatible with conflicts", func(r *fineTuneReview) { r.Status = "compatible" }},
		{"conflict without locations", func(r *fineTuneReview) { r.Conflicts = []fineTuneConflict{} }},
		{"unknown conflict source", func(r *fineTuneReview) { r.Conflicts[0].Path = "../main.nut" }},
		{"zero line", func(r *fineTuneReview) { r.Conflicts[0].StartLine = 0 }},
		{"reversed range", func(r *fineTuneReview) { r.Conflicts[0].EndLine = 2 }},
		{"past final line", func(r *fineTuneReview) { r.Conflicts[0].EndLine = 4 }},
		{"wrong quote", func(r *fineTuneReview) { r.Conflicts[0].Quote = "Print hello everyone." }},
		{"missing reason", func(r *fineTuneReview) { r.Conflicts[0].Reason = " " }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := conflictReview()
			tc.change(&r)
			if err := r.validate(p); err == nil {
				t.Fatal("accepted invalid review")
			}
		})
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "fine-tune.json")
	for _, payload := range []string{"null", "{}", `{"version":1,"status":"compatible","reviewed_sources":["main.nut","features/greeting.nut"]}`, `{"version":1,"status":"compatible","reviewed_sources":["main.nut","features/greeting.nut"],"conflicts":[],"extra":true}`, `{} {}`} {
		put(t, file, payload)
		if err := checkFineTuneReview(dir, p); err == nil {
			t.Fatalf("accepted %s", payload)
		}
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := checkFineTuneReview(dir, p); err == nil {
		t.Fatal("accepted missing review")
	}
	r := fineTuneReview{Version: 1, Status: "compatible", ReviewedSources: []string{"main.nut", "features/greeting.nut"}, Conflicts: []fineTuneConflict{}}
	if err := writeJSON(file, r); err != nil {
		t.Fatal(err)
	}
	if err := checkFineTuneReview(dir, p); err != nil {
		t.Fatal(err)
	}
}

func TestTimeExpectationPrompt(t *testing.T) {
	p := Program{Entry: "main.nut"}
	for _, request := range []string{"", "fix spacing"} {
		prompt := promptFor(p, request, false, 5)
		for _, want := range []string{"no longer than 5 minutes", "separate hard cutoff", "Do not\nskip checks"} {
			if !strings.Contains(prompt, want) {
				t.Fatalf("missing %q", want)
			}
		}
		if strings.Contains(promptFor(p, request, false, 0), "USER TIME EXPECTATION") {
			t.Fatal("invented user time limit")
		}
	}
}
