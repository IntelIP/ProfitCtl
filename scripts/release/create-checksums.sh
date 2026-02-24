#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-${CI_COMMIT_TAG:-}}"
if [[ -z "${TAG}" ]]; then
  echo "TAG is required (arg1 or CI_COMMIT_TAG)"
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${ROOT}/dist/${TAG}"

cd "${OUT_DIR}"
ls profitctl_* > /dev/null

sha256sum profitctl_* > SHA256SUMS
sha256sum --check --status SHA256SUMS
echo "Checksums written: ${OUT_DIR}/SHA256SUMS"
