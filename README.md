# ProfitCtl

`profitctl` is a CLI for profit-first unit economics simulations.

It helps teams model fixed + variable costs, simulate growth and stress scenarios, validate profitability covenants, and (optionally) detect services/dependencies from source config files using an LLM.

## Project Identity

- Canonical repository: `github.com/IntelIP/ProfitCtl`
- Go module path: `github.com/IntelIP/ProfitCtl`

## Features

- `simulate`: run 12-month scale + Monte Carlo stress simulations
- `validate`: validate config structure and rules
- `detect`: scan repository config files and return JSON service/dependency analysis
- Output formats: human CLI, JSON, Markdown
- CI/CD via Woodpecker (Hostinger VPS + builder pool)

## Install

### Build from source

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
go build -o profitctl .
./profitctl --help
```

### Go install

```bash
go install github.com/IntelIP/ProfitCtl@latest
```

## Quick Usage

```bash
# simulate
profitctl simulate -f examples/valid_profit.yml

# strict config validation
profitctl validate -f examples/valid_profit.yml

# detect services/dependencies from repo files
OPENROUTER_API_KEY=... profitctl detect --path . --out detect-report.json
```

## Release Downloads

Release artifacts are published to:

- `https://downloads.intelip.co/profitctl/releases/<tag>/`
- `https://downloads.intelip.co/profitctl/current/`

## Documentation

- [Documentation Index](docs/DOCS_INDEX.md)
- [Install Guide](docs/INSTALL.md)
- [Quick Start](docs/QUICK_START.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Woodpecker + Hostinger Runbook](docs/deployment/WOODPECKER_HOSTINGER_SETUP.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
