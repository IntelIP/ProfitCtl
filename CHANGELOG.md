# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Woodpecker/Hostinger CI/CD pipeline and release/deploy scripts.
- `profitctl detect` command for scanner/LLM analysis output.
- OSS baseline docs and governance files.
- Published-asset release smoke verification.
- Design-partner issue routing and maintainer workflow guidance.
- GitHub Actions PR/main verification workflow with stable `verify-go` and `verify-install-smoke` checks.
- Homebrew formula and tap-publish script for the public release channel.
- Committed benchmark comparison reports for open-core and hybrid pricing scenarios.
- Release SBOM generation, detached Cosign signatures, and published verification key assets.

### Changed
- Canonical module/repository identity aligned to `IntelIP/ProfitCtl`.
- CLI exit behavior standardized.
- Installer checksum matching is more robust across common checksum formats.
- Simulation benchmarks are now discoverable through `go test -bench`.
- Public installer defaults now use GitHub Releases as the canonical source, with the Hostinger mirror available as an explicit override.
- Open-core packaging docs now define who the product is for, what stays free, and what the first paid layer should cover.
- Install docs now include Homebrew and the Quick Start includes concrete output snippets.
- Release docs now include explicit archive, SBOM, and checksum verification steps.
