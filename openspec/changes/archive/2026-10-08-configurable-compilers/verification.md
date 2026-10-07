# Verification

- Strict OpenSpec validation passed before implementation.
- GOWORK=off go test -race ./..., go vet ./..., release metadata, whitespace and documentation link checks passed.
- Tests cover XDG fallback and alternate roots, default initialization, concurrent creation, user override/removal with no fallback, generic wrappers, all prompt transports, legacy migration and preservation, malformed configs, config commands without providers, plus existing native compilation and guardrails.
- Local compilers.json contains codex, grok, claude (Claude Code) and the user's installed glm wrapper. No GLM entry is shipped in the starter template. All four configured command lines accepted --help using the installed CLIs.
- A live Nutshell -c glm fine-tune reached the wrapper and completed conflict review in 36.7 seconds, with exact diagnostics across two .nut files and byte-for-byte preservation of the test project's sources and executable. This checks the actual configured CLI path; it is not a benchmark or a claim that all possible generated programs work.
- Installed nutshell and ns from this checkout; both report 0.12.0. Reinstallation and config init left the local four-entry configuration byte-for-byte unchanged.
- No provider, credential, skill bundle or release tag was installed or published. Private configuration remains outside the public repository.
