# Compare Reporting Guide

This guide turns `profitctl compare` into a repeatable artifact for demos, blog posts, and design-partner conversations.

## When To Use It

Use `compare` when the question is about:

- tiered vs mix pricing
- pilot vs steady-state contract health
- booked revenue vs operating economics
- covenant risk under recurring assumptions

Do not use it as a generic dump of every metric. Keep the comparison narrow.

## Recommended Inputs

Use the benchmark scenarios already in the repo:

```bash
./profitctl compare benchmark_scenarios/open_core_tiered.yml benchmark_scenarios/open_core_mix.yml
./profitctl compare benchmark_scenarios/hybrid_steady_contract.yml benchmark_scenarios/hybrid_pilot_contract.yml
./profitctl compare benchmark_scenarios/hybrid_operating_safe.yml benchmark_scenarios/hybrid_operating_breach.yml
```

## What To Call Out In The Output

In every compare result, lead with:

- recurring margin
- operating cost per user
- covenant status
- delta vs baseline

Only mention booked or one-time revenue after the recurring economics are clear.

## Shareable Output Rules

- keep the comparison to one decision
- include the baseline and candidate file names
- do not paste raw output without a sentence that says what changed
- if the scenario uses a pilot, say whether the pilot is part of steady-state economics or a one-time ramp cost

## Simple Post Format

Use this structure for a short public write-up:

1. State the question.
2. Show the two configs.
3. State the recurring-margin result.
4. State the covenant result.
5. Say which shape is safer and why.

## Internal Checklist Before Sharing

- the configs run locally
- the result is reproducible
- the comparison is easy to explain in one paragraph
- the conclusion survives without the one-time revenue line

