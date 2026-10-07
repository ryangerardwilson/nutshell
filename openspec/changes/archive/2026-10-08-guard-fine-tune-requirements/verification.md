# Verification

- Strict OpenSpec validation passed before implementation, including the added -f and -l requirements.
- GOWORK=off go test -race ./... and go vet ./... passed; release metadata, formatting, whitespace and documentation links checked.
- Tests cover -f/--fine-tune and -l/--limit ordering and equals forms, invalid limits and old-flag guidance, unchanged hard-timeout behavior, prompt expectations, compatible native builds, multi-file diagnostics, CRLF line ranges, source drift, rejected missing/invalid/symlink reports, nonzero provider exits, blocked build execution and preserved outputs.
- A live Grok invocation with -f "replace world with everyone" -l 1 rejected the request in 56.3 seconds. It cited main.nut:3 and features/greeting.nut:2 and :3 with exact original lines. All .nut inputs, implementation files and the binary were byte-for-byte unchanged. The report was produced in the existing AI session; no build manifest was needed.
- Installed both local commands from this checkout. nutshell and ns report 0.11.0. Installed help, old -ft rejection and invalid minute limits checked.
- No release tag was created. Semantic conflict detection remains AI judgment; the driver validates report structure, source coverage and citations. The time limit is a prompt expectation, not a guaranteed duration.
