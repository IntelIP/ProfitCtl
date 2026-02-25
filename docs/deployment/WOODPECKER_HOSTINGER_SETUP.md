# Woodpecker + Hostinger Setup for ProfitCtl

This runbook onboards `IntelIP/ProfitCtl` into your existing Woodpecker infrastructure.

## Pipeline Model

`/.woodpecker.yml` defines:

- PR workflow (`pool=shared-kvm`): checks for `main`
- Main workflow (`pool=shared-kvm`): push checks for `main`
- Tag release workflow (`pool=shared-kvm`): semver tag release + VPS publish

## Required Secrets (Doppler: `profitctl` / `prd_ci_woodpecker`)

- `GITHUB_TOKEN_RELEASE`
- `VPS_HOST`
- `VPS_USER`
- `VPS_SSH_PRIVATE_KEY`
- `VPS_RELEASES_DIR=/opt/profitctl/releases`
- `DOWNLOAD_BASE_URL=https://downloads.intelip.co/profitctl`

## Required Secrets (Woodpecker)

- Repo secret name: `DOPPLER_TOKEN`
- Secret value: Doppler service token for `profitctl` / `prd_ci_woodpecker`
- Events: include `tag`
- Image filters: leave empty (`[]`) so command steps can consume it
- Repo trust: set `IntelIP/ProfitCtl` as trusted in Woodpecker so `from_secret` works in command steps

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

## GitHub Actions

Actions are disabled to enforce Woodpecker-only CI/CD.

```bash
printf '{"enabled":false}' | gh api repos/IntelIP/ProfitCtl/actions/permissions -X PUT --input -
```

## Rollback

To rollback published "current":

```bash
ln -sfn /opt/profitctl/releases/<previous-tag> /opt/profitctl/current
```
