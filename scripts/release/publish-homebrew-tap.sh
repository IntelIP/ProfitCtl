#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TAP_REPO="${HOMEBREW_TAP_REPO:-IntelIP/homebrew-profitctl}"
TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

if [[ ! -f "${ROOT}/Formula/profitctl.rb" ]]; then
  echo "Formula/profitctl.rb is missing" >&2
  exit 1
fi

gh repo clone "${TAP_REPO}" "${TMP_DIR}/tap" -- --depth=1 >/dev/null

mkdir -p "${TMP_DIR}/tap/Formula"
cp "${ROOT}/Formula/profitctl.rb" "${TMP_DIR}/tap/Formula/profitctl.rb"

if [[ -z "$(git -C "${TMP_DIR}/tap" status --short -- Formula/profitctl.rb)" ]]; then
  echo "Homebrew tap already matches Formula/profitctl.rb"
  exit 0
fi

git -C "${TMP_DIR}/tap" add Formula/profitctl.rb
git -C "${TMP_DIR}/tap" commit -m "chore: update profitctl formula"
git -C "${TMP_DIR}/tap" push origin HEAD

echo "Published formula to ${TAP_REPO}"
