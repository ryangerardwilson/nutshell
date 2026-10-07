# Keep fine tuning within the program contract

## Why
Fine tuning currently overrides explicit .nut requirements. A small edit must not silently change the human-authored program contract.

## What Changes
- Make .nut requirements authoritative in fine tuning.
- Rename -ft to -f with migration guidance and add -l / --limit for a positive integer AI time expectation in minutes; --timeout remains the independent hard cutoff.
- Require a conflict review in the existing AI session before implementation.
- Validate structured reviews and show source-backed file/line diagnostics for conflicts.
- Block missing or invalid reviews before native verification and publication.

## Impact
Pre-1.0 behavior change, version 0.11.0. A conflicting change requires editing the .nut requirements first. All discovered source files must be reviewed, which can add context work; no extra AI session or embedded binary metadata is introduced. English conflict detection remains the selected AI's semantic judgment.
