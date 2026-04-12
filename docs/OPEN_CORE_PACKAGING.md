# ProfitCtl Open-Core Packaging

`ProfitCtl` should be easy to install, useful on day one, and commercially expandable without crippling the open product.

## Open-Source Core

The open-source core is the local CLI and the repo assets that make it immediately useful:

- local binary and install flow
- config schema and examples
- `simulate`, `compare`, `validate`, `calibrate`, and `detect`
- benchmark scenarios and repeatable compare outputs
- CI-friendly exit codes and JSON / Markdown output

This layer needs to stay strong enough that a team can model a pricing decision, contract shape, or recurring-margin risk without talking to us first.

## Paid Product Boundary

The paid layer should emerge around operating leverage and team workflow, not artificial feature removal from the CLI.

The first credible paid surfaces are:

- hosted scenario registry and saved runs
- shared team workspaces and approvals
- policy packs and org-level guardrails
- managed PR checks and hosted runners
- SSO, RBAC, audit logs, and private support
- managed or single-tenant enterprise deployment

## What Should Not Be Paywalled

These should remain part of the open product:

- local scenario authoring
- local comparison and simulation
- calibration imports
- benchmark scenarios
- recurring-margin and covenant modeling

If we remove those from the core, we weaken adoption and destroy the open-core story.

## Product Story

The product should be explained in one sentence:

> ProfitCtl models pricing, unit economics, and recurring-margin risk as code before a team ships a pricing change or signs a contract.

The open product proves that value locally.

The paid product helps teams operationalize it across people, repos, approvals, and production workflows.

## Near-Term Packaging

For the next release window, the packaging should stay simple:

1. Open-source CLI:
   install, run examples, compare scenarios, calibrate assumptions
2. Design-partner motion:
   benchmark -> install -> compare -> calibrate -> follow-up
3. Future paid layer:
   collaboration, governance, hosted execution, support

That is enough to ship and market without inventing a fake enterprise platform before demand exists.
