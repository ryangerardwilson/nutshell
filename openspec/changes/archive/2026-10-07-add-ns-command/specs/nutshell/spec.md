## ADDED Requirements

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
