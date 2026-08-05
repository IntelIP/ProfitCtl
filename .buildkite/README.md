# ProfitCtl self-hosted Buildkite CI

This pipeline is a manual, exact-commit validation lane for the current-project
self-hosted queue. Buildkite coordinates the job; a locally managed agent runs
the substantive checks inside a pinned multi-architecture Go container. The
runner installs only the small release-tool set ProfitCtl needs (`curl`, `zip`,
and CA certificates) before validation.

The lane mirrors the portable `verify-go` and `verify-install-smoke` GitHub
jobs. GitHub Actions remains authoritative for the secret-backed AppSec upload
and policy gate. No AppSec, provider, billing, publication, deployment, or
production credentials are attached to this Buildkite lane.

Every run requires an operator lease in the stored Buildkite bootstrap. Manual
retries are disabled, the shared current-project concurrency limit is one, and
the queue remains paused outside an explicitly approved run.

## Local checks

```bash
bk pipeline validate --file .buildkite/pipeline.yml
BUILDKITE_COMMIT="$(git rev-parse HEAD)" bash .buildkite/scripts/self-hosted-validation.sh
```
