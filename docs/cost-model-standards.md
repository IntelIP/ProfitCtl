# Cost Model Standards

## Purpose

These standards are the judge criteria for cost-aware agent work. A scenario can be useful while still being approximate, but the output must make assumptions, provenance, and confidence visible.

## Required Scenario Standards

- Every fixed and variable cost should include `source.type` and `source.confidence`.
- `source.type` must be one of `template`, `user_supplied`, `repo_detected`, `telemetry`, `invoice`, or `provider_catalog`.
- `source.confidence` must be one of `low`, `medium`, or `high`.
- Template assumptions must not use `high` confidence.
- `provider_catalog` assumptions should include `url` or `note`.
- `telemetry` and `invoice` assumptions should include `captured_at` or `note`.
- Productization and adoption costs should be modeled separately from delivery costs using `economics_layer`.
- Non-delivery costs should include an `allocation` rule when the allocation is not obvious.
- Scenarios should define covenants for margin and cost per user.
- Agent recommendations should report assumptions, margin, p95 margin, cost per user, and covenant status.

## Accuracy Judge

Run the local judge from the repository root:

```bash
go run scripts/judge_cost_standards.go
go run scripts/judge_cost_standards.go path/to/scenario.yml
go run scripts/judge_cost_standards.go path/to/scenario-directory
```

The standards judge checks scenario quality, not provider-price truth. It should fail when:

- the scenario does not parse or validate
- cost line provenance is missing
- source type or confidence is invalid
- a template claims high confidence
- non-delivery cost lacks allocation
- margin or cost-per-user covenant is missing

The judge should warn when:

- all costs are template-derived
- no actual telemetry or invoice inputs are present
- provider catalog entries lack source URLs

## Calibration Path

Accuracy improves as sources move from estimates to actuals:

1. `template`: first-pass planning
2. `repo_detected`: architecture-aware planning
3. `user_supplied`: business-context planning
4. `provider_catalog`: sourced planning defaults
5. `telemetry`: measured usage
6. `invoice`: actual spend/revenue

The product goal is not perfect prediction. It is disciplined decision quality: explain assumptions, test stress cases, and update scenarios when actuals arrive.
