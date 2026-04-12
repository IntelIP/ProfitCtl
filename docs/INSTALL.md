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

## Exit Codes

- `0`: success
- `1`: covenant breach
- `2`: config/validation/usage error
- `3`: external runtime/provider failure
