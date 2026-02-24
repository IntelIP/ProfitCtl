# How ProfitCtl Works

`profitctl simulate` runs a deterministic orchestration pipeline:

1. Parse and validate `profit.yml`.
2. Compute fixed and variable costs.
3. Run scale simulation across growth steps.
4. Run Monte Carlo stress testing for p95/p99 metrics.
5. Compute revenue and margins.
6. Validate covenants.
7. Render CLI/JSON/Markdown output.
8. Return deterministic exit code.

`profitctl detect` runs an analysis pipeline:

1. Recursively collect config-relevant files.
2. Build a structured detection prompt.
3. Call OpenRouter LLM provider.
4. Parse and validate JSON response.
5. Emit JSON report to stdout or `--out` file.
