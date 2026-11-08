# ProfitCtl Product Development Roadmap

Updated: 2026-10-07

Status: Proposed capability roadmap. The [Versioned Sprint Plan](SPRINT_PLAN.md) maps the stages to proposed releases from v0.4.0 through v1.0.0. Delivery owners, dates, and spending budgets remain to be assigned.

## Product direction

Help finance, product, and engineering understand how technical decisions affect product economics, test alternatives, and compare predictions with actual performance.

ProfitCtl should help teams answer:

> Which change improves this product's financial performance while preserving the quality, speed, and reliability customers need?

The scope includes code, architecture, hosting, AI models, pricing, demand, operational effort, and development priorities. Intended benefits include better decisions and less manual investigation. These benefits must be measured.

Keep one trusted calculation engine. Add evidence collection, experiment workers, and agent coordination around it. The local CLI, scenario formats, and cost-aware skill remain useful entry points.

This roadmap describes product capability development. The [Open-Core Roadmap](OPEN_CORE_ROADMAP.md) covers distribution, adoption, and commercial packaging. Its day-based milestones do not establish dates for this plan. Use the existing [Full-Economics Cost Layers](full-economics-cost-layers.md) proposal when extending delivery costs into productization and adoption costs.

## Starting point

The October 7 investigation established the following local evidence:

- The system economics commands inspect source signals, evaluate declared workload dependencies, compare scenarios, and summarize sanitized usage receipts.
- Numeric assumptions carry evidence. Missing inputs retain acquisition questions, sources, and owners.
- The current system model uses a monthly USD planning window. Its calculations follow supplied relationships; they do not discover causal effects automatically.
- The Condere production planning template returned 19 missing inputs and no total cost. Workload volume, usage factors, service rates, and revenue still require evidence.
- Existing synthetic comparisons demonstrate calculation behavior. They do not establish production prices, real savings, model quality, or deployment readiness.

The current design does not claim automatic live topology discovery, learned causal relationships, capacity prediction, autoscaler simulation, or quality evaluation. See [System Economics v1](design/system-economics-v1.md), [Cost Intelligence System Design](cost-intelligence-system-design.md), and the [Condere walkthrough](../examples/system/README.md).

## Development principles

1. Start with one business outcome and a representative workload. Expand after the result helps someone make a real decision.
2. Keep calculations reproducible. Agents gather evidence, propose scenarios, and explain results; ProfitCtl owns validated arithmetic and provenance.
3. Distinguish calculated consequences, measured behavior, and business hypotheses. Code inspection alone cannot establish a change in retention or revenue.
4. Preserve unknowns. Missing charges, unsuccessful runs, and unallocated costs must remain visible.
5. Compare equivalent service requirements. Include quality, latency, reliability, security, operational effort, and migration cost.
6. Keep delivery margin distinct from broader product economics. State included costs, allocation rules, and the financial time horizon.
7. Give parallel workers clear ownership. Add agents when their independent work improves the investigation.
8. Humans own product intent, quality requirements, pricing, risk tolerance, and consequential commitments.

## Proposed release stages

These are capability stages. The [Versioned Sprint Plan](SPRINT_PLAN.md#release-sequence) defines their proposed release sequence, stories, tasks, and testing objectives. Versions remain targets until their release gates pass and publication is authorized. Later stages depend on the evidence established by earlier work.

| Stage | Business outcome | Completion test |
| --- | --- | --- |
| 1. Measured economics | Explain what delivering a customer outcome actually costs | Reconcile a representative baseline and expose missing or unallocated costs |
| 2. Change impact | Understand financial consequences of proposed changes | Trace predictions to explicit assumptions, dependencies, and sources |
| 3. Tested alternatives | Compare implementations and deployment options using experiments | Measure cost and customer-facing performance on equivalent workloads |
| 4. Coordinated agent investigations | Complete investigations across teams and interruptions | Recover work, coordinate specialists, and produce a reviewable decision |
| 5. Continuous economic improvement | Revisit decisions as the business and its costs change | Compare forecasts with actuals and measure realized impact |

### Stage 1: Measured economics

Connect usage, billing, revenue, and human effort to a defined customer outcome. Use the existing Condere assessment workflow as the first case.

Deliver:

- Imports for representative sanitized usage receipts and matching cost evidence. Begin with files before building live connectors.
- A consistent record of source identity, capture time, currency, billing period, units, allocation method, and whether each value is observed or assumed.
- Separate treatment of successful, failed, cancelled, retried, and unknown outcomes.
- Cost per accepted outcome and declared business assumptions for revenue, demand, and human effort.
- A missing-information report that names the source and owner needed to resolve each material gap.

Ready when the baseline reconciles to its source records, no charge is counted twice, and missing costs remain explicit. A receipt sample must not silently become a monthly forecast. Actual bills and token-price estimates for the same usage must not be added together.

### Stage 2: Change impact

Connect proposed code, architecture, pricing, and demand changes to their economic consequences.

Deliver:

- Versioned baseline and candidate scenarios tied to source revisions and deployment assumptions.
- Explicit mappings from workload behavior to resource use and cost, including calls, retries, cache paths, fan-out, runtime, storage, and transfer.
- Growth scenarios and sensitivity analysis showing which assumptions change the recommendation.
- Separate views of recurring delivery cost, implementation and migration effort, and broader product economics.
- Explanations that distinguish measured relationships from untested hypotheses.

Ready when a reviewer can trace each material result back to changed inputs and reproduce the calculation. Comparisons must show the cost scope and time horizon. Price reductions must not imply savings when workload behavior remains unknown.

### Stage 3: Tested alternatives

Use controlled experiments to measure uncertain behavior before recommending changes.

Deliver:

- Sandbox workers with bounded file, network, credential, compute, time, and spending access.
- Versioned experiment records containing the code revision, environment, input workload, pricing snapshot, and acceptance criteria.
- Baseline and candidate measurements for resource use, output quality, latency, failures, retries, and human correction effort.
- Load profiles that represent actual arrival patterns and expose capacity limits.
- Comparable AWS, Azure, and local-hosting scenarios, with provider-specific calibration, billing rules, operational effort, and migration costs.

Ready when repeated comparable runs support the claimed difference, uncertainty and extrapolation remain visible, and another operator can reproduce the comparison. Keep experiment artifacts outside disposable sandboxes so work survives interruption.

A container makes an experiment portable; it does not guarantee equal performance across hardware or providers. Local hosting comparisons must include hardware, power, spare capacity, maintenance, backups, and recovery obligations. Quality and reliability requirements remain constraints on the cost comparison.

### Stage 4: Coordinated agent investigations

Let agents gather evidence, test assumptions, and prepare decisions across a persistent workflow.

Deliver:

- A saved decision case with its goal, constraints, task ownership, evidence, experiments, decisions, progress, and unresolved questions.
- A coordinator that assigns bounded work to product, engineering, infrastructure, experiment, and economics roles as needed.
- Focused messages and artifacts between workers, with explicit ownership of files and business records.
- Recovery after interruption, bounded retries, cancellation, duplicate prevention, and visible cost accounting.
- Human decision points for material assumption changes, product tradeoffs, and consequential actions.

Ready when a representative investigation survives interruption and produces a decision record with reproducible calculations and inspectable experiments. Compare its quality, elapsed time, cost, and human effort with a simpler baseline.

Separate agent context does not create separate security boundaries. A worktree separates code changes; a sandbox controls execution access. Keep one workflow owner responsible for coordination and final integration.

### Stage 5: Continuous economic improvement

Revisit approved decision cases as prices, demand, workload behavior, and product requirements change.

Deliver:

- Comparisons of forecasts with subsequent usage, billing, and accepted customer outcomes.
- Detection of material drift, with recommendations to refresh assumptions or investigate a change.
- Human-approved triggers and budgets for recurring investigations.
- Versioned model updates that preserve prior decisions and explain revised recommendations.
- Reporting on realized financial impact, including implementation effort and ongoing human supervision.

Ready when the system identifies useful changes, stays quiet on non-actionable variation, and demonstrates improved decision quality against a recorded baseline. Commercial savings and forecast accuracy require measured results over an agreed observation period.

## Two development tracks

| Track | Capabilities | Evidence of progress |
| --- | --- | --- |
| Economic capability | Source reconciliation, allocation, workload models, growth scenarios, forecasting, and comparison | More accurate and useful decisions against real outcomes |
| Agent capability | Filesystems, sandboxes, saved progress, communication, parallel work, and human oversight | Less human effort without reduced decision quality or uncontrolled cost |

Use the [Walking Labs harness course](https://walkinglabs.github.io/learn-harness-engineering/en/) for state, feedback, and recovery experiments. Use the [Deep Agents course](https://academy.langchain.com/courses/foundation-introduction-to-deepagents) for filesystem, context, and delegation patterns. Agno Workflows/Teams and LangGraph/Deep Agents are implementation candidates. Select one coordination owner after testing the first workflow.

## First execution slice: Condere baseline

Business question:

> What is the most economical way to deliver an accepted assessment as demand grows, while meeting agreed quality and response-time requirements?

Sequence:

1. Define an accepted assessment, required quality and response time, the observation period, and the costs included.
2. Collect existing sanitized receipts, matching billed costs, representative outcomes, and human effort. Separate observed demand from future volume assumptions.
3. Reconstruct the baseline in ProfitCtl. Resolve or explicitly retain material gaps.
4. Identify the dominant measured cost and choose one candidate change. Possibilities include model selection, retry behavior, caching, batching, or deployment configuration.
5. Specify a matched experiment, its limits, and the evidence needed. Assign a budget before any paid run.
6. Compare predicted and observed results. Record whether the evidence changes the recommendation.

First deliverable: a sourced baseline and a bounded experiment specification. Broader provider comparisons follow once the workload and acceptance criteria are understood.

## Success measures

Targets and acceptable error levels must be agreed before judging a pilot.

| Measure | Definition and use |
| --- | --- |
| Cost per accepted outcome | Attributable cost, including unsuccessful attempts, divided by accepted outcomes. Keep unknown costs explicit; the ratio is undefined when none are accepted. |
| Forecast accuracy | Compare predicted cost and resource use with actuals for an agreed workload and period, including work not used to calibrate the model. |
| Decision quality | Record whether evidence changes the decision and whether subsequent outcomes support it. |
| Human effort per decision | Measure investigation, planned approval, correction, and rescue time separately. |
| Realized margin | Compare revenue with the explicitly included cost scope; report delivery margin separately from broader economics. |
| Implementation payback | Compare measured recurring benefit with implementation and migration cost over a declared horizon. |
| Agent operating cost | Include coordination, experiments, failed attempts, storage, and ongoing monitoring in the system's own economics. |

## Integration strategy

Reuse existing tools where they fit:

- [FOCUS](https://focus.finops.org/) for billing normalization.
- [AWS Price List](https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/using-price-list-query-api.html) and [Azure Retail Prices](https://learn.microsoft.com/en-us/rest/api/cost-management/retail-prices/azure-retail-prices) for sourced planning rates. Actual contracts and billing remain separate evidence.
- [OpenCost](https://opencost.io/docs/) for relevant infrastructure and container allocation.
- [Infracost](https://www.infracost.io/docs/) for infrastructure-change estimates.
- [OpenTelemetry GenAI conventions](https://github.com/open-telemetry/semantic-conventions-genai/blob/main/docs/gen-ai/gen-ai-metrics.md) for usage and timing measurements. Pin adapter assumptions while these conventions evolve.
- [FinOps unit economics](https://www.finops.org/framework/capabilities/unit-economics/) for aligning resource measurements with business outcomes.

Start with the smallest integration needed for the baseline. Expand live connectors, shared services, and user interfaces after repeated decision cases establish their value.

## Decisions before implementation

- Who owns the first pilot, its acceptance criteria, and its financial assumptions?
- Which representative data can be used, and what access and retention rules apply?
- What quality, latency, reliability, and forecast-error limits define an acceptable result?
- What budget and execution authority apply to experiments?
- Which runtime best fits the first workflow, and who operates it?
- Is the first audience internal IntelIP product teams, external design partners, or both?

The roadmap records direction. Individual milestones still need scoped acceptance criteria, owners, and delivery decisions. Deployment, infrastructure changes, billing changes, paid experiments, production-data writes, and merges retain their separate authorization boundaries.
