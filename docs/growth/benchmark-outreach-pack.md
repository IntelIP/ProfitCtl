# Benchmark Outreach Pack

This is the approved benchmark-led outreach set for early `ProfitCtl` conversations.

Use these assets when the goal is to get an evaluator from interest to one concrete `compare` run.

## Operating Rule

Pick one benchmark pair only.

Do not send multiple comparisons in the first conversation. The point is to prove one decision, not showcase every feature.

## Pack 1: Open-Core Packaging

- Question:
  should the team stay with simple tiered packaging or move toward a clearer mix/open-core shape?
- Lead metric:
  recurring margin
- Benchmark command:

```bash
profitctl compare benchmark_scenarios/open_core_tiered.yml benchmark_scenarios/open_core_mix.yml
```

- Shareable report:
  [open_core_tiered_vs_mix.md](../../benchmark_scenarios/reports/open_core_tiered_vs_mix.md)
- Best fit:
  founder, product, or growth-led packaging decisions

## Pack 2: Pilot Vs Steady State

- Question:
  is pilot revenue making the contract look healthier than it will be in steady state?
- Lead metric:
  operating margin
- Benchmark command:

```bash
profitctl compare benchmark_scenarios/hybrid_steady_contract.yml benchmark_scenarios/hybrid_pilot_contract.yml
```

- Shareable report:
  [hybrid_steady_vs_pilot.md](../../benchmark_scenarios/reports/hybrid_steady_vs_pilot.md)
- Best fit:
  founder, sales, CS, or finance conversations around enterprise deal shape

## Pack 3: Covenant Safety

- Question:
  does the contract still hold up once recurring costs are modeled honestly?
- Lead metric:
  covenant status
- Benchmark command:

```bash
profitctl compare benchmark_scenarios/hybrid_operating_safe.yml benchmark_scenarios/hybrid_operating_breach.yml
```

- Shareable report:
  [hybrid_safe_vs_breach.md](../../benchmark_scenarios/reports/hybrid_safe_vs_breach.md)
- Best fit:
  finance, RevOps, or product/finance alignment conversations

## How To Use This Pack

1. choose the benchmark pair that matches the evaluator's question
2. send one report or one command only
3. state the lead metric before any secondary metrics
4. ask whether they want to run the closest version of their own scenario next

## What Not To Do

- do not attach all benchmark reports at once
- do not lead with Monte Carlo or implementation details
- do not talk about one-time revenue before recurring economics are clear
- do not move to calibration until the evaluator agrees the first comparison was useful
