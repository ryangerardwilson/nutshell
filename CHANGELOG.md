# Changelog

Versions follow Semantic Versioning. Dates use YYYY-MM-DD. The matching Git tag
has a `v` prefix; published tags are never moved.

## [0.9.0] - 2026-10-07

### Added
- Generated executables carry the original `.nut` bundle, compiler identity,
  implementation language and assumptions as verified embedded provenance.
- Static `ns inspect` and `ns diff`, with source recovery, JSON output and source
  comparisons that never execute the inspected program.
- Prior-output snapshots, provenance and source diffs in compilation-agent context.

### Changed
- Source discovery now supplies the entry and available `.nut` files beneath its
  directory. File composition and all syntax, including `from`, `import` and
  `Include`, belong to the AI compiler.
- Agents must embed the supplied provenance resource during native compilation;
  binaries missing it are rejected before publication.
- Migration: keep intended `.nut` context beneath the entry directory and move
  unrelated inputs outside it. Hidden and dependency/build directories are skipped.
  Existing executables need recompilation to gain inspectable provenance.

## [0.8.0] - 2026-10-07

### Added
- `ns` as a short command for Nutshell, installed beside the same executable.
- Installer conflict checks that preserve unrelated existing `ns` commands.
- Documentation for adding the alias to Go-installed binaries.

## [0.7.0] - 2026-10-07

### Added
- One-command installation through Go with latest or pinned version selection.
- Explicit local installation with `install.sh from <checkout>`.
- Embedded VERSION, release consistency checks, guarded tag publication and
  automatic GitHub release notes.
- Documented agent maintenance and local reinstall workflow.

## [0.6.0] - 2026-10-07

### Initial public source
- English-first `.nut` programs compiled through Codex, Grok, Claude or a custom CLI.
- Explicit implementation context through `-s`, temporary build work, dynamic
  progress summaries, native artifact verification and source publication.
- Mandatory rgw-ast instructions and dependency preflight.
- Independent MIT-licensed repository, examples, documentation and CI.

0.6.0 was the initial source snapshot; 0.7.0 is the first tagged public release.
