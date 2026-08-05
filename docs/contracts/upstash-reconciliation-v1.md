# Upstash Reconciliation v1

Status: PCTL-20 candidate

`profitctl upstash reconcile` is a local, provider-free vertical slice for
explaining Redis polling cost from deterministic synthetic fixtures. It models
polling interval, worker replicas, active time, commands, misses, and a
declared command rate. It does not call Upstash, use credentials, change a
provider plan, spend money, or make a current-price claim.

## Command

```bash
profitctl upstash reconcile \
  --input test/fixtures/upstash/v1/idle_polling.json \
  --ledger /tmp/profitctl-ledger.json \
  --output /tmp/upstash-reconciliation.json
```

The command emits the same stable JSON to stdout and, when `--output` is set,
to the local artifact path. It also normalizes the derived observation into the
PCTL-19 local ledger. Retrying identical input returns `already_present`;
changed data with the same observation ID fails closed.

## Fixture cases

- `normal_load.json`: modeled active workload.
- `idle_polling.json`: declared `idle_polling` driver with materially higher
  predicted command cost.
- `missing_billing.json`: billing evidence is `unavailable`, never `$0`.
- `duplicate_export.json`: identical telemetry records are counted once;
  conflicting records with the same ID fail.

Every telemetry and billing record must exactly match the fixture billing
window. The reconciliation preserves artifact identity, capture time, and the
exact input SHA-256 in `ledger_normalization.raw_source`.

## Interpretation boundary

`dominant_operational_driver` is the fixture's declared classification. The
artifact explicitly says that it does not infer causal proof from temporal
correlation. `derived_cost` is observed command count multiplied by the
declared synthetic rate; it is not an invoice claim. Billing evidence is only
reported when the synthetic fixture supplies it.

The output schema is
[`schemas/upstash-reconciliation/v1/reconciliation.schema.json`](../../schemas/upstash-reconciliation/v1/reconciliation.schema.json).
