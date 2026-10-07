## MODIFIED Requirements

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

## ADDED Requirements

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
