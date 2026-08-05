---
name: profitctl-cost-aware
description: Use when making architecture, infrastructure, AI model/tool, pricing, launch, or product-development decisions where cost, unit economics, recurring margin, or cost-aware tradeoffs matter. Runs ProfitCtl scenarios before recommending providers, services, or implementation approaches.
---

# ProfitCtl Cost-Aware

## Purpose

Make development and system-design advice cost-aware. Use ProfitCtl to ground architecture choices in scenario economics instead of vague provider preference.

## Trigger This Skill

Use this skill when the user asks about:

- architecture/provider choices such as Cloudflare, Vercel, Cloud Run, Modal, Neon, Supabase, or managed services
- AI model, search, tool-call, agent-run, or API-cost tradeoffs
- new product, pricing, launch, or rollout assumptions
- adding paid infrastructure, billing, auth, storage, observability, or integration vendors
- whether a feature, contract, workflow, or implementation is worth its cost

## Workflow

1. Inspect repo context first: stack, runtime, hosting, database, auth, billing, paid APIs, expected usage, and existing docs.
2. Pick the closest template from `references/templates/` and copy it to a temp working directory. Do not edit repo-tracked scenario files unless the user asks to save one.
3. Adjust temp assumptions only when they are inferable from context or explicitly supplied by the user. If missing assumptions materially change the conclusion, ask before sounding precise.
4. Preserve `source.type`, `source.confidence`, `captured_at`, and `note` on every cost line. Do not upgrade template confidence to `high` without telemetry, invoices, or explicit user confirmation.
5. Run `profitctl validate`, then `profitctl simulate --json` or `profitctl compare`.
6. Summarize recommendation with evidence:
   - assumed users, growth, ARPU, and pricing shape
   - monthly fixed cost and top variable cost drivers
   - gross margin, p95 margin, cost per active user, and covenant failures
   - cheaper viable alternative when one exists
7. State assumptions plainly. Treat template prices as editable estimates, not current provider pricing guarantees.

## Helper Script

Use the helper for quick scenario runs:

```bash
python3 /Users/hudson/.codex/skills/profitctl-cost-aware/scripts/run_profitctl_scenarios.py \
  --template cloudflare-workers-ai-saas --template cloud-run-ai-saas --compare
```

The script locates `profitctl` on `PATH`, or falls back to `go run .` in `/Users/hudson/Documents/GitHub/IntelIP/ProfitCtl`.

## Defaults

The bundled v1 templates use these guardrails:

- gross margin must be at least `60%`
- p95 margin must be at least `40%`
- cost per active user must stay at or below `$18`

Do not use `profitctl detect` as a required step. It needs `OPENROUTER_API_KEY`, so it is optional only.

Before treating a scenario as decision-grade, run the standards judge:

```bash
go run /Users/hudson/Documents/GitHub/IntelIP/ProfitCtl/scripts/judge_cost_standards.go \
  /path/to/scenario.yml
```

## Output Pattern

For cost-aware recommendations, lead with:

```text
Recommendation: [choice] because [ProfitCtl evidence].
Assumptions: [users/growth/ARPU/top usage assumptions].
Economics: [revenue/cost/margin/p95/cost per user/covenants].
Tradeoff: [what gets cheaper or riskier].
Source provenance: [source.type values and supporting URLs or artifacts].
Confidence: [confidence level and why].
Next step: [scenario to refine or check to run].
```
