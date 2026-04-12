#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-${CI_COMMIT_TAG:-}}"
if [[ -z "${TAG}" ]]; then
  echo "TAG is required (arg1 or CI_COMMIT_TAG)" >&2
  exit 1
fi

REPO="${PROFITCTL_RELEASE_REPO:-IntelIP/ProfitCtl}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

verify_checksum() {
  local checksum_file="$1"
  local archive_file="$2"
  local archive_name
  archive_name="$(basename "$archive_file")"
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
    echo "checksum verification failed" >&2
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

ARCHIVE_EXT="tar.gz"

need_cmd gh
need_cmd tar

ARCHIVE_NAME="profitctl_${TAG}_${OS_NAME}_${ARCH_NAME}.${ARCHIVE_EXT}"

gh release download "${TAG}" \
  --repo "${REPO}" \
  --pattern "${ARCHIVE_NAME}" \
  --pattern "SHA256SUMS" \
  --dir "${TMP_DIR}"

ARCHIVE_PATH="${TMP_DIR}/${ARCHIVE_NAME}"
CHECKSUM_PATH="${TMP_DIR}/SHA256SUMS"

verify_checksum "${CHECKSUM_PATH}" "${ARCHIVE_PATH}"

EXTRACT_DIR="${TMP_DIR}/extract"
mkdir -p "${EXTRACT_DIR}"

case "${ARCHIVE_NAME}" in
  *.zip)
    need_cmd unzip
    unzip -q "${ARCHIVE_PATH}" -d "${EXTRACT_DIR}"
    ;;
  *.tar.gz)
    tar -xzf "${ARCHIVE_PATH}" -C "${EXTRACT_DIR}"
    ;;
  *)
    echo "unsupported archive format: ${ARCHIVE_NAME}" >&2
    exit 1
    ;;
esac

BIN_PATH="$(find "${EXTRACT_DIR}" -type f -name profitctl -perm -111 | head -n 1)"
[[ -n "${BIN_PATH}" ]] || {
  echo "installed archive did not contain profitctl" >&2
  exit 1
}

"${BIN_PATH}" --help >/dev/null
"${BIN_PATH}" validate -f "${ROOT}/examples/mix_profit.yml" >/dev/null
"${BIN_PATH}" compare \
  "${ROOT}/benchmark_scenarios/open_core_tiered.yml" \
  "${ROOT}/benchmark_scenarios/open_core_mix.yml" >/dev/null

echo "Release smoke passed for ${TAG}"
