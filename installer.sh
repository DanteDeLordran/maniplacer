#!/usr/bin/env bash
set -Eeuo pipefail

umask 077

TOOL_NAME="maniplacer"
REPO_URL="https://github.com/dantedelordran/maniplacer"
API_URL="https://api.github.com/repos/dantedelordran/maniplacer/releases/latest"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TEMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/maniplacer-install.XXXXXX")"

cleanup() {
    rm -rf -- "$TEMP_DIR"
}
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

info() {
    printf 'info: %s\n' "$1"
}

warning() {
    printf 'warning: %s\n' "$1" >&2
}

error() {
    printf 'error: %s\n' "$1" >&2
    exit 1
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

CURL_OPTIONS=(
    --fail
    --location
    --silent
    --show-error
    --connect-timeout "${MANIPLACER_CONNECT_TIMEOUT:-10}"
    --max-time "${MANIPLACER_MAX_TIME:-120}"
)

check_prerequisites() {
    command_exists curl || error "curl is required"

    local curl_help
    curl_help="$(curl --help all 2>/dev/null || true)"
    if [[ "$curl_help" == *"--proto"* ]]; then
        CURL_OPTIONS+=(--proto '=https')
    fi

    if ! command_exists jq && ! command_exists python3 && ! perl -MJSON::PP -e 1 >/dev/null 2>&1; then
        error "release metadata requires jq, python3, or Perl with JSON::PP"
    fi

    if ! command_exists sha256sum && ! command_exists shasum && ! command_exists openssl; then
        error "SHA-256 verification requires sha256sum, shasum, or openssl"
    fi
}

detect_platform() {
    local raw_os raw_arch
    raw_os="$(uname -s)"
    raw_arch="$(uname -m)"

    case "$raw_os" in
        Linux) OS="linux" ;;
        Darwin) OS="darwin" ;;
        MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
        *) error "unsupported operating system: $raw_os" ;;
    esac

    case "$raw_arch" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) error "unsupported architecture: $raw_arch" ;;
    esac

    case "$OS/$ARCH" in
        linux/amd64|linux/arm64|darwin/amd64|darwin/arm64|windows/amd64) ;;
        *) error "unsupported platform: $OS/$ARCH" ;;
    esac

    BINARY_NAME="$TOOL_NAME-$OS-$ARCH"
    INSTALLED_NAME="$TOOL_NAME"
    if [[ "$OS" == "windows" ]]; then
        BINARY_NAME="${BINARY_NAME}.exe"
        INSTALLED_NAME="${INSTALLED_NAME}.exe"
    fi

    info "detected platform $OS/$ARCH"
}

parse_release_metadata() {
    local input="$1" output="$2"

    if command_exists jq; then
        jq -er '.tag_name, (.assets[] | [.name, (.digest // ""), .browser_download_url] | join("\u001c"))' \
            "$input" > "$output"
    elif command_exists python3; then
        python3 - "$input" "$output" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as source:
    release = json.load(source)
with open(sys.argv[2], "w", encoding="utf-8") as target:
    print(release["tag_name"], file=target)
    for asset in release["assets"]:
        print(asset["name"], asset.get("digest") or "", asset["browser_download_url"], sep="\x1c", file=target)
PY
    else
        perl -MJSON::PP -0777 -e '
            my $release = decode_json(<>);
            print "$release->{tag_name}\n";
            for my $asset (@{$release->{assets}}) {
                print join("\x1c", $asset->{name}, $asset->{digest} // "", $asset->{browser_download_url}), "\n";
            }
        ' "$input" > "$output"
    fi
}

get_release() {
    local metadata="$TEMP_DIR/release.json"
    local parsed="$TEMP_DIR/release.tsv"
    local name digest url

    info "fetching latest release metadata"
    curl "${CURL_OPTIONS[@]}" --output "$metadata" "$API_URL"
    parse_release_metadata "$metadata" "$parsed" || error "invalid GitHub release metadata"

    IFS= read -r LATEST_VERSION < "$parsed"
    [[ -n "$LATEST_VERSION" ]] || error "release metadata has no tag"

    DOWNLOAD_URL=""
    EXPECTED_DIGEST=""
    CHECKSUM_URL=""
    {
        IFS= read -r _
        while IFS=$'\034' read -r name digest url; do
            if [[ "$name" == "$BINARY_NAME" ]]; then
                DOWNLOAD_URL="$url"
                EXPECTED_DIGEST="${digest#sha256:}"
            fi
            case "$name" in
                checksums.txt|sha256sums.txt|SHA256SUMS|maniplacer-checksums.txt)
                    CHECKSUM_URL="$url"
                    ;;
            esac
        done
    } < "$parsed"

    [[ -n "$DOWNLOAD_URL" ]] || error "release $LATEST_VERSION has no asset named $BINARY_NAME"
    info "latest version is $LATEST_VERSION"
}

check_existing_installation() {
    local current clean_current clean_latest reply
    if ! command_exists "$TOOL_NAME"; then
        return
    fi

    current="$("$TOOL_NAME" --version 2>/dev/null || true)"
    warning "$TOOL_NAME is already installed (version: ${current:-unknown})"
    clean_current="${current#v}"
    clean_latest="${LATEST_VERSION#v}"
    if [[ "$clean_current" != "$clean_latest" ]]; then
        return
    fi

    if [[ -t 0 ]]; then
        read -r -p "Reinstall the same version? (y/N): " reply
        [[ "$reply" =~ ^[Yy]$ ]] || exit 0
    else
        info "reinstalling in non-interactive mode"
    fi
}

calculate_sha256() {
    local file="$1"
    if command_exists sha256sum; then
        sha256sum "$file" | awk '{print $1}'
    elif command_exists shasum; then
        shasum -a 256 "$file" | awk '{print $1}'
    else
        openssl dgst -sha256 "$file" | awk '{print $NF}'
    fi
}

checksum_from_release_asset() {
    local checksum_file="$TEMP_DIR/checksums.txt"
    curl "${CURL_OPTIONS[@]}" --output "$checksum_file" "$CHECKSUM_URL"
    awk -v target="$BINARY_NAME" '
        NF >= 2 {
            name = $2
            sub(/^\*/, "", name)
            if (name == target) { print $1; exit }
        }
        $1 == "SHA256" && $2 == "(" target ")" && $3 == "=" { print $4; exit }
    ' "$checksum_file"
}

verify_download() {
    local file="$1" actual

    if [[ -n "$CHECKSUM_URL" ]]; then
        EXPECTED_DIGEST="$(checksum_from_release_asset)"
        [[ -n "$EXPECTED_DIGEST" ]] || error "checksum asset has no entry for $BINARY_NAME"
    fi

    EXPECTED_DIGEST="$(printf '%s' "$EXPECTED_DIGEST" | tr '[:upper:]' '[:lower:]')"
    [[ "$EXPECTED_DIGEST" =~ ^[0-9a-f]{64}$ ]] || error "release has no valid SHA-256 digest for $BINARY_NAME"

    actual="$(calculate_sha256 "$file")"
    actual="$(printf '%s' "$actual" | tr '[:upper:]' '[:lower:]')"
    [[ "$actual" == "$EXPECTED_DIGEST" ]] || error "SHA-256 verification failed for $BINARY_NAME"
    info "verified SHA-256 checksum"
}

download_binary() {
    DOWNLOADED_FILE="$TEMP_DIR/$BINARY_NAME"
    info "downloading $BINARY_NAME"
    curl "${CURL_OPTIONS[@]}" --output "$DOWNLOADED_FILE" "$DOWNLOAD_URL"
    [[ -s "$DOWNLOADED_FILE" ]] || error "downloaded asset is empty"
    verify_download "$DOWNLOADED_FILE"
}

install_binary() {
    [[ "$INSTALL_DIR" != *$'\n'* && "$INSTALL_DIR" != *$'\r'* ]] || error "INSTALL_DIR contains a newline"
    mkdir -p -- "$INSTALL_DIR"
    chmod 0755 "$DOWNLOADED_FILE"
    mv -f -- "$DOWNLOADED_FILE" "$INSTALL_DIR/$INSTALLED_NAME"
    info "installed $INSTALL_DIR/$INSTALLED_NAME"
}

shell_quote() {
    local value="$1"
    printf '%q' "$value"
}

file_has_line() {
    local file="$1" expected="$2" line
    [[ -f "$file" ]] || return 1
    while IFS= read -r line || [[ -n "$line" ]]; do
        [[ "$line" == "$expected" ]] && return 0
    done < "$file"
    return 1
}

update_path() {
    local marker="# Added by maniplacer installer: $INSTALL_DIR"
    local shell_name="${SHELL##*/}" rc_file="" quoted_dir

    [[ "${MANIPLACER_NO_MODIFY_PATH:-0}" =~ ^(1|true|yes)$ ]] && {
        info "skipping PATH modification because MANIPLACER_NO_MODIFY_PATH is set"
        return
    }
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) return ;;
    esac

    quoted_dir="$(shell_quote "$INSTALL_DIR")"
    case "$shell_name" in
        bash) rc_file="$HOME/.bashrc" ;;
        zsh) rc_file="$HOME/.zshrc" ;;
        fish)
            rc_file="$HOME/.config/fish/conf.d/maniplacer.fish"
            mkdir -p -- "${rc_file%/*}"
            if ! file_has_line "$rc_file" "$marker"; then
                printf '\n%s\nfish_add_path -- %s\n' "$marker" "$quoted_dir" >> "$rc_file"
            fi
            info "PATH configured in $rc_file"
            return
            ;;
        *)
            warning "could not determine a shell startup file; add $INSTALL_DIR to PATH manually"
            return
            ;;
    esac

    if ! file_has_line "$rc_file" "$marker"; then
        printf '\n%s\nexport PATH=%s:"$PATH"\n' "$marker" "$quoted_dir" >> "$rc_file"
    fi
    info "PATH configured in $rc_file"
}

verify_installation() {
    local installed_version
    installed_version="$("$INSTALL_DIR/$INSTALLED_NAME" --version)"
    [[ "${installed_version#v}" == "${LATEST_VERSION#v}" ]] || \
        error "installed version '$installed_version' does not match release '$LATEST_VERSION'"
    info "verified installed version $installed_version"
}

main() {
    check_prerequisites
    detect_platform
    get_release
    check_existing_installation
    download_binary
    install_binary
    update_path
    verify_installation
    info "successfully installed $TOOL_NAME $LATEST_VERSION"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
