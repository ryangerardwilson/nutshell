# The executable carries its source

Every program successfully compiled with Nutshell 0.9.0 or newer contains a
provenance resource. The resource preserves the entry point and the exact text
and relative paths of the `.nut` files supplied to the AI, together with:

- A canonical SHA-256 of the complete input bundle.
- The Nutshell version and selected AI compiler name.
- The implementation language and the AI's recorded assumptions.

The bundle is the compiler's input context. It is not a claim that every available
file was semantically used. Provider model identity is not invented: adapters use
the tool's default model, and the driver does not receive a reliable model identity.

## Inspect without running

```sh
ns inspect ./app
ns inspect ./app --source
ns inspect ./app --json
```

The default summary lists files, compiler information and assumptions. `--source`
prints every original file under a path heading; it does not write extracted files
to disk. `--json` returns the full versioned record, including exact source strings.
It works even if the original source directory no longer exists.

These commands only read file data. They do not execute the application, start an
AI session, require `rgw-ast`, or depend on the application's runtime arguments.
Both `ns` and `nutshell` work. Inspection exits 0 on success and 2 on errors.

## Compare the recorded program with current source

```sh
ns diff ./app main.nut
ns diff ./app main.nut --json
```

The entry defaults to `main.nut` if omitted. Diff discovers the current `.nut`
bundle using the same boundary as compilation, then reports added, removed and
modified files and changes to the selected entry. A change in `feature1.nut`
is visible even if `main.nut` is unchanged. Equal files discovered in a different
order do not count as a change.

Human output uses whole-file unified hunks. JSON returns `version`, `changed`,
old/new entries and source hashes, and changes with `path`, `kind`, `before` and
`after`. Exit codes are **0 equal**, **1 different**, **2 error**. An agent must
not mistake exit 1 for a failed inspection.

This compares requirements, not native instructions or runtime behavior.

## What the compilation agent receives

When the requested output already exists, the driver copies it into the temporary
workspace as read-only `previous-program`. `previous-context.json` contains its
recovered provenance and current source comparison, or an explicit reason the
record is unavailable. A legacy executable does not block recompilation.

The initial prompt explains inspect/diff, gives commands using the running
Nutshell executable, and points to these staged files. The AI can understand which
requirements changed without accessing the original project or executing the old
program. Current `.nut` source remains authoritative.

## How embedding works

The driver owns `src/.nutshell-provenance.bin`. Before the agent starts it supplies
the source resource; after receiving build.json it rewrites the resource with the
final language and assumptions, then runs the declared build and tests.

The AI must embed that file through its chosen toolchain and keep the full bytes
in the executable. Examples include Go embed, Rust include_bytes!, or a native
resource object for C. An unused constant may be stripped by a linker: the agent
must retain the resource, for example through a startup check. Inlining a stale
copy is not sufficient because the driver updates metadata before the final build.
Reading the resource from disk at runtime does not satisfy the contract.

Nutshell statically verifies the exact expected record in the final binary before
publication. Missing, stale or conflicting provenance fails the build and leaves
the old binary and implementation in place. The resource is published with the
implementation source, so ordinary rebuilds can include it. Regenerating the
resource for changed `.nut` requirements is Nutshell's job.

## Format and boundaries

The v1 frame is a marker (`\0NUTSHELL-PROVENANCE-v1\0`), an unsigned 64-bit
little-endian JSON byte length, the UTF-8 JSON payload, and its 32-byte SHA-256.
The JSON record has `version`, `entry`, `sources`, `source_sha256`,
`nutshell_version`, `compiler`, `language`, and `assumptions` fields. Each source
has `path` and `text`. Payloads are limited to 8 MiB; source limits remain 128
files and 1 MiB of combined text. Inspection validates the schema, relative paths,
checksum and canonical source hash before exposing a record.

Canonical source hashing uses compact JSON with fields `entry` then `sources`,
where sources are sorted by path and each has fields `path` then `text`, encoded
with Go's encoding/json defaults. Identical embedded records may occur more than
once; different valid records are rejected as ambiguous.

The checksum detects accidental resource damage. It is not a signature and does
not authenticate the publisher or bind the program's machine instructions to its
claimed requirements. The driver verifies its generated artifact before publication;
inspection later recovers that claim. Tests and code review still assess behavior.

The bundle is readable plaintext inside the executable. Distribute it with the
same expectations as distributing the original `.nut` text. Older binaries have
no recoverable record; recompile them to enable inspection. External stripping or
rewriting tools may remove the resource and make subsequent inspection fail.
