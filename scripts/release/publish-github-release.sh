#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-${CI_COMMIT_TAG:-}}"
if [[ -z "${TAG}" ]]; then
  echo "TAG is required"
  exit 1
fi

if [[ -z "${GITHUB_TOKEN_RELEASE:-}" ]]; then
  echo "GITHUB_TOKEN_RELEASE is required"
  exit 1
fi

export GH_TOKEN="${GITHUB_TOKEN_RELEASE}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${ROOT}/dist/${TAG}"

if ! gh release view "${TAG}" --repo IntelIP/ProfitCtl >/dev/null 2>&1; then
  gh release create "${TAG}" --repo IntelIP/ProfitCtl --title "${TAG}" --notes "Automated Woodpecker release for ${TAG}."
fi

gh release upload "${TAG}" "${OUT_DIR}"/profitctl_* "${OUT_DIR}"/SHA256SUMS --repo IntelIP/ProfitCtl --clobber
echo "Published GitHub release assets for ${TAG}"
