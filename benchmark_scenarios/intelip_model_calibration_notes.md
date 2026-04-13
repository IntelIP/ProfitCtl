# IntelIP Model Calibration Notes

This note captures the remaining assumptions in the IntelIP pricing model after landing plan-mix support in ProfitCtl.

## Remaining assumptions to calibrate

### 1. Model-routing cost is still synthetic

The current token line is a blended planning number. It is not yet tied to:

- the actual OpenRouter model mix IntelIP will route to
- prompt / completion token split
- tool-call and retry behavior
- staging vs production traffic mix

### 2. Backend compute is still normalized per workflow

The current workflow-compute line is a proxy for:

- Cloud Run request compute
- backend memory / retrieval work
- Redis / Chroma read-write activity
- supporting orchestration overhead

This is directionally useful, but it is not invoice-derived.

### 3. Integration cost is modeled as light variable overhead

This is reasonable while Composio-backed workflows are optional or limited.
If IntelIP's wedge becomes "connect everything quickly," the integration-action line may need to move up before launch.

### 4. Payment processing is only partially calibrated

ProfitCtl now supports `variable_costs[].user_scope: paid_users`, so IntelIP scenarios can stop charging payment processing to free users.

What still remains:

- monthly vs annual buyers
- payment-fee burden by plan
- percentage-fee vs fixed-fee treatment when packaging changes

### 5. Manual onboarding / support reserve is the biggest judgment call

This is currently the most important non-infrastructure assumption in the model.
If onboarding becomes more automated, IntelIP can support a more generous free motion.
If onboarding remains founder-led, the free tier gets riskier fast.

## Plan-mix patch status

ProfitCtl now supports:

1. `pricing.mode` with:
   - `cumulative` as the default
   - `mix` as an optional alternative
2. Optional `share` on each pricing plan when `mode: mix`
3. Share-based user allocation with deterministic rounding
4. Validation that mix shares sum to `1.0` within a small epsilon
5. Existing revenue and output shapes preserved so downstream reporting did not require a broader rewrite

That means IntelIP scenarios can now model:

- realistic free / starter / pro adoption mix
- first-pass pricing experiments without abusing tier caps as a proxy

## Workspace-hybrid patch status

ProfitCtl now supports `pricing.mode: workspace_hybrid` with:

- `workspace.average_users_per_workspace`
- per-plan `share`
- optional `workspace_minimum`
- hybrid revenue that can represent seat + workspace minimum pricing

That means IntelIP can now model:

- workspace minimums
- seat + workspace hybrids
- paid pilot structures using native revenue logic instead of per-user proxies

## Next-best patch after paid-user fee scoping

After `workspace_hybrid` and `user_scope: paid_users`, the next useful improvement is richer package semantics such as:

- included seats before overage seat billing
- pure minimum-vs-seat-floor semantics at the plan level
- annual billing and payment-fee differences by plan
