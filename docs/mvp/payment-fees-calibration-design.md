# Payment Fees and Calibration Design

This doc defines the MVP design for two related gaps in ProfitCtl:

1. Payment-fee modeling that separates free vs paid users, monthly vs annual billing, and plan-aware fee burdens.
2. A practical calibration workflow for ingesting real production cost and usage inputs.

The goal is to keep the model honest enough for MVP pricing decisions without turning ProfitCtl into a full revenue operations platform.

## Current State

ProfitCtl currently has:

- `internal/pricing`: tiered revenue calculation and margin math
- `internal/cost`: fixed and variable cost calculation
- `internal/simulation`: scale simulation and Monte Carlo stress testing
- `internal/output`: CLI, JSON, and markdown renderers
- `internal/config`: YAML parsing and validation

What it does not currently have:

- any first-class payment-fee model
- any distinction between paid and free user cohorts
- any billing cadence awareness in revenue or fee outputs
- any calibration input path for real-world production data

The existing simulation paths only see one blended user count and one revenue number. That is sufficient for basic SaaS pricing, but it is too coarse for open-core MVP planning.

## Design Goals

- Model payment fees as a separate revenue-side cost, not a generic variable cost.
- Let the model distinguish free users from paid users.
- Let the model distinguish monthly and annual billing cohorts.
- Let the model express plan-aware fee burden by billing cadence.
- Support calibration from exports or CSV files without requiring a new data service.
- Keep backwards compatibility with existing `pricing` and `covenants` config.
- Make the rollout incremental so current tests and scenarios keep working.

## Non-Goals

- Full accounting ledger support.
- Real-time ingest from Stripe webhooks.
- User-level revenue recognition.
- Per-transaction charge reconciliation with every processor edge case.
- A generalized BI pipeline.

## Proposed Config Shape

Add a new top-level section:

```yaml
payment_fees:
  processor: stripe
  currency: usd
  free_user_fee:
    monthly_percent: 0.0
    per_transaction: 0.00
  paid_user_fee:
    monthly_percent: 2.9
    per_transaction: 0.30
  billing_mix:
    monthly_share: 0.80
    annual_share: 0.20
  plan_burden:
    - plan: starter
      applies_to: monthly
      fee_multiplier: 1.0
    - plan: pro
      applies_to: annual
      fee_multiplier: 0.6
  assumptions:
    annual_discount_percent: 15
    annual_prepaid_months: 12
```

Recommended schema rules:

- `payment_fees` is optional for backwards compatibility.
- `billing_mix.monthly_share + annual_share` must equal `1.0`.
- `free_user_fee` and `paid_user_fee` are both optional, but at least one must be present if `payment_fees` is enabled.
- `fee_multiplier` should default to `1.0` when omitted.
- `applies_to` should accept `monthly`, `annual`, or `all`.
- `processor` should default to `stripe` for MVP, but stay extensible.

## Revenue And Fee Model

The current revenue engine treats pricing as a tiered user ladder. That can stay as the default path for existing configs.

For payment fees, add a second layer that computes fees from revenue and billing mix:

```text
paid_monthly_revenue
paid_annual_revenue
free_revenue
payment_fees = monthly_fee_component + annual_fee_component
net_revenue = gross_revenue - payment_fees
```

Core rules:

- Free users contribute zero fee-bearing revenue unless the scenario explicitly models donations or add-ons.
- Monthly and annual cohorts are modeled separately because processor fees and discounting differ.
- Annual plans should be recognized as prepaid revenue in the scenario but can still be amortized for monthly reporting if needed.
- Fee burden should be plan-aware, so fee assumptions can vary by billing cadence or plan tier.

Suggested calculation shape:

```text
for each plan:
  paid_users = total_users * plan.share
  monthly_users = paid_users * billing_mix.monthly_share
  annual_users = paid_users * billing_mix.annual_share

  monthly_fee = monthly_users * monthly_revenue * percent_fee + monthly_users * transaction_fee
  annual_fee = annual_users * annual_revenue * percent_fee + annual_users * transaction_fee * annual_payment_events

  adjusted_fee = (monthly_fee + annual_fee) * fee_multiplier
```

Notes:

- `annual_payment_events` is typically `1` per prepaid customer per year, then amortized per month in reports.
- If the model cannot distinguish payments by plan, it should fall back to a blended fee burden rather than fail.
- The fee model should not replace revenue calculation; it should decorate it.

## Free Vs Paid Users

The MVP needs an explicit split because open-core and freemium products have very different fee burdens:

- free users may create support, infra, and usage cost with no direct payment fee
- paid users should carry the processor and billing fee burden
- annual users usually lower payment-fee burden per month but raise cash-flow complexity

Proposed scenario inputs:

```yaml
pricing:
  plans:
    - name: free
      price: 0
      limits:
        users: 5000
      cohort: free
    - name: starter
      price: 10
      limits:
        users: 1000
      cohort: paid
    - name: pro
      price: 29
      limits:
        users: 5000
      cohort: paid

payment_fees:
  billing_mix:
    monthly_share: 0.75
    annual_share: 0.25
```

The `cohort` field is optional for MVP but is the cleanest way to avoid mixing free and paid users into one tier. If omitted, the engine can infer `price == 0` as free and `price > 0` as paid.

## Calibration Workflow

Calibration should be a file-based workflow, not a service dependency.

### Inputs

Recommended ingest sources:

- Stripe exports for gross revenue, invoice totals, fee totals, refunds, and billing cadence
- payment processor CSVs for fee line items
- product analytics exports for active users, free-to-paid conversion, and plan mix
- infrastructure billing exports for cost grounding
- support or ticketing exports for load assumptions if needed

### Calibration File Shape

Add a calibration artifact under `benchmark_scenarios/` or `test/fixtures/` for the MVP:

```yaml
calibration:
  period: 2026-03
  source: stripe_export
  gross_revenue: 125000.00
  payment_fees: 4380.00
  free_users: 12000
  paid_users:
    monthly: 1800
    annual: 400
  plan_mix:
    starter: 0.65
    pro: 0.35
  usage:
    api_calls: 3200000
    storage_gb: 1820
    llm_tokens: 9400000
```

Calibration rules:

- Reject negative counts, percentages outside `0..1`, or mix totals that do not sum to `1.0`.
- Accept missing fields when the relevant model path is not enabled.
- Prefer explicit source values over inferred values.
- Store calibration snapshots by period so the same config can be re-run over time.

### Workflow

1. Import a calibration file or CSV export.
2. Normalize counts and monetary amounts into the ProfitCtl units.
3. Compare modeled gross revenue, fee burden, and cost per user against actuals.
4. Emit deltas and suggested assumption corrections.
5. Re-run the scenario with updated assumptions.

### Validation Outputs

Calibration should produce:

- modeled vs actual revenue delta
- modeled vs actual payment-fee delta
- modeled vs actual cost-per-user delta
- plan mix delta
- billing cadence delta

## Validation Rules

Add validation at config parse time and at calibration ingest time.

For config:

- `billing_mix` shares must sum to `1.0` within a small rounding tolerance.
- annual pricing assumptions must not be used without a prepaid billing cadence.
- payment fee percentages must be non-negative.
- transaction fees must be non-negative.
- `cohort` values should be limited to `free` or `paid` if present.

For calibration:

- revenue, fee, and cost totals must be non-negative.
- user counts must be non-negative integers.
- monthly and annual paid counts must not exceed total paid users.
- if both processor fee data and payment fee assumptions exist, the calibration file should state which one is authoritative.

## Output Changes

The simulation output should expose payment fees explicitly rather than bury them inside a generic cost bucket.

Suggested additions:

- CLI summary: show `Payment Fees` as a separate line.
- JSON output: add `payment_fees` under costs or revenue side, with monthly and annual breakdowns.
- markdown output: include a fee section near revenue and margin.

Example output fields:

```json
{
  "payment_fees": {
    "monthly": 4380.00,
    "annual_amortized": 920.00,
    "total": 5300.00
  },
  "revenue": {
    "gross": 125000.00,
    "net": 119700.00
  }
}
```

The important point is that gross revenue, payment fees, and operating costs should remain distinct.

## Phased Rollout

### Phase 1

- Add config schema and validation for `payment_fees`.
- Implement a simple blended payment-fee calculator.
- Surface payment fees in output.
- Add fixtures and tests.

### Phase 2

- Add free vs paid user split.
- Add monthly vs annual billing mix.
- Add plan-aware fee burden multipliers.
- Add calibration file ingest and delta reporting.

### Phase 3

- Add stronger plan-mix modeling and scenario calibration assistance.
- Add optional CSV import support for finance exports.
- Add richer reporting for model fit and assumption confidence.

## Exact Files Likely To Change

These are the files most likely to need code changes when implementing this design:

- `internal/config/parser.go`
- `internal/config/parser_test.go`
- `internal/config/validator.go`
- `internal/config/validator_test.go`
- `internal/pricing/calculator.go`
- `internal/pricing/calculator_test.go`
- `internal/output/types.go`
- `internal/output/json.go`
- `internal/output/json_test.go`
- `internal/output/cli.go`
- `internal/output/cli_test.go`
- `internal/output/markdown.go`
- `internal/output/markdown_test.go`
- `cmd/simulate.go`
- `cmd/simulate_test.go`
- `test/integration/simulate_test.go`
- `test/fixtures/valid_config.yml`
- `test/fixtures/no_pricing_config.yml`
- `test/fixtures/golden/go-webapp-expected.yml`
- `benchmark_scenarios/*.yml`

## Test Plan

The MVP implementation should add tests for:

- config parsing of `payment_fees`
- validation of billing mix sums and non-negative fee inputs
- revenue/fee calculations for:
  - free-only scenarios
  - paid monthly scenarios
  - paid annual scenarios
  - mixed monthly/annual scenarios
- output rendering of payment-fee sections
- calibration ingest and delta reporting
- one end-to-end integration scenario using a calibration-backed fixture

## Open Questions

- Should annual revenue be reported as cash collected or amortized revenue in the default CLI summary?
- Should free users ever carry any payment fee burden in cases like one-click upgrades or add-ons?
- Should plan-aware fee burden live in pricing config or in the payment-fee block?

Recommendation: keep the MVP answer simple. Report cash collected and amortized fee burden separately, and keep fee assumptions in a dedicated `payment_fees` block so the revenue engine stays focused.
