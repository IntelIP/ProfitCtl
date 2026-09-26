# Cost variance CLI

The commands operate on local JSON only; they make no provider calls and do not modify a ledger.

```sh
profitctl cost_diff --forecast forecast.json --actual actual.json --output diff.json
profitctl cost_explain --input diff.json --output explanation.json
```

Both commands print their JSON report to stdout even when `--output` (`-o`) writes a copy to disk. `cost_diff` accepts either one `profitctl.cost/v1` `CostObservation` per file or an array of observations for multiple periods. Use matching windows, currency, and workload dimensions. Every observation must carry quantity, unit-price, and total-cost evidence with a source identity. The forecast is the percentage denominator: a zero forecast cannot yield a meaningful percent change and is rejected, not silently replaced with zero. `cost_explain` reads the exact report produced by `cost_diff`; never substitute an invoice or a free-form summary.

The versioned output contract is the [cost variance JSON schema](../schemas/cost-variance/v1/schema.json). Reports retain source references (including artifact identity and captured time), units, windows, and any exact delivery receipts. A high-confidence attribution requires direct supporting evidence; medium confidence indicates partial support; low confidence indicates weak or synthetic support. An unknown attribution means evidence does not establish a cause, especially when an exact commit, pull-request, release, or deployment receipt is missing. A source label alone does not prove causation. A nonzero residual is the part of actual-minus-forecast cost not explained by named drivers; inspect it instead of distributing it across unsupported causes. Never treat unavailable actual cost as zero.

Invalid inputs (malformed JSON, missing evidence, unmatched observations, or zero percentage denominator) exit with code **2** and a diagnostic on stderr. File or stdout write failures exit with code **3**. No partial JSON is printed on invalid input; an `--output` path that aliases any input is rejected before reading or writing it.
