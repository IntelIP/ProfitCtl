#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-${CI_COMMIT_TAG:-}}"
if [[ -z "${TAG}" ]]; then
  echo "TAG is required (arg1 or CI_COMMIT_TAG)" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${ROOT}/dist/${TAG}"

command -v syft >/dev/null 2>&1 || {
  echo "syft is required on PATH" >&2
  exit 1
}

mkdir -p "${OUT_DIR}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

export SYFT_CHECK_FOR_APP_UPDATE=false
export XDG_CACHE_HOME="${TMP_DIR}/cache"
mkdir -p "${XDG_CACHE_HOME}"

artifact_stem() {
  local file_name="$1"
  file_name="${file_name%.tar.gz}"
  file_name="${file_name%.zip}"
  printf '%s\n' "${file_name}"
}

shopt -s nullglob
archives=("${OUT_DIR}"/profitctl_"${TAG}"_*.tar.gz "${OUT_DIR}"/profitctl_"${TAG}"_*.zip)
shopt -u nullglob

if [[ "${#archives[@]}" -eq 0 ]]; then
  echo "no release archives found in ${OUT_DIR}" >&2
  exit 1
fi

for archive in "${archives[@]}"; do
  stem="$(artifact_stem "$(basename "${archive}")")"
  syft scan "file:${archive}" \
    --quiet \
    --source-name "${stem}" \
    --source-version "${TAG}" \
    --output "spdx-json=${OUT_DIR}/${stem}.spdx.json"
done

syft scan "dir:${ROOT}" \
  --quiet \
  --exclude "./.git" \
  --exclude "./dist" \
  --source-name "profitctl-source" \
  --source-version "${TAG}" \
  --output "spdx-json=${OUT_DIR}/profitctl_${TAG}_source.spdx.json"

echo "SBOMs written in ${OUT_DIR}"
