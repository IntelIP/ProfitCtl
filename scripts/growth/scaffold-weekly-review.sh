#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scaffold-weekly-review.sh [options]

Options:
  --week-of YYYY-MM-DD       Week start date to use in the review file.
  --owner NAME               Owner name for the review.
  --release-version VERSION  Release version to record.
  --release-downloads COUNT  Release downloads to record.
  --homebrew-installs COUNT  Homebrew installs to record.
  --output PATH              Markdown output file path.
  --force                    Overwrite the output file if it already exists.
  --help                     Show this help text.

If --release-version or --release-downloads are omitted, the script will try to
read the latest GitHub release for IntelIP/ProfitCtl via `gh release view`.

The generated markdown is an operator artifact. Keep the resulting file in a
private system of record rather than committing live evaluator notes to the
public repository.
EOF
}

week_of="$(date -u +%Y-%m-%d)"
owner=""
release_version=""
release_downloads=""
homebrew_installs=""
output_path=""
force="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --week-of)
      week_of="${2:-}"
      shift 2
      ;;
    --owner)
      owner="${2:-}"
      shift 2
      ;;
    --release-version)
      release_version="${2:-}"
      shift 2
      ;;
    --release-downloads)
      release_downloads="${2:-}"
      shift 2
      ;;
    --homebrew-installs)
      homebrew_installs="${2:-}"
      shift 2
      ;;
    --output)
      output_path="${2:-}"
      shift 2
      ;;
    --force)
      force="true"
      shift
      ;;
    --help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -z "${output_path}" ]]; then
  output_path="/tmp/profitctl-weekly-review-${week_of}.md"
fi

if [[ -e "${output_path}" && "${force}" != "true" ]]; then
  echo "Output already exists: ${output_path}. Use --force to overwrite." >&2
  exit 1
fi

if [[ -z "${release_version}" || -z "${release_downloads}" ]]; then
  if command -v gh >/dev/null 2>&1; then
    if release_json="$(gh release view --repo IntelIP/ProfitCtl --json tagName,assets 2>/dev/null)"; then
      if [[ -z "${release_version}" ]]; then
        release_version="$(printf '%s' "${release_json}" | python3 -c 'import json,sys; data=json.load(sys.stdin); print(data.get("tagName",""))')"
      fi
      if [[ -z "${release_downloads}" ]]; then
        release_downloads="$(printf '%s' "${release_json}" | python3 -c 'import json,sys; data=json.load(sys.stdin); print(sum(a.get("downloadCount",0) for a in data.get("assets",[])))')"
      fi
    fi
  fi
fi

mkdir -p "$(dirname "${output_path}")"

cat > "${output_path}" <<EOF
# Weekly Review Template

Use this template every Friday after the design-partner and adoption dashboard review.

## Week Summary

- Week of: ${week_of}
- Owner: ${owner}
- Release version: ${release_version}
- Notes:

## Dashboard Snapshot

- New targets added:
- Targets contacted:
- Replies received:
- Installs completed:
- First compare runs completed:
- Calibration requests received:
- Design-partner commitments:
- Release downloads: ${release_downloads}
- Homebrew installs: ${homebrew_installs}

## What Changed

- What improved this week:
- What regressed:
- What stayed flat:
- Which benchmark pair got the clearest response:

## Friction Review

- Install friction:
- Activation friction:
- Scenario mapping friction:
- Output or trust friction:
- Repeated blocker theme:

## Evaluator Notes

- Evaluators to keep:
- Evaluators to drop:
- Evaluators needing follow-up:
- New signals from the best conversations:

## Source Check

- Is the dashboard current:
- Are evaluator notes current:
- Are Linear issues current:
- Are benchmark reports current:

## Decisions

Choose one primary direction for next week:

- improve onboarding
- improve scenario mapping
- improve output clarity
- continue outreach with the current motion

## Linear Follow-Through

- 1:
- 2:
- 3:

## Next Week

- top 3 targets:
- one product fix:
- one docs fix:
- one outreach change:
EOF

cat <<EOF
Wrote weekly review scaffold to:
${output_path}

Dashboard row:
| ${week_of} | ${release_version} |  |  |  |  |  |  |  | ${release_downloads} | ${homebrew_installs} |  |  |
EOF
