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
case "${TAG}" in
  .|..|*/*) echo "TAG must be a single release name" >&2; exit 1 ;;
esac
mkdir -p "${OUT_DIR}"
STAGE_DIR="$(mktemp -d "${OUT_DIR}/.build.XXXXXX")"
trap 'rm -rf "${STAGE_DIR}"' EXIT

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
  PKG_NAME="profitctl_${TAG}_${GOOS}_${GOARCH}"
  PKG_DIR="${STAGE_DIR}/${PKG_NAME}"

  if [[ "${GOOS}" == "windows" ]]; then
    BIN_NAME="profitctl.exe"
    STANDARDS_BIN_NAME="profitctl-standards.exe"
    ARCHIVE_EXT="zip"
  fi

  mkdir -p "${PKG_DIR}"
  GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w -X main.version=${TAG}" -o "${PKG_DIR}/${BIN_NAME}" "${ROOT}"

  cp "${ROOT}/README.md" "${PKG_DIR}/README.md"
  mkdir -p "${PKG_DIR}/.codex-plugin"
  python3 - "${ROOT}/packages/codex-plugin" "${PKG_DIR}" "${TAG#v}" "${BIN_NAME}" <<'PYTHON'
import json
import sys
from pathlib import Path

source, package = map(Path, sys.argv[1:3])
manifest = Path(".codex-plugin/plugin.json")
plugin = json.loads((source / manifest).read_text())
plugin["version"] = sys.argv[3]
(package / manifest).write_text(json.dumps(plugin, indent=2) + "\n")
servers = json.loads((source / ".mcp.json").read_text())
servers["mcpServers"]["profitctl"]["command"] = "./" + sys.argv[4]
(package / ".mcp.json").write_text(json.dumps(servers, indent=2) + "\n")
PYTHON
  # Assemble the supported skill assets. Finder resource forks and private
  # scratch files in a developer checkout do not belong in release archives.
  SKILL_PACKAGE="${PKG_DIR}/skills/profitctl-cost-aware"
  mkdir -p "${SKILL_PACKAGE}"/{agents,bin,references/templates,scripts}
  cp "${SKILL_SOURCE}/SKILL.md" "${SKILL_PACKAGE}/"
  cp "${SKILL_SOURCE}/agents/openai.yaml" "${SKILL_PACKAGE}/agents/"
  cp "${SKILL_SOURCE}/scripts/run_profitctl_scenarios.py" "${SKILL_PACKAGE}/scripts/"
  cp "${SKILL_SOURCE}/references/templates/"{cloud-run,cloudflare-workers,vercel}-ai-saas.yml "${SKILL_PACKAGE}/references/templates/"
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
    (cd "${STAGE_DIR}" && zip -rq "${PKG_NAME}.zip" "${PKG_NAME}")
  else
    COPYFILE_DISABLE=1 tar --no-xattrs -C "${STAGE_DIR}" -czf "${STAGE_DIR}/${PKG_NAME}.tar.gz" "${PKG_NAME}"
  fi
  mv "${STAGE_DIR}/${PKG_NAME}.${ARCHIVE_EXT}" "${OUT_DIR}/"

  rm -rf "${PKG_DIR}"
done

echo "Artifacts built in ${OUT_DIR}"
