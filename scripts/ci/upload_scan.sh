#!/usr/bin/env sh
set -eu

if [ "$#" -lt 5 ]; then
  echo "usage: upload_scan.sh <api_url> <repo> <commit> <tool> <format> [pr_number] <report_path>"
  exit 1
fi

API_URL="$1"
REPO="$2"
COMMIT="$3"
TOOL="$4"
FORMAT="$5"

if [ "$#" -eq 7 ]; then
  PR_NUMBER="$6"
  REPORT_PATH="$7"
else
  PR_NUMBER=""
  REPORT_PATH="$6"
fi

if [ ! -f "$REPORT_PATH" ]; then
  if [ "${APPSEC_STRICT_API:-true}" = "true" ]; then
    echo "report file not found and APPSEC_STRICT_API=true: $REPORT_PATH"
    exit 1
  fi
  echo "warning: report file not found ($REPORT_PATH), continuing in fail-open bootstrap mode"
  exit 0
fi

if [ -n "${PR_NUMBER}" ]; then
  if ! curl -fsS -X POST "$API_URL/v1/ingestion/scans" \
    -H "Authorization: Bearer ${APPSEC_API_TOKEN:-}" \
    -H "Idempotency-Key: ${CI_PIPELINE_ID:-manual}-${TOOL}-${COMMIT}" \
    -F "repo=$REPO" \
    -F "commit=$COMMIT" \
    -F "pr_number=$PR_NUMBER" \
    -F "scan_scope=pr" \
    -F "tool=$TOOL" \
    -F "format=$FORMAT" \
    -F "report=@$REPORT_PATH"; then
    if [ "${APPSEC_STRICT_API:-true}" = "true" ]; then
      echo "upload failed and APPSEC_STRICT_API=true"
      exit 1
    fi
    echo "warning: upload failed, continuing in fail-open bootstrap mode"
  fi
else
  if ! curl -fsS -X POST "$API_URL/v1/ingestion/scans" \
    -H "Authorization: Bearer ${APPSEC_API_TOKEN:-}" \
    -H "Idempotency-Key: ${CI_PIPELINE_ID:-manual}-${TOOL}-${COMMIT}" \
    -F "repo=$REPO" \
    -F "commit=$COMMIT" \
    -F "scan_scope=main" \
    -F "tool=$TOOL" \
    -F "format=$FORMAT" \
    -F "report=@$REPORT_PATH"; then
    if [ "${APPSEC_STRICT_API:-true}" = "true" ]; then
      echo "upload failed and APPSEC_STRICT_API=true"
      exit 1
    fi
    echo "warning: upload failed, continuing in fail-open bootstrap mode"
  fi
fi
