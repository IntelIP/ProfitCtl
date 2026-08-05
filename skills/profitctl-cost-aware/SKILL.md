---
name: profitctl-cost-aware
description: Use for material architecture, infrastructure, AI model/tool, provider, pricing, launch, or product decisions where cost, unit economics, recurring margin, or scale assumptions can change the recommendation. Preserve current user-supplied prices and scale; when either is missing, run ProfitCtl's source-backed assessment before recommending an option.
---

# ProfitCtl Cost-Aware

## Decision Gate

Use ProfitCtl when the decision can materially change provider spend, recurring cost, unit economics, price, margin, or launch scope.

Do not run ProfitCtl for unrelated coding, local refactors, copy changes, or ordinary offline simulation against a complete existing scenario.

For a material cost decision:

1. Preserve current user-supplied invoices, contracts, rates, budgets, and scale inputs. Never replace them with Exa research.
2. Reuse a saved assessment only when it covers the current code and the user treats its evidence as current.
3. If any material provider price or scale input is missing, run `profitctl assess`. This step is required, not optional.
4. Never make a background provider call. Run assessment only for the active user request.

## Workflow

1. Inspect repository context: runtime, hosting, database, auth, billing, paid APIs, provider configuration, and expected usage.
2. Choose the evidence route:
   - complete current user evidence: preserve it and build a temporary scenario
   - current saved assessment: read it offline
   - missing price or scale: run a live source-backed assessment
3. For a live assessment, require `OPENROUTER_API_KEY` and `EXA_API_KEY`:

   ```bash
   python3 <this-skill-directory>/scripts/run_profitctl_scenarios.py \
     --assess-path /path/to/codebase
   ```

4. For saved evidence, make no provider call:

   ```bash
   python3 <this-skill-directory>/scripts/run_profitctl_scenarios.py \
     --assessment /path/to/cost-assessment.json
   ```

5. When a scenario exists, copy the nearest file from `references/templates/` to a temporary directory. Apply user evidence first, then repository facts, then clearly labeled inference.
6. Run `profitctl validate`, followed by `profitctl simulate --json` or `profitctl compare`.
7. Keep `source.type`, `source.confidence`, `captured_at`, and `note` on every scenario cost line. Keep confidence below `high` without telemetry, invoices, contracts, or explicit user confirmation.
8. Run the version-matched standards judge before calling a scenario decision-grade:

   ```bash
   <this-skill-directory>/bin/profitctl-standards /path/to/scenario.yml
   ```

   In a source checkout before release artifacts exist, use `go run scripts/judge_cost_standards.go /path/to/scenario.yml` from the ProfitCtl package root.

## Portability and Failures

Resolve the helper relative to this `SKILL.md`. The helper resolves `profitctl` from:

1. the installed skill's version-matched `bin/` directory
2. `PATH`
3. a source checkout next to this repository skill, or an explicit `PROFITCTL_REPO`

Never use a developer-specific absolute path.

If the binary is missing, stop with the install action. If either API key is missing for live assessment, name the missing variable and stop. Do not substitute an unsourced price. Saved assessment review and ordinary scenario simulation remain offline.

## Output

Separate the answer into:

```text
Recommendation: [choice and economic reason].
Assumptions: [user inputs, repository facts, and every inferred scale assumption].
Facts: [code-backed providers and user/repository inputs].
Sourced rates: [official URLs, captured rates, and capture times].
Inferred scale: [every inferred volume or growth assumption].
Economics: [fixed cost, variable drivers, total, gross margin, p95 margin, cost per active user, covenants].
Tradeoff: [cheaper viable alternative and what gets riskier].
Source provenance: [source.type values and supporting URLs or artifacts].
Confidence: [confidence level and why].
Next action: [one refinement, measurement, or implementation step].
```

State that assessment values are a source-backed starter model, not actual billing. Include a cheaper viable alternative when one exists.

## Defaults

Use these bundled scenario guardrails:

- gross margin at least `60%`
- p95 margin at least `40%`
- cost per active user at most `$18`
