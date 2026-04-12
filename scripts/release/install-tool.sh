#!/usr/bin/env bash
set -euo pipefail

TOOL="${1:-}"
INSTALL_DIR="${2:-$(pwd)/.release-tools}"

if [[ -z "${TOOL}" ]]; then
  echo "usage: $0 <syft|cosign> [install-dir]" >&2
  exit 1
fi

OS_NAME="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS_NAME}" in
  linux|darwin) ;;
  *)
    echo "unsupported operating system for ${TOOL}: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64|amd64) ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
  *)
    echo "unsupported architecture for ${TOOL}: $(uname -m)" >&2
    exit 1
    ;;
esac

mkdir -p "${INSTALL_DIR}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

download() {
  local url="$1"
  local output="$2"
  curl -fsSL --retry 5 --retry-all-errors "${url}" -o "${output}"
}

case "${TOOL}" in
  syft)
    VERSION="${SYFT_VERSION:-v1.42.4}"
    VERSION_NO_V="${VERSION#v}"
    ARCHIVE="syft_${VERSION_NO_V}_${OS_NAME}_${ARCH_NAME}.tar.gz"
    URL="https://github.com/anchore/syft/releases/download/${VERSION}/${ARCHIVE}"
    download "${URL}" "${TMP_DIR}/${ARCHIVE}"
    tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "${INSTALL_DIR}" syft
    chmod +x "${INSTALL_DIR}/syft"
    ;;
  cosign)
    VERSION="${COSIGN_VERSION:-v3.0.6}"
    URL="https://github.com/sigstore/cosign/releases/download/${VERSION}/cosign-${OS_NAME}-${ARCH_NAME}"
    download "${URL}" "${INSTALL_DIR}/cosign"
    chmod +x "${INSTALL_DIR}/cosign"
    ;;
  *)
    echo "unsupported tool: ${TOOL}" >&2
    exit 1
    ;;
esac

echo "${INSTALL_DIR}/${TOOL}"
