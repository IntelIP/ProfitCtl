# Local Cost Ledger v1

Status: PCTL-19 candidate

Schema version: `profitctl.ledger/v1`

The local cost ledger stores normalized `profitctl.cost/v1` drivers and
observations. It is a local JSON file, not a hosted service or a live provider
integration.

Each observation record preserves its contract-level evidence plus a
`raw_source` object:

- `artifact_identity`: the source artifact identity from the accepted fixture;
- `sha256`: SHA-256 of the exact input file read during ingest.

The observation ID is the idempotency key. Re-ingesting identical input returns
`already_present`; a changed payload using the same ID fails instead of
silently overwriting the prior record.

## Commands

```bash
profitctl ledger ingest \
  --ledger /tmp/profitctl-ledger.json \
  --input test/fixtures/cost_contract/v1/upstash_idle_polling.json

profitctl ledger query --ledger /tmp/profitctl-ledger.json
```

Both commands emit stable JSON. Query output includes the referenced normalized
drivers, is ordered by observation ID, and accepts exact `--observation-id` and
`--workload` filters.

This slice only accepts the deterministic PCTL-18 fixture shape. It does not
contact providers, read credentials, create background work, or make pricing
claims beyond the embedded evidence contract.
