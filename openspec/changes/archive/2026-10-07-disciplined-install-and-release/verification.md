# Verification

- Race-enabled Go tests and go vet passed, including explicit source installation, failed build preservation, invalid arguments and release metadata rejection fixtures.
- Release metadata accepts v0.7.0 and rejects a mismatched tag. Release command rejects a dirty checkout before publishing anything.
- Local source installation succeeded and the installed CLI reports nutshell 0.7.0.
- Documentation links and shell syntax checks passed.
- Isolated local Git repositories verified rejection of a remote duplicate tag, an unpublished HEAD and a dirty checkout.
- Main CI passed (37616881335). The guarded release command published annotated tag v0.7.0; Release CI passed (37617017061) and published its changelog notes.
- A fresh `go install github.com/ryangerardwilson/nutshell@latest` through the default module proxy downloaded v0.7.0, built successfully and reported nutshell 0.7.0. The smoke binary was installed under /tmp, separate from the local source installation.
- Release: https://github.com/ryangerardwilson/nutshell/releases/tag/v0.7.0
