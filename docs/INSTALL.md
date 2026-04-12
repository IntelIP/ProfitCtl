# Install Guide

ProfitCtl ships a release-based install script for macOS and Linux. The script downloads the latest published binary from the release mirror, verifies the checksum when `SHA256SUMS` is available, and installs the binary into a writable prefix.

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
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | PROFITCTL_VERSION=v0.1.0 bash
```

## Custom Prefix

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | PROFITCTL_INSTALL_DIR=/usr/local/bin bash
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
profitctl simulate --help
```

## Exit Codes

- `0`: success
- `1`: covenant breach
- `2`: config/validation/usage error
- `3`: external runtime/provider failure
