#!/usr/bin/env bash
set -euo pipefail

candidate="${BUILDKITE_COMMIT:?BUILDKITE_COMMIT is required}"
resolved_candidate="$(git rev-parse --verify 'HEAD^{commit}')"
if [[ "$resolved_candidate" != "$candidate" ]]; then
  printf 'checkout mismatch: expected %s, found %s\n' "$candidate" "$resolved_candidate" >&2
  exit 1
fi

apt-get update
apt-get install --yes --no-install-recommends ca-certificates curl zip
rm -rf /var/lib/apt/lists/*

format_drift="$(gofmt -l .)"
if [[ -n "$format_drift" ]]; then
  printf 'gofmt drift:\n%s\n' "$format_drift" >&2
  exit 1
fi

go mod tidy
git diff --exit-code -- go.mod go.sum
go vet ./...
go test ./...

bash -n scripts/install.sh
tag="v0.0.0-ci"
bash scripts/release/build-artifacts.sh "$tag"
bash scripts/release/create-checksums.sh "$tag"

temporary_root="$(mktemp -d)"
cleanup() {
  rm -rf -- "$temporary_root"
}
trap cleanup EXIT

mirror_root="$temporary_root/mirror"
install_root="$temporary_root/bin"
mkdir -p "$mirror_root/releases/$tag" "$mirror_root/current" "$install_root"
cp "dist/$tag/profitctl_${tag}_"* "dist/$tag/SHA256SUMS" "$mirror_root/releases/$tag/"
printf '{"latest":"%s"}\n' "$tag" > "$mirror_root/current/index.json"

PROFITCTL_DOWNLOAD_BASE_URL="file://$mirror_root" \
PROFITCTL_VERSION="$tag" \
PROFITCTL_INSTALL_DIR="$install_root" \
  bash scripts/install.sh
"$install_root/profitctl" --help >/dev/null
"$install_root/profitctl" validate -f examples/mix_profit.yml >/dev/null

go_version="$(go env GOVERSION)"
architecture="$(uname -m)"
case "$go_version:$architecture" in
  *[!A-Za-z0-9._:-]*)
    printf 'non-portable runtime identity\n' >&2
    exit 1
    ;;
esac
printf '{"schemaVersion":"profitctl-self-hosted/v0.1","commit":"%s","goVersion":"%s","architecture":"%s","status":"passed"}\n' \
  "$resolved_candidate" "$go_version" "$architecture" > profitctl-self-hosted-evidence.json
