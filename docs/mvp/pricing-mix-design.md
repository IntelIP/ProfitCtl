# Pricing Mix Design

## Problem Statement

ProfitCtl currently treats pricing plans as cumulative user bands. In `internal/pricing/calculator.go`, each plan absorbs a range of users based on `limits.users`, and the last plan can absorb all remaining users. That works for classic tier ladders, but it is the wrong abstraction for product mix modeling.

For the IntelIP MVP, we need to model direct adoption shares instead:

- free vs paid conversion
- starter/pro adoption mix
- pricing experiments where the plan mix is the assumption, not the result of tier caps

The current tier model forces authors to encode plan mix as fake user ceilings. That makes the YAML harder to read, makes revenue assumptions less explicit, and blocks honest modeling of products that are not sold as cumulative bands.

## Proposed Schema

Add an explicit pricing mode on `pricing`:

```yaml
pricing:
  mode: mix
  plans:
    - name: Free
      price: 0
      share: 0.70
    - name: Starter
      price: 10
      share: 0.20
    - name: Pro
      price: 29
      share: 0.10
```

### Schema Rules

- `pricing.mode` is optional.
- Default mode is `tiered` to preserve current behavior.
- `mode: tiered` keeps the existing `limits.users` shape.
- `mode: mix` uses `share` instead of cumulative limits.
- `share` is a decimal fraction in the range `(0, 1]`.
- Shares must sum to `1.0` within a small epsilon.
- `price` remains required in both modes.
- `price: 0` is valid for free plans.
- `limits` should not be used in mix mode.

### Example: Tiered Backward-Compatible Config

```yaml
pricing:
  mode: tiered
  plans:
    - name: Basic
      price: 10
      limits:
        users: 1000
    - name: Pro
      price: 29
      limits:
        users: 5000
    - name: Enterprise
      price: 99
```

The final tier may omit `limits` in tiered mode to represent an unlimited final plan.

## Validation Rules

Validation should become mode-aware in `internal/config/validator.go`.

### Tiered Mode

- Require ascending `limits.users`.
- Preserve the existing cumulative allocation semantics.
- Allow the last plan to omit `limits` to represent an unlimited tail.

### Mix Mode

- Require `share` on every plan.
- Reject `limits` on mix plans.
- Reject negative or zero shares.
- Reject shares greater than `1`.
- Require the total share sum to equal `1.0` within tolerance, for example `1e-9`.
- Preserve plan order as the deterministic tie-breaker for rounding.

### Suggested Validation Errors

- `pricing mode mix requires share on each plan`
- `pricing mode mix does not support limits`
- `pricing shares must sum to 1.0`
- `only the last pricing plan may omit limits`

## Engine Changes

The revenue engine should stop encoding plan semantics directly inside `CalculateRevenue` and instead delegate to a shared allocation step.

### Recommended Shape

- Add `mode` to `internal/config.PricingConfig`.
- Add `Share *float64` or `Share float64` to `internal/config.PricingPlan`.
- Add a helper such as `AllocatePlanUsers(users int, plans []PricingPlan, mode string)`.
- Keep `CalculateRevenue` as the public entry point, but branch by mode.
- In mix mode, allocate users from shares using a deterministic largest-remainder method:
  - compute `users * share` for each plan
  - floor each allocation
  - distribute the remaining users to the plans with the largest fractional remainders
  - use config order to break ties

### Why Largest Remainder

Using naive rounding can leave the total off by one or more users. Largest-remainder allocation keeps the totals exact and deterministic, which matters for tests, output stability, and covenant checks.

### Revenue Result Shape

`internal/pricing.RevenueResult` can stay mostly the same, but mix mode should expose the share assumption alongside the allocated user count. The later implementation should consider adding:

- `Share` on `PlanRevenue`
- `PricingMode` on the top-level simulation output

That keeps CLI, JSON, and Markdown output interpretable without forcing the user to infer whether counts came from cumulative bands or direct mix shares.

## Output and Reporting

The current output layer already renders plan-level revenue, margin, and covenant results. Mix mode needs additional clarity, not a different reporting model.

### CLI

- Keep the summary and covenant behavior the same.
- For verbose or standard output, add plan mix share percentages next to each plan row.
- Make the allocation method explicit, e.g. `mode=mix`.

### JSON

`internal/output/json.go` should include:

- pricing mode
- plan share for each `revenue.by_plan` item

The existing fields `plan_name`, `price`, `users`, and `revenue` should remain unchanged so downstream consumers do not break.

### Markdown

`internal/output/markdown.go` should render a plan mix table when mix mode is active, including:

- plan name
- share
- allocated users
- price
- revenue

## Migration Path

This should be a backward-compatible rollout.

1. Default `pricing.mode` to `tiered` when absent.
2. Keep existing YAML files valid without changes.
3. Add a new mix example config and document it in the product docs.
4. Update `cmd/init.go` to keep the starter config on tiered mode, but make the new mode visible in docs and examples.
5. Add a release note or changelog entry explaining that mix mode is opt-in.

The goal is to let current users continue using cumulative tiers while new MVP scenarios can move to explicit adoption shares.

## File Touchpoints

These are the files the implementation should touch later:

- `internal/config/parser.go`
- `internal/config/validator.go`
- `internal/config/parser_test.go`
- `internal/config/validator_test.go`
- `internal/pricing/calculator.go`
- `internal/pricing/calculator_test.go`
- `internal/output/types.go`
- `internal/output/json.go`
- `internal/output/json_test.go`
- `internal/output/markdown.go`
- `internal/output/markdown_test.go`
- `cmd/simulate.go`
- `cmd/init.go`
- `test/integration/simulate_test.go`
- `test/fixtures/valid_config.yml`
- `test/fixtures/no_pricing_config.yml`
- `test/fixtures` new mix-mode fixture

## Test Plan

The implementation should add coverage for:

- parsing a mix-mode config
- rejecting mix mode without shares
- rejecting mix mode with cumulative limits
- rejecting shares that do not sum to `1.0`
- deterministic user allocation for fractional shares
- CLI output for mix mode
- JSON output for mix mode
- markdown output for mix mode
- one integration test that runs `profitctl simulate` with a mix fixture

## Recommended Order

1. Land parser and validator support for `pricing.mode` and `share`.
2. Add the allocation helper and branch `CalculateRevenue` by mode.
3. Update output layers to surface mode and share details.
4. Add fixtures and integration tests.
5. Update the starter template and docs.

