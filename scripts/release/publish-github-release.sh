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
PUBLIC_KEY="${ROOT}/keys/profitctl-release-cosign.pub"

[[ -f "${PUBLIC_KEY}" ]] || {
  echo "missing release public key: ${PUBLIC_KEY}" >&2
  exit 1
}

if ! gh release view "${TAG}" --repo IntelIP/ProfitCtl >/dev/null 2>&1; then
  gh release create "${TAG}" --repo IntelIP/ProfitCtl --title "${TAG}" --notes "Automated ProfitCtl release for ${TAG}."
fi

gh release upload "${TAG}" "${OUT_DIR}"/profitctl_* "${OUT_DIR}"/SHA256SUMS "${OUT_DIR}"/SHA256SUMS.sig "${PUBLIC_KEY}" --repo IntelIP/ProfitCtl --clobber
echo "Published GitHub release assets for ${TAG}"
