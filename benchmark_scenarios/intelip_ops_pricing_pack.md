# IntelIP Ops Pricing Scenario Pack

This scenario pack models IntelIP as an agentic business operations platform rather than a generic chatbot or coding assistant.

## Modeling intent

The goal is to test whether IntelIP can support an initial pricing posture that emphasizes:

- faster workflow connection
- lower operations setup burden
- reusable operational context
- recurring business execution

## Important modeling limitations

- ProfitCtl now supports `pricing.mode: mix`, but the scenario pack still uses simple plan shares rather than conversion funnels or time-based transitions.
- Every pricing plan can still carry a `limits.users` field for reference, but in mix mode revenue is allocated by `share`, not by cumulative tier bands.
- The scenarios below treat `users` as active platform users or seats participating in the pricing band model.
- Annual billing is not modeled separately yet.
- Stripe fee modeling is now grounded by live test-mode IntelIP charges, but the current scenario pack still approximates billing cost as a per-paid-user reserve rather than a per-invoice or per-workspace fee.
- These numbers are planning assumptions, not authoritative finance inputs.

## Scenario files

- `benchmark_scenarios/intelip_ops_conservative.yml`
- `benchmark_scenarios/intelip_ops_target.yml`
- `benchmark_scenarios/intelip_ops_stress.yml`
- `benchmark_scenarios/intelip_tooling_cost_inventory.md`

## Assumption categories

Fixed costs:

- Cloudflare Workers / DNS baseline
- Neon baseline
- Clerk baseline
- backend runtime baseline
- telemetry / product ops tooling as `productization`
- manual onboarding and support reserve as `adoption`

Variable costs:

- agent execution token spend as `delivery`
- workflow execution compute and memory operations as `delivery`
- integration sync and action overhead as `delivery`
- billing and payment processing as `delivery` with `user_scope: paid_users`
- no upload/storage baseline; the old R2-backed chat upload path is no longer part of the active IntelIP app architecture

Note:

- For normally or uniformly distributed usage inputs, the scenario pack sets `units_per_user: 1` and treats the distribution as the effective unit count per user because ProfitCtl currently multiplies both values together.
- Recent Stripe test-mode evidence confirms the live IntelIP Starter Monthly price is `$49` and that observed processing fees on sampled Starter charges are `$1.72` per invoice. Those fee observations should now be treated as the anchor for Stripe reserve assumptions until production billing data is available.

## Packaging assumptions

The modeled package shape is:

- small free tier to reduce evaluation friction
- Starter tier for solo operators and very small teams
- Pro tier for higher-touch operational usage and collaboration

The current pricing hypothesis used across the revised scenarios is:

- Free: trial / evaluation only
- Starter: $49/month
- Pro: $119/month

Scenarios using `mode: mix` currently encode first-pass adoption assumptions instead of cumulative tier occupancy:

- Conservative: `Free 65% / Starter 25% / Pro 10%`
- Target: `Free 45% / Starter 35% / Pro 20%`
- Stress: `Free 25% / Starter 35% / Pro 40%`
- Tight free: `Free 15% / Starter 60% / Pro 25%`

This is not a final launch recommendation. It is a working hypothesis chosen to pressure-test whether IntelIP can afford a free entry point while still supporting a high-touch onboarding motion.

This is meant to support pricing exploration for:

- `INT-148` pricing packages and rollout hypotheses
- `INT-149` unit economics and launch scenarios in ProfitCtl CLI
- `INT-154` market benchmark alignment
- `INT-155` onboarding and activation design

## Run commands

```bash
go run . simulate -f benchmark_scenarios/intelip_ops_conservative.yml
go run . simulate -f benchmark_scenarios/intelip_ops_target.yml
go run . simulate -f benchmark_scenarios/intelip_ops_stress.yml
```

## First local results

Measured on Apr 13, 2026 after adding `pricing.mode: mix`, economics-layer reporting, `user_scope: paid_users` for payment fees, removing the stale upload/storage line, and recalibrating the IntelIP scenario pack against the current stack:

- Conservative: `FAILED margin=42.56% cost_per_user=13.40 violations=1`
- Target: `PASSED margin=83.63% cost_per_user=6.66 violations=0`
- Stress: `PASSED margin=91.41% cost_per_user=5.56 violations=0`

Interpretation of the current mix-based model:

- the conservative scenario remains the most important signal because it approximates the first real paid cohort
- the conservative scenario now fails once the launch cohort is modeled as a free-heavy mix instead of cumulative tier occupancy
- the target and stress scenarios still stay healthy unless agent execution or human support grows faster than expected
- the meaningful early-stage risk is not infrastructure alone; it is the combination of free users, support burden, and slow conversion into monetized accounts
- at the current conservative assumptions, IntelIP has a fragile low-scale zone; the first 10 to 30 active users are where the free tier and support burden can erase pricing power

## Modeling posture after the vendor inventory

The revised scenario pack is grounded in the currently observed IntelIP stack:

- Cloudflare Workers for the app
- Neon for the app database
- Clerk for auth
- Stripe for billing
- PostHog for analytics
- Cloud Run-style backend runtime for Condere
- OpenRouter, Composio, Chroma, and optional Exa for agent workflows
- Redis remains in the backend runtime path for the worker / queue flow

Notably removed from the current frontend baseline on April 13, 2026:

- R2 upload/storage
- OpenPanel
- Grafana Faro browser instrumentation

The cost inventory in `benchmark_scenarios/intelip_tooling_cost_inventory.md` should be treated as the source of truth for why the scenario assumptions changed.
The scenario files now also carry `economics_layer` tags so future ProfitCtl output can separate direct delivery COGS from productization and adoption overhead without redefining the current gross-margin path.

## Rollout implication

The revised model suggests a clear launch constraint:

- IntelIP can sustain a `$49 Starter / $119 Pro` test if activation quality is strong and the paid mix develops quickly
- IntelIP should avoid broad, indefinite free usage during the earliest rollout
- the first pricing experiment should protect against low-scale, high-touch pilots, because the first 10 to 20 active users are where support costs hurt most
- if onboarding remains manual, IntelIP may need either a paid pilot, a workspace minimum, or a tighter free-tier cap
- the scenario pack now shows the free-heavy conservative mix as underpriced even after payment fees are restricted to monetized users

## Rollout-shape comparison

Additional rollout-shape files:

- `benchmark_scenarios/intelip_rollout_paid_pilot.yml`
- `benchmark_scenarios/intelip_rollout_tight_free.yml`
- `benchmark_scenarios/intelip_rollout_workspace_minimum_proxy.yml`
- `benchmark_scenarios/intelip_model_calibration_notes.md`

Measured on Apr 13, 2026:

- Paid pilot hybrid: `PASSED margin=84.84% cost_per_user=27.14 violations=0`
- Tight free tier: `PASSED margin=77.20% cost_per_user=13.94 violations=0`
- Workspace minimum hybrid: `PASSED margin=76.98% cost_per_user=20.60 violations=0`

Interpretation:

- paid pilot hybrid is the financially safest launch shape while onboarding is still high-touch
- a tighter free tier materially improves the early margin profile without abandoning self-serve evaluation
- a workspace minimum can now be modeled natively with user-share allocation plus a per-workspace minimum floor, and the paid-pilot motion can now be modeled with the same workspace-aware revenue path
- paid-user fee scoping improves the free-tier scenarios slightly, but it does not change the core launch recommendation

Current recommendation:

- if IntelIP wants fastest learning with least risk, start with a paid pilot
- if IntelIP wants broader top-of-funnel learning, use a tighter free tier instead of the current 10-seat free band
- keep workspace minimum in consideration and recalibrate it now that ProfitCtl has workspace-aware revenue support

## Next modeling improvements

- calibrate workspace-minimum and hybrid revenue assumptions against real design-partner packaging
- model annual billing separately from monthly
- add scenario-specific onboarding conversion assumptions beyond static plan shares
- separate workspace pricing from seat pricing if IntelIP chooses a hybrid package
