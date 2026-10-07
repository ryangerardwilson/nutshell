#!/usr/bin/env bash
set -euo pipefail
case "$#" in
  0) source_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)" ;;
  2)
    if [[ "$1" != from || -z "$2" ]]; then
      printf 'Usage: %s [from <checkout>]\n' "$0" >&2
      exit 2
    fi
    source_dir="$2"
    ;;
  *) printf 'Usage: %s [from <checkout>]\n' "$0" >&2; exit 2 ;;
esac
cd -- "$source_dir"
if [[ ! -f go.mod || ! -f VERSION ]] || ! grep -qx 'module github.com/ryangerardwilson/nutshell' go.mod; then
  printf 'Expected a Nutshell source checkout: %s\n' "$source_dir" >&2
  exit 2
fi
install_dir="${NUTSHELL_INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p -- "$install_dir"
build_file="$(mktemp "$install_dir/.nutshell-install-XXXXXX")"
trap 'rm -f -- "$build_file"' EXIT
GOWORK=off go build -trimpath -o "$build_file" .
chmod 755 "$build_file"
# Inspect the staged executable before replacing a working installation.
"$build_file" --version
mv -f -- "$build_file" "$install_dir/nutshell"
printf 'Installed %s\n' "$install_dir/nutshell"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) printf 'Add %s to PATH to run nutshell.\n' "$install_dir" ;;
esac
