# Benchmark Positioning

ProfitCtl should be positioned as a decision tool, not just a simulator.

## Core Message

Use this message in demos, intros, and benchmark content:

> ProfitCtl is unit economics as code for AI-native software teams.

Expanded:

> ProfitCtl models pricing, AI/tool costs, unit economics, and recurring-margin risk so teams can compare product, architecture, API, and contract choices before they commit.

## What To Show

Lead with the scenarios that already exist in the repo:

- AI SaaS templates from the `profitctl-cost-aware` skill for Vercel, Workers, and Cloud Run comparisons
- `open_core_tiered.yml` vs `open_core_mix.yml`
- `hybrid_steady_contract.yml` vs `hybrid_pilot_contract.yml`
- `hybrid_operating_safe.yml` vs `hybrid_operating_breach.yml`

These examples answer the questions people actually have:

- which pricing shape is healthier in steady state
- which architecture or paid API choice keeps margin and cost per user inside target
- whether an agent feature should be gated, tiered, or rejected
- whether pilot revenue is masking weak operating economics
- whether the contract survives recurring-margin scrutiny

## Audience-Specific Angle

- AI SaaS teams: compare hosting, model, search, database, and runtime choices before shipping
- Agent builders: compare deep research, tool-call, and model-routing cost before rollout
- Founders: compare pricing and product-cost shapes before shipping a plan
- Product teams: separate recurring economics from one-time revenue
- Finance: check operating margin and covenant health before a contract is signed
- Sales / CS: use the compare output to support pricing conversations with numbers

## Positioning Rules

- say "compare pricing shapes" instead of "run Monte Carlo"
- say "compare architecture and AI API choices" instead of "optimize cloud spend"
- say "recurring margin" instead of "model output"
- say "decision-ready comparison" instead of "simulation"
- keep the demo focused on one comparison, not a tour of every feature

## Short Demo Script

1. Open one pair of benchmark scenarios or one cost-aware agent template pair.
2. Run `profitctl compare` or the ProfitCtl cost-aware skill helper.
3. Point at recurring margin, p95 margin, operating cost per user, and covenant status.
4. Explain why one scenario is easier to defend operationally and economically.
5. Offer the install link and a path to calibrate their own numbers.
