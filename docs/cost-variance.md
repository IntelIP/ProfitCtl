# Local cost variance

`profitctl diff --input variance-input.json --output diff.json` compares cost observations and modeled driver pairs. `profitctl explain --input diff.json --output explanation.json` adds evidence-preserving, non-automatic review suggestions. Both print JSON to stdout; `--output` writes the same JSON to disk. Neither command calls a provider or changes a ledger. For totals-only comparisons, `profitctl diff --forecast forecast.json --actual actual.json` accepts two `profitctl.cost/v1` observations (no driver attribution; the difference remains residual).

Input is a `variance/v1.Input` JSON object:

```json
{
  "forecast": {"schema_version": "profitctl.cost/v1", "id": "...", "driver_ids": ["polls"], "window": {"start": "...", "end": "..."}, "quantity": {"value": 1, "unit": "command"}, "unit_price": {"amount": {"amount": 1, "currency": "USD"}, "per": {"value": 1, "unit": "command"}}, "total_cost": {"amount": 1, "currency": "USD"}, "dimensions": {"workload": "worker"}, "evidence": {"quantity": "...", "unit_price": "...", "total_cost": "..."}},
  "actual": null,
  "unavailable_reason": "invoice not delivered",
  "drivers": []
}
```

The abbreviated evidence placeholders above must be replaced by complete `cost/v1` evidence objects with source identity, capture date, confidence and rationale. For an available comparison, set `actual` to a validated observation and omit `unavailable_reason`. To explain drivers supply `drivers` as pairs of full `{ "forecast": CostDriver, "actual": CostDriver }` records referenced by both observations. Optional `exchange_rates` provide snapshot currency conversions and `receipts` provide exact observation-paired delivery evidence. Missing exchange rates fail; no live rates are fetched. The forecast and actual windows and workload must match.

The diff JSON matches [`result.schema.json`](../schemas/cost-variance/v1/result.schema.json): `profitctl.cost-variance/v1`, status, window, observation IDs, currency-denominated forecast/actual/absolute variance and residual, percentage variance, contributions, missing evidence, rounding tolerance, and timestamped source references. A zero forecast gives `percentage_variance: null`; unavailable actual gives null actual, variance and residual, not a fabricated zero. To explain an unavailable actual, `missing_evidence` must include a direct `actual cost: <unavailable_reason>` entry (for example, `actual cost: invoice not delivered`). Negation such as `actual cost: not invoice not delivered`, or a reason appearing only in another clause such as `actual cost: ledger pending; forecast invoice not delivered`, does not count. Contributions are ranked by absolute amount; the residual is never hidden. `explain` returns the same amounts under `variance_absolute` and `variance_percent`, plus ranked `driver_contributions` containing their evidence, confidence score, receipts, and review actions. Review actions are suggestions, never automatic remediation or asserted root causes. Without an exact receipt, delivery attribution is `unknown` and missing evidence is named. A source label alone cannot establish causation.

Malformed/invalid input and aliased output paths return exit code 2; output failures return code 3. No partial JSON is printed for invalid input. The exported `cmd.Diff(ctx, forecast, actual)` offers a totals-only comparison returning `*variancev1.Result`; `cmd.Explain(ctx, result)` returns `*cmd.ExplainReport`. For full attribution call `variancev1.Compare(variancev1.Input)` first. The peer engine's concrete exported type is `Result` (not `DiffReport`) and its comparison entry point is `Compare`.
