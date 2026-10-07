package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ryangerardwilson/nutshell/internal/compiler"
)

const version = "0.6.0"

const help = `nutshell — compile English-first .nut programs with your AI tool

Usage: nutshell [main.nut] -c <codex|grok|claude|custom> -s <path> [-o output]

  -c, --compiler TOOL     AI compilation agent (required)
  -s, --source PATH       Implementation directory to read and update (required)
  -o, --output PATH       Native executable; defaults to entry without .nut
      --timeout DURATION  Total compilation deadline (default 30m)
  -h, --help              Show help
  -v, --version           Show version

Options work before or after the entry file. Default entry: main.nut.
Compilation agents run unattended with full permissions and their default models.
The AI makes assumptions automatically; no clarification questions.
Build work and logs stay under /tmp. Progress appears while it compiles.
The directory selected by -s supplies context and receives updated source.
Relative -s paths resolve from your current directory; its parent must exist.
Other tools: ~/.config/nutshell/interpreters.json (respects XDG_CONFIG_HOME).

Example: nutshell main.nut -c codex -s ./src
`

type arguments struct {
	compiler.Options
	Timeout time.Duration
	Action  string
}

func parse(args []string) (arguments, error) {
	o := arguments{Options: compiler.Options{Entry: "main.nut"}, Timeout: 30 * time.Minute}
	entrySeen := false
	literal := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !literal && arg == "--" {
			literal = true
			continue
		}
		if !literal {
			switch arg {
			case "help", "--help", "-h":
				o.Action = "help"
				return o, nil
			case "version", "--version", "-v":
				o.Action = "version"
				return o, nil
			}
			key, value, hasValue := strings.Cut(arg, "=")
			switch key {
			case "-i", "--interpreter":
				return o, fmt.Errorf("%s was replaced by -c (or --compiler)", key)
			case "-c", "--compiler", "-s", "--source", "-o", "--output", "--timeout":
				if !hasValue {
					i++
					if i >= len(args) {
						return o, fmt.Errorf("%s requires a value", key)
					}
					value = args[i]
				}
				if value == "" {
					return o, fmt.Errorf("%s requires a nonempty value", key)
				}
				switch key {
				case "-c", "--compiler":
					o.Interpreter = value
				case "-s", "--source":
					o.SourceDir = value
				case "-o", "--output":
					o.Output = value
				case "--timeout":
					d, err := time.ParseDuration(value)
					if err != nil || d <= 0 {
						return o, fmt.Errorf("timeout must be a positive duration, e.g. 30m")
					}
					o.Timeout = d
				}
				continue
			}
			if strings.HasPrefix(arg, "-") {
				return o, fmt.Errorf("unknown option %s", arg)
			}
		}
		if entrySeen {
			return o, fmt.Errorf("expected one .nut entry file, got %q", arg)
		}
		o.Entry, entrySeen = arg, true
	}
	if o.Interpreter == "" {
		return o, fmt.Errorf("choose a compilation agent with -c codex, -c grok, -c claude, or a configured tool")
	}
	if o.SourceDir == "" {
		return o, fmt.Errorf("choose an implementation source directory with -s <path>, for example -s ./src")
	}
	return o, nil
}

func execute(ctx context.Context, args []string, out, errOut io.Writer) int {
	o, err := parse(args)
	if err != nil {
		fmt.Fprintf(errOut, "nutshell: %v\nRun nutshell --help for usage.\n", err)
		return 2
	}
	switch o.Action {
	case "help":
		fmt.Fprint(out, help)
		return 0
	case "version":
		fmt.Fprintln(out, "nutshell "+version)
		return 0
	}
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)
	defer cancel()
	display := newProgressDisplay(errOut, isTerminal(errOut))
	o.OnProgress = display.update
	result, err := compiler.Compile(ctx, o.Options)
	display.finish(err)
	if err != nil {
		fmt.Fprintf(errOut, "nutshell: %v\n", err)
		if result.Directory != "" {
			fmt.Fprintf(errOut, "Details: %s\n", result.Directory)
		}
		if errors.Is(err, context.Canceled) {
			return 130
		}
		return 1
	}
	fmt.Fprintln(out, result.Output)
	return 0
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(execute(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
