# nutshell Specification

## Purpose

Compile English-first .nut source into verified native executables through user-selected local AI compilers, with focused follow-up edits and inspectable implementation evidence.
## Requirements
### Requirement: English-first source programs
Nutshell SHALL accept UTF-8 .nut source with main.nut as the default entry point and SHALL allow another explicit entry. The driver SHALL supply that entry and all available .nut files recursively beneath its directory, preserving exact text and relative paths, bounded to 128 files and 1 MiB of combined source text. Hidden directories and node_modules, vendor, target, dist and build SHALL be excluded. Directory symlinks SHALL NOT be followed and .nut file symlinks SHALL be rejected. The entry itself SHALL always be included. Invalid, empty, unreadable or over-limit source SHALL fail before invoking an AI. Include, from, import and other composition syntax SHALL be interpreted by the selected AI compiler, not by the driver. Files outside the discovery boundary SHALL NOT be implicitly read; their absence SHALL be explained in the compiler prompt. The complete input bundle SHALL be retained in the temporary build workspace, without claiming every available file was semantically used. It SHALL NOT be embedded in the generated executable by Nutshell.

#### Scenario: Compiler-defined imports
- **GIVEN** main.nut says from feature1.nut import abc and feature1.nut is available beneath the entry directory
- **WHEN** the user compiles main.nut
- **THEN** both original files SHALL be supplied, with interpretation of abc and the import wording left to the AI.

#### Scenario: Alternative composition wording
- **GIVEN** an entry references another available .nut file using ordinary English or another syntax
- **WHEN** source is loaded
- **THEN** discovery SHALL supply the available file without requiring a recognized keyword or parsing its symbols.

### Requirement: User-selected AI interpreter
The command SHALL accept `nutshell main.nut -c <tool> -s <path>` with flags after or before the entry path. Compiler selection SHALL use -c or --compiler. Former -i and --interpreter flags SHALL fail with guidance to use the compiler flags before starting an AI session. Compilation SHALL require an explicit implementation source directory through -s or --source; omission SHALL fail before starting an AI session. Help and version SHALL not require compilation options. Compiler commands SHALL be resolved exclusively from the user configuration, with no reserved names or provider-specific runtime dispatch. Starter configuration SHALL provide Codex, Grok and Claude Code commands using maximum-permission unattended settings and default models without a model override. Users SHALL be able to add, replace or remove any compiler entry. Missing tools or invalid adapters SHALL fail with actionable errors.

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
Nutshell SHALL use only the directory explicitly supplied by -s as implementation context and as the updated source destination. Relative paths SHALL resolve from the caller's working directory. It SHALL NOT discover or fall back to src/ beside the entry or output. The source directory MAY be absent if its parent exists; it SHALL be created only on successful publication. The AI SHALL be instructed to inspect and adapt existing code, tests, architecture and assets, preserving unrelated work rather than rebuilding from scratch. The .nut program SHALL remain authoritative in both full compilation and fine-tune mode. Fine-tune requests SHALL NOT override conflicting .nut instructions. Existing source SHALL NOT require a generated-ownership receipt. Source context SHALL remain bounded and SHALL reject symlinks and non-regular files. Publication SHALL compare the selected source snapshot under publication locks before replacing source and binary. The executable SHALL be outside the selected source directory.

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
Nutshell SHALL require an executable rgw-ast on PATH before launching a compilation agent and SHALL fail with actionable guidance when it is unavailable. Every configured compilation agent SHALL receive instructions requiring rgw-ast for implementation inspection and mutation, even when its status reports enforcement disabled. The instructions SHALL scope rgw-ast to the temporary src directory, require status before implementation work, tracing before editing existing code, bounded reads, hash-checked mutations and explicit creation of new files. Nutshell SHALL create the temporary src directory before agent launch. Progress and build metadata outside src MAY use ordinary file writes. Existing global policies SHALL NOT be disabled or changed. This requirement SHALL be documented as an agent instruction contract rather than a permission sandbox or independently verified proof of agent compliance.

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

### Requirement: Focused fine tuning
Former -ft invocations SHALL fail with guidance to use -f before starting an AI. Nutshell SHALL accept -f or --fine-tune with a nonblank change request, including flags before or after the entry and equals-form values. Fine tuning SHALL require existing implementation files in the explicitly selected -s directory and SHALL reject absent, empty or metadata-only context before invoking an AI. The initial prompt SHALL prominently supply the request, instruct the agent to inspect and edit only relevant code, retain the existing language and architecture, preserve unrelated behavior, and avoid reimplementing the program or auditing unrelated implementation. The prompt SHALL supply .nut snapshot paths instead of inlining their full text and SHALL require checking all supplied .nut files for conflicts before implementation. The .nut instructions SHALL take precedence over the request, and .nut inputs SHALL remain unchanged. Fine tuning SHALL retain unattended compilation, rgw-ast instructions, activity summaries, native build/test verification, source drift checks and protected publication. It SHALL NOT promise a fixed compilation duration.

#### Scenario: Targeted follow-up
- **GIVEN** an existing implementation supplied via -s
- **WHEN** the user passes -f "replace x with y"
- **THEN** the agent SHALL receive that focused request with relevant source context available and be instructed to preserve unrelated code and behavior.

#### Scenario: Missing implementation or request
- **GIVEN** a blank fine-tune request or an absent or empty implementation directory
- **WHEN** fine tuning is requested
- **THEN** Nutshell SHALL fail before starting an AI session with guidance to provide a request or compile an implementation first.

### Requirement: Lean generated executables
Nutshell SHALL NOT generate, embed or validate a source provenance resource, stage prior binaries for inspection, or expose inspect/diff subcommands. When existing source contains the former .nutshell-provenance.bin resource, the compilation prompt SHALL instruct the AI to remove it and its Nutshell-only embedding hooks together, preserving application resources and behavior. A leftover legacy resource SHALL prevent publication with an actionable error. Ordinary native binaries without provenance SHALL be publishable after build and test verification.

#### Scenario: Rebuild a former provenance-bearing implementation
- **GIVEN** existing source contains the legacy Nutshell provenance resource and embedding code
- **WHEN** full compilation or fine tuning runs
- **THEN** the agent SHALL be instructed to remove both from the temporary source before rebuilding, and failed migration SHALL leave original source and output intact.

### Requirement: Fine-tune conflict diagnostics
Before modifying implementation, a fine-tune agent SHALL review all supplied .nut snapshots for conflicts with the request, respecting AI-defined file composition, and write a versioned fine-tune.json report. A compatible report SHALL identify every reviewed source and contain no conflicts. A conflict report SHALL identify every observed conflict across files with an exact relative source path, one-based inclusive start/end lines, the exact quoted lines and an explanation. The agent SHALL stop without generating a build when conflicts are observed, and SHALL NOT resolve them by assumptions or rewriting .nut inputs. Requests to bypass this check SHALL NOT override the guardrail.

The driver SHALL require and validate the report before executing declared build/test commands or publishing outputs. Missing, malformed, incomplete or inconsistent reports SHALL fail closed. Reported paths, line bounds and quoted text SHALL be validated against the original source snapshot, and diagnostics SHALL be rendered using original source paths, line numbers and source text rather than AI-invented quotations. Source drift SHALL invalidate diagnostics rather than presenting outdated coordinates as current. A valid conflict report SHALL produce a nonzero exit and visible diagnostics even when the agent exits nonzero; cancellation SHALL retain cancellation behavior. Conflict failures SHALL preserve existing implementation, .nut files and binary. Semantic conflict detection SHALL be documented as an AI judgment, not deterministic proof of English semantics. No review metadata SHALL be embedded in the executable.

#### Scenario: Literal greeting conflict
- **GIVEN** main.nut line 3 requires printing hello world
- **WHEN** a fine-tune requests replacing world with everyone and the AI reports that conflict
- **THEN** Nutshell SHALL reject it, cite main.nut:3 with the original instruction and reason, and leave outputs intact.

#### Scenario: Conflicts across files
- **GIVEN** the entry and an imported .nut file both constrain the requested change
- **WHEN** the AI observes conflicts in both
- **THEN** diagnostics SHALL identify each file and its exact line or line range.

#### Scenario: Invalid or absent review
- **GIVEN** an agent supplies a build but omits the review, claims compatibility while listing conflicts, omits a reviewed source, or cites a nonexistent path, wrong quote or out-of-range line
- **WHEN** the driver checks the report
- **THEN** compilation SHALL fail before build/test execution or publication.

#### Scenario: Compatible change
- **GIVEN** the request does not conflict with the .nut instructions and a valid compatible review is supplied
- **WHEN** the agent implements the focused change
- **THEN** normal native build/test and publication checks SHALL run.

### Requirement: XDG compiler configuration
Nutshell SHALL use $XDG_CONFIG_HOME/nutshell/compilers.json when XDG_CONFIG_HOME is absolute, otherwise ~/.config/nutshell/compilers.json. The version-1 JSON object SHALL contain a compilers map of arbitrary names to command argv arrays and prompt transports: stdin, file using {prompt_file}, or argument using {prompt}. Optional unsafe_args and {unsafe_args} SHALL remain supported for existing adapter configuration, but no mandatory provider-specific permission flag SHALL be inferred. Commands SHALL run directly without implicit shell, tilde or environment expansion. Executables SHALL resolve from PATH or an absolute path, before invoking an AI. Invalid configuration and unknown names SHALL fail with the config path and actionable guidance; removed starter names SHALL NOT fall back to hidden defaults.

The first compilation or config init SHALL create a missing configuration using an editable starter template with codex, grok and claude, without overwriting an existing file. Existing interpreters.json entries SHALL be migrated into that new file with user entries taking precedence over starter entries, leaving the legacy file intact. Invalid legacy data SHALL fail without creating a replacement. Existing compilers.json SHALL always take precedence. Creation SHALL publish a complete file without overwriting a concurrently created configuration. Config path SHALL print the effective path without creating configuration, and config init SHALL require no provider, source file or compiler credentials. Help and version SHALL NOT initialize config. Reinstallation SHALL preserve user configuration. User-specific wrappers such as GLM SHALL be configured locally, not shipped as built-in commands or starter entries.

#### Scenario: User-selected wrapper
- **GIVEN** a custom entry points to an installed wrapper CLI
- **WHEN** the user selects that name with -c
- **THEN** Nutshell SHALL pass configured arguments and the compilation prompt to the wrapper without provider-specific handling.

#### Scenario: Override or remove a starter compiler
- **GIVEN** the user replaces or removes codex in compilers.json
- **WHEN** codex is selected
- **THEN** the configured replacement SHALL run, or the missing name SHALL fail without a built-in fallback.

#### Scenario: Configuration initialization and migration
- **GIVEN** compilers.json is absent and valid legacy interpreter entries exist
- **WHEN** config init or compilation first initializes configuration
- **THEN** starter entries and user entries SHALL be written to compilers.json without changing the legacy file, with user entries winning name collisions.

#### Scenario: Malformed configuration
- **GIVEN** compilers.json is malformed or a selected adapter has an invalid command or transport
- **WHEN** a compiler is selected
- **THEN** Nutshell SHALL report an actionable error and SHALL NOT invoke an AI or silently reset configuration.

#### Scenario: Alternate XDG root
- **GIVEN** XDG_CONFIG_HOME names an absolute directory
- **WHEN** configuration is initialized or read
- **THEN** Nutshell SHALL use that root; an unset, empty or relative value SHALL use ~/.config instead.

### Requirement: Compilation without time-pressure prompts
Nutshell SHALL NOT offer a prompt-level AI time budget or inject user time expectations into compilation prompts. Former -l and --limit flags, including equals forms and invocations before or after the entry, SHALL fail before invoking an AI with guidance to remove the flag. The independent --timeout hard cutoff and its default SHALL remain unchanged. Full compilation and fine tuning SHALL continue to use fresh private temporary build workspaces with explicitly selected -s source context; no persistent AI session or retained build recipe SHALL be added by this change. Fine tuning SHALL retain its focused scope and authoritative .nut conflict checks.

#### Scenario: Removed time expectation
- **WHEN** a user supplies -l 2 or --limit=2
- **THEN** parsing SHALL fail with guidance to remove the option without invoking an AI.

#### Scenario: Focused fresh build
- **WHEN** a user compiles with -f and existing -s source
- **THEN** the AI SHALL receive the focused change and conflict-review instructions in a fresh temporary build workspace without an injected time expectation.

#### Scenario: Explicit hard cutoff
- **WHEN** a user supplies --timeout 10m without removed options
- **THEN** the process deadline SHALL be ten minutes and the AI prompt SHALL carry no time-pressure instruction.
