#!/usr/bin/env sh
set -eu

if [ "$#" -ne 3 ]; then
  echo "usage: evaluate_gate.sh <api_url> <repo_name> <pr_number>"
  exit 1
fi

API_URL="$1"
REPO_NAME="$2"
PR_NUMBER="$3"
STRICT_API="${APPSEC_STRICT_API:-false}"

ENCODED_REPO="$(printf '%s' "$REPO_NAME" | jq -sRr @uri)"
if ! RESP="$(curl -fsS -H "Authorization: Bearer ${APPSEC_API_TOKEN:-}" "$API_URL/v1/prs/$PR_NUMBER/summary?repo=$ENCODED_REPO")"; then
  if [ "$STRICT_API" = "true" ]; then
    echo "gate evaluation failed and APPSEC_STRICT_API=true"
    exit 1
  fi
  echo "warning: gate evaluation unavailable, continuing in fail-open bootstrap mode"
  exit 0
fi

if ! MODE="$(echo "$RESP" | jq -r '.gate_mode // "report_only"')" || ! FAIL="$(echo "$RESP" | jq -r '.should_fail // false')"; then
  if [ "$STRICT_API" = "true" ]; then
    echo "gate response parse failed and APPSEC_STRICT_API=true"
    exit 1
  fi
  echo "warning: invalid gate response, continuing in fail-open bootstrap mode"
  exit 0
fi

echo "gate_mode=$MODE should_fail=$FAIL"

if [ "$MODE" = "enforce" ] && [ "$FAIL" = "true" ]; then
  echo "Policy failed in enforce mode"
  exit 1
fi

echo "Policy gate passed"
