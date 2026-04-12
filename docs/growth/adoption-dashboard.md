# Adoption Dashboard

This is the lightweight weekly dashboard for ProfitCtl adoption.

It is intentionally manual. Until we have product telemetry, the dashboard should be easy enough to update in 10 minutes from release stats, Linear, and evaluator notes.

## Dashboard Purpose

Track whether ProfitCtl is becoming easier to:

- install
- activate
- trust
- calibrate
- recommend to another evaluator

## Dashboard Schema

Use one row per week.

| Week Of | Release | New Targets | Contacted | Replied | Installed | Activated | Calibrated | Design Partners | Release Downloads | Homebrew Installs | Key Friction Theme | Next Fix |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|  |  |  |  |  |  |  |  |  |  |  |  |  |

## Metric Definitions

- `New Targets`: new companies added to the first 10-target list
- `Contacted`: targets that received first touch or follow-up
- `Replied`: targets that responded with a real evaluation signal
- `Installed`: targets that completed a local or Homebrew install
- `Activated`: targets that ran one compare or benchmark pair
- `Calibrated`: targets that shared real inputs or scenario details
- `Design Partners`: targets committed to a follow-up loop
- `Release Downloads`: release page downloads for the current version
- `Homebrew Installs`: installs observed through the tap or manual confirmations
- `Key Friction Theme`: the main blocker that repeated during the week
- `Next Fix`: the one change we should make before the next outreach cycle

## Manual Tracking Process

Update the dashboard every Friday after the weekly evaluator review.

Fastest path:

```bash
bash scripts/growth/scaffold-weekly-review.sh --owner "<name>"
```

That script scaffolds the weekly review markdown and prints a dashboard row with the current release version and GitHub release download count when `gh` is available.

1. Pull the current release version and download trend from GitHub Releases.
2. Check the Linear issues for the active design-partner pipeline.
3. Count completed installs, activations, calibrations, and commitments from evaluator notes.
4. Record the dominant friction theme from the week.
5. Pick one next fix only.

## What To Source From

- GitHub Releases for release version and download trend
- Homebrew tap activity for install signal
- Linear for target states and execution follow-ups
- evaluator notes for install, activation, and calibration outcomes
- benchmark reports for evidence that the product proof holds

## Operating Rule

Do not wait for perfect tracking.

If a field is missing, leave it blank and note why. The value is in weekly movement, not in pretending we have full telemetry.
