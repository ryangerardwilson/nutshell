# Add the ns command

## Why
The user wants a shorter command without changing Nutshell's behavior.

## What Changes
- Local installation creates an ns symlink beside nutshell, targeting the same executable.
- Reinstallation preserves the alias and refuses unrelated existing ns commands before changing the installation.
- Document the optional one-time symlink for Go-installed binaries and show ns in CLI help.
- Advance the development version to 0.8.0; no public release is part of this change.

## Impact
Capability: nutshell. All existing flags, version output and compiler behavior remain unchanged. Removing the symlink removes the short command without affecting nutshell. No shell configuration or provider settings change.
