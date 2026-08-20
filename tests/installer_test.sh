#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Sourcing exposes installer helpers without performing an installation.
source "$ROOT_DIR/installer.sh"

fixture="$TEMP_DIR/release-fixture.json"
parsed="$TEMP_DIR/release-fixture.txt"
printf '%s\n' '{"tag_name":"1.4.0","assets":[{"name":"maniplacer-linux-amd64","digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","browser_download_url":"https://github.com/example/maniplacer"}]}' > "$fixture"
parse_release_metadata "$fixture" "$parsed"

IFS= read -r tag < "$parsed"
[[ "$tag" == "1.4.0" ]]
IFS=$'\034' read -r name digest download_url < <(sed -n '2p' "$parsed")
[[ "$name" == "maniplacer-linux-amd64" ]]
[[ "$digest" == "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" ]]
[[ "$download_url" == "https://github.com/example/maniplacer" ]]

payload="$TEMP_DIR/payload"
printf 'maniplacer\n' > "$payload"
[[ "$(calculate_sha256 "$payload")" == "b5eaca8fb661151deec0e20b170e289eab3ea2d93cfdd4a07fba7f5c90c597e8" ]]

HOME="$TEMP_DIR/home"
INSTALL_DIR="$HOME/bin with spaces"
SHELL=/bin/zsh
PATH=/usr/bin:/bin
mkdir -p "$HOME"
update_path
file_has_line "$HOME/.zshrc" "# Added by maniplacer installer: $INSTALL_DIR"
