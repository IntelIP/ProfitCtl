# ProfitCtl Bug Backlog From IntelIP Modeling

This note captures the concrete ProfitCtl bugs and modeling gaps exposed while using ProfitCtl to price IntelIP.

The goal is to separate:

- real bugs or inconsistencies in ProfitCtl behavior
- meaningful modeling limitations that need roadmap work

## 1. Monthly margin summary bug

### Status

Fixed locally, not yet committed or released.

### Problem

`profitctl simulate` was comparing monthly revenue against costs calculated on a 12-month horizon in the top-level summary path.

That produced misleading margin and covenant outputs for otherwise valid monthly pricing scenarios.

### Evidence

- implementation fix is currently local in `cmd/simulate.go`
- regression coverage is currently local in `test/fixtures/monthly_margin_regression.yml` and `test/integration/simulate_test.go`

### Why it matters

This is a true correctness bug. It can make a scenario look unprofitable even when the monthly economics are actually healthy.

## 2. Unlimited-plan inconsistency between calculator and config validation

### Status

Open bug.

### Problem

The revenue calculator explicitly supports a plan without limits and treats it as applying to all remaining users:

- `internal/pricing/calculator.go`

But the parsed config schema and validator still require `limits` on every plan:

- `internal/config/parser.go`
- `internal/config/validator.go`

### Why it matters

This means the CLI path rejects a model shape that the revenue engine itself already knows how to handle.

Practical consequence:

- users cannot model a real unlimited plan cleanly through YAML
- scenario authors are forced to use large fake user caps instead of an honest unlimited top tier

## 3. Plan mix support landed, but scenario calibration is still open

### Status

Fixed in the engine, still open as a modeling calibration task.

### What changed

ProfitCtl now supports:

- `pricing.mode: cumulative`
- `pricing.mode: mix`
- per-plan `share` allocation with validation and deterministic rounding

That removes the biggest prior pricing limitation in the IntelIP scenario pack.

### What still matters

The remaining issue is not engine capability. It is calibration:

- what free / starter / pro mix is realistic for IntelIP
- how fast that mix shifts over time
- how design-partner and paid-pilot motions should be represented

## 4. Workspace minimum and seat-plus-workspace pricing support landed

### Status

Fixed in the engine, still open as a scenario-calibration task.

### What changed

ProfitCtl now supports a native `workspace_hybrid` pricing mode with:

- `workspace.average_users_per_workspace`
- per-plan `share`
- optional `workspace_minimum`

That means revenue no longer has to be approximated as pure `price * users` for workspace-oriented packaging.

### Why it still matters

The remaining work is calibration, not engine support:

- what workspace minimum should IntelIP actually charge
- what average active seats per workspace are realistic
- when paid-pilot pricing should use hybrid logic versus pure minimum pricing

## 5. Payment-fee modeling is partially fixed, but still incomplete

### Status

Partially fixed in the engine, still open as a calibration and package-semantics gap.

### Problem

ProfitCtl now supports `variable_costs[].user_scope: paid_users`, so scenario authors can scope billing and payment processing to monetized users instead of all active users.

It still does not distinguish:

- monthly vs annual buyers
- fee burden by plan
- billing percentage fees versus pure card-processing fees

### Why it matters

This forces blended assumptions that are directionally useful but too rough for launch-decision confidence.

## 6. Real usage calibration still depends on external telemetry or invoices

### Status

Open calibration gap.

### Problem

Several scenario lines remain synthetic:

- model-routing cost
- backend compute
- integration-action volume
- support reserve

### Why it matters

This is not a ProfitCtl correctness bug by itself, but it limits the confidence of any pricing recommendation produced from the tool.

## Recommended follow-up order

1. Commit and release the monthly-margin bug fix.
2. Fix the unlimited-plan parser/validator inconsistency.
3. Improve payment-fee modeling by separating paid-user revenue from free-user volume.
4. Add a calibration workflow that can ingest real cost and usage inputs from production systems.
5. Add richer package semantics such as included seats and annual billing by plan.
