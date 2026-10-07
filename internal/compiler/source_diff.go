package compiler

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type SourceChange struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

type SourceDiff struct {
	Version         int            `json:"version"`
	Changed         bool           `json:"changed"`
	OldEntry        string         `json:"old_entry"`
	NewEntry        string         `json:"new_entry"`
	OldSourceSHA256 string         `json:"old_source_sha256"`
	NewSourceSHA256 string         `json:"new_source_sha256"`
	Changes         []SourceChange `json:"changes"`
}

func DiffSources(previous Provenance, current Program) SourceDiff {
	d := SourceDiff{Version: 1, OldEntry: previous.Entry, NewEntry: current.Entry, OldSourceSHA256: previous.SourceSHA256, NewSourceSHA256: SourceHash(current), Changes: []SourceChange{}}
	d.Changed = d.OldSourceSHA256 != d.NewSourceSHA256
	before, after := map[string]string{}, map[string]string{}
	paths := map[string]bool{}
	for _, s := range previous.Sources {
		before[s.Path], paths[s.Path] = s.Text, true
	}
	for _, s := range current.Sources {
		after[s.Path], paths[s.Path] = s.Text, true
	}
	names := make([]string, 0, len(paths))
	for name := range paths {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		old, had := before[name]
		newText, has := after[name]
		if had && has && old == newText {
			continue
		}
		kind := "modified"
		if !had {
			kind = "added"
		} else if !has {
			kind = "removed"
		}
		d.Changes = append(d.Changes, SourceChange{name, kind, old, newText})
	}
	return d
}

// WriteText uses whole-file unified hunks to keep memory bounded for large lines.
func (d SourceDiff) WriteText(w io.Writer) {
	if !d.Changed {
		fmt.Fprintln(w, "Source bundle matches the executable.")
		return
	}
	if d.OldEntry != d.NewEntry {
		fmt.Fprintf(w, "Entry: %s -> %s\n", d.OldEntry, d.NewEntry)
	}
	for _, c := range d.Changes {
		old, newName := "a/"+c.Path, "b/"+c.Path
		if c.Kind == "added" {
			old = "/dev/null"
		}
		if c.Kind == "removed" {
			newName = "/dev/null"
		}
		before, after := textLines(c.Before), textLines(c.After)
		oldStart, newStart := 1, 1
		if len(before) == 0 {
			oldStart = 0
		}
		if len(after) == 0 {
			newStart = 0
		}
		fmt.Fprintf(w, "--- %s\n+++ %s\n@@ -%d,%d +%d,%d @@\n", old, newName, oldStart, len(before), newStart, len(after))
		for _, part := range []struct{ prefix, text string }{{"-", c.Before}, {"+", c.After}} {
			for _, line := range textLines(part.text) {
				fmt.Fprintf(w, "%s%s\n", part.prefix, line)
			}
			if part.text != "" && !strings.HasSuffix(part.text, "\n") {
				fmt.Fprintln(w, "\\ No newline at end of file")
			}
		}
	}
}

func textLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}
