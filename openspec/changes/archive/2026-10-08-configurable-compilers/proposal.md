# User-owned compiler commands

## Why
Hard-coded provider commands cannot represent user wrappers or changes to installed CLIs. Users should own compiler definitions in XDG configuration.

## What Changes
- Resolve every compiler from versioned compilers.json, with editable starter defaults and no reserved names.
- Add config init/path, first-use initialization and preservation of existing configuration.
- Migrate legacy interpreter entries without deleting their source file.
- Configure this user's GLM wrapper locally, keeping it out of the public product defaults.

## Impact
Version 0.12.0. Users can replace all command arguments and prompt transports. Existing configuration wins; installers do not rewrite it. No provider, skill bundle or credential installation occurs. Existing guardrails, progress, source reuse and verification stay intact.
