# Installation, versions and releases

## Public installation

With Go 1.26 or newer:

```sh
go install github.com/ryangerardwilson/nutshell@latest
```

Go puts the binary in `GOBIN`, or `$(go env GOPATH)/bin` when GOBIN is unset.
Put that directory on `PATH`. To use Nutshell's usual local installation location:

```sh
GOBIN="$HOME/.local/bin" go install github.com/ryangerardwilson/nutshell@latest
```

Repeat the command to upgrade. Pin a release for repeatable installation:

```sh
GOBIN="$HOME/.local/bin" go install github.com/ryangerardwilson/nutshell@v0.7.0
```

The same command with an older published tag is the rollback path. Installation
builds the Go driver; compiling `.nut` programs still requires an authenticated AI
CLI and `rgw-ast`. No command here installs providers, credentials, skills or plugins.

## Local development installation

After completing a feature or fix, run the checks and install the checkout you
actually changed:

```sh
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
./scripts/check-release.sh
./install.sh from "$PWD"
"${NUTSHELL_INSTALL_DIR:-$HOME/.local/bin}/nutshell" --version
```

From another directory, use an absolute path to the script and checkout. Paths
with spaces are supported. `./install.sh` without arguments selects its own
checkout. `NUTSHELL_INSTALL_DIR` changes the destination, defaulting to
`~/.local/bin`. The installer builds to a staged file, checks it can report a
version, then replaces the installed binary. Failed builds preserve the old one.

Local installation works with uncommitted source and publishes nothing. Never
download a release to validate code you just changed locally. Read the installed
version from the exact destination path; another `nutshell` earlier on PATH may
be a different installation.

## One version source

`VERSION` contains `X.Y.Z` and is embedded directly into the Go binary. There is
no second version constant to update. Public tags are `vX.Y.Z`; the CLI prints
`nutshell X.Y.Z`. A local build reports the VERSION of its checkout, not proof that
it is identical to that published release.

- **Patch:** compatible fixes and maintenance changes.
- **Minor:** backward-compatible capabilities. Before 1.0, breaking changes also
  advance the minor version and must explain migration in the changelog.
- **Major:** breaking changes after 1.0.
- Documentation-only changes need no version bump unless released separately.

For a user-visible change, update VERSION and add a dated CHANGELOG entry before
the next release. Multiple commits for one unreleased version share that entry.
Choose a version greater than the latest release. Never reuse or move a published
tag. The current release tooling accepts stable numeric versions, not prereleases.

CI checks the version format and matching dated changelog entry. Release CI also
requires the tag to match VERSION and verifies the binary's reported version.

## Publish a release

Release publication is separate from local installation. Run it only when asked
to publish a release or as part of an explicitly authorized release task.

1. Update VERSION and CHANGELOG, including migration notes where needed.
2. Complete the OpenSpec lifecycle for behavior changes. Run tests and vet.
3. Commit the reviewed changes, push main, and confirm normal CI passes.
4. From the clean main checkout, run:

   ```sh
   ./scripts/release.sh
   ```

The script requires HEAD to match origin/main, rejects an existing version tag,
runs formatting checks, tests and vet, then creates and pushes an annotated tag.
The Release workflow revalidates metadata and code and publishes GitHub release
notes from that changelog entry. Go users install directly from the tag; no
prebuilt binary assets are promised by this release process.

Verify the workflow, release page, and a fresh version-pinned installation before
calling the release complete. Go module proxy discovery can lag a newly pushed
tag; retry after propagation or use `GOPROXY=direct` to retrieve from GitHub.

If pushing the tag fails, the local tag remains: inspect the remote before retrying
that push. If release CI fails, fix the problem and publish a new version rather
than moving the failed tag. Re-running a workflow for an already published release
does not overwrite it.
