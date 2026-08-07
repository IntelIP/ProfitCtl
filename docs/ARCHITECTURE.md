# Architecture

## Core Layers

- `cmd/`: CLI command orchestration
- `internal/config`: parse + validation
- `internal/cost`: fixed/variable/total cost engines
- `internal/simulation`: scale + Monte Carlo engines
- `internal/pricing`: revenue + margin calculations
- `internal/covenant`: covenant rules and violations
- `internal/output`: CLI/JSON/Markdown rendering
- `internal/scanner`: file collection + LLM detection stack
- `pkg/types`: shared model types

## CI/CD Architecture

- GitHub Actions provides required public repo checks: `verify-go`, `verify-install-smoke`, and `security-scan`
- Canonical public release channel: GitHub Releases
- Public releases include detached Cosign signatures, SPDX JSON SBOMs, and the release verification public key
- Release artifacts mirrored to Hostinger VPS at `/opt/profitctl/releases`
- Optional mirror endpoint: `https://downloads.intelip.co/profitctl`
