# PCTL-3 / PCTL-14 Owner Decision Packet

Status: accepted by Hudson in Codex task `019fb11b-6f69-7d41-aabb-c51588d50989`

Prepared: 2026-07-31

Current tracker: AgentShift. The PCTL identifiers below are legacy planning references, not verified AgentShift story IDs.

Historical execution authority: PCTL-3, PCTL-14, and PCTL-2

Implementation authority: PCTL-2 only; merge, release, deployment, live provider access, spend, and secrets remain excluded

## Decision Accepted

The accepted contract:

1. ProfitCtl becomes a local-first cost-intelligence control loop for founder-led developers and operators of AI-native software.
2. Existing simulation remains the forecast engine. First vertical slice adds synthetic Upstash idle-polling observations, variance attribution, and one bounded recommendation.
3. Provider catalog values remain sourced planning defaults, never guaranteed current prices.
4. Supported package identity is repository and Go module `github.com/IntelIP/ProfitCtl`, binary `profitctl`, GitHub Release install script as primary distribution, a Bun/npm package as secondary distribution, and `go run` / `go install` as explicit development fallbacks.

This acceptance clears product and provenance definition. It authorizes PCTL-2 implementation without selecting, publishing, merging, or deploying a release.

## Known Facts

- `main` presents ProfitCtl as a local CLI for pricing, unit economics, and recurring-margin simulation.
- The legacy PCTL-3 record and the proposed portfolio roadmap define a broader control loop: forecast, actual, variance, attribution, and action, with Upstash idle polling as first slice.
- Dirty canonical branch `docs/profitctl-oss-company-plan` proposes a different public center: “unit economics as code for AI-native software teams,” agent-decision demos, and a 120-day adoption plan. PCTL-10 and PCTL-16 preserve its work, but no active task, PR, or accepted product decision owns that direction.
- Canonical checkout is modified across product, open-core, growth, legal, standards-judge, skill, and website surfaces. It remains read-only.
- Repository identity is already `github.com/IntelIP/ProfitCtl`; installed binary name is already `profitctl`.
- The former secondary tap drifted from release assets and was retired; the replacement Bun/npm package remains private until its publication gate is approved.
- Release builds inject `main.version`, but current CLI does not expose it. A clean local build rejects both `profitctl --version` and `profitctl doctor`.
- Seed provider catalog was captured 2026-05-29. Entries have type, confidence, capture date, and notes, but no refresh owner, cadence, expiry behavior, or Upstash entry.
- PCTL-17 validation manifest exists only on local branch `codex/pctl-17-product-validation`; it requires deterministic, zero-provider-call schema, semantic, workflow, operational, and security validation.

## Canonical Dirty Ownership Inventory

Read-only snapshot at base SHA `cd2e095f0dd02da4a9fdbb060adaed543cfea636`, branch `docs/profitctl-oss-company-plan`:

- Product/open-core docs: `README.md`, `docs/DOCS_INDEX.md`, `docs/OPEN_CORE_PACKAGING.md`, `docs/OPEN_CORE_ROADMAP.md`, `docs/OPEN_SOURCE_COMPANY_PLAN.md`.
- Growth/design-partner work: `.github/ISSUE_TEMPLATE/design_partner_request.yml`, `docs/growth/README.md`, `docs/growth/benchmark-outreach-pack.md`, `docs/growth/benchmark-positioning.md`, `docs/growth/design-partner-offer.md`, `docs/growth/source-of-truth-guidance.md`, `docs/growth/value-proposition.md`, `docs/growth/agent-cost-os-demo-playbook.md`.
- Legal/project policy: `.gitignore`, `LICENSE`, `NOTICE`, `TRADEMARKS.md`.
- Standards and skill work: `docs/cost-model-standards.md`, `scripts/judge_cost_standards.go`, `scripts/judge_cost_standards_test.go`, `test/fixtures/agent_recommendation_valid.md`, `skills/profitctl-cost-aware/SKILL.md`.
- Local tooling/site surfaces: `.codex/`, `.entire/`, `website/`.

All listed paths belong to the canonical company-plan work and remain untouched. PCTL-2 changes overlap only `README.md`, limited to supported install and CLI identity text on this isolated branch. No canonical content was copied or reconciled.

## Option A — Control Loop with Simulation Core

Recommended.

Primary user: founder-led developer or operator responsible for cost and margin decisions in AI-native software.

Job: forecast a change, ingest local or synthetic observations, explain variance, and choose a bounded engineering action before cost or margin drift becomes operational debt.

First slice:

1. model expected Upstash idle-polling cost;
2. ingest deterministic synthetic observations into a local ledger;
3. compare forecast with observations over an explicit time window;
4. attribute variance to polling volume and unit rate;
5. recommend one bounded change with confidence and provenance.

Benefits:

- preserves current simulation investment;
- gives PCTL-18, PCTL-19, PCTL-20, and PCTL-21 one coherent dependency chain;
- proves cost intelligence without live provider access;
- keeps install and exact-head validation foundational.

Tradeoff: broader than current public CLI story. Requires explicit acceptance before implementation.

## Option B — Simulation and Decision Advisor Only

Keep current public product boundary. Ship stable CLI identity and agent-facing comparisons. Defer local ledger, observation, variance, and Upstash work.

Benefits: smallest product change; closest to committed docs and current runtime.

Tradeoff: invalidates then-current Definition-to-MVP planning sequence and leaves “cost intelligence” as planning language rather than an executable control loop.

## Option C — Agent Cost OS Public Wedge First

Adopt dirty company-plan direction: AI-native cost-aware recommendations, public demos, and design-partner adoption before local ledger work.

Benefits: strongest distribution narrative; uses existing dirty branch work.

Tradeoff: collides with unowned canonical changes, defers Upstash control-loop proof, and risks treating proposed customer/GTM claims as accepted evidence.

## Recommended Product Contract

Product goal: help one technical operator turn explicit workload assumptions and local evidence into a trustworthy cost decision.

Core workflow:

```text
forecast -> observe -> normalize -> compare -> attribute -> recommend
```

Capability boundaries:

- Forecast: deterministic scenario engine using explicit workload and rate assumptions.
- Observation: local, synthetic, imported telemetry, runtime ledger, or billing evidence.
- Normalization: units, currency, provider, service, region/tier, and time window.
- Variance: forecast versus observed cost for the same normalized boundary.
- Attribution: named workload, unit-rate, or fixed-cost drivers; no causal claim beyond evidence.
- Policy: visible covenants and thresholds; no autonomous infrastructure or billing mutation.
- Skill adoption: agents may construct scenarios and explain results; ProfitCtl owns deterministic contracts and calculations.

Measurable MVP outcome: fresh user installs `profitctl`, ingests synthetic Upstash observations, explains one variance, and receives one bounded recommendation without editing code or contacting a live provider.

Non-goals:

- finance or FinOps system of record;
- provider-pricing oracle;
- live provider query or authenticated billing access;
- autonomous remediation;
- hosted SaaS, customer evidence, publication, release, merge, or deployment.

## Recommended Provenance Policy

### Authority by Claim

| Claim | Preferred evidence | Default confidence |
| --- | --- | --- |
| Published unit rate | Official provider page or API, stored as `provider_catalog` with URL and capture time | medium |
| Billed amount | Authenticated billing export or invoice | high |
| Usage quantity | Product telemetry or runtime ledger | high when measured and scoped |
| Business assumption | Explicit `user_supplied` value | medium; high only when owner-confirmed |
| Planning fallback | Versioned template | low |

No source type is universally “most authoritative.” Authority depends on claim. Invoice proves billed amount, not future rate. Telemetry proves observed usage, not provider pricing. Official pricing proves published rate, not the user’s realized bill.

### Required Entry Fields

- provider, service, region/tier where applicable;
- value, currency, unit, and applicable time window;
- `source.type`, source URL or artifact identity, `captured_at`, and effective date when known;
- confidence and confidence rationale;
- refresh owner, cadence, and stale-after date;
- note identifying exclusions, discounts, taxes, minimums, or unresolved ambiguity.

### Refresh and Stale Behavior

- Official pricing / provider catalog: refresh every 30 days; recheck within 7 days before a decision-grade release claim; refresh immediately after a known provider change.
- Templates: review every 90 days or at each ProfitCtl release, whichever comes first.
- Billing exports / invoices: refresh per import or billing close; never silently carry forward as a current rate.
- Telemetry / runtime ledgers: retain observation window and collector identity; never silently substitute for missing rate data.
- Stale planning data: remain usable only with a visible stale warning and downgraded confidence.
- Stale data blocks high-confidence, “current price,” or release-readiness claims.
- Refresh is controlled and maintainer-owned. No random scraping, background live query, spend, or credential use.

### Upstash Demonstration

PCTL-20 should add a source-backed Upstash catalog entry only after policy acceptance. Demo uses official pricing documentation plus deterministic synthetic polling observations. It must show source, age, confidence, owner, cadence, and stale behavior. No live Upstash account or query.

## Recommended Package and Release Path for PCTL-2

- Canonical identity: `github.com/IntelIP/ProfitCtl`.
- Supported executable: `profitctl`.
- Primary install: checksum-verified GitHub Release install script into a user-selected or temporary prefix.
- Secondary install: Bun/npm package assembled from the same release binaries.
- Development fallback: explicit `go run .` or `go install github.com/IntelIP/ProfitCtl/cmd/profitctl@<version>`; never silent. The command package path is required because installing the mixed-case module root would create a `ProfitCtl` binary instead of the supported `profitctl` identity.
- Go-install availability: select a commit or future tag containing `cmd/profitctl`; the current published `v0.2.0` tag predates that package.
- Version behavior: `profitctl version` and `profitctl --version` report injected package tag; development builds report an explicit development identity.
- Doctor behavior: inspect current executable identity, configuration, provider catalog presence/age, and command prerequisites; explain failures; exit nonzero on missing required dependency; never switch binaries or mutate machine-global installation.
- Release number, publication, merge, and deployment remain separate decisions.

## Acceptance Effect

If accepted:

- PCTL-3 product decision becomes authoritative.
- PCTL-14 provenance policy becomes authoritative.
- PCTL-3 may unblock PCTL-2 and PCTL-18 according to native dependencies.
- PCTL-14 may unblock PCTL-19 according to native dependencies.
- PCTL-2 may start in an isolated `codex/` branch and this worktree.

If amended: record exact changed clauses before implementation.

If rejected: choose Option B or C and re-sequence dependencies in AgentShift before implementation.

## Acceptance Record

Hudson accepted Option A in full on 2026-07-31:

- ProfitCtl is a local-first cost-intelligence control loop.
- Existing simulation remains the forecast engine.
- Synthetic Upstash idle polling is the first forecast → observation → variance → recommendation slice.
- Official pricing is medium-confidence catalog data; invoices own billed amounts; telemetry and runtime ledgers own usage.
- Stale data is warned and downgraded, and cannot support current-price claims.
- Canonical package/repository is `github.com/IntelIP/ProfitCtl`; binary is `profitctl`.
- Checksum-verified GitHub Release installer is primary; Bun/npm is secondary; Go paths are explicit developer fallbacks.
- `--version` and `doctor` are required.

The legacy PCTL-3 and PCTL-14 records capture the accepted outcome and evidence contracts. PCTL-2 implementation proceeds on isolated branch `codex/pctl-2-stable-cli-identity`; the dirty canonical company-plan checkout remains separate and read-only.
