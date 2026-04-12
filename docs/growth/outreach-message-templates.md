# Outreach Message Templates

Use these templates for the first live `ProfitCtl` evaluator loops.

Keep them short. The goal is to move a target to one install and one `compare` run, not to explain the whole product.

## Rules

- pick one benchmark pair only
- name the evaluator's likely decision
- lead with recurring economics, not simulation internals
- end with one next step

## Cold Or Warm Intro

Use when the target likely has a pricing or contract decision in flight.

```text
We built ProfitCtl to help SaaS teams compare pricing and contract shapes before they commit to them.

The reason I thought of you is that your team looks like a fit for one concrete question: [insert pricing or contract question].

The closest benchmark pair is [insert benchmark pair]. If it is useful, the fastest path is to install ProfitCtl and run one compare command against that benchmark.

If that sounds relevant, I can point you to the exact install path and the one benchmark run that matches your decision.
```

## Follow-Up After Interest

Use when they reply and want to know what the evaluation looks like.

```text
The simplest path is:

1. install ProfitCtl
2. run one compare command against the closest benchmark pair
3. decide whether the output is strong enough to justify calibrating your own inputs

For your case, I would start with:
[insert benchmark command]

What I want to learn first is whether the comparison changes how you think about the decision you are about to make.
```

## Install Push

Use when the target is qualified and you want a concrete action.

```text
Best next step is to install and run one compare.

Install:
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | bash

Or:
brew tap IntelIP/profitctl
brew install profitctl

Then run:
[insert benchmark command]

If the result is useful, we can map the closest version of your own scenario next.
```

## Calibration Ask

Use only after the target says the first comparison was useful.

```text
If you want to test your own shape next, I only need the minimum real inputs:

- current pricing or contract shape
- whether users behave differently across free, paid monthly, and paid annual
- one recurring-cost assumption
- the success criterion you care about most

That is enough to calibrate a first pass without turning this into a long consulting loop.
```

## Design-Partner Commitment Ask

Use when the evaluator saw value and is open to one follow-up loop.

```text
If this first comparison helped, the next useful step is a design-partner loop around one real decision.

The scope is intentionally small:
- one pricing or contract question
- one calibrated scenario
- one follow-up review
- one success criterion

If we can help you keep, revise, or reject the change with more confidence, the loop is doing its job.
```
