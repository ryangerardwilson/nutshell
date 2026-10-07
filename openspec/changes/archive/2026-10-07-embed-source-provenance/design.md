# Design

## Source boundary
The driver discovers bounded local .nut files, with the selected entry first and
other files ordered by path. Directory traversal skips hidden directories and
node_modules, vendor, target, dist and build. It never follows symlinks. The AI
receives every discovered file as available context, and chooses references and
symbol semantics from the original text. No import grammar is introduced.

## Embedded resource
The driver owns src/.nutshell-provenance.bin. It writes a placeholder resource
before the AI session, then replaces it with the final source bundle, entry,
source hash, driver version, AI compiler name, implementation language and
manifest assumptions before running build commands. The agent must embed the
file's bytes as a retained resource in its native executable, not inline a stale
copy or read it from disk at runtime. The driver checks the final embedded record
matches its expected record before publishing.

Embedding during native compilation avoids appending bytes to signed Mach-O or
PE executables after their toolchain has built/signed them. Native code signing
is outside Nutshell's contract; source provenance itself is not a signature.

## Static inspection
A versioned marker, bounded length, JSON payload and SHA-256 checksum frame the
resource. Inspection scans regular files in bounded chunks, validates records,
and rejects conflicting valid records. Identical copies are allowed. It never
loads or executes code from the inspected binary. Source hashes use a canonical
entry and path-sorted source list. Limits apply before allocation and parsing.

Diff compares the complete recorded bundle with the newly discovered bundle:
entry changes, added/removed files and exact text changes. JSON carries before
and after text; human output uses whole-file unified hunks. Exit codes are 0 for
equal, 1 for differences and 2 for inspection/diff errors.

## Incremental context and failure
An existing output is copied into the temporary workspace as previous-program.
Static metadata and source diff, or an explicit unavailable reason for a legacy
binary, are recorded for the agent. The prompt names inspect/diff and supplies
argv examples using the running driver's absolute path. No previous program is
executed. Missing, corrupt, ambiguous or mismatched provenance in newly generated
output prevents publication and preserves prior source and executable.
