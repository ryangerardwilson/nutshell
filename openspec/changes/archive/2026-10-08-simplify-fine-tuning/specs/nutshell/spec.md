## MODIFIED Requirements

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

### Requirement: Incremental implementation context
Nutshell SHALL use only the directory explicitly supplied by -s as implementation context and as the updated source destination. Relative paths SHALL resolve from the caller's working directory. It SHALL NOT discover or fall back to src/ beside the entry or output. The source directory MAY be absent if its parent exists; it SHALL be created only on successful publication. The AI SHALL be instructed to inspect and adapt existing code, tests, architecture and assets, preserving unrelated work rather than rebuilding from scratch. The .nut program SHALL remain authoritative for full compilation. In fine-tune mode, the explicit requested change SHALL take precedence only within its stated scope. Existing source SHALL NOT require a generated-ownership receipt. Source context SHALL remain bounded and SHALL reject symlinks and non-regular files. Publication SHALL compare the selected source snapshot under publication locks before replacing source and binary. The executable SHALL be outside the selected source directory.

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

## ADDED Requirements

### Requirement: Focused fine tuning
Nutshell SHALL accept -ft or --fine-tune with a nonblank change request, including flags before or after the entry and equals-form values. Fine tuning SHALL require existing implementation files in the explicitly selected -s directory and SHALL reject absent, empty or metadata-only context before invoking an AI. The initial prompt SHALL prominently supply the request, instruct the agent to inspect and edit only relevant code, retain the existing language and architecture, preserve unrelated behavior, and avoid reimplementing the program or auditing all requirements. The prompt SHALL supply .nut paths with snapshots available on demand instead of inlining their full text. The request SHALL override conflicting .nut requirements only for the requested change, and .nut inputs SHALL remain unchanged. Fine tuning SHALL retain unattended compilation, rgw-ast instructions, activity summaries, native build/test verification, source drift checks and protected publication. It SHALL NOT promise a fixed compilation duration.

#### Scenario: Targeted follow-up
- **GIVEN** an existing implementation supplied via -s
- **WHEN** the user passes -ft "replace x with y"
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

## REMOVED Requirements

### Requirement: Embedded program provenance
**Reason**: Remove binary provenance overhead and complexity.
**Migration**: Keep .nut files separately; use -s and -ft for follow-up edits. Existing embedding hooks and resource are removed together during recompilation.

### Requirement: Static provenance inspection and diff
**Reason**: Remove binary provenance overhead and complexity.
**Migration**: Keep .nut files separately; use -s and -ft for follow-up edits. Existing embedding hooks and resource are removed together during recompilation.

### Requirement: Agent access to prior program provenance
**Reason**: Remove binary provenance overhead and complexity.
**Migration**: Keep .nut files separately; use -s and -ft for follow-up edits. Existing embedding hooks and resource are removed together during recompilation.
