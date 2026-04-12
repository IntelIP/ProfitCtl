#!/usr/bin/env bash
set -euo pipefail

DOWNLOAD_BASE_URL="${PROFITCTL_DOWNLOAD_BASE_URL:-https://downloads.intelip.co/profitctl}"
INSTALL_DIR="${PROFITCTL_INSTALL_DIR:-}"
VERSION="${PROFITCTL_VERSION:-}"
BIN_NAME="${PROFITCTL_BIN_NAME:-profitctl}"

usage() {
  cat <<'EOF'
Usage: install.sh [--version TAG] [--prefix DIR] [--download-base-url URL]

Environment variables:
  PROFITCTL_VERSION           Release tag to install, for example v0.1.0
  PROFITCTL_DOWNLOAD_BASE_URL Release mirror root, default https://downloads.intelip.co/profitctl
  PROFITCTL_INSTALL_DIR       Install prefix, default $HOME/.local/bin
  PROFITCTL_BIN_NAME          Binary name to install, default profitctl
EOF
}

log() {
  printf '%s\n' "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

extract_latest_tag() {
  local manifest="$1"
  sed -n 's/.*"latest"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$manifest" | head -n 1
}

verify_checksum() {
  local checksum_file="$1"
  local archive_file="$2"
  local archive_dir
  archive_dir="$(dirname "$archive_file")"

  if command -v shasum >/dev/null 2>&1; then
    local expected_line
    expected_line="$(grep -F " $(basename "$archive_file")" "$checksum_file" || true)"
    [[ -n "$expected_line" ]] || die "checksum entry missing for $(basename "$archive_file")"
    (
      cd "$archive_dir"
      printf '%s\n' "$expected_line" | shasum -a 256 -c >/dev/null 2>&1
    ) || die "checksum verification failed"
    return 0
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    local expected_line
    expected_line="$(grep -F " $(basename "$archive_file")" "$checksum_file" || true)"
    [[ -n "$expected_line" ]] || die "checksum entry missing for $(basename "$archive_file")"
    (
      cd "$archive_dir"
      printf '%s\n' "$expected_line" | sha256sum -c >/dev/null 2>&1
    ) || die "checksum verification failed"
    return 0
  fi

  die "neither sha256sum nor shasum is available"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      [[ $# -ge 2 ]] || die "--version requires an argument"
      VERSION="$2"
      shift 2
      ;;
    --prefix)
      [[ $# -ge 2 ]] || die "--prefix requires an argument"
      INSTALL_DIR="$2"
      shift 2
      ;;
    --download-base-url)
      [[ $# -ge 2 ]] || die "--download-base-url requires an argument"
      DOWNLOAD_BASE_URL="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

need_cmd curl
need_cmd tar

case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
  linux) OS_NAME="linux" ;;
  darwin) OS_NAME="darwin" ;;
  *) die "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
  x86_64|amd64) ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
  *) die "unsupported architecture: $(uname -m)" ;;
esac

if [[ -z "$INSTALL_DIR" ]]; then
  INSTALL_DIR="$HOME/.local/bin"
fi

mkdir -p "$INSTALL_DIR"

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

BASE_URL="${DOWNLOAD_BASE_URL%/}"
if [[ -n "$VERSION" ]]; then
  FOUND_VERSION=""
  for candidate in "$VERSION" "v${VERSION#v}"; do
    [[ -n "$candidate" ]] || continue
    if curl -fsSL "${BASE_URL}/releases/${candidate}/SHA256SUMS" -o "${TMP_DIR}/SHA256SUMS"; then
      FOUND_VERSION="${candidate}"
      break
    fi
  done
  [[ -n "$FOUND_VERSION" ]] || die "unable to find release assets for version ${VERSION}"
  VERSION="${FOUND_VERSION}"
else
  curl -fsSL "${BASE_URL}/current/index.json" -o "${TMP_DIR}/index.json"
  VERSION="$(extract_latest_tag "${TMP_DIR}/index.json")"
  [[ -n "$VERSION" ]] || die "unable to determine latest release version"
  curl -fsSL "${BASE_URL}/releases/${VERSION}/SHA256SUMS" -o "${TMP_DIR}/SHA256SUMS"
fi

ARCHIVE_NAME="profitctl_${VERSION}_${OS_NAME}_${ARCH_NAME}.tar.gz"
ARCHIVE_PATH="${TMP_DIR}/${ARCHIVE_NAME}"
CHECKSUM_PATH="${TMP_DIR}/SHA256SUMS"
ASSET_URL="${BASE_URL}/releases/${VERSION}/${ARCHIVE_NAME}"

log "Downloading ${ARCHIVE_NAME}"
curl -fsSL "$ASSET_URL" -o "$ARCHIVE_PATH"

log "Verifying checksum"
verify_checksum "$CHECKSUM_PATH" "$ARCHIVE_PATH"

EXTRACT_DIR="${TMP_DIR}/extract"
mkdir -p "$EXTRACT_DIR"
tar -xzf "$ARCHIVE_PATH" -C "$EXTRACT_DIR"

BIN_PATH="$(find "$EXTRACT_DIR" -type f -name "$BIN_NAME" -perm -111 | head -n 1)"
[[ -n "$BIN_PATH" ]] || die "installed archive did not contain ${BIN_NAME}"

TARGET_PATH="${INSTALL_DIR}/${BIN_NAME}"
cp "$BIN_PATH" "$TARGET_PATH"
chmod 755 "$TARGET_PATH"

log "Installed ${BIN_NAME} to ${TARGET_PATH}"

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    log "Add ${INSTALL_DIR} to PATH if it is not already present."
    ;;
esac
