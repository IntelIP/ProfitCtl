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

- Woodpecker pipeline file: `.woodpecker.yml`
- PR/main verification jobs on `pool=shared-kvm`
- Tag release jobs on `pool=builder`
- Release artifacts mirrored to Hostinger VPS at `/opt/profitctl/releases`
- Public artifact endpoint: `https://downloads.intelip.co/profitctl`
