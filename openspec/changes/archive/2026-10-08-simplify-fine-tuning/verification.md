# Verification

- Strict change validation passed before implementation; accepted specs validate with --all --strict --no-interactive after archival.
- GOWORK=off go test -race ./... and go vet ./... passed.
- Release metadata, formatting, diff whitespace and documentation links checked.
- Deterministic tests cover flag order/aliases/equals forms, blank requests, missing/empty/metadata-only source, focused versus full prompts, preserved user assets and .nut inputs, native fine-tune build/test, legacy embedding removal, failed generation and blocked legacy publication.
- Live Grok smoke used an existing 0.9 Go greeting implementation with embedded provenance. A targeted Hello-to-Hi change completed in 134.8 seconds including native verification. No-argument and named greetings passed; excess arguments retained usage output and exit 2. Original .nut inputs were unchanged; legacy resource and binary marker were absent. This is one small-app observation, not a general latency benchmark.
- install.sh from the checkout installed nutshell 0.10.0 and the ns symlink in ~/.local/bin. Both exact installed commands report 0.10.0; installed help and blank-request rejection checked.
- No release tag was created; the latest tagged release remains v0.7.0.
