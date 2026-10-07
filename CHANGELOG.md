# Changelog

Versions follow Semantic Versioning. Dates use YYYY-MM-DD. The matching Git tag
has a `v` prefix; published tags are never moved.

## [0.12.0] - 2026-10-08

### Added
- Versioned XDG `nutshell/compilers.json` for every compiler command, including
  arbitrary user CLIs and wrappers, with stdin/file/argument prompt transports.
- `config init` and `config path`, automatic first-use initialization, and editable
  Codex, Grok and Claude Code starter entries with unattended full-permission flags.
- Legacy `interpreters.json` migration preserving user entries and the original file.
- Non-overwriting configuration initialization, including concurrent first use.

### Changed
- Removed provider-specific command dispatch and reserved names. Existing config
  is authoritative; missing names never fall back to built-in commands.
- Permission arguments can live directly in command argv. `unsafe_args` remains
  optional for older adapters. Models and private profiles stay with the CLI.
- Reinstalling Nutshell preserves user configuration; private wrappers are not
  part of the public starter template.

## [0.11.0] - 2026-10-08

### Added
- Required fine-tune conflict reviews across supplied `.nut` files, with validated
  original source paths, one-based line ranges, quotations and explanations.
- Missing, incomplete or invalid reviews block builds and publication. Conflicting
  requests preserve existing source and binary and produce visible errors.
- `-l` / `--limit` sets a positive integer AI time expectation in minutes, without
  changing the independent `--timeout` hard deadline or relaxing verification.

### Changed
- `-f` replaces `-ft`; old invocations fail with migration guidance. The long form
  remains `--fine-tune`.
- `.nut` instructions take precedence over fine-tune requests. Update requirements
  first when intentionally changing specified behavior. Semantic conflict detection
  remains the AI's judgment; the driver validates report completeness and citations.

## [0.10.0] - 2026-10-08

### Added
- `-ft` / `--fine-tune` for focused follow-up changes against existing `-s` source.
- Fine-tune prompts supply the request and available `.nut` paths, reading relevant
  requirements on demand instead of inlining the whole bundle.
- Early errors for blank requests and absent, empty or metadata-only implementation.

### Removed
- Embedded `.nut` provenance, resource generation/verification, `inspect`/`diff`,
  and prior-output snapshots and source comparisons.

### Changed
- Fine-tune requests override conflicting `.nut` text only within the requested
  scope. `.nut` files remain unchanged; update lasting requirements separately.
- Recompilation instructs agents to remove legacy provenance resources and their
  embedding hooks together; a leftover resource prevents publication.
- Native builds, tests, rgw-ast, progress and protected publication remain required.
- This checkout must be installed from source; the latest tagged release is 0.7.0.

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
