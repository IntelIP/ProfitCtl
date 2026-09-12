# ProfitCtl Open-Core Roadmap

This roadmap turns ProfitCtl from a technically credible CLI into the public wedge for a cost-aware development company.

## Product Position

`ProfitCtl` should lead with one clear promise:

> Unit economics as code for AI-native software teams.

The practical use case is decision support before spend becomes real:

- compare hosting and runtime choices
- compare model, search, and paid API usage
- compare pricing and contract shapes
- validate recurring-margin covenants in CI
- give agents evidence-backed recommendations instead of vague architecture advice

## Open Product

The open-source product must stay genuinely useful on its own:

- local CLI
- scenario config schema
- `simulate`, `compare`, `validate`, `calibrate`, and `detect`
- benchmark scenarios and AI SaaS templates
- cost-source provenance and confidence labels
- standards judge for scenarios and recommendation artifacts
- Codex skill and future agent adapter docs
- CI-friendly exit codes and JSON / Markdown output

## Paid Product

The paid product should emerge around workflow and operating leverage, not artificial core limits:

- hosted scenario registry and saved runs
- team collaboration and approvals
- org policy packs and governance
- managed PR checks and hosted runners
- calibration connectors for invoices, telemetry, Stripe revenue, cloud bills, and runtime ledgers
- SSO, audit logs, RBAC, private support
- single-tenant or managed enterprise deployment

## Success Criteria Before Serious Marketing

We are ready to push harder on adoption when all of the following are true:

1. A new user can install `ProfitCtl` in under five minutes.
2. A user can run one convincing benchmark or cost-aware agent comparison without editing code.
3. The repo has clear contribution and review surfaces.
4. A versioned release exists and binaries are easy to download.
5. The README and docs make the open-core value proposition obvious.
6. Agent recommendations cite ProfitCtl output, assumptions, source confidence, and covenants.
7. There is a visible path from free local value to paid workflow value.

## Days 0-30: Public Wedge

Goal: make the agent cost-aware development story installable, demoable, and easy to explain.

Stories:

- Keep install, Bun package, GitHub release, and branded download paths working.
- Make README, docs index, open-core packaging, and company plan agree on the public wedge.
- Publish the first three demos: Vercel vs Workers, Cloud Run vs Workers, and deep research API risk.
- Keep the Codex skill and repo packaging copy aligned.
- Run the standards judge against bundled templates and recommendation fixtures.

Exit condition:

- someone outside the current team can install ProfitCtl, run one demo, and understand why the recommendation is cost-aware

## Days 31-60: Adoption Loop

Goal: turn demos into repeatable OSS adoption.

Stories:

- Publish benchmark-backed comparison content from existing scenarios.
- Add starter issues around templates, provider catalog entries, docs, and agent adapters.
- Add GitHub Action guidance for recurring-margin checks.
- Add standards judge acceptance examples for scenarios and recommendation artifacts.
- Package every useful demo into a short decision artifact.

Exit condition:

- a contributor or evaluator can identify where to help and what a good cost-aware answer looks like

## Days 61-90: Design Partners

Goal: prove that real teams use ProfitCtl before a decision.

Stories:

- Recruit 5-10 design partners from AI SaaS, agent builders, devtools founders, and FinOps-forward teams.
- Run one architecture, AI API, pricing, or contract comparison per partner.
- Capture minimum calibration inputs: users, ARPU, usage shape, paid API calls, and target margin.
- Track whether the recommendation changed the decision.

Exit condition:

- at least five design-partner loops produce decision artifacts with real inputs and clear next actions

## Days 91-120: First Paid Surface

Goal: build only the paid surface that repeated usage proves.

Choose one:

- hosted saved runs and scenario registry
- managed PR checks and policy packs
- calibration connectors for invoices, telemetry, Stripe revenue, cloud bills, or runtime ledgers

Exit condition:

- the first paid surface maps to repeated demand from design partners, not internal speculation

## Non-Goals Right Now

Do not spend this phase on:

- broad SaaS platform work before repeated paid workflow pain
- enterprise-only feature buildout before open product adoption
- random scraping as a provider-price trust path
- a generic FinOps dashboard
- advanced simulation complexity that does not improve install, demo, calibration, or decision quality

The constraint is distribution and decision trust, not model sophistication.
