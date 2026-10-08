# Architecture

Nutshell has two compilation layers. The AI translates English into implementation
source. A native toolchain turns that source into an executable. The Go driver
coordinates the process and controls publication.

```text
entry + available .nut files       selected implementation (-s)
             │                                  │
             └─────────── snapshot ──────────────┘
                             │
                    /tmp/nutshell-build-*
                             │
                   selected AI CLI (-c)
                 source + tests + build.json
                             │
                    run build → run tests
                             │
                validate artifact and snapshots
                             │
               publish binary + updated source
```

## Driver stages

1. Parse the CLI, load the bounded `.nut` bundle, validate destinations, and
   snapshot the selected implementation. Fine tuning requires existing source.
2. Load the named command from XDG compilers.json (initializing if absent) and
   require its executable and `rgw-ast` on `PATH`.
3. Create a private build directory under `/tmp`, copy source context, and write
   the shared compilation prompt.
4. Start the AI with the configured argv and prompt transport.
   Read activity updates while its output goes to a temporary log.
5. For fine tuning, validate fine-tune.json and reject conflicts or invalid reviews.
   Then parse the strict versioned build manifest and run its build/test commands.
6. Validate a native executable for the host and check for concurrent source
   changes. Publish source and binary under publication locks. Roll back source
   if binary publication fails.

The driver uses the Go standard library. It has no embedded model, editor,
terminal emulator, or language-specific code generator.

## Compiler configuration

All names resolve through the user's version-1 `compilers.json`. The embedded
`default-compilers.json` is an initialization template only: it is never consulted
to fill missing names in an existing config. There is no provider-name switch.
The driver honors an absolute XDG_CONFIG_HOME or falls back to ~/.config.

Initialization validates and merges legacy interpreter entries if present, then
publishes a completed temporary JSON file using a non-replacing hard link. Existing
and concurrently created configurations remain untouched. Runtime resolution
validates the config and resolves the executable before creating a build attempt.
Installers do not modify this configuration. Private wrappers remain local entries.

## Agent contract

All adapters receive the same prompt. The agent must choose reasonable assumptions,
write implementation code and tests in `src/`, and provide `build.json`:

```json
{
  "version": 1,
  "language": "Go",
  "summary": "Print a greeting",
  "assumptions": ["Names are supplied as one command-line argument"],
  "artifact": "bin/program",
  "build": [["go", "-C", "src", "build", "-o", "../bin/program", "."]],
  "test": [["go", "-C", "src", "test", "./..."]]
}
```

Commands are argv arrays executed from the build directory. Both command lists
are required. The driver runs build before test and does not evaluate command
strings as shell syntax; an agent can still explicitly invoke a shell. Artifact
paths must stay inside the build directory and cannot traverse symlinks.

The agent uses `rgw-ast --root src` for code inspection and changes, including
status, tracing, bounded reads and hash-checked edits. The prompt requires this
even below the tool's enforcement threshold. The driver checks availability,
not a tamper-proof trace of every action.

The first agent action writes a progress update. Each meaningful task should
produce another specific 3–7-word summary:

```json
{"version":1,"stage":"implementing","message":"Adding empty input validation tests"}
```

Stages are `understanding`, `implementing`, `checking`, and `ready`. The driver
owns later build, test, publication, and completion milestones. Invalid updates
are ignored; missing updates do not invent progress. A later task can change the
summary without moving the bar backward. Raw chat is not the progress protocol.

## Full compilation and fine tuning

Full compilation supplies the complete `.nut` bundle in the agent prompt.
`-f` / `--fine-tune` instead supplies a focused change request and a list of
snapshot paths. The AI must review all supplied `.nut` instructions against the
request before implementation. Only relevant implementation is inspected and
changed after compatibility is established. `.nut` instructions take precedence;
unrelated behavior, language and architecture are preserved.

The agent writes `fine-tune.json` in the temporary build directory. Version 1 has
`status` (`compatible` or `conflict`), `reviewed_sources` (every supplied relative
`.nut` path exactly once) and `conflicts` (an array, empty only for compatible).
Each conflict has `path`, `start_line`, `end_line`, `quote` and `reason`. Lines are
one-based and inclusive, including blank lines. Quotes preserve complete lines
joined by LF with CRLF normalized and the last terminator omitted.

The driver requires a regular bounded JSON file and validates schema, source
coverage, paths, ranges and exact quotations against its original in-memory input
snapshot. It rejects missing or invalid reviews before build/test commands or
publication. Valid conflicts produce diagnostics using original project paths and
source text, even if the agent exits nonzero. Changed inputs invalidate the review;
cancellation keeps its normal behavior. The source report stays under `/tmp`.
This validates evidence, not English semantics: identifying conflicts remains the
AI's responsibility. Review and implementation use the same AI session.

Fine tuning rejects missing, empty or metadata-only implementation before invoking
a provider. Both modes retain progress, rgw-ast, native build/test and protected
publication. Neither mode rewrites `.nut` inputs.

Every invocation creates a fresh build workspace and uses explicitly selected
source as context. The driver retains no AI session or build recipe between runs.
Neither mode injects a user time expectation. `--timeout` controls the hard
process cutoff independently of fine-tune scope.

The driver no longer generates source provenance, stages prior binaries, or offers
inspect/diff. A legacy `.nutshell-provenance.bin` in the selected source triggers
instructions to remove it and its embedding hooks together. Publication rejects a
leftover resource so failed migration preserves existing outputs. No new metadata
resource is embedded; ordinary application resources can still be embedded.

## Source and output ownership

`-s` explicitly selects both the initial source context and the source destination.
Nutshell never searches for a nearby `src/`. The selected directory can be absent
if its parent exists; creation at that location happens only after successful
verification. The temporary working directory is always named `src/`.

Existing context is bounded to 4,096 entries and 100 MiB, rejecting symlinks and
special files. The temporary `original-src/` preserves its initial snapshot.
Source fingerprints detect changes before publication. Existing `.nut` inputs
inside the selected source tree must survive publication unchanged.

The output binary must live outside the selected source directory. Individual
replacement steps use atomic filesystem operations; binary and source are not a
single filesystem transaction. Source rollback handles binary publication failure.

## Temporary evidence

Each attempt may retain source snapshots, `prompt.txt`, `source-context.json`,
`interpreter.log`, `verification.log`, `build.json`, `progress.json`, and
`result.json` (plus `fine-tune.json` for fine tuning) under `/tmp`. Failure output names the attempt directory. Success
prints the executable path without requiring the user to inspect build metadata.

Child processes inherit authentication and provider configuration. The driver
sets `GOWORK=off` and gives Go builds temporary cache/module directories. Unix
cancellation kills the process group; Windows cancellation targets the child.
The default compilation deadline is 30 minutes.

## Verification boundaries

The driver checks that declared commands succeeded, the output is a regular
native executable for the host, and publication inputs have not changed. Native
format checks cover ELF, Mach-O and PE. Linux is the primary tested runtime;
the Unix compiler integration fixtures also target macOS.

These checks do not prove the English requirements, the quality of generated
tests, freedom from malicious behavior, or reproducible generation. The agent
and build commands run with the caller's permissions. The `/tmp` layout isolates
working files by convention; it does not restrict process access to the machine.

## Code map

| Path | Responsibility |
| --- | --- |
| `main.go` | CLI options, cancellation, exit codes |
| `progress.go` | Terminal bar and redirected activity output |
| `internal/compiler/source.go` | Syntax-independent `.nut` discovery and validation |
| `internal/compiler/provider.go` | Generic configured command resolution and prompt transport |
| `internal/compiler/config.go` | XDG paths, initialization and legacy migration |
| `internal/compiler/default-compilers.json` | Editable starter template copied on first use |
| `internal/compiler/prompt.go` | Shared instructions for compilation agents |
| `internal/compiler/fine_tune.go` | Required conflict review validation and source diagnostics |
| `internal/compiler/compiler.go` | Generation, build/test orchestration, result records |
| `internal/compiler/context.go` | Selected implementation snapshots and seeding |
| `internal/compiler/export.go` | Source/binary publication, locking, rollback |
| `internal/compiler/artifact.go` | Native artifact validation |
| `internal/compiler/process*.go` | Child environment and process cancellation |
| `internal/compiler/progress.go` | Reading structured agent activity |

The [accepted specification](../openspec/specs/nutshell/spec.md) defines current
behavior. Proposed changes belong under `openspec/changes/` until implemented.
