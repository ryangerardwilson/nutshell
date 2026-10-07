## MODIFIED Requirements

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

## ADDED Requirements

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

### Requirement: User-stated AI time budget
Nutshell SHALL accept -l or --limit with a positive integer number of minutes, before or after the entry and in equals form, for full compilation or fine tuning. Missing, zero, negative, fractional, nonnumeric or overflowing values SHALL fail before invoking an AI. When specified, the initial prompt SHALL tell the AI that the user expects completion in no longer than that many minutes and instruct it to keep work focused and avoid unnecessary complexity while retaining conflict review, correctness, tests and verification. Omitting the flag SHALL omit this expectation. The limit SHALL be a prompt-level expectation, not a guaranteed duration or process deadline; --timeout SHALL retain its independent hard cutoff and default. Help and documentation SHALL distinguish these controls.

#### Scenario: Five-minute expectation
- **WHEN** the user passes -f "fix spacing" -l 5
- **THEN** the AI prompt SHALL state an expectation of no longer than 5 minutes while retaining authoritative .nut conflict checks and build/test requirements.

#### Scenario: Independent hard timeout
- **WHEN** the user supplies -l 5 --timeout 10m
- **THEN** the prompt SHALL carry a five-minute expectation and the process deadline SHALL be ten minutes.

#### Scenario: Invalid budget
- **WHEN** -l has no value or a value such as 0, -1, 1.5 or five
- **THEN** parsing SHALL fail with guidance to supply a positive integer number of minutes without invoking an AI.
