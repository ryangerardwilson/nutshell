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
install_dir="$(cd -- "$install_dir" && pwd)"
check_alias() {
  if [[ -e "$install_dir/ns" || -L "$install_dir/ns" ]]; then
    if [[ -L "$install_dir/ns" ]]; then
      alias_target="$(readlink -- "$install_dir/ns")"
      if [[ "$alias_target" == nutshell || "$alias_target" == "$install_dir/nutshell" ]]; then
        return
      fi
    fi
    printf 'Refusing to replace unrelated command %s; move it before installing Nutshell.\n' "$install_dir/ns" >&2
    exit 1
  fi
}
check_alias
build_file="$(mktemp "$install_dir/.nutshell-install-XXXXXX")"
trap 'rm -f -- "$build_file"' EXIT
GOWORK=off go build -trimpath -o "$build_file" .
chmod 755 "$build_file"
# Inspect the staged executable before replacing a working installation.
"$build_file" --version
check_alias
mv -f -- "$build_file" "$install_dir/nutshell"
if [[ ! -L "$install_dir/ns" ]]; then
  ln -s nutshell "$install_dir/ns"
fi
printf 'Installed %s and %s\n' "$install_dir/nutshell" "$install_dir/ns"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) printf 'Add %s to PATH to run nutshell.\n' "$install_dir" ;;
esac
