# Embedded source provenance and compiler-defined composition

## Why
An executable should retain the English source that governed its implementation. Users and compilation agents need to inspect that source and compare it with current files without running the executable. File composition wording belongs to the AI compiler, not a hard-coded import grammar.

## What Changes
- Bundle the entry and available .nut files recursively beneath its directory, preserving paths and exact text. Do not parse Include, from, import or symbol syntax.
- Prepare a framed, checksummed provenance resource for the agent to embed using its chosen native toolchain. Fill final compiler assumptions before the driver build and verify the resource in the resulting executable before publication.
- Add static inspect and diff commands with human-readable and JSON output.
- Stage a previous output for inspection and explain these options in the initial agent prompt.
- Advance the checkout/local version to 0.9.0, with tests, docs and local installation. No release tag is included in this change.

## Impact
Capability: nutshell. Existing Include wording is still supplied verbatim; the AI now decides its semantics. The discovery boundary is the entry directory, excluding hidden directories and common dependency/build directories; symlinks are never followed. Old binaries remain usable but have no recoverable provenance until recompiled. This is source provenance, not publisher authentication or a proof of behavior. Required source resource embedding can cause a previously incomplete agent implementation to fail verification. Runtime arguments of generated programs do not change.
