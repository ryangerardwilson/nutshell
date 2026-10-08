## REMOVED Requirements

### Requirement: User-stated AI time budget
**Reason**: A prompt-level time expectation does not control execution duration and is no longer part of the product.
**Migration**: Omit -l / --limit and its value. Keep -f for focused changes. The independent --timeout process cutoff remains available.

## ADDED Requirements

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
