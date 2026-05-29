# Woodpecker + Hostinger Setup for ProfitCtl

This runbook onboards `IntelIP/ProfitCtl` into your existing Woodpecker infrastructure.

## Pipeline Model

`/.woodpecker.yml` defines:

- Tag release workflow (`pool=shared-kvm`): semver tag release + VPS publish

PR and main verification now run in GitHub Actions:

- `verify-go`
- `verify-install-smoke`
- `security-scan` with AppSec upload and PR gate evaluation

## Required Secrets (Doppler: `profitctl` / `prd_ci_woodpecker`)

- `GITHUB_TOKEN_RELEASE`
- `COSIGN_PRIVATE_KEY`
- `COSIGN_PASSWORD`
- `VPS_HOST`
- `VPS_USER`
- `VPS_SSH_PRIVATE_KEY`
- `VPS_RELEASES_DIR=/opt/profitctl/releases`
- `DOWNLOAD_BASE_URL=https://downloads.intelip.co/profitctl`

## Required Secrets (Woodpecker)

- Repo secret name: `doppler_token`
- Secret value: Doppler service token for `profitctl` / `prd_ci_woodpecker`
- Events: include `tag`
- Image filters: leave empty (`[]`) so command steps can consume it
- Repo trust: set `IntelIP/ProfitCtl` as trusted in Woodpecker so `from_secret` works in command steps

## Required Secrets (GitHub Actions)

- Repo secret name: `APPSEC_API_TOKEN`
- Secret value: bearer token used by `appsec-mvp` ingestion and summary endpoints
- Repo secret name: `APPSEC_API_URL`
- Secret value: reachable base URL for the deployed `appsec-mvp` API
- Optional repo variable name: `APPSEC_STRICT_API`
- Rollout default: leave unset or set to `false` so AppSec transport failures do not block PRs while `/.appsec.yml` is still `report_only`
- Post-rollout: set to `true` when the service is stable and you want CI to fail closed on AppSec API errors

## Required AppSec MVP Runtime Configuration

`appsec-mvp` must already be deployed and reachable from the Woodpecker runner network.

- `GITHUB_APP_ID`
- `GITHUB_APP_PRIVATE_KEY_PEM` or `GITHUB_APP_PRIVATE_KEY_PATH`
- `GITHUB_WEBHOOK_SECRET`
- `APPSEC_API_TOKEN`
- DB migration `0004_github_webhook_deliveries.sql` applied

Install the AppSec GitHub App on `IntelIP/ProfitCtl` with:

- `checks:write`
- `pull_requests:write`
- `contents:read`
- `metadata:read`

Subscribe the App to:

- `pull_request`
- `installation`
- `installation_repositories`

## Artifact Layout on VPS

- `/opt/profitctl/releases/<tag>/...`
- `/opt/profitctl/current -> /opt/profitctl/releases/<tag>`
- `/opt/profitctl/index.json`

## Release Trigger

Tag push only (semver):

- `vMAJOR.MINOR.PATCH`

## Webhooks

- Keep existing Composio webhook unchanged.
- Activate repository in Woodpecker to add/refresh Woodpecker webhook.
- Add/refresh the AppSec GitHub App installation webhook on `IntelIP/ProfitCtl`.

## GitHub Actions

GitHub Actions is enabled for repository-facing verification (`verify-go` and `verify-install-smoke`), while Woodpecker remains the authoritative tag release pipeline.

```bash
gh api repos/IntelIP/ProfitCtl/actions/permissions
```

## Rollback

To rollback published "current":

```bash
ln -sfn /opt/profitctl/releases/<previous-tag> /opt/profitctl/current
```
