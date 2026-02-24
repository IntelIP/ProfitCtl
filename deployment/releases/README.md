# ProfitCtl Release Publishing (VPS)

## Script

- `publish-to-vps.sh <tag>`

## Required Environment

- `VPS_HOST`
- `VPS_USER`
- `VPS_SSH_PRIVATE_KEY`
- `VPS_RELEASES_DIR` (default `/opt/profitctl/releases`)
- `DOWNLOAD_BASE_URL` (default `https://downloads.intelip.co/profitctl`)

## Resulting Layout

- `/opt/profitctl/releases/<tag>/...artifacts...`
- `/opt/profitctl/current` symlink
- `/opt/profitctl/index.json`
