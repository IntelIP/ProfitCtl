# Install Guide

ProfitCtl ships a release-based install script for macOS and Linux. The script downloads the latest published binary from GitHub Releases by default, verifies the checksum from the matching `SHA256SUMS` asset, and installs the binary into a writable prefix. If you operate a private mirror, override the source with `PROFITCTL_DOWNLOAD_BASE_URL`.

## Requirements

- `curl`
- `tar`
- `sha256sum` or `shasum`

## Recommended Install

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | bash
```

The script installs to `~/.local/bin` by default. If that directory is not on your `PATH`, add it before running `profitctl`.

## Pin a Version

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | PROFITCTL_VERSION=v0.1.2 bash
```

## Homebrew

```bash
brew tap IntelIP/profitctl
brew install profitctl
```

For explicit tap-qualified installs:

```bash
brew install IntelIP/profitctl/profitctl
```

## Custom Prefix

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | PROFITCTL_INSTALL_DIR=/usr/local/bin bash
```

## Custom Mirror

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | \
  PROFITCTL_DOWNLOAD_BASE_URL=https://downloads.intelip.co/profitctl bash
```

## Install from Source

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
go build -o profitctl .
./profitctl --help
```

## Install via go install

```bash
go install github.com/IntelIP/ProfitCtl@latest
```

## Verify

```bash
profitctl --help
profitctl validate -f examples/mix_profit.yml
```

## Verify Release Integrity

Public GitHub releases include:

- detached Cosign signatures for each archive
- SPDX JSON SBOMs for each archive
- a signed `SHA256SUMS` manifest
- `profitctl-release-cosign.pub`

Example verification flow for `darwin_arm64`:

```bash
TAG=<tag>
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl_${TAG}_darwin_arm64.tar.gz"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl_${TAG}_darwin_arm64.tar.gz.sig"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl_${TAG}_darwin_arm64.spdx.json"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl_${TAG}_darwin_arm64.spdx.json.sig"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/SHA256SUMS"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/SHA256SUMS.sig"
curl -fsSLO "https://github.com/IntelIP/ProfitCtl/releases/download/${TAG}/profitctl-release-cosign.pub"

cosign verify-blob --key profitctl-release-cosign.pub --signature "profitctl_${TAG}_darwin_arm64.tar.gz.sig" --insecure-ignore-tlog=true "profitctl_${TAG}_darwin_arm64.tar.gz"
cosign verify-blob --key profitctl-release-cosign.pub --signature "profitctl_${TAG}_darwin_arm64.spdx.json.sig" --insecure-ignore-tlog=true "profitctl_${TAG}_darwin_arm64.spdx.json"
cosign verify-blob --key profitctl-release-cosign.pub --signature SHA256SUMS.sig --insecure-ignore-tlog=true SHA256SUMS
shasum -a 256 -c SHA256SUMS
```

Use `brew install cosign` or the upstream Sigstore install path if `cosign` is not already available.

## Exit Codes

- `0`: success
- `1`: covenant breach
- `2`: config/validation/usage error
- `3`: external runtime/provider failure
