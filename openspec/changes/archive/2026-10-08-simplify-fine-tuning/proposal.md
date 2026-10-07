# Focused follow-up compilation

## Why
Minor changes currently pay for a full requirements review and binary provenance machinery. Users want to state a small edit directly against their existing implementation.

## What Changes
- Add -ft / --fine-tune for focused changes against required existing -s source.
- Supply requirements on demand in this mode; explicit changes override conflicting requirements within their scope.
- Remove embedded .nut provenance, inspect/diff, and prior-output inspection. Keep multi-file discovery and protected publication.
- Migrate legacy embedding through the agent and require its resource to be removed before publication.

## Impact
Pre-1.0 breaking change, version 0.10.0. Users keep .nut files separately; binaries no longer carry them. inspect/diff are removed. Full compilation remains .nut-driven. Fine tuning does not rewrite .nut files. Install locally after verification; do not publish a release tag.
