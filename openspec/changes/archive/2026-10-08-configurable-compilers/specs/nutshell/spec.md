## MODIFIED Requirements

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

## ADDED Requirements

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
