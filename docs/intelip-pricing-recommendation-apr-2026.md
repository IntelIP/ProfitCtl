# IntelIP Pricing Recommendation (Apr 2026)

This note turns the merged IntelIP scenario pack into an executable launch recommendation.

## What changed

Before running the scenarios, ProfitCtl needed one follow-up fix: `pricing.mode: mix` now accepts `limits.users` as reference metadata instead of rejecting those files during validation. That keeps the scenario pack aligned with its documented modeling posture.

## Scenarios run

All commands were run from the ProfitCtl repo on Apr 13, 2026:

```bash
go run . simulate -f benchmark_scenarios/intelip_ops_conservative.yml --quiet
go run . simulate -f benchmark_scenarios/intelip_ops_target.yml --quiet
go run . simulate -f benchmark_scenarios/intelip_ops_stress.yml --quiet
go run . simulate -f benchmark_scenarios/intelip_rollout_paid_pilot.yml --quiet
go run . simulate -f benchmark_scenarios/intelip_rollout_tight_free.yml --quiet
go run . simulate -f benchmark_scenarios/intelip_rollout_workspace_minimum_proxy.yml --quiet
```

## Results

| Scenario | Users | Revenue | Delivery Cost | Full Cost | Full Margin | Status |
|---|---:|---:|---:|---:|---:|---|
| Conservative mix | 30 | $700 | $92.10 | $402.10 | 42.56% | Failed |
| Target mix | 90 | $3,661 | $164.35 | $599.35 | 83.63% | Passed |
| Stress mix | 180 | $11,655 | $325.91 | $1,000.91 | 91.41% | Passed |
| Paid pilot | 15 | $2,685 | $97.05 | $407.05 | 84.84% | Passed |
| Tight free tier | 30 | $1,834 | $108.10 | $418.10 | 77.20% | Passed |
| Workspace minimum hybrid | 20 | $1,790 | $102.07 | $412.07 | 76.98% | Passed |

Notes:

- The conservative mix is still the most important launch signal because it approximates the first real cohort rather than a later, healthier mix.
- The stress mix remains useful as an upper-bound sanity check, but it is not the launch decision case.
- The paid-pilot, tight-free, and workspace-minimum shapes all clear the modeled thresholds. The difference is packaging and go-to-market complexity, not raw survivability.

## Recommendation

Use a paid pilot as the default first launch motion.

Why:

- It is the strongest model outcome at the smallest active-user count.
- It protects IntelIP from the exact failure zone exposed by the conservative free-heavy mix.
- It aligns with the current onboarding reality: IntelIP still carries meaningful adoption overhead, and that cost is what erodes the early free-heavy scenario.
- It gives IntelIP the fastest clean path to collect real pricing, billing, and onboarding calibration data.

## Backup recommendation

If IntelIP wants broader top-of-funnel learning instead of the safest launch shape, use the tight free tier as the fallback.

Why:

- It preserves self-serve evaluation.
- It materially improves the low-scale economics versus the current free-heavy mix.
- It is simpler to explain than a workspace-minimum package during the first rollout.

## What not to do

Do not launch with the current free-heavy conservative mix as the default posture.

Reason:

- The scenario fails the margin covenant at the earliest meaningful cohort.
- Delivery cost is healthy, but adoption overhead dominates the full-economics view.
- That means the real problem is not runtime COGS. The problem is allowing too much unpaid evaluation while onboarding is still high-touch.

## When to use workspace minimum

Workspace minimum remains a viable second-stage package, not the best first-stage package.

Why:

- It is economically safe in the model.
- It captures team-level value better than seat-only pricing.
- But it adds packaging complexity before IntelIP has enough live calibration data to know whether the extra pricing surface is worth it.

## Recommended launch sequence

1. Start with a paid pilot.
2. Define the pilot exit criteria into Starter or Pro packaging.
3. Keep the tight free tier ready as the fallback if top-of-funnel learning becomes more important than early monetization certainty.
4. Revisit workspace minimum after the first real design-partner and billing observations are in hand.

## Immediate product and GTM follow-up

1. Write the paid-pilot package in user-facing terms:
   - onboarding scope
   - included workflows
   - expected time-to-value
   - conversion target into recurring packaging
2. Record the first real calibration inputs:
   - paid invoice count
   - observed Stripe fee burden
   - onboarding/support hours
   - workspace size and active-seat shape
3. Re-run the scenario pack after the first real cohort so the recommendation moves from modeled launch guidance to measured pricing guidance.
