# How ProfitCtl Works

`profitctl simulate` runs a deterministic orchestration pipeline:

1. Parse and validate `profit.yml`.
2. Load optional external calibration data via `calibration_file`.
3. Compute fixed and variable costs.
4. Run scale simulation across growth steps.
5. Run Monte Carlo stress testing for p95/p99 metrics.
6. Compute revenue and margins, including hybrid billable-seat overrides.
7. Validate gross and operating covenant fields.
8. Render CLI/JSON/Markdown output.
9. Return deterministic exit code.

`profitctl calibrate` runs a normalization pipeline:

1. Read a YAML, JSON, or CSV calibration artifact.
2. Normalize it into ProfitCtl's `CalibrationConfig` shape.
3. Emit YAML or JSON that can be referenced via `calibration_file`.

`profitctl detect` runs an analysis pipeline:

1. Recursively collect config-relevant files.
2. Build a structured detection prompt.
3. Call OpenRouter LLM provider.
4. Parse and validate JSON response.
5. Emit JSON report to stdout or `--out` file.
