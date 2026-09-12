#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-${CI_COMMIT_TAG:-}}"
if [[ -z "${TAG}" ]]; then
  echo "TAG is required (arg1 or CI_COMMIT_TAG)" >&2
  exit 1
fi

VERSION="${TAG#v}"
if [[ ! "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  echo "TAG must contain a valid semantic version: ${TAG}" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${ROOT}/dist/${TAG}"
TEMPLATE_DIR="${ROOT}/packages/profitctl"
TMP_DIR="$(mktemp -d)"
PACKAGE_DIR="${TMP_DIR}/profitctl"

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

command -v bun >/dev/null 2>&1 || {
  echo "bun is required to prepare the package" >&2
  exit 1
}

[[ -f "${TEMPLATE_DIR}/package.json" ]] || {
  echo "missing Bun package template: ${TEMPLATE_DIR}/package.json" >&2
  exit 1
}

mkdir -p "${OUT_DIR}" "${PACKAGE_DIR}"
cp -R "${TEMPLATE_DIR}/." "${PACKAGE_DIR}/"
rm -rf "${PACKAGE_DIR}/native"

MATRIX=(
  "linux amd64 tar.gz"
  "linux arm64 tar.gz"
  "darwin amd64 tar.gz"
  "darwin arm64 tar.gz"
  "windows amd64 zip"
)

for entry in "${MATRIX[@]}"; do
  read -r GOOS GOARCH EXT <<<"${entry}"
  ARCHIVE="${OUT_DIR}/profitctl_${TAG}_${GOOS}_${GOARCH}.${EXT}"
  UNPACK_DIR="${TMP_DIR}/unpack-${GOOS}-${GOARCH}"
  PACKAGE_BINARY_DIR="${PACKAGE_DIR}/native/${GOOS}-${GOARCH}"
  BINARY_NAME="profitctl"
  STANDARDS_NAME="profitctl-standards"

  if [[ "${GOOS}" == "windows" ]]; then
    BINARY_NAME="profitctl.exe"
    STANDARDS_NAME="profitctl-standards.exe"
  fi

  [[ -f "${ARCHIVE}" ]] || {
    echo "missing release archive: ${ARCHIVE}" >&2
    exit 1
  }

  mkdir -p "${UNPACK_DIR}" "${PACKAGE_BINARY_DIR}"
  if [[ "${EXT}" == "zip" ]]; then
    unzip -q "${ARCHIVE}" -d "${UNPACK_DIR}"
  else
    tar -xzf "${ARCHIVE}" -C "${UNPACK_DIR}"
  fi

  SOURCE_BINARY="${UNPACK_DIR}/profitctl_${TAG}_${GOOS}_${GOARCH}/${BINARY_NAME}"
  [[ -f "${SOURCE_BINARY}" ]] || {
    echo "missing binary in release archive: ${SOURCE_BINARY}" >&2
    exit 1
  }

  cp "${SOURCE_BINARY}" "${PACKAGE_BINARY_DIR}/${BINARY_NAME}"
  cp "${UNPACK_DIR}/profitctl_${TAG}_${GOOS}_${GOARCH}/skills/profitctl-cost-aware/bin/${STANDARDS_NAME}" "${PACKAGE_BINARY_DIR}/${STANDARDS_NAME}"
  if [[ "${GOOS}" != "windows" ]]; then
    chmod 755 "${PACKAGE_BINARY_DIR}/${BINARY_NAME}"
    chmod 755 "${PACKAGE_BINARY_DIR}/${STANDARDS_NAME}"
  fi
done

(
  cd "${PACKAGE_DIR}"
  bun pm pkg set "version=${VERSION}" >/dev/null
  bun pm pack \
    --destination "${OUT_DIR}" \
    --ignore-scripts
)

echo "Prepared ${OUT_DIR}/profitctl-${VERSION}.tgz"
