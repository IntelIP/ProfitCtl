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
For hybrid and pilot-style contracts, the comparison output separates booked margin from operating margin so one-time setup revenue does not masquerade as steady-state unit economics.

## 6. Normalize a calibration export

```bash
profitctl calibrate --input examples/calibration_exports/hybrid_profit_calibration.csv
profitctl calibrate --input examples/calibration_exports/hybrid_profit_calibration.csv --out /tmp/hybrid-calibration.yml
```

The normalized output can be referenced from a scenario via `calibration_file: ...`. Supported input formats are YAML, JSON, and CSV using `field,value` rows such as `paid_users.monthly,1800` or `plan_mix.starter,0.65`.

## 7. Starter config

```bash
profitctl init
profitctl validate -f profit.yml
```

The starter template now defaults to `pricing.mode: mix` and includes payment-fee assumptions so new scenarios begin from a realistic open-core baseline.

## 8. Pricing modes

- `tiered`: cumulative user bands with an optional unlimited final plan
- `mix`: explicit plan shares for free-to-paid and plan adoption mix
- `hybrid`: workspace minimums, base-fee-plus-seat contracts, and optional pilots

## 9. Hybrid seat-based contracts

Hybrid scenarios can set `simulation.billable_users` when billable seats differ from total modeled users. This keeps contract revenue honest for open-core or shared-workspace deployments where active users and paid seats diverge.

## 10. Calibration

You can add a top-level `calibration:` block or a `calibration_file:` reference to compare modeled revenue and payment fees against actual exported values. The simulator surfaces deltas in JSON, CLI, and Markdown output.
