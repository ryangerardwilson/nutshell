# Verification

- Race-enabled Go tests and go vet passed, including explicit source installation, failed build preservation, invalid arguments and release metadata rejection fixtures.
- Release metadata accepts v0.7.0 and rejects a mismatched tag. Release command rejects a dirty checkout before publishing anything.
- Local source installation succeeded and the installed CLI reports nutshell 0.7.0.
- Documentation links and shell syntax checks passed.
- Public tag, CI and fresh public installation verification will be recorded after publication.
