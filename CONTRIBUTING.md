# Contributing

Bring a concrete program, a reproducible failure, or a change that makes the
compiler easier to trust. A tiny `.nut` file that exposes a bad assumption is
more useful than a promise to make the model smarter.

## Work locally

You need Go 1.26 or newer. The driver has no third-party Go dependencies.

```sh
git clone https://github.com/ryangerardwilson/nutshell.git
cd nutshell
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go build -o /tmp/nutshell-dev .
/tmp/nutshell-dev --help
```

Tests use fake compilation agents and real native build/test commands. They do
not require provider credentials, paid AI calls, or an installed `rgw-ast`.
Actual `.nut` compilation requires both a configured provider and `rgw-ast`.
See the current dependency distribution limitation in the [README](README.md#install).

Do not add live model calls to the default test suite. They introduce cost,
latency and nondeterminism into tests meant to verify driver behavior.

## Change behavior deliberately

Read [AGENTS.md](AGENTS.md) and the [accepted specification](openspec/specs/nutshell/spec.md).
OpenSpec is developer tooling, separate from the compiled Go application.
Use an existing OpenSpec CLI or install the dependency declared in `package.json`:

```sh
npm install
npm run specs:list
npm run specs:changes
npm run specs:validate
```

For a behavior change, create a named change under `openspec/changes/` with a
proposal, requirement deltas and tasks. Validate it before implementing:

```sh
npm exec -- openspec validate <change-name> --strict --no-interactive
```

After implementation and verification, update the accepted spec and archive the
change. Conformance fixes, documentation corrections, formatting, and tests of
existing behavior do not require a new proposal.

## Install and version your changes

After feature or fix work passes tests and vet, run `./scripts/check-release.sh`,
then `./install.sh from "$PWD"`. Verify the version at the exact installation path.
This updates your local command from your checkout without publishing anything.
Read [the release guide](docs/releases.md) for VERSION, changelog, semantic version
rules, public one-command installation and the separate release workflow.

## Submit a useful pull request

Describe the observable problem and the resulting behavior. Include a small
before/after example when it helps, and name the checks you ran. Add regression
tests for behavior that could fail silently. Keep changes focused enough to review.

AI-assisted contributions are welcome. You are responsible for the submitted
code, its provenance, and its tests. Do not submit provider credentials, private
source, generated transcripts, temporary build directories, or compiled binaries.

For runtime reports, include the Nutshell version, host OS and architecture,
provider CLI/version, a minimal `.nut` program, and sanitized relevant errors.
Temporary logs may contain sensitive data; review them before sharing.

Useful work includes better examples, stronger publication guarantees, clearer
compiler contracts, provider compatibility, and a public installation path for
the required `rgw-ast` dependency. Propose new behavior before building it.

Contributions are made under the project's [MIT license](LICENSE).
