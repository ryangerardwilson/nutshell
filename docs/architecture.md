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
   snapshot the selected implementation.
2. Resolve the AI adapter and require `rgw-ast` on `PATH`.
3. Create a private build directory under `/tmp`, copy source context, and write
   the shared compilation prompt.
4. Start the AI with unattended full-permission settings and no model override.
   Read activity updates while its output goes to a temporary log.
5. Parse the strict versioned build manifest and run its build and test commands.
6. Validate a native executable for the host, verify its embedded provenance, and check for concurrent source
   changes. Publish source and binary under publication locks. Roll back source
   if binary publication fails.

The driver uses the Go standard library. It has no embedded model, editor,
terminal emulator, or language-specific code generator.

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

## Embedded provenance and inspection

Before the AI starts, the driver writes `src/.nutshell-provenance.bin`. The agent
embeds its opaque bytes as a retained resource using the selected native toolchain.
After reading build.json, the driver updates that resource with final assumptions
and implementation language before running build commands. The completed executable
must contain the exact expected record or publication fails.

The binary resource is part of compilation; the driver does not append metadata
to a finished signed executable. This respects native formats whose signing data
lives in the executable itself. See [Apple's signing procedures](https://developer.apple.com/library/archive/documentation/Security/Conceptual/CodeSigningGuide/Procedures/Procedures.html)
and [Microsoft's PE format](https://learn.microsoft.com/en-us/windows/win32/debug/pe-format).

Static inspection scans regular files in chunks, checks framed JSON payloads and
source hashes, and rejects conflicting valid records. It never executes the target.
The previous output is staged read-only and inspected before the next AI session;
its source diff or unavailable reason is in `previous-context.json`. The initial
prompt describes inspect/diff and supplies argv examples using the driver's path.
See [the provenance format](provenance.md) for details and limitations.

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
`result.json` under `/tmp`. Failure output names the attempt directory. Success
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
| `internal/compiler/provenance.go` | Resource framing, static inspection and publication verification |
| `internal/compiler/source_diff.go` | Canonical bundle comparison and human diff output |
| `internal/compiler/previous.go` | Static prior-output context for compilation agents |
| `inspect.go` | inspect/diff CLI and output modes |
| `internal/compiler/provider.go` | Built-in and custom CLI adapters |
| `internal/compiler/prompt.go` | Shared instructions for compilation agents |
| `internal/compiler/compiler.go` | Generation, build/test orchestration, result records |
| `internal/compiler/context.go` | Selected implementation snapshots and seeding |
| `internal/compiler/export.go` | Source/binary publication, locking, rollback |
| `internal/compiler/artifact.go` | Native artifact validation |
| `internal/compiler/process*.go` | Child environment and process cancellation |
| `internal/compiler/progress.go` | Reading structured agent activity |

The [accepted specification](../openspec/specs/nutshell/spec.md) defines current
behavior. Proposed changes belong under `openspec/changes/` until implemented.
