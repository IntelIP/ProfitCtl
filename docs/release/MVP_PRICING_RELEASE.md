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
go test ./internal/simulation -run='^$' -bench=. -count=1
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
3. Run the simulation benchmark command so the core economics path has a fresh baseline.
4. Review benchmark scenario outputs in `benchmark_scenarios/README.md`.
5. Tag the release and publish binaries, SBOMs, and detached signatures.
6. Run the published-artifact smoke path:
   - `bash scripts/release/smoke-published-release.sh <tag>`
7. Verify the release contains:
   - `profitctl_<tag>_<os>_<arch>.spdx.json`
   - `profitctl_<tag>_<os>_<arch>.tar.gz.sig` or `profitctl_<tag>_<os>_<arch>.zip.sig`
   - `SHA256SUMS.sig`
   - `profitctl-release-cosign.pub`
8. Confirm the public verification instructions in `docs/INSTALL.md` still match the published assets.
9. Point downstream docs or GTM collateral to:
   - `compare` for pricing review
   - `calibrate` plus `calibration_file` for assumption grounding
   - `operating_margin` covenants for contract safety checks
10. Verify the exact public install path from the README against the new tag:
   - `curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | env -u PROFITCTL_DOWNLOAD_BASE_URL PROFITCTL_VERSION=<tag> bash`
11. Update and publish the Homebrew tap if `Formula/profitctl.rb` changed:
   - `bash scripts/release/publish-homebrew-tap.sh`

## Verification Model

The current release pipeline uses Cosign key-pair signing because the authoritative release pipeline runs in Woodpecker, not GitHub Actions. That means releases do not currently use GitHub OIDC keyless signing or Rekor-backed transparency bundles. Consumers verify with the committed and published `profitctl-release-cosign.pub` key instead.

This is a deliberate tradeoff:

- it fits the current release infrastructure
- it provides deterministic offline verification for archives, SBOMs, and checksum manifests
- it keeps the upgrade path open if release publishing moves to an OIDC-capable environment later

## Suggested Release Notes

`ProfitCtl` now supports open-core pricing comparison and calibration workflows end to end. Teams can model tiered, mix, and hybrid contracts; separate booked from operating economics; ingest normalized calibration exports; and enforce recurring-margin guardrails in covenant checks.
