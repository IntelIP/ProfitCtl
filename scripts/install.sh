#!/usr/bin/env bash
set -euo pipefail

DOWNLOAD_BASE_URL="${PROFITCTL_DOWNLOAD_BASE_URL:-}"
INSTALL_DIR="${PROFITCTL_INSTALL_DIR:-}"
VERSION="${PROFITCTL_VERSION:-}"
BIN_NAME="${PROFITCTL_BIN_NAME:-profitctl}"
RELEASE_REPO="${PROFITCTL_RELEASE_REPO:-IntelIP/ProfitCtl}"
CODEX_HOME_DIR="${CODEX_HOME:-${HOME}/.codex}"

usage() {
  cat <<'EOF'
Usage: install.sh [--version TAG] [--prefix DIR] [--download-base-url URL]

Environment variables:
  PROFITCTL_VERSION           Release tag to install, for example v0.1.0
  PROFITCTL_RELEASE_REPO      GitHub release repository, default IntelIP/ProfitCtl
  PROFITCTL_DOWNLOAD_BASE_URL Optional release mirror root override
  PROFITCTL_INSTALL_DIR       Install prefix, default $HOME/.local/bin
  PROFITCTL_BIN_NAME          Binary name to install, default profitctl
  CODEX_HOME                  Codex home, default $HOME/.codex
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

extract_latest_github_tag() {
  local manifest="$1"
  sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$manifest" | head -n 1
}

verify_checksum() {
  local checksum_file="$1"
  local archive_file="$2"
  local archive_dir
  archive_dir="$(dirname "$archive_file")"

  if command -v shasum >/dev/null 2>&1; then
    local expected_line
    expected_line="$(grep -E "[[:space:]]\\*?$(basename "$archive_file")$" "$checksum_file" || true)"
    [[ -n "$expected_line" ]] || die "checksum entry missing for $(basename "$archive_file")"
    (
      cd "$archive_dir"
      printf '%s\n' "$expected_line" | shasum -a 256 -c >/dev/null 2>&1
    ) || die "checksum verification failed"
    return 0
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    local expected_line
    expected_line="$(grep -E "[[:space:]]\\*?$(basename "$archive_file")$" "$checksum_file" || true)"
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
CHECKSUM_URL=""
ASSET_URL=""
if [[ -n "$VERSION" ]]; then
  FOUND_VERSION=""
  for candidate in "$VERSION" "v${VERSION#v}"; do
    [[ -n "$candidate" ]] || continue
    if [[ -n "$BASE_URL" ]]; then
      checksum_candidate_url="${BASE_URL}/releases/${candidate}/SHA256SUMS"
    else
      checksum_candidate_url="https://github.com/${RELEASE_REPO}/releases/download/${candidate}/SHA256SUMS"
    fi

    if curl -fsSL "${checksum_candidate_url}" -o "${TMP_DIR}/SHA256SUMS"; then
      FOUND_VERSION="${candidate}"
      CHECKSUM_URL="${checksum_candidate_url}"
      break
    fi
  done
  [[ -n "$FOUND_VERSION" ]] || die "unable to find release assets for version ${VERSION}"
  VERSION="${FOUND_VERSION}"
else
  if [[ -n "$BASE_URL" ]]; then
    curl -fsSL "${BASE_URL}/current/index.json" -o "${TMP_DIR}/index.json"
    VERSION="$(extract_latest_tag "${TMP_DIR}/index.json")"
    CHECKSUM_URL="${BASE_URL}/releases/${VERSION}/SHA256SUMS"
  else
    curl -fsSL "https://api.github.com/repos/${RELEASE_REPO}/releases/latest" -o "${TMP_DIR}/release.json"
    VERSION="$(extract_latest_github_tag "${TMP_DIR}/release.json")"
    CHECKSUM_URL="https://github.com/${RELEASE_REPO}/releases/download/${VERSION}/SHA256SUMS"
  fi
  [[ -n "$VERSION" ]] || die "unable to determine latest release version"
fi

curl -fsSL "${CHECKSUM_URL}" -o "${TMP_DIR}/SHA256SUMS"

ARCHIVE_NAME="profitctl_${VERSION}_${OS_NAME}_${ARCH_NAME}.tar.gz"
ARCHIVE_PATH="${TMP_DIR}/${ARCHIVE_NAME}"
CHECKSUM_PATH="${TMP_DIR}/SHA256SUMS"
if [[ -n "$BASE_URL" ]]; then
  ASSET_URL="${BASE_URL}/releases/${VERSION}/${ARCHIVE_NAME}"
else
  ASSET_URL="https://github.com/${RELEASE_REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"
fi

log "Downloading ${ARCHIVE_NAME}"
curl -fsSL "$ASSET_URL" -o "$ARCHIVE_PATH"

log "Verifying checksum"
verify_checksum "$CHECKSUM_PATH" "$ARCHIVE_PATH"

EXTRACT_DIR="${TMP_DIR}/extract"
mkdir -p "$EXTRACT_DIR"
tar -xzf "$ARCHIVE_PATH" -C "$EXTRACT_DIR"

BIN_PATH="$(find "$EXTRACT_DIR" -type f -name "$BIN_NAME" -perm -111 | head -n 1)"
[[ -n "$BIN_PATH" ]] || die "installed archive did not contain ${BIN_NAME}"
SKILL_SOURCE="$(find "$EXTRACT_DIR" -type d -path "*/skills/profitctl-cost-aware" | head -n 1)"

TARGET_PATH="${INSTALL_DIR}/${BIN_NAME}"
cp "$BIN_PATH" "$TARGET_PATH"
chmod 755 "$TARGET_PATH"

# Keep the MCP companion beside the installed CLI. Older archives may only
# contain it inside the bundled skill; pre-MCP releases have no companion.
STANDARDS_PATH="$(dirname "$BIN_PATH")/profitctl-standards"
if [[ ! -f "$STANDARDS_PATH" && -n "$SKILL_SOURCE" ]]; then
  STANDARDS_PATH="${SKILL_SOURCE}/bin/profitctl-standards"
fi
if [[ -f "$STANDARDS_PATH" ]]; then
  cp "$STANDARDS_PATH" "${INSTALL_DIR}/profitctl-standards"
  chmod 755 "${INSTALL_DIR}/profitctl-standards"
fi

log "Installed ${BIN_NAME} to ${TARGET_PATH}"

if [[ -n "$SKILL_SOURCE" ]]; then
  SKILL_TARGET="${CODEX_HOME_DIR}/skills/profitctl-cost-aware"
  mkdir -p "$SKILL_TARGET"
  cp -R "${SKILL_SOURCE}/." "$SKILL_TARGET/"
  chmod 755 "${SKILL_TARGET}/scripts/run_profitctl_scenarios.py"
  if [[ -f "${SKILL_TARGET}/bin/${BIN_NAME}" ]]; then
    chmod 755 "${SKILL_TARGET}/bin/${BIN_NAME}"
  fi
  if [[ -f "${SKILL_TARGET}/bin/profitctl-standards" ]]; then
    chmod 755 "${SKILL_TARGET}/bin/profitctl-standards"
  fi
  log "Installed profitctl-cost-aware skill to ${SKILL_TARGET}"
else
  log "Release ${VERSION} predates the bundled Codex skill; installed CLI only."
fi

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    log "Add ${INSTALL_DIR} to PATH if it is not already present."
    ;;
esac
