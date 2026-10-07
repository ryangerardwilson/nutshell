# Verification

- Strict proposal and accepted spec validation passed.
- Race-enabled Go tests and go vet passed. Targeted publication tests additionally verified provenance on the staged artifact and source rollback when that verification fails.
- Coverage includes syntax-independent discovery, nested files, limits and symlink boundaries; exact source recovery; canonical hashing; added/removed/modified files and changed entry points; absent, corrupted, oversized and conflicting records; inspect/diff exit codes; and inspection without executing the target.
- Real native Go fixture builds embedded the resource, remained executable, published source metadata and supplied previous provenance/diff on incremental compilation.
- A live Grok compilation in /tmp/nutshell-provenance-live-osa19knq produced a working native greeting program with three exact .nut files and seven assumptions embedded. Default/named greetings and invalid arguments behaved as specified. Changing only greetings.nut produced one source diff, and a detached binary copy retained the same provenance.
- Local installation from this checkout succeeded. Both nutshell and ns report 0.9.0; installed ns inspected the live artifact and confirmed its source bundle matches.
- Release metadata, shell syntax, local documentation links and git diff checks passed. Main publication/CI are verified as the final task step; no release tag is part of this feature change.
