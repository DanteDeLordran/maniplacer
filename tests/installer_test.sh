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
[[ "$(calculate_sha256 "$payload")" == "64882ed52843f3b78d62e2aa82d2be4dedc70c4a9f65c74d915e68b91086de35" ]]

HOME="$TEMP_DIR/home"
INSTALL_DIR="$HOME/bin with spaces"
SHELL=/bin/zsh
PATH=/usr/bin:/bin
mkdir -p "$HOME"
update_path
file_has_line "$HOME/.zshrc" "# Added by maniplacer installer: $INSTALL_DIR"
file_has_line "$HOME/.zshrc" "export PATH=$(shell_quote "$INSTALL_DIR"):\"\$PATH\""
