# ProfitCtl

`profitctl` is a CLI for profit-first unit economics simulations.

It helps teams model fixed + variable costs, simulate growth and stress scenarios, validate profitability covenants, and (optionally) detect services/dependencies from source config files using an LLM.

## Project Identity

- Canonical repository: `github.com/IntelIP/ProfitCtl`
- Go module path: `github.com/IntelIP/ProfitCtl`

## Features

- `simulate`: run 12-month scale + Monte Carlo stress simulations
- `compare`: evaluate multiple pricing scenarios side by side
- `calibrate`: normalize YAML, JSON, or CSV calibration exports into ProfitCtl calibration artifacts
- `ledger`: ingest deterministic cost-contract fixtures into a local normalized ledger and query stable JSON
- `upstash reconcile`: reconcile synthetic Redis polling fixtures into the local ledger with a machine-readable variance artifact
- `validate`: validate config structure and rules
- `detect`: scan repository config files and return JSON service/dependency analysis
- `assess`: detect code-backed providers, retrieve official pricing through Exa, and create a GPT-5.6 Terra starter cost model that includes its own OpenRouter and Exa execution cost
- portable `profitctl-cost-aware` Codex skill: requires source-backed assessment when material price or scale context is missing
- `doctor`: check local binary, runtime, config, and catalog readiness without changing local state
- pricing modes: `tiered`, `mix`, and `hybrid`
- workspace-aware pricing and minimum-floor scenario modeling
- payment-fee modeling with monthly vs annual billing mix
- calibration deltas for modeled vs actual revenue and fees
- `calibration_file` support for external calibration artifacts
- explicit `billable_users` support for hybrid seat-based contracts
- booked vs operating margin reporting for one-time-heavy contract scenarios
- full-economics reporting across delivery, productization, and adoption cost layers
- Output formats: human CLI, JSON, Markdown
- CI/CD via Woodpecker (Hostinger VPS + builder pool)

## Install

### Recommended install

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | bash
```

The installer downloads a GitHub Release archive, verifies it against that release's `SHA256SUMS`, installs `profitctl` to `~/.local/bin`, and installs the portable `profitctl-cost-aware` Codex skill to `${CODEX_HOME:-$HOME/.codex}/skills`.

### Homebrew

```bash
brew tap IntelIP/profitctl
brew install profitctl
```

See [Install Guide](docs/INSTALL.md) for pinned versions, Homebrew, custom prefixes, and source-based installs.

### Developer build from source

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
go build -o profitctl .
./profitctl --version
./profitctl doctor -f examples/valid_profit.yml --catalog /path/to/provider-catalog.yml
```

### Developer install with Go

```bash
go install github.com/IntelIP/ProfitCtl/cmd/profitctl@<version>
```

Choose a commit or future tag that contains `cmd/profitctl`; the current published
`v0.2.0` tag predates this package. Go-based paths are explicit developer fallbacks.
GitHub Release installation is the supported user path; Homebrew is secondary.

## Quick Usage

```bash
# simulate
profitctl simulate -f examples/valid_profit.yml

# mix-mode open-core pricing
profitctl simulate -f examples/mix_profit.yml

# hybrid contract pricing with pilot
profitctl simulate -f examples/hybrid_profit.yml --json

# compare steady-state vs pilot hybrid contracts
profitctl compare examples/hybrid_steady_profit.yml examples/hybrid_profit.yml

# normalize a calibration export
profitctl calibrate --input examples/calibration_exports/hybrid_profit_calibration.csv

# ingest and query a deterministic cost observation locally
profitctl ledger ingest --ledger /tmp/profitctl-ledger.json --input test/fixtures/cost_contract/v1/upstash_idle_polling.json
profitctl ledger query --ledger /tmp/profitctl-ledger.json --workload idle_worker_polling

# reconcile synthetic Redis polling telemetry without a provider call
profitctl upstash reconcile --input test/fixtures/upstash/v1/idle_polling.json --ledger /tmp/profitctl-ledger.json --output /tmp/upstash-reconciliation.json

# strict config validation
profitctl validate -f examples/valid_profit.yml

# detect services/dependencies from repo files
OPENROUTER_API_KEY=... profitctl detect --path . --out detect-report.json

# build a source-backed starter cost model from the codebase
OPENROUTER_API_KEY=... EXA_API_KEY=... profitctl assess --path . --out cost-assessment.json
```

`assess` makes two Terra requests and one bounded official-domain Exa pricing lookup per distinct receipt. Provider domains and pricing paths come from ProfitCtl's trusted registry, while provider discovery must quote exact code evidence. Its JSON separates target-code costs from assessment runtime costs, prices both model input and output, and records inferred scale, exact pricing excerpts, cost-line math, the estimated monthly total, and the recommendation. It does not claim forecast values are actual billing.

## Release Downloads

Canonical public release artifacts are published on GitHub Releases:

- `https://github.com/IntelIP/ProfitCtl/releases`
- each release includes:
  - per-platform archives
  - per-platform SPDX JSON SBOMs
  - detached Cosign signatures
  - `SHA256SUMS` plus a detached signature
  - `profitctl-release-cosign.pub` for offline verification

Operational mirrors may also publish to:
- `https://downloads.intelip.co/profitctl/releases/<tag>/`
- `https://downloads.intelip.co/profitctl/current/`

## Open-Core Product

The open-source core gives teams a local CLI to model pricing, recurring-margin risk, and contract safety before they ship pricing or sign deals.

The paid layer should sit on top of that core through hosted workflows, collaboration, policy enforcement, and support, not by weakening the local product.

## IntelIP Modeling Pack

The repo now includes IntelIP-specific economics scenario packs for conservative, target, stress, paid-pilot, and tight-free rollout analysis.

- calibration notes: `benchmark_scenarios/intelip_model_calibration_notes.md`
- cost inventory: `benchmark_scenarios/intelip_tooling_cost_inventory.md`
- rollout recommendation pack: `benchmark_scenarios/intelip_ops_pricing_pack.md`

## Documentation

- [Documentation Index](docs/DOCS_INDEX.md)
- [Install Guide](docs/INSTALL.md)
- [Quick Start](docs/QUICK_START.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Open-Core Packaging](docs/OPEN_CORE_PACKAGING.md)
- [Open-Core Roadmap](docs/OPEN_CORE_ROADMAP.md)
- [Upstash Reconciliation v1](docs/contracts/upstash-reconciliation-v1.md)
- [Benchmark Scenarios](benchmark_scenarios/README.md)
- [Full Economics Cost Layers](docs/full-economics-cost-layers.md)
- [IntelIP Ops Pricing Pack](benchmark_scenarios/intelip_ops_pricing_pack.md)
- [Growth Assets](docs/growth/README.md)
- [Benchmark Outreach Pack](docs/growth/benchmark-outreach-pack.md)
- [Value Proposition](docs/growth/value-proposition.md)
- [Design-Partner Offer](docs/growth/design-partner-offer.md)
- [Outreach Message Templates](docs/growth/outreach-message-templates.md)
- [Evaluator Session Template](docs/growth/evaluator-session-template.md)
- [Adoption Dashboard](docs/growth/adoption-dashboard.md)
- [Design-Partner Operating System](docs/growth/design-partner-operating-system.md)
- [MVP Pricing Release Guide](docs/release/MVP_PRICING_RELEASE.md)
- [Woodpecker + Hostinger Runbook](docs/deployment/WOODPECKER_HOSTINGER_SETUP.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## AppSec PR Review

`ProfitCtl` can publish PR security findings into the separate `appsec-mvp` service.

- PR and main scans run in GitHub Actions via `semgrep`, `trivy`, `gitleaks`, and `osv-scanner`
- findings are uploaded to `appsec-mvp`
- GitHub PRs receive:
  - commit status `appsec/mvp`
  - check run `AppSec MVP Review`
  - selective inline comments for high-confidence findings

Required GitHub repository secrets:

- `APPSEC_API_URL`
- `APPSEC_API_TOKEN`

Optional GitHub repository variable:

- `APPSEC_STRICT_API`

Required `appsec-mvp` runtime config:

- `GITHUB_APP_ID`
- `GITHUB_APP_PRIVATE_KEY_PEM` or `GITHUB_APP_PRIVATE_KEY_PATH`
- `GITHUB_WEBHOOK_SECRET`

The initial `/.appsec.yml` policy is set to `report_only` for rollout validation, and the helper scripts default `APPSEC_STRICT_API` to `false` so AppSec service outages do not hard-fail PRs during onboarding. Switch the policy to `enforce` and set `APPSEC_STRICT_API=true` once the integration is stable.

## License

MIT. See [LICENSE](LICENSE).
