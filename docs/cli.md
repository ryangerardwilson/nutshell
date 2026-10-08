# CLI and compiler reference

Write a program in English. Choose an AI compilation agent. Compile a native executable.

Nutshell's source files end in `.nut`; its default entry point is `main.nut`.
The Go driver gives your source to an installed AI CLI, which chooses an
implementation, writes code and tests, and supplies build commands. Nutshell runs
those commands and delivers the executable plus updated code in your chosen implementation directory.
All generation, compilation and testing happen under `/tmp`.

## Start

For a public release, use `go install github.com/ryangerardwilson/nutshell@latest`.
See [installation and versioning](releases.md) for PATH setup, pinned versions and
upgrades. The instructions below install your local checkout.

From the repository root, install with Go 1.26 or newer:

```sh
./install.sh from "$PWD"
```

The installer writes `~/.local/bin/nutshell`. Set `NUTSHELL_INSTALL_DIR` to change
that directory. Your selected AI CLI must already be installed, authenticated,
and available on `PATH`. `rgw-ast` must also be executable on `PATH`; compilation
fails before starting the AI if it is missing. Nutshell does not install providers,
rgw-ast, skills or plugins.

Write `main.nut`:

```text
Create a command-line program that writes "Hello from Nutshell!"
followed by a newline to standard output, then exits successfully.
```

Compile and run:

```sh
nutshell main.nut -c codex -s ./src
./main
```

Select the AI tool with `-c` or `--compiler`.

Other invocations:

```sh
nutshell main.nut -c grok -s ./implementation
nutshell main.nut -c claude -s ./src -o hello
nutshell -c codex -s ./src               # reads main.nut
nutshell main.nut -c codex -s ./src --timeout 10m
nutshell --help
```

Compilation does not automatically run your program except through its declared
tests. The default output is the entry path without `.nut`; Windows adds `.exe`.
An explicit `-o` path is relative to your shell's current directory. Its parent
directory must exist. Standard output contains the final executable path;
a compact progress display goes to standard error. Raw AI and build output stay
in temporary logs. There is no question-and-answer phase.

The local installer also provides `ns` as a short command. For example:

```sh
ns main.nut -c codex -s ./src
```

Both names run the same binary with the same arguments. For Go installations,
see the [one-time alias setup](releases.md#the-ns-command).

## The language

Use English to describe observable behavior, data, rules, examples, and errors.
Implementation constraints are allowed when you care about them. The AI chooses
routine details you leave open, including the native implementation language.
The implementation lives in the directory you select with `-s`; no particular framework is required.

The driver supplies the selected entry and available `.nut` files recursively
beneath its directory, preserving names and exact text. It does not parse an
import language. `from feature1.nut import abc`, `Include "rules.nut"`, and ordinary
English references are all interpreted by the AI. Discovery skips hidden
directories and `node_modules`, `vendor`, `target`, `dist`, and `build`; it never
follows directory symlinks and rejects `.nut` file symlinks. Source is bounded to
128 nonempty UTF-8 files and 1 MiB of combined text. See the
[language guide](language.md) for the input boundary.

The compilation agent makes whatever assumptions it needs, including resolving missing
or conflicting requirements. It records those decisions internally and continues.
It does not ask you questions or stop solely to request clarification.

This is an initial English-first language implementation. The generated implementation can vary
between models or runs. Passing generated tests establishes that the chosen
implementation builds and passes those tests; it does not formally prove that
every English requirement was implemented correctly. Review the source, tests,
and assumptions when that distinction matters.

See [hello](../examples/hello/main.nut) and the multi-file
[greeting example](../examples/greet/main.nut).

## Fine tuning

For a targeted follow-up edit, use existing implementation source:

```sh
nutshell main.nut -c grok -o app -s ./src -f "replace x with y"
ns main.nut -c codex -s ./src --fine-tune "Reject negative quantities"
```

`-f` and `--fine-tune` accept one quoted, nonblank change request. Flags work
before or after the entry; `--fine-tune="replace x with y"` also works. The `-s`
directory must already contain implementation files. Missing, empty or
metadata-only source fails before invoking the AI; compile without `-f` first.

The AI receives the request and `.nut` snapshot paths. It reviews all supplied
`.nut` instructions for conflicts before editing relevant implementation code.
It preserves unrelated behavior, language and architecture. A compatible request
continues through the normal build, tests, rgw-ast and publication checks.

`.nut` instructions always take precedence. For a source line that requires
`hello world`, requesting `replace world with everyone` is rejected. Diagnostics
identify every observed conflict across files with one-based line numbers, exact
source text and a reason, for example:

```text
nutshell: fine-tune conflicts with .nut instructions:
/project/main.nut:3: Replacing world changes the required literal output.
  3 | Print exactly "hello world" followed by a newline.
/project/features/greeting.nut:2: The request changes the required greeting.
  2 | The output text must be exactly "hello world".
Change the fine-tune request, or update the .nut instructions first.
```

A conflict returns exit status 1 without running declared build commands or
publishing outputs. `.nut` inputs, implementation and binary stay unchanged.
Missing or invalid reviews also fail; the AI cannot proceed merely by omitting
its review. The driver validates paths, line ranges and quoted source text.
Semantic conflict detection still depends on the selected AI. Source changes
during review invalidate diagnostics rather than reporting stale line numbers.
The former `-ft` flag is rejected with guidance to use `-f` or `--fine-tune`.

Every compilation starts in a fresh private `/tmp` workspace, with implementation
context copied from the explicitly selected `-s` directory. Fine tuning limits
scope, not duration. There is no retained AI session or build recipe between runs.

The former `-l` / `--limit` options are removed. Omit the flag and its value from
old commands; they fail before starting an AI. `--timeout` controls the hard
process deadline (default 30 minutes), without adding time-pressure instructions
to the AI prompt.

Version 0.10.0 removes `inspect`, `diff`, prior-binary context and embedded `.nut`
provenance. Existing 0.9 source can be reused; the agent is told to remove the old
`.nutshell-provenance.bin` and its embedding hooks together, preserving application
resources. A leftover resource prevents publication. Keep your `.nut` files
separately; the binary no longer serves as their archive.

## Compilation agents

All compiler commands live in `compilers.json`. Use an absolute `XDG_CONFIG_HOME`
for `$XDG_CONFIG_HOME/nutshell/compilers.json`; otherwise the path is
`~/.config/nutshell/compilers.json`. Empty or relative XDG values use that fallback,
as specified by the [XDG base directory standard](https://specifications.freedesktop.org/basedir/latest/).

```sh
ns config path   # print the effective path without creating anything
ns config init   # create editable starter entries if the file is absent
```

The first compilation also initializes a missing file. The starter template
contains `codex`, `grok`, and `claude` (Claude Code), using unattended full-permission
flags and each CLI's default model. It is a starting point, not a runtime fallback.
All names can be changed, overridden or deleted; Nutshell executes the selected
entry exactly as configured. Reinstallation never resets your file. Help and
version do not initialize it. No providers, credentials or skills are installed.

| Starter name | Invocation settings |
| --- | --- |
| `codex` | `exec --dangerously-bypass-approvals-and-sandbox --skip-git-repo-check --ephemeral --color never -`; prompt on stdin |
| `grok` | `--always-approve --sandbox off --no-plan --prompt-file …` |
| `claude` | `--print --dangerously-skip-permissions --output-format text --no-session-persistence`; prompt on stdin |

Authentication, model profiles and provider policies remain with the CLI. Build
directories are separate working directories, not permission sandboxes. Generated
commands run with the current user's permissions. `GOWORK=off` keeps generated Go
modules independent of an enclosing workspace.

Every compiler receives a mandatory `rgw-ast` workflow: check status against the
temporary `src/`, trace existing code before changes, inspect with bounded reads,
and use hash-checked edits and explicit file creation. This applies even when
`rgw-ast` reports enforcement disabled for a small project. Progress and build
metadata outside `src/` can use ordinary writes. Agents must honor existing policy
and stop on unresolved tool failures rather than bypass it.
This is an agent instruction contract; Nutshell checks tool availability but does
not independently verify every agent operation or install provider enforcement hooks.

Adapter references: [Codex CLI](https://developers.openai.com/codex/cli/reference),
[Grok scripting](https://docs.x.ai/build/cli/headless-scripting),
[Grok sandbox profiles](https://docs.x.ai/build/features/sandbox).
Installed CLI help is the source for the tested flag spelling.

Add a tool to the `compilers` object in your file, for example:

```json
{
  "version": 1,
  "compilers": {
    "my-ai": {
      "command": ["my-ai-cli", "generate", "--its-yolo-flag", "--prompt-file", "{prompt_file}"],
      "prompt": "file"
    },
    "my-wrapper": {
      "command": ["/absolute/path/to/my-wrapper", "exec", "-"],
      "prompt": "stdin"
    }
  }
}
```

Replace the example flags with your tool's actual unattended options. The example
shows a complete file; merge entries into your existing map to keep its other
compilers. Select any entry with `-c my-ai` or `-c my-wrapper`. User-specific wrappers
and profiles belong here, not in Nutshell's code or shipped defaults.

`command` is an argv array, not a shell command string. Its executable must be a
name on `PATH` or an absolute path. There is no automatic shell, `~`, or environment
variable expansion. Paths with spaces work as one array element. Arguments are
passed literally except for the documented prompt placeholders:

- `stdin`: the complete compilation prompt is piped to standard input.
- `file`: `{prompt_file}` is replaced with the temporary prompt file path.
- `argument`: `{prompt}` is replaced with the complete prompt as argument text.

`unsafe_args` is optional for compatibility with older adapters. When supplied,
its arguments replace a standalone `{unsafe_args}` token, or are inserted after
the executable if that token is absent. Permission flags can instead be written
directly in `command`; wrappers that already set permissions need no separate flags.
Nutshell neither appends provider-specific options nor selects a model at runtime.
Unknown names and invalid definitions fail with the configuration path.

If only the old `interpreters.json` exists, initialization copies its entries into
the new versioned `compilers` map alongside starter defaults. User entries win name
collisions. The legacy file stays untouched. Invalid legacy data stops migration;
an existing `compilers.json` always takes precedence and is never silently reset.

## What you get

After compiling `main.nut` with `-s ./src`, your project contains:

```text
main.nut            your Nutshell program
main                compiled native executable
src/                generated implementation, tests and project manifests
```

`-s <path>` (or `--source <path>`) is required for compilation. It selects both
implementation context and the destination for updated code. Relative paths resolve
from your current working directory, independently of the entry file and `-o`.
Nutshell does not discover or fall back to any nearby `src/` directory.

If the selected directory exists, Nutshell copies it into the temporary workspace
and tells the AI to inspect and adapt it, including manual edits, tests and assets.
If it does not exist, the AI starts a new implementation and Nutshell creates the
selected directory only after successful compilation. Its parent must already exist.
No ownership receipt is required.

`-o` selects only the executable. For example, with both parent directories present:

```sh
nutshell main.nut -c grok -s ./implementation -o ./bin/app
```

This updates `implementation/` and delivers `bin/app`. The executable must be outside
the selected source directory. Use a quoted path for names containing spaces.

Only a verified build is published. If a source file is added, removed or changed
while compilation is running, Nutshell stops publication and preserves your latest
files and previous binary. Existing `.nut` inputs within the selected directory are also protected.
Source trees are limited to 4,096 entries and 100 MiB, without symlinks or special
files. Old `.nutshell-generated.json` receipts are ignored and omitted from the
updated source.

The progress bar advances through compilation milestones: generation, checks, native build,
tests and publication. It represents milestones, not a prediction of remaining
time. A spinner and elapsed time remain visible while the AI works. Redirected
output shows each changed activity without terminal control codes. Descriptions
are 3–7 words, such as "Writing command line argument parser" or "Adding empty
input validation tests". The AI reports individual tasks within a phase, not just
phase changes. If no update arrives for 30 seconds, the terminal shows "Waiting
for AI progress update"; it does not simulate activity.

## Temporary build work

Every attempt uses a private `/tmp/nutshell-build-*` directory. Nutshell does not
create a `.nutshell/` folder in your project. Old `.nutshell/` folders from earlier
versions are left alone. The temporary directory contains original source
snapshots, working `src/`, prompts, AI output, build/test logs, assumptions and
result metadata. When reusing source, `original-src/` keeps its initial snapshot
and `source-context.json` records its origin and fingerprints. Child commands use temporary scratch and Go cache directories
inside that attempt. Authentication and other provider configuration are inherited.
Temporary evidence remains available until it or `/tmp` is cleaned up; its location
is only printed when compilation fails.

The AI reports current work by writing `progress.json` in the build directory:

```json
{"version":1,"stage":"implementing","message":"Writing the command-line program"}
```

AI stages: `understanding`, `implementing`, `checking`, `ready`. The first tool
action writes an update, and the AI updates before each meaningful task with
a specific 3–7-word message. Repeated stages can carry new summaries. Write via
a temporary file and rename into place. Invalid or missing updates do not
interrupt compilation. Returning to implementation after checking can update
the activity without moving the bar backward. Nutshell owns later
build, test, publication and completion milestones. This protocol works independently
of the selected AI tool's chat output format.

The compilation agent still writes a version-1 `build.json` with language, summary,
assumptions, native artifact and nonempty build/test argv lists. Generated code,
tests, module manifests and application resources go under the temporary
`src/` working tree; binaries,
logs, caches and scratch files stay outside `src/` in the temporary build directory.
Build commands run from that directory (for Go, use `go -C src build ...`). The
driver publishes the working tree back to the explicit `-s` path, regardless of
the private workspace name.

Nutshell verifies builds and tests, checks native executable format and unchanged
`.nut` inputs and implementation snapshots, then publishes the binary and updated
source. Failed compilation preserves the previous output. Source publication is rolled back if binary
publication fails. Ctrl-C, SIGTERM or the deadline (30 minutes by default) stops
the running process group on Linux/macOS; Windows cancels the direct child.

Native validation supports host ELF on Linux, Mach-O on macOS, and PE on Windows.
The compilation agent is instructed to embed required application resources so the binary
runs independently of the temporary build directory. System runtime libraries may
still be needed; state portability requirements in your `.nut` source.

## Development

```sh
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
```

Tests cover source loading, command adapters, a fake compilation agent with a real Go
build/test cycle, native output, live progress updates, quiet terminal rendering,
cancellation, source changes, source publication rollback and previous-output
preservation. Unit tests do not call paid AI services.
