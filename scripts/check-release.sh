#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
if (( $# > 1 )); then
  printf 'Usage: %s [vX.Y.Z]\n' "$0" >&2
  exit 2
fi
version="$(cat VERSION)"
if [[ ! "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  printf 'VERSION must contain a stable X.Y.Z semantic version.\n' >&2
  exit 1
fi
if (( $# == 1 )) && [[ "$1" != "v$version" ]]; then
  printf 'Tag %s does not match VERSION v%s.\n' "$1" "$version" >&2
  exit 1
fi
header="$(grep -F "## [$version] - " CHANGELOG.md || true)"
if [[ ! "$header" =~ ^##\ \[$version\]\ -\ [0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
  printf 'CHANGELOG.md needs one dated entry for %s.\n' "$version" >&2
  exit 1
fi
notes="$(awk -v heading="$header" '
  $0 == heading { active=1; next }
  active && /^## / { exit }
  active { print }
' CHANGELOG.md)"
if [[ -z "${notes//[[:space:]]/}" ]]; then
  printf 'Release notes must not be empty.\n' >&2
  exit 1
fi
printf '%s\n' "$notes"
