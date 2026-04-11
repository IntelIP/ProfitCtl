# Quick Start

## 1. Simulate unit economics

```bash
profitctl simulate -f examples/valid_profit.yml
profitctl simulate -f examples/mix_profit.yml
profitctl simulate -f examples/hybrid_profit.yml
profitctl simulate -f examples/hybrid_steady_profit.yml
```

## 2. Validate config

```bash
profitctl validate -f examples/valid_profit.yml
profitctl validate -f examples/mix_profit.yml
profitctl validate -f examples/hybrid_profit.yml
profitctl validate -f examples/hybrid_steady_profit.yml
```

## 3. Detect dependencies/services (LLM)

```bash
OPENROUTER_API_KEY=... profitctl detect --path . --out detect-report.json
```

## 4. Output modes

```bash
profitctl simulate -f examples/valid_profit.yml --json
profitctl simulate -f examples/valid_profit.yml --markdown
profitctl simulate -f examples/valid_profit.yml --quiet
profitctl simulate -f examples/hybrid_profit.yml --json
```

## 5. Compare scenarios

```bash
profitctl compare examples/hybrid_steady_profit.yml examples/hybrid_profit.yml
profitctl compare examples/hybrid_steady_profit.yml examples/hybrid_profit.yml --json
profitctl compare examples/hybrid_steady_profit.yml examples/hybrid_profit.yml --markdown
```

The first config acts as the baseline. `compare` exits non-zero if any scenario breaches its covenants, which makes it usable in pricing reviews and CI gates.

## 6. Starter config

```bash
profitctl init
profitctl validate -f profit.yml
```

The starter template now defaults to `pricing.mode: mix` and includes payment-fee assumptions so new scenarios begin from a realistic open-core baseline.

## 7. Pricing modes

- `tiered`: cumulative user bands with an optional unlimited final plan
- `mix`: explicit plan shares for free-to-paid and plan adoption mix
- `hybrid`: workspace minimums, base-fee-plus-seat contracts, and optional pilots

## 8. Calibration

You can add a top-level `calibration:` block to compare modeled revenue and payment fees against actual exported values. The simulator surfaces deltas in JSON, CLI, and Markdown output.
