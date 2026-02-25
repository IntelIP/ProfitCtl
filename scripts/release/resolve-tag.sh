#!/usr/bin/env bash
set -euo pipefail

tag="${CI_COMMIT_TAG:-${CI_COMMIT_REF:-}}"
tag="${tag#refs/tags/}"
if [[ -z "${tag}" ]]; then
  echo "TAG is required" >&2
  exit 1
fi

printf '%s\n' "${tag}"
