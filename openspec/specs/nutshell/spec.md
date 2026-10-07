# nutshell Specification

## Purpose

Compile English-first .nut source into verified native executables through user-selected local AI interpreters, preserving source provenance and inspectable implementation evidence.

## Requirements

### Requirement: English-first source programs
Nutshell SHALL accept UTF-8 .nut source with main.nut as the default entry point and SHALL allow another explicit entry. The driver SHALL supply that entry and all available .nut files recursively beneath its directory, preserving exact text and relative paths, bounded to 128 files and 1 MiB of combined source text. Hidden directories and node_modules, vendor, target, dist and build SHALL be excluded. Directory symlinks SHALL NOT be followed and .nut file symlinks SHALL be rejected. The entry itself SHALL always be included. Invalid, empty, unreadable or over-limit source SHALL fail before invoking an AI. Include, from, import and other composition syntax SHALL be interpreted by the selected AI compiler, not by the driver. Files outside the discovery boundary SHALL NOT be implicitly read; their absence SHALL be explained in the compiler prompt. The complete input bundle SHALL be retained as provenance, without claiming every available file was semantically used.

#### Scenario: Compiler-defined imports
- **GIVEN** main.nut says from feature1.nut import abc and feature1.nut is available beneath the entry directory
- **WHEN** the user compiles main.nut
- **THEN** both original files SHALL be supplied, with interpretation of abc and the import wording left to the AI.

#### Scenario: Alternative composition wording
- **GIVEN** an entry references another available .nut file using ordinary English or another syntax
- **WHEN** source is loaded
- **THEN** discovery SHALL supply the available file without requiring a recognized keyword or parsing its symbols.

### Requirement: User-selected AI interpreter
The command SHALL accept `nutshell main.nut -c <tool> -s <path>` with flags after or before the entry path. Compiler selection SHALL use -c or --compiler. Former -i and --interpreter flags SHALL fail with guidance to use the compiler flags before starting an AI session. Compilation SHALL require an explicit implementation source directory through -s or --source; omission SHALL fail before starting an AI session. Help and version SHALL not require compilation options. Built-in Codex, Grok and Claude adapters SHALL use installed CLIs, their documented maximum-permission unattended settings by default, and their default models without an explicit model override. Other tools SHALL be supported through an explicit local command adapter. Missing tools or invalid adapters SHALL fail with actionable errors.

#### Scenario: Compile with the user's choice
- **GIVEN** an installed and authenticated Grok or Codex CLI
- **WHEN** the user selects it with -c or --compiler and supplies -s
- **THEN** Nutshell SHALL invoke that tool with the source program, configured unattended full-permission arguments and no model-selection argument.

#### Scenario: Migrate an old invocation
- **GIVEN** the user supplies -i or --interpreter
- **WHEN** the command is parsed
- **THEN** it SHALL fail with guidance to use -c or --compiler without starting an AI session.

#### Scenario: Missing source selection
- **GIVEN** a compilation invocation omits -s
- **WHEN** the command is parsed
- **THEN** it SHALL fail with guidance to supply a source directory without starting an AI session.

### Requirement: Structured generation contract
The interpreter SHALL be instructed to preserve .nut source, resolve ambiguity and conflicting requirements using its own assumptions, record those assumptions internally, and produce a versioned build manifest with native artifact, build and test commands. It SHALL NOT ask the user questions or stop solely for clarification. Missing or invalid manifests and failed implementation commands SHALL still fail compilation.

#### Scenario: Underspecified behavior
- **GIVEN** the source leaves behavior unspecified or contradictory
- **WHEN** the interpreter implements it
- **THEN** it SHALL choose a coherent interpretation, record assumptions and continue without user input.

### Requirement: Verified native output
Nutshell SHALL run the declared build and test commands, require their success, and validate that the declared artifact is a regular native executable for the host. It SHALL publish the output atomically, defaulting to the entry path without .nut and supporting -o for another output. It SHALL reject source-path overwrites and preserve an earlier output when generation, verification or publication fails.

#### Scenario: Failed replacement build
- **GIVEN** a previously compiled executable exists
- **WHEN** the new interpreter output fails tests or is a script instead of a native executable
- **THEN** Nutshell SHALL fail and retain the existing executable.

### Requirement: Inspectable isolated build attempts
Each invocation SHALL perform generation, building and testing in a distinct private directory under /tmp, containing source snapshots, prompt, logs, generated implementation and available execution evidence. It SHALL NOT create build metadata in the source directory. The finished executable SHALL be published to its output path and implementation source SHALL be published to the directory explicitly selected with -s. Existing selected source SHALL be copied as implementation context, including unmanaged and user-edited files, rather than rejected on ownership grounds. Temporary evidence MAY remain under /tmp; successful output SHALL NOT require the user to inspect it. Changes to .nut inputs or the selected source directory during compilation SHALL prevent publication. Existing .nut input files inside the published source directory SHALL remain unchanged. Cancellation SHALL stop running process groups on supported Unix systems. These work directories SHALL NOT be described as permission sandboxes.

#### Scenario: Separate source and binary locations
- **GIVEN** the user specifies -s implementation and -o bin/program
- **WHEN** compilation succeeds
- **THEN** updated code SHALL be delivered to implementation and the executable to bin/program, with scratch work remaining under /tmp.

#### Scenario: Failed incremental build
- **GIVEN** an existing binary and selected source directory
- **WHEN** generation, verification or publication fails
- **THEN** the previous binary and source SHALL remain intact.

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

### Requirement: Quiet progress reporting
The compilation agent SHALL be instructed to send structured stage updates with specific 3–7-word descriptions of current work, beginning with its first tool action and continuing before meaningful actions within each stage. Nutshell SHALL display changing summaries in a compact terminal progress bar with elapsed time, and as concise lines when redirected. Valid summaries SHALL NOT be truncated at a fixed 35-character boundary. Raw AI and build transcripts SHALL remain in temporary logs. Invalid, missing or duplicate updates SHALL NOT fail compilation or flood the display. Later activities in earlier stages MAY update the summary but SHALL NOT regress milestone progress. Progress SHALL represent milestones rather than a time estimate, and completion SHALL only be shown after verification and publication. Compilation SHALL have no interactive question-and-answer phase.

#### Scenario: Multiple tasks in one stage
- **GIVEN** the agent is implementing a program
- **WHEN** it reports Writing command line argument parser and then Adding empty input validation tests in the same stage
- **THEN** Nutshell SHALL display each summary as it arrives in terminal and redirected output.

#### Scenario: Iterative implementation
- **GIVEN** the agent has begun checking its work
- **WHEN** it returns to implementation and reports a valid description of a fix
- **THEN** the summary SHALL change without moving the progress bar backward.

#### Scenario: Provider remains quiet
- **GIVEN** no accepted activity update has arrived for 30 seconds
- **WHEN** the agent is still running
- **THEN** the interactive terminal SHALL show Waiting for AI progress update with elapsed time instead of inventing tasks or completion, and resume displaying activity on the next valid update.

### Requirement: Incremental implementation context
Nutshell SHALL use only the directory explicitly supplied by -s as implementation context and as the updated source destination. Relative paths SHALL resolve from the caller's working directory. It SHALL NOT discover or fall back to src/ beside the entry or output. The source directory MAY be absent if its parent exists; it SHALL be created only on successful publication. The AI SHALL be instructed to inspect and adapt existing code, tests, architecture and assets, preserving unrelated work rather than rebuilding from scratch. The .nut program SHALL remain authoritative for requested behavior. Existing source SHALL NOT require a generated-ownership receipt. Source context SHALL remain bounded and SHALL reject symlinks and non-regular files. Publication SHALL compare the selected source snapshot under publication locks before replacing source and binary. The executable SHALL be outside the selected source directory.

#### Scenario: User-selected implementation
- **GIVEN** -s selects a directory containing user-edited code and another src/ directory exists nearby
- **WHEN** Nutshell compiles the program
- **THEN** only the selected implementation SHALL be staged for the AI and the unrelated src/ SHALL remain untouched.

#### Scenario: New implementation
- **GIVEN** -s selects a nonexistent directory whose parent exists
- **WHEN** compilation succeeds
- **THEN** that directory SHALL receive the generated implementation without using any implicit source directory.

#### Scenario: Concurrent source edit
- **GIVEN** a build began from a snapshot of the selected implementation directory
- **WHEN** another process modifies, adds or removes a source file before publication
- **THEN** compilation SHALL fail without replacing the current binary or source directory.

### Requirement: AST-aware compilation workflow
Nutshell SHALL require an executable rgw-ast on PATH before launching a compilation agent and SHALL fail with actionable guidance when it is unavailable. Every built-in and custom compilation agent SHALL receive instructions requiring rgw-ast for implementation inspection and mutation, even when its status reports enforcement disabled. The instructions SHALL scope rgw-ast to the temporary src directory, require status before implementation work, tracing before editing existing code, bounded reads, hash-checked mutations and explicit creation of new files. Nutshell SHALL create the temporary src directory before agent launch. Progress and build metadata outside src MAY use ordinary file writes. Existing global policies SHALL NOT be disabled or changed. This requirement SHALL be documented as an agent instruction contract rather than a permission sandbox or independently verified proof of agent compliance.

#### Scenario: Missing dependency
- **GIVEN** rgw-ast is not executable on PATH
- **WHEN** a valid compilation is requested
- **THEN** Nutshell SHALL fail before launching the agent or publishing any outputs and explain the missing dependency.

#### Scenario: Fresh or incremental compilation
- **GIVEN** an installed compiler and rgw-ast
- **WHEN** Nutshell launches the selected agent for a new or existing implementation
- **THEN** the shared prompt SHALL require the rgw-ast workflow against the existing temporary src directory, independent of the enforcement threshold.

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

### Requirement: Short command alias
Local installation SHALL provide ns beside nutshell as a symbolic link to the same executable. Both command names SHALL accept the same arguments and produce the same behavior, including Nutshell's canonical version output. Reinstallation SHALL keep ns pointing at the installed nutshell binary. An unrelated existing ns path SHALL cause installation to fail before replacing either command. Failed builds SHALL leave existing commands intact. Help and documentation SHALL describe the alias and a one-time symlink setup for binaries installed directly with Go. Adding the alias SHALL NOT require shell configuration changes.

#### Scenario: Use the short command
- **GIVEN** Nutshell has been installed with the local installer
- **WHEN** the user runs ns with compilation arguments or --version
- **THEN** the installed Nutshell executable SHALL run with those arguments.

#### Scenario: Preserve another ns tool
- **GIVEN** the installation directory contains an unrelated ns file, directory or symlink
- **WHEN** the installer runs
- **THEN** it SHALL report the conflict and leave both existing command paths unchanged.

### Requirement: Embedded program provenance
Every successfully published generated executable SHALL contain a validated versioned provenance resource with the entry path, complete original .nut source bundle, canonical source SHA-256, Nutshell version, selected AI compiler, implementation language and recorded assumptions. The driver SHALL supply the resource for toolchain embedding and verify that the final executable contains the exact expected record before publication. It SHALL NOT rewrite signed executable bytes to append metadata. Missing, corrupted, conflicting or mismatched provenance SHALL prevent publication. Provenance SHALL be documented as a record of inputs and assumptions, not authentication or proof of executable behavior. The source resource SHALL travel with published implementation source so subsequent builds can embed it.

#### Scenario: Recover source from an executable
- **GIVEN** a successfully compiled executable and no original source directory
- **WHEN** the user inspects the executable
- **THEN** all original .nut filenames and text SHALL be recoverable without executing the program.

#### Scenario: Agent omits the source resource
- **GIVEN** a generated binary builds and passes tests but lacks the expected provenance
- **WHEN** publication verification runs
- **THEN** the build SHALL fail and prior source and binary SHALL remain intact.

### Requirement: Static provenance inspection and diff
Both command names SHALL support inspect <binary> with --source or --json, and diff <binary> [entry.nut] with optional --json. These commands SHALL require neither an AI provider nor -c or -s. They SHALL only read the target and source files, never execute the inspected binary. Inspection SHALL bound allocations and validate the provenance schema, checksum, paths and source hash. Legacy binaries without provenance SHALL produce an actionable error. Diff SHALL report changed entry points and added, removed or modified source files, returning 0 for equality, 1 for differences and 2 for errors.

#### Scenario: Imported feature changes
- **GIVEN** a binary records main.nut and feature1.nut and only feature1.nut changes
- **WHEN** the user runs diff against main.nut
- **THEN** the change in feature1.nut SHALL be reported and the command SHALL return 1.

### Requirement: Agent access to prior program provenance
The initial compilation prompt SHALL describe inspect and diff, provide callable examples, explain source discovery and delegate composition semantics to the AI. When the selected output exists, the driver SHALL stage a read-only copy under the temporary workspace and provide its available provenance and source diff or a reason provenance is unavailable. The prompt SHALL instruct the AI to use prior requirements to understand changes while treating current source as authoritative. The AI SHALL NOT be asked to execute the previous program for inspection.

#### Scenario: Incremental compilation from an existing output
- **GIVEN** the output from an earlier compilation exists
- **WHEN** the next AI session starts
- **THEN** its initial context SHALL identify the staged prior binary and the inspect/diff tools without requiring access to the original project directory.
