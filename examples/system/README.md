# Condere production economics walkthrough

Start with the business question: “What will running Condere cost, and which
other costs change if we alter its assessment workflow?”

Condere uses Agno inside an AgentOS API and assessment Job. The Job depends on
OpenRouter, Exa, Modal, Neon and GCS; web/CLI/MCP are access surfaces. Codex can
inspect source and explain assumptions. ProfitCtl calculates declared economic
consequences. Do not turn a detected package into a claim of a deployed service.

## 1. Inspect source evidence locally

```bash
profitctl system inspect /path/to/Condere
```

Output contains sorted paths, hashes and lexical provider signals, never file
contents. It excludes generated trees, documentation and symlinks. Read current
deployment and runtime evidence to distinguish active services, internal agents,
development settings and retired workers. Source inspection is bounded to 2,048
files, 1 MiB per file and 32 MiB total; failure is explicit, not partial success.

## 2. See which production inputs are missing

```bash
profitctl system evaluate examples/system/condere-production.json
```

Exit 1 is expected: each missing quantity/rate has a question, acquisition source
and owner. Total cost and margin remain null. The production template includes
unallocated frontend and supporting infrastructure so they are not silently
excluded. Its simplified units must be reconciled to provider billing units and
tiers before filling prices. See the design's information-acquisition table.

Root drivers are independent values. Derived drivers sum parent quantities times
edge factors. For example, assessments → calls → tokens and calls → Job seconds.
Every numeric input needs existing cost/v1 evidence. Unknown input metadata
requires a question, source and owner. Declared service dependencies explain the
topology; driver edges, not topology alone, define the arithmetic.

## 3. Reproduce a cascading change with artificial inputs

```bash
profitctl system compare examples/system/condere-synthetic-baseline.json examples/system/condere-synthetic-cheaper.json
profitctl system compare examples/system/condere-synthetic-baseline.json examples/system/condere-synthetic-adverse.json
```

Both examples are explicitly synthetic. Baseline total $191.50; changing model
unit price alone yields $141.50. The adverse scenario also changes calls per
assessment from 10 to 25: token cost rises, Job seconds rise, and total becomes
$276.50. Job cost alone increases $60. Inspect changed inputs, driver dependencies
and per-service deltas. ProfitCtl propagates the entered call multiplier; it did
not predict that a cheaper model would require more calls.

Keep model-quality and latency evidence separate. A model swap's output length,
retry rate and accepted-dossier rate require independent measurements. Cache and
fan-out changes need stage-specific factors. Change the input file, never merely
the final dollar total, so the causal explanation remains inspectable.

## 4. Summarize existing receipts without inventing monthly usage

```bash
profitctl system receipts examples/system/condere-sanitized-receipts.json
```

This synthetic failed assessment includes one reported charge and one unknown
Exa charge. Expected exit 1, known reported cost $0.02, total null. Metadata with
`cost_status=reported` is required to classify a number as reported; absent,
`not_reported`, `estimated` or `unknown_spend=true` retains uncertainty.
Current Condere's analysis UsageReceipt stores status in metadata.

Prepare a sanitized `condere.usage-export/v1` envelope from an already authorized
export: assessment_id, outcome (succeeded/failed/cancelled/unknown), receipts. Each
receipt supports provider, model, region, input/cached/output tokens, cost_usd,
latency_ms, tool_calls and metadata. Do not include conversations or credentials.
Receipt latency is summed only when every receipt supplies it; overlapping calls mean this sum is not Job wall time. Cached token counts remain separate.

The output omits assessment IDs and arbitrary metadata and includes the input
digest. This adapter does not pull live records or modify scenarios. Do not
re-ingest both aggregate and nested per-agent receipts for the same charge.

Use representative successful, failed and cancelled samples. Map reported usage
to the corresponding driver/unit and record period, source, model version, sample
selection and confidence. A single sample's sum is neither monthly spend nor a
success-rate estimate. Never add a reported model charge to its token-derived
estimate. Incomplete prices or measurements remain unknown.

## Demo placement

Record question → source map → missing inputs → baseline → intervention →
downstream driver effects → cost comparison → limitations. Use a short recording
on the website, reproducible instructions for onboarding and the full data pack
for design-partner discussions. Obtain publication/outreach approval separately.
CTA: “Bring one AI production decision; compare its economics with ProfitCtl.”
