# Install Guide

## Requirements

- Go 1.21+

## Install from source

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
profitctl simulate -f examples/valid_profit.yml
```

## Exit Codes

- `0`: success
- `1`: covenant breach
- `2`: config/validation/usage error
- `3`: external runtime/provider failure
