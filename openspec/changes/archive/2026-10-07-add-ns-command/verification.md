# Verification

- Strict change validation passed before implementation; accepted spec validation passed after merging the delta.
- go test -race ./... and go vet ./... passed. Installer tests cover ns creation, following an updated executable, failed build preservation and rejection of unrelated files, directories and dangling symlinks.
- Local install from this checkout succeeded. Both installed names report nutshell 0.8.0 and produce identical help, error output and exit codes; ns resolves to the installed nutshell executable.
- Release metadata and shell syntax checks passed. This feature updates main and the local development version; it does not publish a release tag. The latest tagged release remains v0.7.0.
