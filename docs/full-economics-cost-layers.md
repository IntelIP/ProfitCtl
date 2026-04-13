# Full-Economics Cost Layers

This document proposes the next ProfitCtl economics model: keep the current delivery-cost engine intact, then add higher-level layers that capture the real cost of building, launching, and supporting the product.

## Why This Change

Current scenarios are strong at modeling delivery economics:

- infrastructure
- service/runtime
- application-level operating costs

That is enough for gross margin analysis, but it misses the full cost of getting a product to market and keeping it there. For ProfitCtl, the missing pieces are:

- AI dev tooling used to build and maintain the product
- engineering time spent on productization
- release hardening and trust work
- design-partner onboarding and adoption support

If those are left out, the model can look healthy on paper while hiding the actual cost of shipping and selling the product.

## Cost Layer Model

Use two dimensions, not one:

1. Delivery layer
   Existing `layer` values remain the operational breakdown:
   - `infrastructure`
   - `service`
   - `application`

2. Economics layer
   Add a higher-level business bucket that answers what the cost is for:
   - `delivery`
   - `productization`
   - `adoption`

This keeps the current cost engine useful while adding the context needed for full-economics analysis.

### Delivery

Delivery is the cost to serve a user or workspace after the product exists.

Typical items:

- cloud hosting
- database
- API/runtime execution
- storage
- usage-based vendor fees

This is the layer most aligned with the current `simulate` output.

### Productization

Productization is the cost to make the product usable, shippable, and trustworthy.

Typical items:

- AI dev tooling subscriptions and usage
- engineering time allocated to product development
- QA and review time
- release hardening
- docs and packaging
- benchmark generation and scenario maintenance

This should usually be modeled as fixed monthly overhead, even if the actual work is bursty.

### Adoption

Adoption is the cost to turn a credible product into a real customer relationship.

Typical items:

- design-partner onboarding
- implementation support
- calibration help
- benchmark interpretation
- buyer-facing collateral and enablement

This layer is especially important for early customers, concierge pilots, and paid design-partner motions.

## Representation In ProfitCtl

The implementation should be additive, not disruptive.

Recommended shape:

- keep `fixed_costs` and `variable_costs`
- add `economics_layer` to each item
- default `economics_layer` to `delivery` when omitted
- preserve the existing `layer` field for infra/service/application breakdowns

Suggested per-item fields:

- `name`
- `amount` or `cost_per_unit`
- `period`
- `layer`
- `economics_layer`
- optional allocation metadata for non-delivery costs

## Core Design Decision

Do not redefine the current gross-margin path.

The safe implementation is:

- keep current delivery-cost math and output stable
- keep current covenant fields stable
- add a second, sibling `full economics` view

This avoids breaking:

- existing scenario files
- current `simulate` output expectations
- covenant semantics around `margin`, `cost_per_user`, `p95_margin`, and `p99_margin`

It also keeps the current delivery economics useful for scale questions while adding a fuller business view for early-stage planning.

## Data Model Recommendation

Use two independent dimensions:

1. Operational layer
   Existing values:
   - `infrastructure`
   - `application`
   - `service`

2. Economics layer
   New values:
   - `delivery`
   - `productization`
   - `adoption`

Do not overload the current `layer` enum to carry both meanings. That would blur the reports and make the API harder to reason about.

## Primary Implementation Surfaces

### Schema and Parsing

- `internal/config/parser.go`
- `internal/config/validator.go`
- `pkg/types/cost.go`
- `pkg/types/structures.go`

Recommended change:

- add a separate economics-layer enum and optional allocation metadata
- default missing `economics_layer` to `delivery`
- preserve backward compatibility for all existing config files

### Cost Aggregation

- `internal/cost/fixed.go`
- `internal/cost/variable.go`
- `internal/cost/engine.go`
- `internal/simulation/scale.go`

Recommended change:

- keep current delivery aggregation intact
- add parallel aggregation totals by economics layer
- avoid rewriting current fixed/variable delivery breakdowns

### Margin and Economics Calculation

- `internal/pricing/margin.go`
- `cmd/simulate.go`

Recommended change:

- leave `CalculateMargins` untouched for current gross-margin behavior
- add a sibling calculator for full economics
- compute the second view after the current delivery-cost path has completed

### Output Surfaces

- `internal/output/types.go`
- `internal/output/cli.go`
- `internal/output/markdown.go`
- `internal/output/json.go`

Recommended change:

- keep the current output contract stable by default
- add a sibling `economics` or `full_economics` block
- label the new view explicitly so users do not confuse it with gross margin

### Scenario and Template Surfaces

- `cmd/init.go`
- `examples/valid_profit.yml`
- `benchmark_scenarios/**`
- `test/fixtures/**`

Recommended change:

- update templates and fixtures only after schema support exists
- add a canonical `early design-partner` example alongside the current delivery-first examples

### Allocation Guidance

Productization and adoption costs should not be treated like ordinary per-request runtime costs.

Use explicit allocation rules so the economics stay readable:

- `fixed_monthly`
- `per_active_user`
- `per_design_partner`
- `per_workspace`
- `per_release`

The goal is not accounting-grade precision. The goal is to make the assumptions visible and consistent.

### AI Dev Tooling

AI dev tooling should be represented as productization cost unless it directly serves production traffic.

Examples:

- Codex subscriptions
- other AI-assisted development tools
- eval-generation runs used to improve the product
- code generation / refactoring support used during implementation

Recommended treatment:

- model as fixed monthly productization overhead
- optionally allocate across active engineering capacity, design partners, or release cycles
- do not fold it into delivery COGS unless it is clearly production-serving inference

### Engineering Time

Engineering time is also productization unless it is directly tied to a live customer implementation or support motion.

Recommended treatment:

- use a fixed monthly cost line item
- set it high enough to reflect the real burn of building and maintaining the product
- keep it separate from cloud/runtime delivery cost

### Adoption Support

Adoption cost should track the real work of onboarding and enabling users.

Recommended treatment:

- separate fixed monthly support reserve
- optionally add a per-design-partner or per-workspace onboarding line item
- keep this separate from productization so support intensity can be measured independently

## Phased Implementation

### Phase 1: Schema and Parsing

Add full-economics metadata without breaking existing scenarios.

- accept `economics_layer` on cost items
- default missing values to `delivery`
- update parser, validator, and tests
- keep current scenario files working unchanged

Success criteria:

- old scenarios still parse
- new scenarios can represent delivery, productization, and adoption costs
- CLI output remains stable for current users

### Phase 2: Reporting and Output

Expose the new economics layer in the CLI output and JSON/Markdown reports.

- show layer totals by `economics_layer`
- show delivery gross margin separately from full-economics margin
- surface per-user and per-design-partner cost where relevant

Success criteria:

- users can see what portion of cost is delivery versus productization versus adoption
- reports make it obvious when productization or adoption dominates early economics

### Phase 3: Canonical Scenario Pack

Add a small set of standard scenarios that use the new model.

- `scale_gross_margin`
- `early_design_partner`
- `post_productization`

These should become the reference scenarios for roadmap, pricing, and launch decisions.

Success criteria:

- the scenario pack covers scale economics and early-customer economics
- one scenario can be used for product planning and another for buyer conversations

### Phase 4: Workflow Integration

Use the new model in the repo’s decision process.

- tie cost-sensitive changes to ProfitCtl runs
- add PR or release review checks for economics impact
- make design-partner and pricing decisions compare against the same cost model

Success criteria:

- cost-sensitive work consistently passes through ProfitCtl
- engineering, product, and adoption work are evaluated with the same assumptions

## Suggested Sequence Against Current Work

Use the current repo and Linear direction to stage this work in the right order:

1. Keep adoption-proof work moving.
   Current active work is still design-partner pipeline, buyer-facing positioning, evaluator review, and adoption dashboard.
2. Land pricing realism improvements before full-economics polish if they unblock realistic scenarios.
   This includes plan mix, hybrid workspace pricing, unlimited plan handling, and payment-fee realism.
3. Add schema support for economics layers.
4. Add additive reporting for full economics.
5. Dogfood the model on IntelIP scenarios that include AI dev tooling, engineering allocation, and evaluator support.

That keeps the implementation aligned with the actual business questions being asked right now.

## Sample Scenario Shape

This is the proposed shape for a full-economics scenario. It keeps the current file structure, but adds a higher-level business view.

```yaml
project:
  name: "ProfitCtl - early design-partner economics"

fixed_costs:
  - name: "Cloud hosting baseline"
    amount: 40
    period: monthly
    layer: infrastructure
    economics_layer: delivery

  - name: "Production runtime baseline"
    amount: 55
    period: monthly
    layer: service
    economics_layer: delivery

  - name: "AI dev tooling"
    amount: 900
    period: monthly
    layer: application
    economics_layer: productization
    allocation:
      mode: fixed_monthly

  - name: "Engineering productization allocation"
    amount: 12000
    period: monthly
    layer: application
    economics_layer: productization
    allocation:
      mode: fixed_monthly

  - name: "Design-partner onboarding reserve"
    amount: 2500
    period: monthly
    layer: application
    economics_layer: adoption
    allocation:
      mode: per_design_partner

variable_costs:
  - name: "Production API and inference"
    cost_per_unit: 0.000006
    units_per_user: 1
    distribution: normal
    mean: 180000
    stddev: 50000
    layer: service
    economics_layer: delivery

  - name: "Benchmark generation and eval runs"
    cost_per_unit: 0.15
    units_per_user: 1
    distribution: normal
    mean: 12
    stddev: 4
    layer: application
    economics_layer: productization

pricing:
  plans:
    - name: Design Partner
      price: 199
      limits:
        users: 10

    - name: Pro
      price: 399
      limits:
        users: 1000

covenants:
  - type: threshold
    field: delivery_margin
    operator: gte
    value: 60
    message: Delivery gross margin should stay above 60%

  - type: threshold
    field: full_economics_margin
    operator: gte
    value: 20
    message: Full economics margin should stay above 20% during design-partner mode

simulation:
  base_users: 5
  growth_factor: 1.08
  iterations: 10000
```

## Decision Rule

Use delivery economics to answer whether the product can serve profitably.

Use full economics to answer whether the current phase is actually sustainable.

If productization and adoption dominate the model, that is not a failure. It is a signal that the product is still in build-and-prove mode, and the pricing or support motion should reflect that.
