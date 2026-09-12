#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-${CI_COMMIT_TAG:-}}"
if [[ -z "${TAG}" ]]; then
  echo "TAG is required (arg1 or CI_COMMIT_TAG)"
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${ROOT}/dist/${TAG}"
SKILL_SOURCE="${ROOT}/skills/profitctl-cost-aware"
mkdir -p "${OUT_DIR}"

[[ -f "${SKILL_SOURCE}/SKILL.md" ]] || {
  echo "missing bundled skill: ${SKILL_SOURCE}" >&2
  exit 1
}

MATRIX=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
  "windows amd64"
)

for entry in "${MATRIX[@]}"; do
  GOOS="${entry%% *}"
  GOARCH="${entry##* }"

  BIN_NAME="profitctl"
  STANDARDS_BIN_NAME="profitctl-standards"
  ARCHIVE_EXT="tar.gz"
  PKG_DIR="${OUT_DIR}/profitctl_${TAG}_${GOOS}_${GOARCH}"

  if [[ "${GOOS}" == "windows" ]]; then
    BIN_NAME="profitctl.exe"
    STANDARDS_BIN_NAME="profitctl-standards.exe"
    ARCHIVE_EXT="zip"
  fi

  mkdir -p "${PKG_DIR}"
  GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w -X main.version=${TAG}" -o "${PKG_DIR}/${BIN_NAME}" "${ROOT}"

  cp "${ROOT}/README.md" "${PKG_DIR}/README.md"
  mkdir -p "${PKG_DIR}/skills"
  cp -R "${SKILL_SOURCE}" "${PKG_DIR}/skills/profitctl-cost-aware"
  mkdir -p "${PKG_DIR}/skills/profitctl-cost-aware/bin"
  cp "${PKG_DIR}/${BIN_NAME}" "${PKG_DIR}/skills/profitctl-cost-aware/bin/${BIN_NAME}"
  GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" \
      -o "${PKG_DIR}/skills/profitctl-cost-aware/bin/${STANDARDS_BIN_NAME}" \
      "${ROOT}/scripts/judge_cost_standards.go"
  chmod 755 "${PKG_DIR}/skills/profitctl-cost-aware/scripts/run_profitctl_scenarios.py"
  cp "${PKG_DIR}/skills/profitctl-cost-aware/bin/${STANDARDS_BIN_NAME}" "${PKG_DIR}/${STANDARDS_BIN_NAME}"
  if [[ "${GOOS}" != "windows" ]]; then
    chmod 755 "${PKG_DIR}/skills/profitctl-cost-aware/bin/${BIN_NAME}"
    chmod 755 "${PKG_DIR}/skills/profitctl-cost-aware/bin/${STANDARDS_BIN_NAME}"
  fi

  if [[ "${ARCHIVE_EXT}" == "zip" ]]; then
    (cd "${OUT_DIR}" && zip -rq "profitctl_${TAG}_${GOOS}_${GOARCH}.zip" "profitctl_${TAG}_${GOOS}_${GOARCH}")
  else
    tar -C "${OUT_DIR}" -czf "${OUT_DIR}/profitctl_${TAG}_${GOOS}_${GOARCH}.tar.gz" "profitctl_${TAG}_${GOOS}_${GOARCH}"
  fi

  rm -rf "${PKG_DIR}"
done

echo "Artifacts built in ${OUT_DIR}"
