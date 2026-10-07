## MODIFIED Requirements

### Requirement: Local command installation
The compiler SHALL be written in Go, installed as nutshell and provide help and version without starting an AI session. Public users SHALL be able to install a tagged version with go install github.com/ryangerardwilson/nutshell@<version>, or select the latest tagged version with @latest. Local installation SHALL build the selected checkout via install.sh from <path>, defaulting to the script checkout when no arguments are given. It SHALL install atomically into NUTSHELL_INSTALL_DIR or ~/.local/bin and preserve an existing binary when a build fails. Invalid installer arguments SHALL fail before building. Installation SHALL NOT retrieve Codex skills, plugin bundles, AI providers or rgw-ast. Documentation SHALL explain prerequisites, permissions and compiler verification limits.

#### Scenario: Install updated local source
- **GIVEN** a local checkout and a Go toolchain
- **WHEN** the installer is invoked with from and that checkout's path
- **THEN** it SHALL build that source, replace the local binary only on success and report its version without invoking a provider.

#### Scenario: Public version install
- **GIVEN** a published semantic-version tag and a Go toolchain
- **WHEN** a user installs that module version
- **THEN** the binary SHALL report the version associated with the tag.

## ADDED Requirements

### Requirement: Disciplined versioned releases
VERSION SHALL be the authoritative semantic version embedded in the binary. Each release SHALL have a dated CHANGELOG entry and an immutable matching v-prefixed Git tag. CI SHALL validate version and changelog consistency; release CI SHALL also match the tag. The release command SHALL require a clean main branch matching origin/main, successful tests and checks, and an unused version tag before tagging and pushing. Public release notes SHALL derive from that version's changelog entry. Documentation SHALL distinguish patch fixes, minor additions and pre-1.0 breaking changes, and post-1.0 major breaking changes. Local agents SHALL verify and reinstall completed feature/fix changes from source; local installation SHALL NOT itself publish a release.

#### Scenario: Mismatched release tag
- **GIVEN** a tag differs from VERSION
- **WHEN** release validation runs
- **THEN** validation SHALL fail and no GitHub release SHALL be published.

#### Scenario: Local update without release
- **GIVEN** an agent has completed a verified feature or fix
- **WHEN** it installs that checkout locally
- **THEN** the installed command SHALL reflect the checkout without creating or moving any release tag.
