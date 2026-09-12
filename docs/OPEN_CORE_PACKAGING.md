# ProfitCtl Open-Core Packaging

`ProfitCtl` should be easy to install, useful on day one, and commercially expandable without crippling the open product.

The public wedge is ProfitCtl first: cost-aware development for AI-native software teams.

## Current Value Proposition

`ProfitCtl` gives an agent, developer, product, finance, or founder-led team a fast way to answer a hard question before they ship pricing, choose infrastructure, add an AI API, or sign a contract:

> Do the modeled revenue, fees, AI/tool costs, and recurring costs still produce a healthy business under normal and stressed conditions?

That is the product we should lead with publicly. It is specific, immediate, and already supported by the CLI we ship today.

## Who Should Evaluate It Now

The current open product is best for:

- B2B SaaS teams testing pricing changes
- AI SaaS teams comparing hosting, model, search, database, and agent-runtime decisions
- agent builders adding paid APIs, deep research, model routing, or tool calls
- startups selling hybrid or seat-based contracts
- operators who want recurring-margin checks in CI
- founders comparing pilot-heavy deals against steady-state economics

It is not yet trying to be a finance system of record, billing platform, provider-pricing oracle, or full SaaS collaboration product.

## Open-Source Core

The open-source core is the local CLI and the repo assets that make it immediately useful:

- local binary and install flow
- config schema and examples
- `simulate`, `compare`, `validate`, `calibrate`, and `detect`
- benchmark scenarios and repeatable compare outputs
- cost-aware agent skill and reusable AI SaaS templates
- provider catalog seeds with provenance and confidence metadata
- standards judge for scenario quality and recommendation artifacts
- CI-friendly exit codes and JSON / Markdown output

This layer needs to stay strong enough that a team can model a pricing decision, contract shape, or recurring-margin risk without talking to us first.

## Paid Product Boundary

The paid layer should emerge around operating leverage and team workflow, not artificial feature removal from the CLI.

The first credible paid surfaces are:

- hosted scenario registry and saved runs
- shared team workspaces and approvals
- policy packs and org-level guardrails
- managed PR checks and hosted runners
- calibration connectors for invoices, telemetry, Stripe revenue, cloud bills, and runtime ledgers
- SSO, RBAC, audit logs, and private support
- managed or single-tenant enterprise deployment

## Free vs Paid Boundary

Use this rule:

- free: anything a single team can run locally or in their own CI to model, compare, calibrate, and validate unit economics
- paid: anything that reduces coordination cost across people, repos, approvals, policies, or enterprise controls

Individual developer and agent-decision value stays open. Manager, organization, finance, compliance, and governance workflow becomes paid.

That keeps the open product credible while preserving a real commercial layer.

## What Should Not Be Paywalled

These should remain part of the open product:

- local scenario authoring
- local comparison and simulation
- calibration imports
- benchmark scenarios
- recurring-margin and covenant modeling
- cost-aware agent guidance
- standards checks for local scenarios and recommendation artifacts

If we remove those from the core, we weaken adoption and destroy the open-core story.

## Product Story

The product should be explained in one sentence:

> ProfitCtl is unit economics as code for AI-native software teams.

Expanded:

> ProfitCtl models pricing, AI/tool costs, unit economics, and recurring-margin risk before a team ships a product change, chooses infrastructure, adds a paid API, or signs a contract.

The open product proves that value locally.

The paid product helps teams operationalize it across people, repos, approvals, and production workflows.

## Why Someone Would Buy Later

The first strong buy signals are not "I need more simulation math." They are:

- multiple people need to review and approve architecture, pricing, or AI-spend decisions
- leadership wants saved runs and benchmark history
- finance or product wants policy guardrails enforced automatically
- the company needs RBAC, audit logs, SSO, or managed support
- teams need invoice or telemetry calibration without maintaining local glue

If we monetize those needs, the open-core line stays honest.

## Near-Term Packaging

For the next release window, the packaging should stay simple:

1. Open-source CLI:
   install, run examples, compare scenarios, calibrate assumptions
2. Design-partner motion:
   benchmark -> install -> compare -> calibrate -> decision artifact
3. Future paid layer:
   saved runs, managed PR checks, calibration connectors, collaboration, governance, hosted execution, support

That is enough to ship and market without inventing a fake enterprise platform before demand exists.
