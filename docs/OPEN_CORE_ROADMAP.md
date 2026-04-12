# ProfitCtl Open-Core Roadmap

This roadmap is the shortest path from "technically credible CLI" to "shippable open-core product that people can install, evaluate, contribute to, and eventually buy around."

## Product Position

`ProfitCtl` should lead with one clear promise:

> Model unit economics, pricing scenarios, and recurring-margin risk as code before shipping product or signing contracts.

The open-source product should stay strong enough to be genuinely useful on its own:

- local CLI
- scenario config schema
- `simulate`, `compare`, `validate`, `calibrate`, `detect`
- benchmark scenarios
- CI-friendly exit codes and output formats

The paid product should emerge around workflow and operating leverage, not artificial core limits:

- hosted scenario registry and saved runs
- team collaboration and approvals
- org policy packs and governance
- managed PR checks and hosted runners
- SSO, audit logs, RBAC, private support
- single-tenant or managed enterprise deployment

## Success Criteria Before Serious Marketing

We are ready to push harder on adoption when all of the following are true:

1. A new user can install `ProfitCtl` in under five minutes.
2. A user can run one convincing benchmark or pricing comparison without editing code.
3. The repo has clear contribution and review surfaces.
4. A versioned release exists and binaries are easy to download.
5. The README and docs make the open-core value proposition obvious.
6. There is a visible path from free CLI value to paid operational value.

## Phase 1: Release Baseline

Goal: make the project easy to install, release, and evaluate immediately.

Stories:

- Cut and verify `v0.1.0` from `main`.
- Ensure GitHub release artifacts and VPS mirrors are actually published and downloadable.
- Add one frictionless install path beyond `go install`.
- Tighten README and Quick Start around the core user journey: install, init, compare, calibrate.
- Confirm benchmark scenarios are stable and representative.

Exit condition:

- someone outside the current team can install and run `ProfitCtl` from docs alone

## Phase 2: Open for PRs

Goal: make the repo legible and safe for outside contribution.

Stories:

- Add issue templates, pull request template, and `CODEOWNERS`.
- Label and curate a starter set of contributor issues.
- Decide and document PR policy for CI, AppSec, and Greptile.
- Reduce backlog noise so active work is obvious.

Exit condition:

- a contributor can identify where to help and what "done" means without asking for process

## Phase 3: Open-Core Packaging

Goal: make the product boundary obvious and commercially defensible.

Stories:

- Publish a concise open-core packaging document.
- Define what is permanently free in the CLI.
- Define the first paid workflow layer.
- Align messaging across README, docs, and release notes.

Exit condition:

- we can explain in one paragraph why the open product is useful and what the paid layer eventually buys

## Phase 4: Benchmark-Led Growth

Goal: make the project easy to talk about, share, and evaluate.

Stories:

- Publish benchmark-backed comparison content from existing scenarios.
- Add a public roadmap and design-partner messaging.
- Package `compare` outputs into shareable artifacts for founders and product teams.
- Create a simple adoption loop: benchmark -> install -> compare -> calibrate -> contact / design-partner interest.

Exit condition:

- we have a repeatable way to turn technical utility into inbound conversations

## Immediate Next 30 Days

Week 1:

- ship `v0.1.0`
- verify release distribution
- tighten install and evaluation docs

Week 2:

- add OSS repo hygiene and contributor entry points
- create curated starter issues

Week 3:

- publish open-core packaging and public roadmap
- sharpen benchmark scenarios and shareable outputs

Week 4:

- start direct outreach with benchmark-led demos
- collect design-partner feedback and turn it into roadmap inputs

## Non-Goals Right Now

Do not spend this phase on:

- enterprise-only feature buildout
- premature SaaS platform work
- overbuilt community process
- advanced performance optimization that does not improve install/evaluate/adopt loops

The constraint is distribution, not model sophistication.
