#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
install_dir="${NUTSHELL_INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p -- "$install_dir"
build_file="$(mktemp "$install_dir/.nutshell-install-XXXXXX")"
trap 'rm -f -- "$build_file"' EXIT
GOWORK=off go build -trimpath -o "$build_file" .
chmod 755 "$build_file"
mv -f -- "$build_file" "$install_dir/nutshell"
printf 'Installed %s\n' "$install_dir/nutshell"
"$install_dir/nutshell" --version
