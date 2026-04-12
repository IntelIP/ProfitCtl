# MVP Pricing Release Guide

This release packages the MVP pricing work that landed in `main`:

- `pricing.mode: mix`
- `pricing.mode: hybrid`
- payment-fee modeling with monthly vs annual mix
- booked vs operating economics
- `compare` workflow
- calibration ingest via inline config or `calibration_file`
- explicit `billable_users` support for seat-based hybrid contracts
- recurring covenant fields such as `operating_margin`

## Recommended Smoke Tests

```bash
go test ./...
go run . validate -f examples/mix_profit.yml
go run . validate -f examples/hybrid_profit.yml
go run . simulate -f examples/hybrid_profit.yml
go run . compare examples/hybrid_steady_profit.yml examples/hybrid_profit.yml
go run . calibrate --input examples/calibration_exports/hybrid_profit_calibration.csv
```

Use `GOCACHE` and `GOTMPDIR` overrides in restricted environments.

## Release Checklist

1. Confirm `main` is green locally with `go test ./...`.
2. Run the smoke tests above.
3. Review benchmark scenario outputs in `benchmark_scenarios/README.md`.
4. Tag the release and publish binaries.
5. Point downstream docs or GTM collateral to:
   - `compare` for pricing review
   - `calibrate` plus `calibration_file` for assumption grounding
   - `operating_margin` covenants for contract safety checks

## Suggested Release Notes

`ProfitCtl` now supports open-core pricing comparison and calibration workflows end to end. Teams can model tiered, mix, and hybrid contracts; separate booked from operating economics; ingest normalized calibration exports; and enforce recurring-margin guardrails in covenant checks.
