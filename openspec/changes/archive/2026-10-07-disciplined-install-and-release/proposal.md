# Disciplined installation and releases

## Why
Public users need a single install command. Maintainers and local agents need a clear distinction between installing their checkout and publishing an immutable release, with one authoritative version.

## What Changes
- Support the standard Go install command using published semantic-version tags.
- Embed VERSION as the CLI version; maintain a dated CHANGELOG and validate release consistency in CI.
- Support install.sh from <checkout> for atomic local source installation, retaining no-argument checkout installation.
- Add a guarded release command and tag workflow that validate, test and publish release notes.
- Document local reinstall after feature/fix work and explicit release publication.

## Impact
Capability: nutshell. This is version 0.7.0, a backward-compatible installation/tooling addition. Runtime compiler behavior is unchanged. Rollback installs a prior tag or checkout. No skills, providers or rgw-ast are installed by these commands.
