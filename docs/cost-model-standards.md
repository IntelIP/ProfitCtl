# Cost Model Standards

## Purpose

These standards are the judge criteria for cost-aware agent work. A scenario, assessment, or recommendation can be useful while still being approximate, but the output must make assumptions, provenance, and confidence visible.

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
go run scripts/judge_cost_standards.go path/to/assessment.json
go run scripts/judge_cost_standards.go path/to/recommendation.md
```

The standards judge checks artifact quality, not provider-price truth. It should fail when:

- the scenario does not parse or validate
- cost line provenance is missing
- source type or confidence is invalid
- a template claims high confidence
- non-delivery cost lacks allocation
- margin or cost-per-user covenant is missing
- an assessment receipt lacks an official-domain URL, request ID, capture time, or source excerpt
- an assessment price does not match its exact source excerpt
- assessment math, provider coverage, request counts, inference labels, or reported total are inconsistent
- a recommendation artifact omits recommendation, assumptions, fixed cost, variable drivers, margin, p95 margin, cost per user, covenant status, cheaper alternative, provenance, or confidence
- a recommendation artifact makes precise cost or margin claims without provenance and confidence
- a recommendation artifact uses high-certainty language without telemetry or invoice evidence

The judge should warn when:

- all costs are template-derived
- no actual telemetry or invoice inputs are present
- provider catalog entries lack source URLs
- a recommendation artifact does not name a supported source type

## Source-Backed Assessment Standards

`profitctl.assessment/v1` artifacts must include:

- the selected model, analyzed-file count, two model requests, and bounded Exa request count
- every code-backed provider with an exact identifying code excerpt and a matching pricing receipt
- HTTPS receipt URLs matching the declared official domain
- Exa request ID, RFC3339 capture time, title, and source highlights
- repository-derived or explicitly inferred assumptions
- one source-backed, mathematically valid target-code cost line per detected provider
- separate OpenRouter input, OpenRouter output, and Exa assessment-runtime costs
- a stable recommendation and a total equal to the sum of monthly cost lines

The fixture-backed launch check performs no live OpenRouter or Exa call:

```bash
go run scripts/judge_cost_standards.go test/fixtures/assessment_valid.json
```

## Agent Recommendation Standards

Cost-aware agent answers should be short but evidence-bearing. Each answer must include:

- recommendation
- assumptions
- monthly fixed cost
- variable cost drivers
- gross margin, p95 margin, and cost per user
- covenant status
- cheaper viable alternative or reason the cheaper option is not viable
- source provenance and confidence

Example validation:

```bash
go run scripts/judge_cost_standards.go test/fixtures/agent_recommendation_valid.md
```

## Calibration Path

Accuracy improves as sources move from estimates to actuals:

1. `template`: first-pass planning
2. `repo_detected`: architecture-aware planning
3. `user_supplied`: business-context planning
4. `provider_catalog`: sourced planning defaults
5. `telemetry`: measured usage
6. `invoice`: actual spend/revenue

The product goal is not perfect prediction. It is disciplined decision quality: explain assumptions, test stress cases, and update scenarios when actuals arrive.
