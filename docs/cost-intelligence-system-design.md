# Cost Intelligence System Design

## Objective

ProfitCtl should be the deterministic economics engine for cost-aware development decisions. Agents may gather context and propose scenarios, but ProfitCtl owns the schema, provenance contract, validation rules, and repeatable calculations.

## Ownership Split

| Owner | Responsibility |
| --- | --- |
| Product / ProfitCtl | Cost schema, formulas, scenario validation, provenance labels, provider catalog format, deterministic output, standards judge |
| User | Business assumptions: users, paid seats, pricing, target margin, risk tolerance, expected usage |
| Agent | Repo inspection, provider/service detection, template selection, temp scenario generation, cost-aware explanation |

Agents must not blur facts and assumptions. Every cost or usage line should identify whether it came from a template, user input, repo detection, telemetry, invoice, or provider catalog.

## Data Flow

1. Agent inspects repository context: framework, hosting, database, auth, paid APIs, billing, background work, and telemetry.
2. ProfitCtl or an agent selects the closest scenario template.
3. User supplies or confirms business assumptions that materially change the answer.
4. Provider catalog entries fill planning defaults when no actuals exist.
5. Telemetry and invoices override defaults when connected.
6. ProfitCtl validates and simulates the scenario.
7. The agent returns a recommendation with assumptions, margins, p95 stress, cost per user, covenant status, and confidence.
8. The final scenario and output become a decision artifact that can later be compared against real actuals.

## Internet Access

V1 does not require internet access. Scenario templates and user inputs are enough for local economics checks.

Productized catalog refresh should use controlled sources:

- official provider pricing pages or APIs
- authenticated billing exports
- product telemetry
- runtime cost ledgers

Do not make random web scraping part of the trust path. Scraped values are brittle, hard to cite, and hard to keep current. If web lookup is used, cache the result with source URL, capture date, and confidence.

## System Changes

### ProfitCtl

- Add cost-source provenance fields to fixed and variable costs.
- Add a provider-catalog seed format for planning defaults.
- Add standards judge tooling that validates provenance and scenario quality.
- Keep core simulation math deterministic and local.

Cost source schema lives at `schemas/cost-source.schema.json`. Seed catalog values live under `provider_catalog/` and must be treated as planning defaults until replaced by sourced provider data, telemetry, or invoices.

### AgentOS / Condere

- Continue using the run ledger as the source for agent/model/search costs.
- Normalize runtime cost observations into project-scoped economics signals when integrating later.
- Treat provider-specific pricing estimates as calibration inputs, not hidden model behavior.

### Web App

- Capture project-scoped business assumptions: users, billable seats, runs per user, tokens per run, search calls per run, pricing plan, and target margin.
- Expose those assumptions to agents and ProfitCtl as scenario inputs.
- Avoid dashboard work until scenario quality and decision usefulness are proven.

## Confidence Model

| Source type | Typical confidence | Meaning |
| --- | --- | --- |
| `template` | low to medium | Default planning assumption, useful for first-pass comparison |
| `repo_detected` | low to medium | Inferred from code/config; needs user or telemetry confirmation |
| `user_supplied` | medium to high | Business assumption from operator or customer |
| `provider_catalog` | medium | Provider default with source and capture date |
| `telemetry` | high | Measured usage from runtime/product systems |
| `invoice` | high | Actual billed cost or revenue export |

High confidence should be reserved for actuals or explicit user/business inputs, not generic templates.
