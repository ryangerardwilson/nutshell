# Working on Nutshell

Nutshell compiles English-first .nut programs into native executables through an
installed AI CLI. The driver is Go and uses the standard library. Read README.md,
then docs/architecture.md for the code paths relevant to a change. Source discovery
is syntax-independent: the AI interprets all file-composition wording. Fine tuning (-f / --fine-tune) targets a requested change against existing -s
source. All supplied .nut instructions must be reviewed first; conflicts cannot
be overridden. The driver requires a validated fine-tune.json review. Each run
uses a fresh temporary workspace; --timeout controls the hard process cutoff.
Generated executables carry no
Nutshell provenance resource. Compiler commands are user-owned in XDG
compilers.json; default-compilers.json is only a first-use template. Do not add
provider-specific runtime branches or private wrappers to the defaults.
See docs/cli.md for configuration, precedence and migration.

## OpenSpec

Accepted specs in openspec/specs/ are product and engineering truth. Active
openspec/changes/ directories are proposals until implemented and archived.
List specs and active changes before non-trivial work:

```sh
openspec list --specs
openspec list
```

For behavior changes, create a named change with proposal, spec deltas and tasks;
validate with `openspec validate <change> --strict --no-interactive` before coding,
then implement, verify, update accepted specs and archive. Bug fixes that restore
accepted behavior, typos, formatting, non-breaking config and tests of existing
behavior do not need a new proposal. Prefer these project-local contracts over
assumptions imported from another workspace. OpenSpec is a development dependency
in package.json; it is not required to run Nutshell.

## Verification

```sh
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go build -o /tmp/nutshell-dev .
/tmp/nutshell-dev --help
```

Keep tests deterministic. Do not invoke paid AI providers in tests. Use the fake
adapter fixtures for compiler integration; they build and test real Go programs.
Keep credentials, generated binaries, provider transcripts and build directories
out of Git. Do not download or install Codex skills, bundles or plugins.

## Local installation and versioning

After completing a feature or fix, run tests and vet, validate release metadata
with `./scripts/check-release.sh`, then run `./install.sh from "$PWD"` and verify
both `"${NUTSHELL_INSTALL_DIR:-$HOME/.local/bin}/nutshell" --version` and
`"${NUTSHELL_INSTALL_DIR:-$HOME/.local/bin}/ns" --version`. Install directly
from the changed checkout; do not download a release as a substitute. Report the
installed version and checks. Documentation-only changes need no reinstall.

Read docs/releases.md before changing VERSION or publishing. VERSION is embedded
in the binary; update its matching CHANGELOG entry for user-visible changes.
Published tags are immutable. Public release publication is separate from local
installation and requires an explicit release task. Do not tag or publish merely
because a local feature or fix is complete.

<!-- rgw-ast:begin -->
# rgw-ast boundary

Run `rgw-ast status --json` before the first edit in a repository. When it reports
`"enforced": true`, the `rgw-ast hook` PreToolUse hook enforces the rules below. Every
change is checked against the file's hash, so concurrent agents and stale context
cannot silently overwrite a file.

- **Trace** code before editing it: `rgw-ast callers|callees|impact <symbol>`, then
  `map` and `show`. `search` is for a literal you already know. `rgw-ast help callers`
  explains confidence levels and partial results.
- **Read** with `rgw-ast read <file> --lines A-B`, or a host Read with `limit` of at
  most `max_read_lines` (200 by default).
- **Edit** with `rgw-ast patch <file> --expect-hash <sha> --old <text> --new <text>`.
  `rgw-ast hash <file>` prints the hash; each patch prints the next one. For several
  files at once: `rgw-ast apply --from-file <manifest.json|git diff>`.
- **Create** with a host Write to a new path, or `rgw-ast create <file> --expect-absent --stdin`.
- **Move** with `rgw-ast move <src> <dst> --expect-hash <sha>`; **delete** with
  `rgw-ast delete <file> --expect-hash <sha>`.
- **Generators** listed in `generators.allow` (for example `openspec new|archive`) run as
  `rgw-ast exec -- <command>`.
- A denied tool call's reason names the exact command to run instead.

Unenforced repositories use normal tools. Policy is global
(`~/.config/rgw-ast/config.toml`). A nested AGENTS.md does not need to repeat this block.
<!-- rgw-ast:end -->
