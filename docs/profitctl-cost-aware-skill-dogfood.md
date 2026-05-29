# ProfitCtl Cost-Aware Skill Dogfood

Date: 2026-05-29

## Purpose

Validate that the Codex skill can turn repo context into cost-aware architecture advice with ProfitCtl evidence, not generic provider preference.

## Repo Context Inspected

- `web-app/package.json`: Next.js 16, OpenNext Cloudflare, Clerk, Stripe, Postgres client, PostHog.
- `web-app/wrangler.jsonc`: Cloudflare Worker deployment with OpenNext assets, Images binding, self-service binding, and 3000 ms CPU limit.
- `Condere/deployment/cloudbuild.yaml`: AgentOS production runtime deploys to Cloud Run Gen2, 1 vCPU, 2 GiB, concurrency 10, min instances 0, max instances 3, private IAM.
- `Condere/condere_src_os/core/cloud_run_production_plan.py`: repo cost expectations say Cloud Run idle scales near zero, steady single instance is `$57.02/month`, peak three-instance cap is `$171.07/month`, before request, egress, build, and free-tier offsets.
- `Condere/condere_src_os/infra/observability/run_cost_estimation.py`: Exa default estimates are `$0.007` quick search, `$0.022` search with summary/deep/answer, `$0.002` URL content, `$0.10` research task, `$0.25` research pro task.

## Scenario Runs

All scenarios used the v1 AI SaaS templates as seeds. Tracked templates were not edited. Repo-specific changes were made only in temp scenario copies.

Command pattern:

```bash
profitctl validate -f <scenario.yml>
profitctl simulate -f <scenario.yml> --json
profitctl compare <scenario-a.yml> <scenario-b.yml> --json
go run scripts/judge_cost_standards.go <scenario.yml>
```

Default guardrails:

- gross margin `>= 60%`
- p95 margin `>= 40%`
- cost per active user `<= $18`

## Prompt 1: Should the Next.js App Use Vercel or Cloudflare Workers?

Repo signal: current web app is already built around OpenNext Cloudflare and Workers bindings.

| Scenario | Revenue | Monthly Fixed | Total Cost | Gross Margin | p95 Margin | p95 Cost/User | Covenants |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Cloudflare Workers | `$4,095` | `$340.00` | `$340.45` | `91.69%` | `87.62%` | `$5.07` | pass |
| Vercel | `$4,095` | `$395.00` | `$395.51` | `90.34%` | `85.93%` | `$5.76` | pass |

Recommendation: keep Cloudflare Workers for this app unless a Vercel-specific workflow advantage outweighs the margin loss. ProfitCtl shows Workers cheaper by `$55.06/month`, `+1.35` margin points, and `$0.56/user` lower cost in the template run.

## Prompt 2: Compare Cloud Run vs Workers for AgentOS Service

Repo signal: Condere production AgentOS is intentionally Cloud Run Gen2 with private IAM, long timeout, container runtime, and scale-to-zero. A Worker is cheaper in the template, but not a drop-in runtime replacement for current AgentOS shape.

Temp edit: Cloud Run service baseline changed from template `$70/month` to repo-detected `$57.02/month`.

| Scenario | Revenue | Monthly Fixed | Total Cost | Gross Margin | p95 Margin | p95 Cost/User | Covenants |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| AgentOS Cloud Run repo-shaped | `$4,095` | `$432.02` | `$432.71` | `89.43%` | `85.08%` | `$6.11` | pass |
| Cloudflare Workers | `$4,095` | `$340.00` | `$340.45` | `91.69%` | `87.62%` | `$5.07` | pass |

Recommendation: keep AgentOS on Cloud Run for now. Workers is cheaper by `$92.26/month`, but the repo shows Cloud Run is buying private IAM, container/runtime compatibility, long request timeout, and existing Cloud Build deployment contract. Revisit Workers only for smaller stateless adapters or edge-facing routes.

## Prompt 3: What Is the Cost Risk of Adding Exa Deep Research?

Repo signal: Condere already tracks Exa billable kinds and defaults explicit `research_pro_task` to `$0.25/task`.

Temp edit: added `Exa Research Pro tasks` at `$0.25/task`, `10` tasks per active user per month, normal stress mean `10`, stddev `5`, source `repo_detected`.

| Scenario | Revenue | Monthly Fixed | Total Cost | Gross Margin | p95 Margin | p95 Cost/User | Covenants |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| AgentOS Cloud Run repo-shaped | `$4,095` | `$432.02` | `$432.71` | `89.43%` | `85.08%` | `$6.11` | pass |
| Cloud Run + Exa Research Pro | `$4,095` | `$432.02` | `$682.71` | `83.33%` | `78.61%` | `$8.76` | pass |

Recommendation: Exa Research Pro is viable under the default covenant at this load, but it is a real unit-cost driver. The modeled addition costs `$250/month`, drops margin by `6.10` points, and raises cost/user by `$2.50`. Ship behind tier gates, per-run budgets, and telemetry-backed calibration before making it broad default behavior.

## Judge Results

All four dogfood scenarios passed the standards judge.

Warnings:

- template-only scenarios correctly warn that costs are not telemetry or invoice backed.
- repo-shaped scenarios still warn that no telemetry or invoice-backed inputs are present.

## Product Readout

The skill is useful now for ranking architecture choices and forcing assumptions into the answer. It is not yet invoice-grade. The strongest next product improvement is a calibration loop:

1. Pull provider telemetry and invoices into scenario inputs.
2. Keep template confidence low/medium until calibrated.
3. Let agents attach ProfitCtl evidence to architecture recommendations by default.
4. Add judge checks that reject recommendations missing assumptions, margins, p95 margin, cost/user, covenant status, and cheaper viable alternative.

