# Hybrid Pricing Design

This document defines the implementation path for workspace minimums, base-fee-plus-seat pricing, and paid pilot contracts in ProfitCtl.

The design keeps the current tiered `pricing.plans` model working as-is and adds an explicit contract model for MVP pricing structures that are not cumulative tiers.

## Why This Exists

ProfitCtl currently assumes pricing is either:

- cumulative tiered plan pricing
- or a last-plan fallback when `limits` are omitted

That is enough for a simple SaaS ladder, but it is not enough for the pricing shapes we need for the MVP:

- workspace minimums
- base platform fee plus per-seat fee
- paid pilot / proof-of-value contracts

Those shapes need to be represented directly so the simulator can report real revenue, margin, and covenant impact instead of forcing everything through fake user caps.

## Design Goals

- Preserve backward compatibility for existing `pricing.plans` configs.
- Add one explicit contract shape for hybrid pricing.
- Make pilot revenue and recurring revenue visible separately.
- Keep the simulator output understandable in CLI, JSON, and Markdown.
- Avoid inventing free-user/paid-user accounting until the model actually needs it.

## Non-Goals

- Replacing the current tiered pricing engine.
- Modeling discount codes, usage-based overages, or invoice-level tax rules.
- Solving free-user vs paid-user split accounting in this change.
- Introducing a second revenue system for historical configs.

## Proposed YAML Schema

The MVP schema should support two mutually exclusive pricing modes:

1. Legacy tiered plans.
2. Explicit contract pricing.

### 1. Legacy Tiered Mode

Keep the current structure unchanged:

```yaml
pricing:
  mode: tiered
  plans:
    - name: basic
      price: 10
      limits:
        users: 1000
    - name: pro
      price: 29
      limits:
        users: 5000
```

Rules:

- `mode` defaults to `tiered` when `plans` is present and no contract block exists.
- This path must remain compatible with current fixtures and tests.

### 2. Hybrid Contract Mode

Use a single contract block for workspace minimums, base fees, and seat fees:

```yaml
pricing:
  mode: hybrid
  contract:
    base_platform_fee: 500
    per_seat_fee: 25
    workspace_minimum: 2000
    included_seats: 0
```

Field rules:

- `base_platform_fee` is the recurring monthly platform charge.
- `per_seat_fee` is the recurring monthly price per billable seat.
- `workspace_minimum` is the minimum monthly recurring revenue floor.
- `included_seats` is optional and defaults to `0`.
- `workspace_minimum` is optional; if omitted, the floor is not applied.

### 3. Paid Pilot Shape

Add a pilot block under the contract so a deal can start as a paid pilot and convert into steady-state hybrid billing:

```yaml
pricing:
  mode: hybrid
  contract:
    base_platform_fee: 500
    per_seat_fee: 25
    workspace_minimum: 2000
    pilot:
      setup_fee: 5000
      monthly_fee: 1500
      duration_months: 3
```

Pilot rules:

- `setup_fee` is billed once, in month 1 of the pilot.
- `monthly_fee` is billed every month during the pilot.
- `duration_months` must be at least `1`.
- After the pilot ends, billing switches to the recurring hybrid contract.

This is intentionally a single pilot-to-production path. If we later need multiple phases, we can generalize to a phase array, but that is not required for the MVP.

## Revenue Calculation Rules

Hybrid revenue should be calculated monthly and then summed across the simulation horizon.

### Recurring Hybrid Revenue

For a month with `billable_seats`:

```text
seat_revenue = max(0, billable_seats - included_seats) * per_seat_fee
recurring_revenue = base_platform_fee + seat_revenue
monthly_recurring_revenue = max(workspace_minimum, recurring_revenue)
minimum_uplift = max(0, workspace_minimum - recurring_revenue)
```

### Paid Pilot Revenue

For pilot months:

```text
pilot_month_revenue = monthly_fee + setup_fee_if_first_month
```

If the contract includes both a pilot and a hybrid steady-state block:

- bill the pilot for `duration_months`
- switch to recurring hybrid revenue after the pilot
- do not double charge the recurring hybrid amount during the pilot unless the config explicitly adds both, which this MVP should not allow

### Recognition Rules

- `total_revenue` is the sum of all billed revenue in the simulation horizon.
- `recurring_revenue` is the steady-state recurring component only.
- `one_time_revenue` is setup or pilot initialization revenue.
- `minimum_uplift` is the amount added to reach the workspace floor.

### Scenario Inputs

For the first implementation, the simulator can keep using the existing user count as the billable seat count.

That keeps the change small and avoids blocking on free-user vs paid-user separation. The design should leave room for a future `billable_users` or `paid_seats` input once the pricing model needs it.

## Validation Rules

The validator should enforce the following:

- `pricing.mode` must be one of `tiered` or `hybrid`.
- `tiered` configs must use `plans` and not `contract`.
- `hybrid` configs must use `contract` and must not use `plans`.
- `base_platform_fee` must be `>= 0`.
- `per_seat_fee` must be `>= 0`.
- `workspace_minimum`, if set, must be `>= 0`.
- `included_seats` must be `>= 0`.
- `pilot.duration_months` must be `>= 1`.
- `pilot.setup_fee` and `pilot.monthly_fee`, if present, must be `>= 0`.
- `workspace_minimum` should not be less than the recurring contract floor if that would make the model ambiguous.

Validation should fail fast on mixed modes so the engine never has to guess whether it is reading tiered or hybrid pricing.

## Reporting And Covenant Impact

The current `margin` covenant can stay intact because it already compares total revenue against total cost.

What needs to change is visibility:

- `margin.gross` should continue to reflect total recognized revenue.
- add `margin.recurring_gross` for steady-state economics
- add `revenue.breakdown` fields for base fee, seat revenue, minimum uplift, pilot monthly fee, and setup fee
- keep `cost_per_user` unchanged

Why this matters:

- A pilot setup fee can make the total margin look healthy even when the recurring contract is weak.
- A workspace minimum can hide underpriced seat economics unless the floor uplift is shown explicitly.

The output should therefore present both:

- total monthly/annual revenue
- recurring run-rate revenue

That gives product and GTM a clean answer to two different questions:

- what do we bill in this contract?
- what is the steady-state unit economics?

## Output Contract

The current output structs only expose total revenue and plan-level revenue. Hybrid pricing needs a breakdown.

Recommended output additions:

- `Revenue.Model`
- `Revenue.RecurringTotal`
- `Revenue.OneTimeTotal`
- `Revenue.MinimumUplift`
- `Revenue.Components[]`
- `Margin.RecurringGross`

CLI output should add one compact line for the contract summary, for example:

```text
Revenue: $2,750.00/month
  Base fee: $500.00
  Seat revenue: $1,250.00
  Minimum uplift: $1,000.00
```

JSON and Markdown should surface the same breakdown so downstream tools and humans see the same numbers.

## Examples

### Workspace Minimum Only

```yaml
pricing:
  mode: hybrid
  contract:
    base_platform_fee: 0
    per_seat_fee: 0
    workspace_minimum: 2500
```

Interpretation:

- recurring revenue is floored at `$2,500`
- seat count does not matter unless we later add overage logic

### Base Fee Plus Seats

```yaml
pricing:
  mode: hybrid
  contract:
    base_platform_fee: 750
    per_seat_fee: 30
    included_seats: 10
```

Interpretation:

- first 10 seats are included
- every additional seat adds `$30`
- recurring revenue is `base fee + paid seats`

### Paid Pilot Then Hybrid

```yaml
pricing:
  mode: hybrid
  contract:
    base_platform_fee: 500
    per_seat_fee: 25
    workspace_minimum: 2000
    pilot:
      setup_fee: 5000
      monthly_fee: 1500
      duration_months: 3
```

Interpretation:

- months 1-3 bill the pilot
- month 1 includes the setup fee
- month 4 onward bills the steady-state hybrid contract

## Rollout Phases

### Phase 1: Schema And Engine Seams

- add explicit pricing mode selection
- add contract parsing and validation
- keep tiered behavior unchanged

### Phase 2: Hybrid Revenue Calculation

- implement base fee + per-seat + workspace minimum
- implement pilot billing and phase transition
- add revenue breakdown fields

### Phase 3: Output And Reporting

- update CLI summary
- update JSON schema
- update Markdown summary
- add recurring-vs-total visibility for covenants and dashboards

### Phase 4: Starter Fixtures And Docs

- add a hybrid starter config
- add integration fixtures
- update user-facing docs and examples

## Files Likely To Change

These are the files that should move in the implementation phase:

- `internal/config/parser.go`
- `internal/config/validator.go`
- `internal/config/parser_test.go`
- `internal/config/validator_test.go`
- `internal/pricing/calculator.go`
- `internal/pricing/calculator_test.go`
- `internal/pricing/margin.go`
- `internal/pricing/margin_test.go`
- `internal/output/types.go`
- `internal/output/cli.go`
- `internal/output/json.go`
- `internal/output/markdown.go`
- `internal/output/cli_test.go`
- `internal/output/json_test.go`
- `internal/output/markdown_test.go`
- `cmd/simulate.go`
- `cmd/simulate_test.go`
- `cmd/init.go`
- `test/fixtures/*.yml`
- `test/integration/simulate_test.go`
- `README.md`
- `docs/QUICK_START.md`
- `docs/HOW_IT_WORKS.md`

## Acceptance Criteria

The hybrid pricing work is ready when:

- legacy tiered configs still pass unchanged
- a workspace minimum config validates and simulates correctly
- a base-fee-plus-seat config reports recurring revenue and minimum uplift clearly
- a paid pilot config reports setup fee and pilot months separately from recurring revenue
- CLI, JSON, and Markdown all show the same core numbers
- full test suite passes

