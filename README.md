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
- pricing modes: `tiered`, `mix`, and `hybrid`
- payment-fee modeling with monthly vs annual billing mix
- calibration deltas for modeled vs actual revenue and fees
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

# mix-mode open-core pricing
profitctl simulate -f examples/mix_profit.yml

# hybrid contract pricing with pilot
profitctl simulate -f examples/hybrid_profit.yml --json

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
- [Benchmark Scenarios](benchmark_scenarios/README.md)
- [Woodpecker + Hostinger Runbook](docs/deployment/WOODPECKER_HOSTINGER_SETUP.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## AppSec PR Review

`ProfitCtl` can publish PR security findings into the separate `appsec-mvp` service.

- PR scans run in Woodpecker via `semgrep`, `trivy`, `gitleaks`, and `osv-scanner`
- findings are uploaded to `appsec-mvp`
- GitHub PRs receive:
  - commit status `appsec/mvp`
  - check run `AppSec MVP Review`
  - selective inline comments for high-confidence findings

Required Woodpecker secret:

- `APPSEC_API_URL`
- `APPSEC_API_TOKEN`

Required `appsec-mvp` runtime config:

- `GITHUB_APP_ID`
- `GITHUB_APP_PRIVATE_KEY_PEM` or `GITHUB_APP_PRIVATE_KEY_PATH`
- `GITHUB_WEBHOOK_SECRET`

The initial `/.appsec.yml` policy is set to `report_only` for rollout validation, and the helper scripts default `APPSEC_STRICT_API` to `false` so AppSec service outages do not hard-fail PRs during onboarding. Switch the policy to `enforce` and set `APPSEC_STRICT_API=true` once the integration is stable.

## License

MIT. See [LICENSE](LICENSE).
