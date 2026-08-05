# Quick Start

## 1. Simulate unit economics

```bash
profitctl simulate -f examples/valid_profit.yml
profitctl simulate -f examples/mix_profit.yml
profitctl simulate -f examples/hybrid_profit.yml
profitctl simulate -f examples/hybrid_steady_profit.yml
```

Example JSON excerpt from `examples/hybrid_steady_profit.yml`:

```json
{
  "revenue": {
    "mode": "hybrid",
    "total": 2000,
    "recurring_total": 2000,
    "minimum_uplift": 500
  },
  "margin": {
    "gross": 95,
    "operating_gross": 95,
    "cost_per_user": 2
  },
  "covenants": {
    "passed": true
  }
}
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

## 4. Build a starter cost model from a codebase

```bash
OPENROUTER_API_KEY=... EXA_API_KEY=... profitctl assess --path . --out cost-assessment.json
```

`assess` scans active configuration, finds code-backed providers with GPT-5.6 Terra, retrieves official-domain pricing through Exa, and emits a labeled starter cost model. Repository facts and inferred scale assumptions stay separate. Every cost line must reproduce an exact pricing excerpt, and the output includes the OpenRouter and Exa execution cost plus an estimated monthly total.

## 5. Output modes

```bash
profitctl simulate -f examples/valid_profit.yml --json
profitctl simulate -f examples/valid_profit.yml --markdown
profitctl simulate -f examples/valid_profit.yml --quiet
profitctl simulate -f examples/hybrid_profit.yml --json
```

## 6. Compare scenarios

```bash
profitctl compare examples/hybrid_steady_profit.yml examples/hybrid_profit.yml
profitctl compare examples/hybrid_steady_profit.yml examples/hybrid_profit.yml --json
profitctl compare examples/hybrid_steady_profit.yml examples/hybrid_profit.yml --markdown
```

The first config acts as the baseline. `compare` exits non-zero if any scenario breaches its covenants, which makes it usable in pricing reviews and CI gates.
For hybrid and pilot-style contracts, the comparison output separates booked margin from operating margin so one-time setup revenue does not masquerade as steady-state unit economics.

Example CLI excerpt:

```text
=== profitctl Scenario Comparison ===

Baseline: open_core_tiered

Scenario            Mode     Revenue    Recurring  Fees       Booked   Op Marg  CPU     Covenants
------------------------------------------------------------------------------------------------
open_core_tiered   tiered   $2425.00   $2425.00   $0.00      75.26    75.26    $6.00    PASS
open_core_mix      mix      $880.00    $880.00    $35.66     27.77    27.77    $6.36    PASS

Delta vs baseline:
  open_core_mix: revenue -1545.00, booked -47.49 pts, operating -47.49 pts, cost/user +0.36
```

For committed shareable compare artifacts, see:

- `benchmark_scenarios/reports/open_core_tiered_vs_mix.md`
- `benchmark_scenarios/reports/hybrid_steady_vs_pilot.md`
- `benchmark_scenarios/reports/hybrid_safe_vs_breach.md`

## 7. Normalize a calibration export

```bash
profitctl calibrate --input examples/calibration_exports/hybrid_profit_calibration.csv
profitctl calibrate --input examples/calibration_exports/hybrid_profit_calibration.csv --out /tmp/hybrid-calibration.yml
```

The normalized output can be referenced from a scenario via `calibration_file: ...`. Supported input formats are YAML, JSON, and CSV using `field,value` rows such as `paid_users.monthly,1800` or `plan_mix.starter,0.65`.

## 8. Starter config

```bash
profitctl init
profitctl validate -f profit.yml
```

The starter template now defaults to `pricing.mode: mix` and includes payment-fee assumptions so new scenarios begin from a realistic open-core baseline.

## 9. Pricing modes

- `tiered`: cumulative user bands with an optional unlimited final plan
- `mix`: explicit plan shares for free-to-paid and plan adoption mix
- `hybrid`: workspace minimums, base-fee-plus-seat contracts, and optional pilots

## 10. Hybrid seat-based contracts

Hybrid scenarios can set `simulation.billable_users` when billable seats differ from total modeled users. This keeps contract revenue honest for open-core or shared-workspace deployments where active users and paid seats diverge.

## 11. Calibration

You can add a top-level `calibration:` block or a `calibration_file:` reference to compare modeled revenue and payment fees against actual exported values. The simulator surfaces deltas in JSON, CLI, and Markdown output.
