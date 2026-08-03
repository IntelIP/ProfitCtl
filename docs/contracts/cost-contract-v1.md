# Cost Driver and Observation Contract v1

Status: PCTL-18 candidate

Schema version: `profitctl.cost/v1`

## Purpose

This contract lets ProfitCtl compare forecast drivers with cost observations
without treating provenance, measurement, billing, or user input as
interchangeable evidence.

Contract code lives in `pkg/contracts/cost/v1`. Structural JSON Schemas live in
`schemas/cost-contract/v1`. Go `Validate` methods enforce cross-field rules that
JSON Schema cannot express reliably, including matching units and currencies,
ordered time windows, and evidence/source compatibility.

This version is additive. It does not change `simulate`, current scenario math,
CLI output, or provider access.

## Driver Taxonomy

| Kind | Meaning | Example |
| --- | --- | --- |
| `fixed` | Commitment independent of workload volume | One platform commitment per month |
| `variable` | Quantity scaling with another unit | 10,000 requests per user |
| `cadence` | Repeated work over time | 1,386,000 polling commands per month |
| `concurrency` | Simultaneously active capacity | Four workers |
| `uptime` | Active duration over a window | 720 worker-hours per month |

Every driver requires:

- canonical quantity unit;
- explicit RFC3339 start and end;
- price amount, ISO 4217 currency, and price basis;
- workload dimension;
- separate quantity and unit-price evidence, each with evidence class,
  measurement status, source identity, capture time, confidence, and rationale.

`variable` and `cadence` drivers also require a `per` scale basis. A variable
driver cannot omit whether it scales per `user`, `paid_user`, or another
canonical subject. A cadence driver cannot omit its time denominator.
Every present `per` denominator must be greater than zero.
Cadence denominators use `second`, `minute`, `hour`, `day`, `week`, `month`,
or `year`.
Uptime quantities use the same canonical time units.

Variable drivers may carry an explicit `normal`, `uniform`, or `exponential`
distribution. Distribution-specific parameters preserve the stochastic basis
used by stress forecasts instead of silently substituting deterministic
`units_per_user`.
Legacy normal and uniform distributions also carry `floor: 0`, matching
ProfitCtl's non-negative sampler.

Units use lowercase canonical identifiers such as `request`, `command`,
`token`, `gibibyte`, `user`, `worker`, `hour`, or `month`. Ambiguous storage
units such as `GB`/`gb`, free-text ratios, and generic `unit`/`units` fail
validation.

## Observation Envelope

An observation binds one or more driver IDs to:

- one explicit time window;
- quantity and unit;
- unit price, price basis, and currency;
- total cost and currency;
- provider, service, region, tier, and workload dimensions where applicable;
- separate evidence for quantity, unit price, and total cost.

Claim-level evidence matters. Telemetry may own quantity while a provider
catalog owns a planning rate and an invoice owns billed total. One source must
not silently stand in for all three claims.

Claim authority is validated:

- quantity: template, user input, repo detection, telemetry, runtime ledger,
  synthetic fixture, or legacy scenario;
- unit price: template, user input, repo detection, invoice, provider catalog,
  synthetic fixture, or legacy scenario;
- total cost: template, user input, invoice, ProfitCtl-derived artifact,
  synthetic fixture, or legacy scenario.

Telemetry and runtime-ledger sources cannot silently become price or total-cost
evidence. Provider catalogs cannot silently become usage or billed-total
evidence.

## Evidence Semantics

Evidence kinds:

- `predicted`: forecast or derived planning value;
- `observed`: measured or explicitly synthetic observation;
- `billed`: billed value backed by invoice evidence;
- `user_supplied`: declared operator value.

Measurement status is separate:

- `measured`: allowed only for telemetry, runtime-ledger, or invoice sources;
- `synthetic`: allowed only for a named synthetic fixture;
- `derived`: calculated from other explicit claims;
- `declared`: asserted assumption.

Rules:

- provenance alone never proves measurement;
- `billed` requires invoice source;
- `user_supplied` requires user-supplied source;
- synthetic observations stay visibly synthetic;
- derived totals must equal quantity multiplied by unit price and divided by
  price-basis quantity within floating-point tolerance;
- missing source identity, capture time, confidence rationale, units, time
  windows, or currency fails closed.

Provider-catalog evidence additionally requires `refresh_owner`,
`refresh_cadence`, and `stale_after`, and cannot claim high confidence.
Catalog evidence stale before the enclosing driver or observation window fails
closed.
Template evidence also cannot claim high confidence.
Repo-detected evidence is likewise capped at medium until confirmed by user or
measurement evidence.

## Current Scenario Compatibility

Legacy scenarios remain authoritative for current simulation behavior.
`MapLegacyCosts` creates additive v1 driver views and does not mutate legacy
inputs or the cost engine.

Mapping rules:

| Legacy field | v1 mapping |
| --- | --- |
| `fixed_costs[].amount` | unit price per `commitment` |
| `fixed_costs[].period` | explicit legacy-normalized commitment count and time basis |
| `variable_costs[].cost_per_unit` | unit-price amount |
| `variable_costs[].units_per_user` | quantity per one `user` |
| `variable_costs[].distribution` and parameters | explicit v1 stochastic distribution |
| absent variable unit | caller must supply explicit name-to-unit mapping |
| absent scenario currency/window | caller must supply both |
| absent source | low-confidence `legacy_scenario`; never measured |

Legacy normalization details:

- daily fixed costs become 30 commitments per month because the current engine
  uses a fixed 30-day simulated month;
- monthly fixed costs become one commitment per month;
- yearly fixed costs become one commitment per year;
- legacy driver IDs derive from cost kind and name, so harmless list reordering
  does not break observation bindings;
- `user_scope: paid_users` maps to `paid_user`; other legacy variable costs map
  to `user`.

The compatibility fixture
`test/fixtures/cost_contract/v1/legacy_valid_config_mapping.json` maps the
current `valid_config.yml` scenario. Tests prove mapped drivers validate and
legacy calculated totals remain unchanged.

## Upstash Idle-Polling Fixture

`test/fixtures/cost_contract/v1/upstash_idle_polling.json` provides one
forecast-to-observation mapping with 1,386,000 synthetic polling commands.

The fixture is not live Upstash evidence:

- quantity is marked `observed` plus `synthetic`;
- rate is marked `user_supplied` plus `declared`;
- total is marked `predicted` plus `derived`;
- confidence remains low;
- fixture text explicitly rejects a current-price or billing claim.

No provider calls, credentials, or live data are required.

## Versioning

Breaking field, enum, or semantic changes require a new schema namespace such
as `profitctl.cost/v2` and new fixture directory. Additive v1 changes must keep
existing v1 fixtures valid. Current scenario behavior changes require separate
versioned product authority.
