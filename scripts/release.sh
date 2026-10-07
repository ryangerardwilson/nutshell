#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
if (( $# != 0 )); then
  printf 'Usage: %s (releases VERSION from clean, published main)\n' "$0" >&2
  exit 2
fi
./scripts/check-release.sh >/dev/null
tag="v$(cat VERSION)"
if [[ "$(git branch --show-current)" != main || -n "$(git status --porcelain)" ]]; then
  printf 'Release requires a clean main branch. Commit and push reviewed changes first.\n' >&2
  exit 1
fi
remote_head="$(git ls-remote origin refs/heads/main)"
if [[ "${remote_head%%[[:space:]]*}" != "$(git rev-parse HEAD)" ]]; then
  printf 'Local HEAD must match origin/main.\n' >&2
  exit 1
fi
remote_tag="$(git ls-remote --tags origin "refs/tags/$tag")"
if git show-ref --verify --quiet "refs/tags/$tag" || [[ -n "$remote_tag" ]]; then
  printf '%s already exists. Never move a published version tag.\n' "$tag" >&2
  exit 1
fi
test -z "$(gofmt -l .)"
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
if [[ -n "$(git status --porcelain)" ]]; then
  printf 'Checkout changed during validation; refusing to tag.\n' >&2
  exit 1
fi
git tag -a "$tag" -m "Nutshell $tag"
git push origin "refs/tags/$tag"
printf 'Published %s. The release workflow runs checks and creates GitHub release notes.\n' "$tag"
