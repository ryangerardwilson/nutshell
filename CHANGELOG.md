# Changelog

Versions follow Semantic Versioning. Dates use YYYY-MM-DD. The matching Git tag
has a `v` prefix; published tags are never moved.

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
