## Why
Prompt-level time expectations do not control compilation duration and complicate a focused fine-tune workflow. Each run already builds in a fresh temporary workspace; no retained AI session or additional build-context cache is needed.

## What Changes
- Remove `-l` / `--limit`, its option field, and AI time-pressure prompt text.
- Reject removed flags before invoking an AI with guidance to omit them.
- Keep `-f`, authoritative .nut conflict review, existing -s implementation context, fresh temporary builds, and the independent --timeout process cutoff.
- Update usage, documentation, tests and pre-1.0 minor version metadata.

## Impact
Breaking CLI change: remove `-l <minutes>` or `--limit <minutes>` from existing commands. No compiler configuration or generated binary format changes.
