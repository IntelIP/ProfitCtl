#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-${CI_COMMIT_TAG:-}}"
if [[ -z "${TAG}" ]]; then
  echo "TAG is required (arg1 or CI_COMMIT_TAG)" >&2
  exit 1
fi

if [[ -z "${COSIGN_PRIVATE_KEY:-}" ]]; then
  echo "COSIGN_PRIVATE_KEY is required" >&2
  exit 1
fi

if [[ -z "${COSIGN_PASSWORD:-}" ]]; then
  echo "COSIGN_PASSWORD is required" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${ROOT}/dist/${TAG}"

command -v cosign >/dev/null 2>&1 || {
  echo "cosign is required on PATH" >&2
  exit 1
}

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT
export HOME="${TMP_DIR}/home"
mkdir -p "${HOME}"

sign_blob() {
  local blob_path="$1"
  rm -f "${blob_path}.sig"
  cosign sign-blob \
    --key env://COSIGN_PRIVATE_KEY \
    --tlog-upload=false \
    --use-signing-config=false \
    --new-bundle-format=false \
    --output-signature "${blob_path}.sig" \
    --yes \
    "${blob_path}" >/dev/null
}

targets=()
while IFS= read -r target; do
  targets+=("${target}")
done < <(
  find "${OUT_DIR}" -maxdepth 1 -type f \
    \( -name "profitctl_${TAG}_*.tar.gz" -o -name "profitctl_${TAG}_*.zip" -o -name "profitctl_${TAG}_*.spdx.json" -o -name "SHA256SUMS" \) \
    | sort
)

if [[ "${#targets[@]}" -eq 0 ]]; then
  echo "no release assets found to sign in ${OUT_DIR}" >&2
  exit 1
fi

for target in "${targets[@]}"; do
  sign_blob "${target}"
done

echo "Signatures written in ${OUT_DIR}"
