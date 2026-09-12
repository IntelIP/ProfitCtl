# System economics v1 — v0.4.0 candidate

## Business objective

Before launching Condere, explain which services incur cost, how workload changes
affect downstream services, and which measurements are still required. The outcome
is a reviewable production decision, not an automatic claim of deployment or price
accuracy. PCTL-28, PCTL-29 and PCTL-30 deliver this refactor; PCTL-26/27 retain the
real-data demo and recording work.

## Current system and refactoring boundary

ProfitCtl currently collects configuration, optionally asks OpenRouter/Exa for a
starter assessment, and evaluates fixed/per-user scenario costs. Its local ledger
stores normalized cost observations. These paths remain compatible. The new
`system` commands add offline inspection, a versioned dependency/driver model,
explicit missing-input requirements, receipt summaries and change comparison.

Condere's current source launch design is: web/CLI/MCP clients → AgentOS Cloud Run
API → assessment Cloud Run Job; API/Job use Neon; the Job uses OpenRouter, Exa,
Modal and GCS. Internal Agno agents are not separate deployed microservices.
API limits (1 CPU, 2 GiB, concurrency 10, 0–3 instances) and Job limits (1 CPU,
2 GiB, one task, no automatic retries) are not measured capacity. Independently
dispatched Jobs can overlap. Current repository explorers are bounded to three.
Retired ARQ worker configuration is not launch authority.

## Information acquisition

| Input | Source to request | Owner | Meaning and limitation |
| --- | --- | --- | --- |
| Deployed services/regions/limits | Read-only Cloud Run service and Job configuration; deployment commit | Launch operator | Reconcile source intent with live state; no secrets needed |
| Assessments/month, success rate, pricing | Launch volume plan and product pricing decision | Product owner | Declared assumptions; never infer from repository size |
| Model tokens, tool calls, retries and reported cost | Sanitized existing Condere usage receipts | Application owner | Preserve missing/not-reported costs and failed/cancelled runs |
| Exa calls and costs | Receipt metadata or approved billing export | Application owner | Cache-hit paths differ; missing charges are unknown |
| Modal CPU/memory time | Existing sandbox usage/billing export | Infrastructure owner | Separate clone/analysis and avoid counting both invoice and estimate |
| Cloud Run CPU/memory/request time | Metrics and current regional billing terms | Infrastructure owner | Job duration, overlap and billing mode matter |
| Neon activity/connections/storage | Current plan and usage metrics | Database owner | Plan tiers and saturation need explicit modeling; not linear by default |
| GCS bytes, retention, operations, egress | Artifact metadata and billing usage | Infrastructure owner | Region and retention determine costs |
| Cache hit rates and fan-out | Stage-specific traces from representative assessments | Application owner | Not one global discount across all services |
| Quality/latency constraints | Separate task evaluation and runtime traces | Product owner | Cost percentiles are not latency or quality percentiles |

No credential or live paid call is needed for offline commands. Production
configuration/usage acquisition remains a separate authorized read. The model
records missing facts with a question, acquisition source and accountable owner.

## Contract

`profitctl.system/v1` is strict JSON, bounded to 2 MiB, 256 services, 512 drivers
and 1,024 edges. Monetary values are USD. The model covers one declared monthly
planning window. Values are nonnegative finite numbers or explicit `null`.
Every numeric assumption has existing `profitctl.cost/v1` evidence. Unknowns
instead require acquisition metadata. There are no arbitrary expressions or
network lookups.

Drivers form a directed acyclic graph. A root driver is a sourced value. A
derived driver is the sum of its inbound parent quantities multiplied by each
edge's sourced conversion factor. Driver units are explicit; each factor records
the parent's input unit and child's output unit. This allows assessments → Jobs,
assessments → model calls → tokens, and time/volume effects to be reviewed.
Missing parents/factors propagate unknown; unknown × zero stays unknown to avoid
concealing missing evidence. Cycles, dangling references, mismatched units,
duplicate IDs, nonfinite arithmetic and undeclared fields are rejected.

Services contain cost lines referencing drivers and sourced USD-per-driver-unit
prices. A modeled-free service requires a reason. Partial known spend is always
separate from total spend; total and margin remain null if any cost is unknown.
Revenue is an explicit input. Changes are compared across two full models with
the same period and currency. Reports retain driver paths and changed inputs;
they describe consequences of declared links, not inferred causal truth.

Condere receipt import accepts a deliberately sanitized envelope with assessment
ID, outcome and UsageReceipt-shaped entries. It produces per-provider/model
totals, observed counters and unknown-cost counts with a SHA-256 source identity.
Nested metadata is not echoed. It neither pulls live data nor mutates a model;
an analyst must choose a representative sample and map units to driver inputs.
This keeps actual reported charges separate from token-price estimates and
prevents a sample total from silently becoming a monthly forecast.

## Delivery and acceptance

1. **PCTL-28:** shared configuration coverage and filtering; local source-signal
   inspection, paths/hashes, no content emission; Modal registry support. Prove
   generated and symlink exclusions plus Condere-shaped coverage.
2. **PCTL-29:** strict model validation, propagation, source evidence, missing
   data report, per-service change results and sanitized receipt summary. Prove
   arithmetic, unknowns, cycles, invalid references, unit conversion, duplicate
   IDs, nonfinite values, missing receipt charges and input bounds.
3. **PCTL-30:** source-backed Condere template with unknown production inputs,
   separate synthetic baseline/change/adverse fixtures, walkthrough, v0.4.0
   candidate notes and draft PR. Validate through actual CLI entry points and
   existing exact-head product gate. No tag, merge, deployment or paid experiment.

## Deliberate limits

No automatic live topology, learned causal graph, queue simulation, autoscaler,
database capacity predictor or quality evaluator is claimed. Nonlinear tiers,
cache behavior, retries and fan-out must be represented by explicitly reviewed
scenario inputs. Unknowns are acquisition work, not permission to guess.
