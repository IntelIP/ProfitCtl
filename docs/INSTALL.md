# Install Guide

ProfitCtl ships a release-based install script for macOS and Linux. The script downloads the latest published package from GitHub Releases by default, verifies the checksum from the matching `SHA256SUMS` asset, installs the binary into a writable prefix, and installs the portable `profitctl-cost-aware` Codex skill. If you operate a private mirror, override the source with `PROFITCTL_DOWNLOAD_BASE_URL`.

## Requirements

- `curl`
- `tar`
- `sha256sum` or `shasum`
- Python 3.10+ for the bundled Codex skill helper

## Recommended Install

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | bash
```

The script installs the CLI to `~/.local/bin` and the skill to `${CODEX_HOME:-$HOME/.codex}/skills/profitctl-cost-aware`. If the binary directory is not on your `PATH`, the installed skill uses its bundled platform binary.

## Pin a Version

The current evaluator prerelease is `v0.4.0-rc.1`. GitHub's latest-release endpoint excludes prereleases, so install it explicitly:

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | PROFITCTL_VERSION=v0.4.0-rc.1 bash
~/.local/bin/profitctl --version
~/.local/bin/profitctl mcp --workspace-root "$PWD" </dev/null
```

Expected version: `profitctl v0.4.0-rc.1`. Native CLI and MCP startup require no provider key. Paid assessment features require separate configuration. The installer supports macOS and Linux on amd64 and arm64. Windows amd64 users can extract the release ZIP and keep `profitctl.exe` and `profitctl-standards.exe` together; the shell installer does not support Windows.

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | PROFITCTL_VERSION=<tag> bash
```

## Bun

Install the release tarball with Bun:

```bash
curl -fSL https://github.com/IntelIP/ProfitCtl/releases/download/v0.4.0-rc.1/profitctl-0.4.0-rc.1.tgz -o profitctl-0.4.0-rc.1.tgz
bun add --global ./profitctl-0.4.0-rc.1.tgz
profitctl --version
profitctl mcp --workspace-root "$PWD" </dev/null
```

The tarball, its detached signature, and `RELEASE-SHA256SUMS` are release assets. Verify the tarball with the pinned public key before installing when verifying release authenticity. Public npm registry publication remains disabled.

For development from source:

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl/packages/profitctl
bun run build:native
bun run profitctl -- --help
```

The package includes a native launcher and can be packed as a Bun/npm tarball during a release:

```bash
bash scripts/release/build-artifacts.sh <tag>
bash scripts/release/prepare-bun-package.sh <tag>
```

Registry publication is intentionally disabled. GitHub visibility does not change this policy; these commands create no public npm package.

## Custom Prefix

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | PROFITCTL_INSTALL_DIR=/usr/local/bin bash
```

## Custom Mirror

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | \
  PROFITCTL_DOWNLOAD_BASE_URL=https://downloads.intelip.co/profitctl bash
```

## Developer Build from Source

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
go build -o profitctl .
./profitctl --version
```

## Developer Install via go install

```bash
go install github.com/IntelIP/ProfitCtl/cmd/profitctl@<version>
```

Choose a commit or future tag that contains `cmd/profitctl`; the current published
`v0.2.0` tag predates this package. Source and Go installs are explicit developer
fallbacks. The checksum-verified GitHub Release installer is the primary supported
path; the private Bun package is the secondary development path.

## Verify the Current Published Install

```bash
profitctl --help
```

## Verify a Release Candidate

Use these checks for an artifact built from a commit containing the stable CLI identity contract. Publishing that artifact and publishing the Bun package remain separate release actions.

```bash
profitctl --version
profitctl --help
profitctl doctor -f /path/to/profit.yml --catalog /path/to/provider-catalog.yml
```

`doctor` checks the current executable identity, injected version, supported runtime, selected config, and selected provider catalog. Missing required inputs produce an actionable message and exit code `2`; the command does not select another binary, create config, or mutate machine-global state.

## Verify Release Integrity

Public GitHub releases include:

- detached Cosign signatures for each archive
- SPDX JSON SBOMs for each archive
- a signed `SHA256SUMS` manifest
- `profitctl-release-cosign.pub`

Example verification flow for `darwin_arm64`:

```bash
TAG=v0.4.0-rc.1
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl_${TAG}_darwin_arm64.tar.gz"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl_${TAG}_darwin_arm64.tar.gz.sig"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl_${TAG}_darwin_arm64.spdx.json"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl_${TAG}_darwin_arm64.spdx.json.sig"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/SHA256SUMS"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/SHA256SUMS.sig"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl-release-cosign.pub"

printf '%s  %s\n' aebe6076c728012a09e7795ac9c76cfb1729db640bdc798932c00bd6b3c3023f profitctl-release-cosign.pub | shasum -a 256 -c -
cosign verify-blob --key profitctl-release-cosign.pub --signature "profitctl_${TAG}_darwin_arm64.tar.gz.sig" --insecure-ignore-tlog=true "profitctl_${TAG}_darwin_arm64.tar.gz"
cosign verify-blob --key profitctl-release-cosign.pub --signature "profitctl_${TAG}_darwin_arm64.spdx.json.sig" --insecure-ignore-tlog=true "profitctl_${TAG}_darwin_arm64.spdx.json"
cosign verify-blob --key profitctl-release-cosign.pub --signature SHA256SUMS.sig --insecure-ignore-tlog=true SHA256SUMS
awk -v asset="profitctl_${TAG}_darwin_arm64.tar.gz" '$2 == asset {print}' SHA256SUMS | shasum -a 256 -c -
```

Use `brew install cosign` or the upstream Sigstore install path if `cosign` is not already available.

## Exit Codes

- `0`: success
- `1`: covenant breach
- `2`: config/validation/usage error
- `3`: external runtime/provider failure
