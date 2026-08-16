# ProfitCtl Open-Source Company Plan

ProfitCtl should become the public wedge for an open-source company around cost-aware development for AI-native software teams.

## Position

> Unit economics as code for AI-native software teams.

The practical promise is simple: agents, developers, and founders can compare architecture, model, API, pricing, and infrastructure choices before spend becomes real.

This is not generic FinOps. Traditional FinOps usually explains spend after it exists. ProfitCtl shifts economics left into planning, code review, agent recommendations, pricing checks, and launch decisions.

## Company Shape

- Public wedge: `ProfitCtl` first.
- Parent context: IntelIP stays the lab and operating context, not the public product umbrella.
- Dogfood source: Condere supplies agent/model/search run telemetry when calibration work is ready.
- Later sibling: GuardCtl can become the policy/security counterpart after ProfitCtl proves repeated cost-decision value.

## Open-Core Boundary

The open product must stay useful without a sales conversation:

- local CLI install and source build
- scenario schema, examples, and benchmark scenarios
- `simulate`, `compare`, `validate`, `calibrate`, and `detect`
- JSON, Markdown, and CI-friendly output
- cost-aware agent skill and future agent adapters
- standards judge for scenario and recommendation quality
- provider catalog seeds as editable planning defaults

The paid product should start only where coordination cost appears:

- hosted scenario registry and saved runs
- shared team workspaces and approvals
- managed PR checks and hosted runners
- organization policy packs and guardrails
- calibration connectors for invoices, telemetry, cloud bills, Stripe revenue, and run ledgers
- SSO, RBAC, audit logs, support, managed deployment, and single-tenant deployment

Use a buyer-based rule: individual developer decision support stays open; org, manager, finance, compliance, and governance workflows become paid.

## Product System

ProfitCtl Core owns deterministic calculations, scenario validation, covenants, provenance, provider catalog format, and repeatable output.

Agent Kit starts with the Codex skill, then adds Claude, Cursor, GitHub Copilot, and MCP documentation only after the Codex workflow proves repeated value.

Standards Judge validates scenario quality and recommendation quality. It should reject missing provenance, missing confidence, vague recommendations, or precise cost claims without source evidence.

Calibration Layer imports actuals and raises confidence only when the source justifies it. Template data stays low or medium confidence. Provider catalog data is medium confidence. Telemetry and invoices can be high confidence.

ProfitCtl Cloud comes last. The first paid surface should be selected from repeated usage evidence, not imagined enterprise demand.

## First 120 Days

| Window | Goal | Exit Criteria |
| --- | --- | --- |
| Days 0-30 | Package the public wedge | README, install, examples, skill docs, judge docs, and first three public demos are usable from repo docs alone. |
| Days 31-60 | Turn demos into adoption | Public benchmark content, GitHub Action guidance, starter issues, and standards judge acceptance examples exist. |
| Days 61-90 | Run design-partner loops | 5-10 AI SaaS, agent, devtools, or FinOps-forward teams produce one decision artifact each. |
| Days 91-120 | Choose first paid surface | Build only the repeated paid pain: saved runs, managed PR checks, or calibration connectors. |

## GTM Motion

Lead with OSS demos, not enterprise pitch.

The repeatable content loop is:

1. inspect real repo context
2. choose or edit a temporary ProfitCtl scenario
3. run `validate`, `simulate --json`, or `compare`
4. publish a short recommendation with evidence
5. capture whether the decision changed

Best first public demos:

- Should this Next.js AI app run on Vercel or Cloudflare Workers?
- What is the cost risk of adding deep research?
- Which agent architecture keeps p95 margin above target?

Design-partner offer:

> Bring one architecture, AI API, pricing, or contract decision. We compare it against one realistic alternative.

## Accuracy Standard

Every decision artifact should include:

- recommendation
- assumed users, growth, ARPU, and usage shape
- monthly fixed cost
- variable cost drivers
- margin, p95 margin, and cost per user
- covenant status
- cheaper viable alternative
- source provenance and confidence

Do not make provider-price truth depend on random scraping. Use official pricing pages or APIs, authenticated billing exports, product telemetry, runtime cost ledgers, and user-supplied assumptions.

## Company Metrics

- install to first useful compare in under 5 minutes
- 10 public benchmark demos
- 5 design partners with real decisions
- 3 public case studies where ProfitCtl changed a product, infra, pricing, or AI-tool decision
- 1 calibrated scenario using real telemetry or invoice data
- 1 paid workflow signal before building broad SaaS

## Source Pattern

This plan follows the existing open-core packaging docs and external open-source company patterns:

- [GitLab stewardship and buyer-based open core](https://handbook.gitlab.com/handbook/company/stewardship/)
- [GitLab pricing handbook](https://handbook.gitlab.com/handbook/company/pricing/)
- [Open Core Ventures open-core model](https://handbook.opencoreventures.com/how-we-work/open-core)
- [FinOps Foundation State of FinOps](https://data.finops.org/2025-report/)
- [FinOps Foundation agentic AI use cases](https://www.finops.org/insights/ai-for-finops-agentic-use-cases/)
- [PostHog open-source startup lessons](https://newsletter.posthog.com/p/the-hidden-benefits-of-being-an-open)
