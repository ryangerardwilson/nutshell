package compiler

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type fineTuneReview struct {
	Version         int                `json:"version"`
	Status          string             `json:"status"`
	ReviewedSources []string           `json:"reviewed_sources"`
	Conflicts       []fineTuneConflict `json:"conflicts"`
}

type fineTuneConflict struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Quote     string `json:"quote"`
	Reason    string `json:"reason"`
}

// Line numbers are one-based physical lines, with CRLF normalized for quotations.
func sourceLines(text string) []string {
	return strings.Split(strings.TrimSuffix(strings.ReplaceAll(text, "\r\n", "\n"), "\n"), "\n")
}

func checkFineTuneReview(dir string, p Program) error {
	var review fineTuneReview
	if err := readJSON(filepath.Join(dir, "fine-tune.json"), &review); err != nil {
		return fmt.Errorf("fine tuning requires a valid fine-tune.json conflict review: %w", err)
	}
	if err := review.validate(p); err != nil {
		return fmt.Errorf("invalid fine-tune.json conflict review: %w", err)
	}
	// Do not attribute a snapshot's coordinates to files that have since changed.
	if err := p.unchanged(); err != nil {
		return err
	}
	if review.Status == "compatible" {
		return nil
	}
	sort.SliceStable(review.Conflicts, func(i, j int) bool {
		a, b := review.Conflicts[i], review.Conflicts[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.StartLine != b.StartLine {
			return a.StartLine < b.StartLine
		}
		return a.EndLine < b.EndLine
	})
	texts := make(map[string][]string, len(p.Sources))
	for _, source := range p.Sources {
		texts[source.Path] = sourceLines(source.Text)
	}
	var message strings.Builder
	message.WriteString("fine-tune conflicts with .nut instructions:")
	for _, conflict := range review.Conflicts {
		location := fmt.Sprintf("%s:%d", filepath.Join(p.Root, filepath.FromSlash(conflict.Path)), conflict.StartLine)
		if conflict.EndLine != conflict.StartLine {
			location += fmt.Sprintf("-%d", conflict.EndLine)
		}
		fmt.Fprintf(&message, "\n%s: %s", location, conflict.Reason)
		for i := conflict.StartLine; i <= conflict.EndLine; i++ {
			fmt.Fprintf(&message, "\n  %d | %s", i, texts[conflict.Path][i-1])
		}
	}
	message.WriteString("\nChange the fine-tune request, or update the .nut instructions first.")
	return fmt.Errorf("%s", message.String())
}

func (review fineTuneReview) validate(p Program) error {
	if review.Version != 1 || (review.Status != "compatible" && review.Status != "conflict") {
		return fmt.Errorf("expected version 1 and status compatible or conflict")
	}
	sources := make(map[string][]string, len(p.Sources))
	for _, source := range p.Sources {
		sources[source.Path] = sourceLines(source.Text)
	}
	seen := make(map[string]bool, len(review.ReviewedSources))
	for _, path := range review.ReviewedSources {
		if _, exists := sources[path]; !exists || seen[path] {
			return fmt.Errorf("unknown or duplicate reviewed source %q", path)
		}
		seen[path] = true
	}
	if len(seen) != len(sources) {
		return fmt.Errorf("reviewed_sources must include every supplied .nut file")
	}
	if review.Conflicts == nil || (review.Status == "compatible") != (len(review.Conflicts) == 0) {
		return fmt.Errorf("compatible requires an empty conflicts array; conflict requires at least one conflict")
	}
	for _, conflict := range review.Conflicts {
		lines, exists := sources[conflict.Path]
		if !exists || conflict.StartLine < 1 || conflict.EndLine < conflict.StartLine || conflict.EndLine > len(lines) {
			return fmt.Errorf("invalid source path or line range for %q:%d-%d", conflict.Path, conflict.StartLine, conflict.EndLine)
		}
		if strings.TrimSpace(conflict.Reason) == "" {
			return fmt.Errorf("conflict explanation is required")
		}
		if conflict.Quote != strings.Join(lines[conflict.StartLine-1:conflict.EndLine], "\n") {
			return fmt.Errorf("conflict quote does not match %q:%d-%d", conflict.Path, conflict.StartLine, conflict.EndLine)
		}
	}
	return nil
}
