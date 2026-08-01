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
- evidence class, measurement status, source identity, capture time,
  confidence, and rationale.

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
- missing source identity, capture time, confidence rationale, units, time
  windows, or currency fails closed.

## Current Scenario Compatibility

Legacy scenarios remain authoritative for current simulation behavior.
`MapLegacyCosts` creates additive v1 driver views and does not mutate legacy
inputs or the cost engine.

Mapping rules:

| Legacy field | v1 mapping |
| --- | --- |
| `fixed_costs[].amount` | unit price per `commitment` |
| `fixed_costs[].period` | driver `per` unit: `day`, `month`, or `year` |
| `variable_costs[].cost_per_unit` | unit-price amount |
| `variable_costs[].units_per_user` | quantity per one `user` |
| absent variable unit | caller must supply explicit name-to-unit mapping |
| absent scenario currency/window | caller must supply both |
| absent source | low-confidence `legacy_scenario`; never measured |

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
