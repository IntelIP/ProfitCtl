#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-}"
ASSET_DIR="${2:-}"
PUBLIC_KEY="${3:-}"

if [[ -z "${TAG}" || -z "${ASSET_DIR}" ]]; then
  echo "usage: $0 <tag> <asset-dir> [public-key-path]" >&2
  exit 1
fi

if [[ -z "${PUBLIC_KEY}" ]]; then
  PUBLIC_KEY="${ASSET_DIR}/profitctl-release-cosign.pub"
fi

command -v cosign >/dev/null 2>&1 || {
  echo "cosign is required on PATH" >&2
  exit 1
}

verify_checksum() {
  local checksum_file="$1"
  local archive_file="$2"
  local archive_name
  archive_name="$(basename "${archive_file}")"
  local expected_hash
  expected_hash="$(awk -v name="${archive_name}" '$2 == name || $2 == "*"name { print $1; exit }' "${checksum_file}")"
  [[ -n "${expected_hash}" ]] || {
    echo "checksum entry missing for ${archive_name}" >&2
    exit 1
  }

  local actual_hash
  if command -v shasum >/dev/null 2>&1; then
    actual_hash="$(shasum -a 256 "${archive_file}" | awk '{ print $1 }')"
  elif command -v sha256sum >/dev/null 2>&1; then
    actual_hash="$(sha256sum "${archive_file}" | awk '{ print $1 }')"
  else
    echo "missing required checksum verifier: shasum or sha256sum" >&2
    exit 1
  fi

  [[ "${actual_hash}" == "${expected_hash}" ]] || {
    echo "checksum verification failed for ${archive_name}" >&2
    exit 1
  }
}

case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
  linux) OS_NAME="linux" ;;
  darwin) OS_NAME="darwin" ;;
  *) echo "unsupported operating system: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
  *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

ARCHIVE_NAME="profitctl_${TAG}_${OS_NAME}_${ARCH_NAME}.tar.gz"
SBOM_NAME="profitctl_${TAG}_${OS_NAME}_${ARCH_NAME}.spdx.json"

ARCHIVE_PATH="${ASSET_DIR}/${ARCHIVE_NAME}"
SBOM_PATH="${ASSET_DIR}/${SBOM_NAME}"
CHECKSUM_PATH="${ASSET_DIR}/SHA256SUMS"

[[ -f "${PUBLIC_KEY}" ]] || { echo "missing public key: ${PUBLIC_KEY}" >&2; exit 1; }
[[ -f "${CHECKSUM_PATH}" ]] || { echo "missing checksum file: ${CHECKSUM_PATH}" >&2; exit 1; }
[[ -f "${CHECKSUM_PATH}.sig" ]] || { echo "missing checksum signature: ${CHECKSUM_PATH}.sig" >&2; exit 1; }
[[ -f "${ARCHIVE_PATH}" ]] || { echo "missing archive: ${ARCHIVE_PATH}" >&2; exit 1; }
[[ -f "${ARCHIVE_PATH}.sig" ]] || { echo "missing archive signature: ${ARCHIVE_PATH}.sig" >&2; exit 1; }
[[ -f "${SBOM_PATH}" ]] || { echo "missing sbom: ${SBOM_PATH}" >&2; exit 1; }
[[ -f "${SBOM_PATH}.sig" ]] || { echo "missing sbom signature: ${SBOM_PATH}.sig" >&2; exit 1; }

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT
export HOME="${TMP_DIR}/home"
mkdir -p "${HOME}"

cosign verify-blob --key "${PUBLIC_KEY}" --signature "${CHECKSUM_PATH}.sig" --insecure-ignore-tlog=true "${CHECKSUM_PATH}" >/dev/null
cosign verify-blob --key "${PUBLIC_KEY}" --signature "${ARCHIVE_PATH}.sig" --insecure-ignore-tlog=true "${ARCHIVE_PATH}" >/dev/null
cosign verify-blob --key "${PUBLIC_KEY}" --signature "${SBOM_PATH}.sig" --insecure-ignore-tlog=true "${SBOM_PATH}" >/dev/null

verify_checksum "${CHECKSUM_PATH}" "${ARCHIVE_PATH}"

echo "Release signatures and checksums verified for ${TAG}"
