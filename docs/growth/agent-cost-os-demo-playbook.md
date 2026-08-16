# Agent Cost OS Demo Playbook

Use this playbook for public demos, design-partner sessions, and agent smoke tests where ProfitCtl answers a development or architecture decision.

## Demo Promise

> Give the agent repo context and one cost-sensitive decision. ProfitCtl returns a recommendation with assumptions, margins, p95 stress, cost per user, covenant status, and a cheaper viable alternative.

## First Three Demos

### Vercel vs Cloudflare Workers

Question:

```text
Should this Next.js AI app use Vercel or Cloudflare Workers?
```

Evidence path:

1. Inspect framework, deployment target, database, auth, billing, observability, and paid APIs.
2. Use the closest AI SaaS templates.
3. Run `profitctl compare`.
4. Report monthly fixed cost, variable drivers, margin, p95 margin, cost per user, covenant status, and deployment tradeoff.

### Cloud Run vs Workers

Question:

```text
Compare Cloud Run vs Workers for this AgentOS service.
```

Evidence path:

1. Inspect runtime contract, CPU/memory, concurrency, timeout, IAM, container needs, and background work.
2. Compare Cloud Run and Workers templates.
3. Do not choose the cheapest option if runtime requirements break.
4. State cost delta and the non-cost constraint.

### Deep Research API Risk

Question:

```text
What is the cost risk of adding Exa deep research?
```

Evidence path:

1. Inspect current search/research calls and feature gate shape.
2. Add deep research as a variable cost in a temporary scenario.
3. Run `simulate --json`.
4. Report added monthly cost, margin impact, cost per user, p95 margin, covenant status, and gating recommendation.

## Standard Agent Workflow

```bash
python3 /Users/hudson/.codex/skills/profitctl-cost-aware/scripts/run_profitctl_scenarios.py \
  --template cloudflare-workers-ai-saas \
  --template cloud-run-ai-saas \
  --compare
```

For saved repo examples, run from the repository root:

```bash
go run scripts/judge_cost_standards.go skills/profitctl-cost-aware/references/templates
```

For a written recommendation artifact:

```bash
go run scripts/judge_cost_standards.go test/fixtures/agent_recommendation_valid.md
```

## Recommendation Format

Every demo answer should include:

```text
Recommendation: [choice] because [ProfitCtl evidence].
Assumptions: [users, growth, ARPU, usage shape, source confidence].
Economics: [monthly fixed cost, variable drivers, margin, p95 margin, cost/user, covenants].
Tradeoff: [what gets cheaper, riskier, slower, or operationally harder].
Alternative: [cheaper viable option, or why cheaper option is not viable].
Next step: [calibration input or scenario change].
```

## Publish Rules

- Say "planning estimate" unless telemetry or invoice data exists.
- Do not upgrade template confidence to high.
- Do not quote provider prices as current unless sourced and captured.
- Keep scenarios temporary unless the user asks to save the scenario.
- Convert useful demos into a decision artifact, not a long architecture essay.
