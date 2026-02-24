#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-${CI_COMMIT_TAG:-}}"
if [[ -z "${TAG}" ]]; then
  echo "TAG is required"
  exit 1
fi

: "${VPS_HOST:?VPS_HOST is required}"
: "${VPS_USER:?VPS_USER is required}"
: "${VPS_SSH_PRIVATE_KEY:?VPS_SSH_PRIVATE_KEY is required}"
: "${VPS_RELEASES_DIR:=/opt/profitctl/releases}"
: "${DOWNLOAD_BASE_URL:=https://downloads.intelip.co/profitctl}"

ROOT_DIR="$(dirname "${VPS_RELEASES_DIR}")"
CURRENT_LINK="${ROOT_DIR}/current"
MANIFEST_PATH="${ROOT_DIR}/index.json"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DIST_DIR="${ROOT}/dist/${TAG}"

mkdir -p "${HOME}/.ssh"
chmod 700 "${HOME}/.ssh"
eval "$(ssh-agent -s)"
trap 'ssh-agent -k >/dev/null 2>&1 || true' EXIT

# Support either raw multiline private keys or base64-encoded key payloads.
if printf "%s" "${VPS_SSH_PRIVATE_KEY}" | grep -q "BEGIN .*PRIVATE KEY"; then
  SSH_KEY_PAYLOAD="${VPS_SSH_PRIVATE_KEY}"
else
  SSH_KEY_PAYLOAD="$(printf "%s" "${VPS_SSH_PRIVATE_KEY}" | base64 -d 2>/dev/null || true)"
  if [[ -z "${SSH_KEY_PAYLOAD}" ]]; then
    echo "VPS_SSH_PRIVATE_KEY must be a raw private key or base64-encoded private key"
    exit 1
  fi
fi

printf "%s\n" "${SSH_KEY_PAYLOAD}" | sed 's/\r$//' | ssh-add -
ssh-keyscan -H "${VPS_HOST}" > "${HOME}/.ssh/known_hosts"
chmod 600 "${HOME}/.ssh/known_hosts"

TARGET_DIR="${VPS_RELEASES_DIR}/${TAG}"
ssh "${VPS_USER}@${VPS_HOST}" "mkdir -p '${TARGET_DIR}' '${VPS_RELEASES_DIR}' '${ROOT_DIR}'"
rsync -az "${DIST_DIR}/" "${VPS_USER}@${VPS_HOST}:${TARGET_DIR}/"
ssh "${VPS_USER}@${VPS_HOST}" "ln -sfn '${TARGET_DIR}' '${CURRENT_LINK}'"

MANIFEST_LOCAL="${DIST_DIR}/index.json"
{
  echo "{"
  echo "  \"latest\": \"${TAG}\"," 
  echo "  \"download_base\": \"${DOWNLOAD_BASE_URL}\"," 
  echo "  \"files\": ["
  first=1
  while IFS= read -r file; do
    name="$(basename "$file")"
    if [[ $first -eq 0 ]]; then echo ","; fi
    first=0
    printf "    {\"name\": \"%s\", \"url\": \"%s/releases/%s/%s\", \"current_url\": \"%s/current/%s\"}" \
      "$name" "$DOWNLOAD_BASE_URL" "$TAG" "$name" "$DOWNLOAD_BASE_URL" "$name"
  done < <(find "${DIST_DIR}" -maxdepth 1 -type f \( -name "profitctl_*" -o -name "SHA256SUMS" \) | sort)
  echo
  echo "  ]"
  echo "}"
} > "${MANIFEST_LOCAL}"

scp "${MANIFEST_LOCAL}" "${VPS_USER}@${VPS_HOST}:${TARGET_DIR}/index.json"
scp "${MANIFEST_LOCAL}" "${VPS_USER}@${VPS_HOST}:${MANIFEST_PATH}"

echo "Published ${TAG} to ${VPS_HOST}:${TARGET_DIR}"
