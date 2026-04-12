# Benchmark Positioning

ProfitCtl should be positioned as a decision tool, not just a simulator.

## Core Message

Use this message in demos, intros, and benchmark content:

> ProfitCtl models pricing, unit economics, and recurring-margin risk as code so teams can compare contract shapes before they commit to them.

## What To Show

Lead with the scenarios that already exist in the repo:

- `open_core_tiered.yml` vs `open_core_mix.yml`
- `hybrid_steady_contract.yml` vs `hybrid_pilot_contract.yml`
- `hybrid_operating_safe.yml` vs `hybrid_operating_breach.yml`

These examples answer the questions people actually have:

- which pricing shape is healthier in steady state
- whether pilot revenue is masking weak operating economics
- whether the contract survives recurring-margin scrutiny

## Audience-Specific Angle

- Founders: compare pricing shapes before shipping a plan
- Product teams: separate recurring economics from one-time revenue
- Finance: check operating margin and covenant health before a contract is signed
- Sales / CS: use the compare output to support pricing conversations with numbers

## Positioning Rules

- say "compare pricing shapes" instead of "run Monte Carlo"
- say "recurring margin" instead of "model output"
- say "decision-ready comparison" instead of "simulation"
- keep the demo focused on one comparison, not a tour of every feature

## Short Demo Script

1. Open one pair of benchmark scenarios.
2. Run `profitctl compare`.
3. Point at recurring margin, operating cost per user, and covenant status.
4. Explain why one scenario is easier to defend operationally.
5. Offer the install link and a path to calibrate their own numbers.

