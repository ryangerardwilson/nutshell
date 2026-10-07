package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ryangerardwilson/nutshell/internal/compiler"
)

const inspectionHelp = `Usage: ns inspect <binary> [--source | --json]
       ns diff <binary> [entry.nut] [--json]

Read embedded .nut source without executing the binary or invoking an AI.
inspect defaults to a provenance summary; --source prints every original file.
diff defaults to main.nut and compares the complete discovered source bundle.
diff exits 0 if equal, 1 if different, and 2 on errors. inspect errors exit 2.
`

func inspectCommand(args []string, out, errOut io.Writer) int {
	action := args[0]
	var paths []string
	jsonOutput, sourceOutput, literal := false, false, false
	fail := func(err error) int {
		fmt.Fprintf(errOut, "nutshell: %v\n", err)
		return 2
	}
	for _, arg := range args[1:] {
		if !literal {
			switch arg {
			case "--":
				literal = true
				continue
			case "--help", "-h":
				fmt.Fprint(out, inspectionHelp)
				return 0
			case "--json":
				jsonOutput = true
				continue
			case "--source":
				sourceOutput = true
				continue
			}
			if strings.HasPrefix(arg, "-") {
				return fail(fmt.Errorf("unknown inspection option %s", arg))
			}
		}
		paths = append(paths, arg)
	}
	if len(paths) == 0 || len(paths) > 2 || (action == "inspect" && len(paths) != 1) || (sourceOutput && (jsonOutput || action == "diff")) {
		return fail(fmt.Errorf("invalid arguments\n%s", inspectionHelp))
	}
	p, err := compiler.Inspect(paths[0])
	if err != nil {
		return fail(err)
	}
	writeJSON := func(value any) error {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	}
	if action == "inspect" {
		if jsonOutput {
			if err := writeJSON(p); err != nil {
				return fail(err)
			}
		} else if sourceOutput {
			for _, s := range p.Sources {
				fmt.Fprintf(out, "--- %s ---\n%s", s.Path, s.Text)
				if !strings.HasSuffix(s.Text, "\n") {
					fmt.Fprintln(out)
				}
			}
		} else {
			fmt.Fprintf(out, "Entry: %s\nNutshell: %s\nCompiler: %s\nLanguage: %s\nSource SHA-256: %s\n", p.Entry, p.NutshellVersion, p.Compiler, p.Language, p.SourceSHA256)
			for _, s := range p.Sources {
				fmt.Fprintf(out, "  %s (%d bytes)\n", s.Path, len(s.Text))
			}
			for _, assumption := range p.Assumptions {
				fmt.Fprintf(out, "Assumption: %s\n", assumption)
			}
		}
		return 0
	}
	entry := "main.nut"
	if len(paths) == 2 {
		entry = paths[1]
	}
	current, err := compiler.Load(entry)
	if err != nil {
		return fail(err)
	}
	d := compiler.DiffSources(p, current)
	if jsonOutput {
		if err := writeJSON(d); err != nil {
			return fail(err)
		}
	} else {
		d.WriteText(out)
	}
	if d.Changed {
		return 1
	}
	return 0
}
