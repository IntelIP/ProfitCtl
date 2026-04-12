# Design-Partner Operating System

This is the operating model for turning `ProfitCtl` from a credible open-core CLI into a product with provable user value.

The goal is not vague “growth.” The goal is to get real teams to install the product, run it on a real decision, and tell us whether it changed the decision quality.

## Operating Goal

Over the next 4 to 6 weeks, drive 5 to 10 serious evaluator loops through this sequence:

benchmark -> install -> compare -> calibrate -> follow-up

For each evaluator, capture:

- whether they installed without help
- whether they ran a benchmark or their own config
- whether they trusted the output enough to discuss a real pricing or contract decision
- what blocked broader internal adoption

## Who We Are Targeting

Start with teams that already have a pricing or contract decision in motion.

Best-fit evaluators:

- B2B SaaS founders deciding between tiered, mix, or hybrid pricing
- product or finance leads reviewing recurring margin before shipping packaging changes
- RevOps, sales, or CS leaders defending pilot-heavy contract structures
- teams that can share a small set of real inputs for calibration

Bad-fit evaluators for this phase:

- users looking for generic financial planning software
- teams without a concrete pricing or contract question
- users who want a hosted collaboration product today
- users who cannot install a CLI or provide inputs

## Core Offer

Lead with one offer only:

> We will help you compare one pricing or contract shape against one realistic alternative before you commit to it.

That keeps the conversation anchored to the product value the CLI already delivers.

## Funnel Definition

The funnel is deliberately short.

1. `Target`
One company/contact with a clear pricing, packaging, or contract question.

2. `Qualified`
They fit the target criteria and can articulate the decision they are trying to make.

3. `Installed`
They installed `profitctl` locally or via Homebrew.

4. `Activated`
They ran one benchmark comparison or a config close to their own scenario.

5. `Calibrated`
They provided enough real input to map or normalize assumptions.

6. `Design Partner`
They agreed to one follow-up loop with a defined success criterion.

## Required Assets

Every evaluator motion should use the same assets:

- [Benchmark Positioning](benchmark-positioning.md)
- [Compare Reporting Guide](compare-reporting-guide.md)
- [Design-Partner Intake](design-partner-intake.md)
- [Design-Partner Evaluation Template](design-partner-evaluation-template.md)
- one benchmark pair from [`benchmark_scenarios`](../../benchmark_scenarios/README.md)

If an evaluator cannot get through the motion with those assets, the product or docs are still too weak.

## Standard Motion

### Step 1: Pick the right benchmark pair

Choose the benchmark pair that matches their question:

- tiered vs mix
- steady-state hybrid vs pilot-heavy hybrid
- covenant-safe vs covenant-breaching contract

Do not demo every feature.

### Step 2: Drive to install

Use the public install path first:

```bash
curl -fsSL https://raw.githubusercontent.com/IntelIP/ProfitCtl/main/scripts/install.sh | bash
```

Or Homebrew:

```bash
brew tap IntelIP/profitctl
brew install profitctl
```

The install itself is part of the product test. If it requires hand-holding, record that as product friction.

### Step 3: Drive to one compare run

Use one command only, matched to the benchmark pair:

```bash
profitctl compare benchmark_scenarios/open_core_tiered.yml benchmark_scenarios/open_core_mix.yml
```

Or the closest pair to their contract shape.

### Step 4: Drive to a real decision

Ask:

- Which result would have been hard to see without this comparison?
- Which metric did you trust first?
- Would you use this before a real pricing or contract decision?

If the answer is weak, do not jump to feature building. Identify whether the problem is positioning, install friction, config mapping, or output clarity.

### Step 5: Drive to calibration

If they see value, ask for the minimum real input needed to calibrate:

- current pricing shape
- free, paid monthly, and paid annual behavior
- payment-fee assumptions
- target contract shape
- one success condition they care about

### Step 6: Drive to a design-partner commitment

The design-partner ask should be simple:

- one contract or packaging question
- one follow-up cycle
- one success criterion

Example success criteria:

- “The tool helps us reject or justify one pilot-heavy deal.”
- “We can compare our current packaging against one proposed change with real numbers.”
- “Product and finance both trust the same recurring-margin output.”

## Weekly Cadence

Run the same cadence every week.

### Monday

- review open evaluator loops
- select 5 to 10 new targets
- choose the benchmark pair for each

### Tuesday to Thursday

- run outbound and follow-up
- capture install, activation, and calibration outcomes
- convert strong evaluators into design-partner calls

### Friday

- review the week using the evaluation template
- identify the top 3 blockers across all conversations
- create or update execution issues in Linear

## Metrics That Matter

Track only the metrics that show whether the product is becoming easier to adopt and easier to trust.

Primary metrics:

- number of qualified evaluator conversations
- install success rate
- first `compare` run success rate
- number of calibrated scenarios
- number of active design partners
- number of evaluator loops that influenced a real pricing or contract decision

Secondary metrics:

- GitHub release download trend
- Homebrew install trend
- docs/help friction themes
- top missing config or output patterns

## What Counts As Product Proof

The product is ready for harder outbound and broader marketing when:

- 5+ evaluators completed install and first compare without hand-holding
- 3+ evaluators ran calibration on real inputs
- 3+ evaluators said they would use `ProfitCtl` before a real decision
- at least 2 design partners agreed to a follow-up operating loop
- the top recurring blockers are documented and actively being addressed

## Failure Modes To Avoid

- turning every evaluator conversation into bespoke services work
- adding new CLI surface before understanding adoption friction
- confusing benchmark interest with real product activation
- optimizing for vanity metrics instead of install/use/decision proof

## What To Do Next After Each Week

Every weekly review must end with exactly these outputs:

- one short written summary of what changed
- updated evaluator notes
- 1 to 3 execution issues in Linear
- a clear choice:
  - improve onboarding
  - improve scenario mapping
  - improve output clarity
  - continue outreach with the current motion

If the work does not feed one of those outcomes, it is probably noise.
