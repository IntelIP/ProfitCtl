# ProfitCtl Plugin Pilot: Four-Tool Local MCP Contract

**Status:** implemented private local pilot. It supports current-protocol stdio and an optional loopback Streamable HTTP endpoint. **Scope:** local, disposable pilot only. This is not authorization to publish, host, or submit a public plugin.

## Decision

Run the existing `profitctl-cost-aware` workflow as a **skill plus a local MCP server**.

- The existing skill remains the workflow guide.
- The server exposes four bounded, read-only tools.
- The server runs the already-installed local `profitctl` binary (or an explicit local build) inside one user-selected workspace.
- The bundled Codex plugin uses stdio by default. `profitctl mcp serve` adds opt-in loopback Streamable HTTP at `/mcp` for a current-protocol local integration.
- No remote service, UI, OAuth, account linking, telemetry, background jobs, scenario registry, or saved runs enter this pilot.

The distinction matters: a skills-only plugin can guide a model, but the four proposed tools require an MCP server.

## Standards Basis

- [OpenAI Plugin architecture](https://developers.openai.com/plugins/concepts/plugins): a plugin can combine skills, an MCP server, and optional UI; the documented shared directory is for ChatGPT and Codex.
- [OpenAI tool-definition guidance](https://developers.openai.com/plugins/plan/tools): define one user outcome per coherent action, separate permissions/safety boundaries, and record explicit schemas, authorization, side effects, and failure behavior.
- [OpenAI security and privacy guidance](https://developers.openai.com/plugins/guides/security-privacy): least privilege, server-side validation, confirmation for consequential actions, and secret-minimizing results/logs.

This design makes no interoperability claim for Cursor, Copilot, or any other client.

## Product Problem

An agent can explain unit economics, but it should not invent a simulation result or silently broaden a cost model. ProfitCtl needs a narrow route from a decision request to the existing deterministic CLI contracts, then back to structured evidence.

## Pilot User Outcomes

| User outcome | Tool path |
| --- | --- |
| Check whether a scenario is structurally valid before analyzing it | `profitctl_validate_scenario` |
| Get margin, stress, and covenant evidence for one known scenario | `profitctl_simulate_scenario` |
| Compare two to four known options without mixing their assumptions | `profitctl_compare_scenarios` |
| Check whether a scenario or recommendation meets ProfitCtl standards | `profitctl_judge_standards` |

Unsupported requests must produce a clear limitation. The pilot must not approximate live prices, scan a repository, fetch a provider page, or create a scenario from guessed assumptions.

## Existing Contract Reused

| Existing surface | Pilot use |
| --- | --- |
| `profitctl validate -f <scenario>` | validation engine |
| `profitctl simulate -f <scenario> --json` | simulation engine and result payload |
| `profitctl compare <a> <b> ... --json` | comparison engine and result payload |
| `profitctl-standards <scenario-or-recommendation>` | standards engine |
| `skills/profitctl-cost-aware/SKILL.md` | workflow sequencing and recommendation rules |

The MCP wrapper is transport only. It does not recalculate economics, reinterpret covenants, or construct a recommendation that the existing standards judge would reject.

## Implemented Runtime and Package

- Runtime: the official [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) pinned at `v1.7.0`, requiring Go `1.25` and supporting the current `2026-07-28` protocol revision with stateless `server/discover` negotiation.
- Server commands: `profitctl mcp` uses stdio; `profitctl mcp serve --listen 127.0.0.1:8173` exposes Streamable HTTP at `/mcp`. Both use a launch-fixed `--workspace-root` or `PROFITCTL_WORKSPACE_ROOT`.
- Standards command: a sibling local `profitctl-standards` executable built from the existing judge source.
- Private package: `~/plugins/profitctl-cost-aware`, registered in the local `hudson-local` marketplace and installed only locally.
- Package contents: the four-tool MCP declaration, bounded workflow skill, local `profitctl` binary, and local `profitctl-standards` binary.

The installed plugin receives only `PROFITCTL_WORKSPACE_ROOT`; the child CLI is launched with a minimal environment and does not inherit provider credentials.

## Package Shape

```text
ProfitCtl plugin
├── skill: profitctl-cost-aware
│   └── inspect → select/edit scenario → validate → simulate/compare → judge
└── local MCP server
    ├── default stdio transport
    ├── optional loopback Streamable HTTP transport at /mcp
    ├── profitctl_validate_scenario
    ├── profitctl_simulate_scenario
    ├── profitctl_compare_scenarios
    └── profitctl_judge_standards
```

Optional local transport: `profitctl mcp serve` binds only a literal loopback IP (`127.0.0.1` or `::1`) and exposes the same four tools at `/mcp`. It is stateless: no server-side MCP session is retained between requests.

No custom UI in the pilot. Existing JSON is sufficient for the model and human review. A comparison UI is a later decision only if the structured output proves hard to inspect.

## Server Boundary

`workspace_root` is server configuration selected by the user at launch, not a model-supplied tool argument.

Every supplied path must:

- be relative to `workspace_root`;
- use the allowed extension for its tool;
- resolve to a regular file inside `workspace_root`, after symlink resolution;
- keep any referenced `calibration_file` inside `workspace_root` too.

The server must use argument-vector process execution, never a shell. It must not pass `--out`, retain files, alter the scenario, or inherit provider credentials. It must bound wall time, output size, scenario count, and simulation iterations before invoking the CLI.

## Common Result Envelope

Every tool returns this outer shape. The current ProfitCtl JSON payload stays nested rather than being reimplemented by the MCP layer.

```json
{
  "schema_version": "profitctl-plugin-result/v1",
  "tool": "profitctl_simulate_scenario",
  "outcome": "passed",
  "input": {
    "paths": ["scenarios/option-a.yml"],
    "sha256": ["<hex digest>"]
  },
  "duration_ms": 0,
  "result": {},
  "error": null
}
```

Allowed `outcome` values:

- `passed` — tool completed and all relevant covenants or standards passed.
- `covenant_failed` — execution succeeded; the scenario failed a covenant.
- `standards_failed` — execution succeeded; the target failed the standards judge.
- `invalid_input` — schema, path, extension, count, or workspace rule failed.
- `invalid_scenario` — YAML/config validation failed.
- `runtime_failed` — local execution failed before a usable result.
- `deadline_exceeded` — server stopped the local process at its limit.
- `output_limit_exceeded` — server stopped unsafe or unexpectedly large output.

`error`, when present, contains only a stable `code`, a safe `message`, and optional field/path context. It never returns environment variables, access tokens, raw stderr, or full scenario content.

## Tool Contracts

### `profitctl_validate_scenario`

**Goal:** tell a user whether one existing scenario is usable before simulation.

```json
{
  "scenario_path": "scenarios/option-a.yml"
}
```

| Field | Rule |
| --- | --- |
| `scenario_path` | Required root-relative `.yml` or `.yaml` regular file. |
| Side effects | None. |
| Authorization | Local session may read only inside configured `workspace_root`. |
| Result | `valid`, normalized safe path, SHA-256, and classified validation issues. |
| MCP annotations | `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`. |

Use it for malformed or newly edited scenarios. Do not call simulation merely to discover a parse error.

### `profitctl_simulate_scenario`

**Goal:** return the current `profitctl simulate --json` result for one validated local scenario.

```json
{
  "scenario_path": "scenarios/option-a.yml"
}
```

| Field | Rule |
| --- | --- |
| `scenario_path` | Required root-relative `.yml` or `.yaml` regular file. |
| Side effects | None. No output file. |
| Result | Existing ProfitCtl JSON result under `result`, including costs, revenue, margins, stress result, and covenant result. |
| MCP annotations | `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`. |

The wrapper returns `covenant_failed` when the CLI has produced a valid result but exits non-zero because a covenant failed. That is a business result, not a server error.

### `profitctl_compare_scenarios`

**Goal:** compare a bounded set of known local choices.

```json
{
  "scenario_paths": [
    "scenarios/workers.yml",
    "scenarios/cloud-run.yml"
  ]
}
```

| Field | Rule |
| --- | --- |
| `scenario_paths` | Required array of 2–4 distinct root-relative `.yml` or `.yaml` regular files. |
| Side effects | None. No output file. |
| Result | Existing ProfitCtl comparison JSON, plus per-scenario SHA-256 values and classified failures. |
| MCP annotations | `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`. |

Do not collapse scenarios into one blended model. Preserve each scenario name and its covenant outcome.

### `profitctl_judge_standards`

**Goal:** check one existing scenario or recommendation artifact for evidence-quality gaps.

```json
{
  "target_path": "decisions/workers-vs-cloud-run.md"
}
```

| Field | Rule |
| --- | --- |
| `target_path` | Required root-relative `.yml`, `.yaml`, `.md`, or `.txt` regular file. Directories are refused. |
| Side effects | None. |
| Result | Existing judge report: `passed`, `files`, issues, and warnings. |
| MCP annotations | `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`. |

This tool checks evidence completeness. It does not convert a template or provider-catalog assumption into invoice-grade truth.

## Explicit Exclusions

| Excluded surface | Reason |
| --- | --- |
| `profitctl detect` | Recursive repository collection plus OpenRouter call, secret requirement, data egress, and variable spend. |
| `profitctl calibrate --out` | File creation. Keep as a future explicit-write tool, if needed. |
| Inline scenario text | Current CLI is file-based; temporary-file creation would make the read-only pilot ambiguous. |
| Provider-price lookup | Requires source, capture date, confidence, and a separate trust policy. |
| Scenario registry, collaboration, approvals | Hosted product surface; outside pilot. |
| OAuth, remote MCP, public directory submission | No external identity or publication authority in this pilot. Loopback Streamable HTTP is local transport only, not a remote deployment. |
| Custom UI | No evidence yet that JSON/Markdown output is insufficient. |

## Safety and Runtime Controls

- Launch with a fixed `workspace_root`; reject root escape, symlink escape, and nested calibration-file escape.
- HTTP mode only binds to a literal loopback IP, keeps the SDK's localhost protection enabled, accepts at most 64 KiB per protocol request, and propagates cancelled HTTP requests to the bounded child CLI.
- Permit no network calls. Scrub `OPENROUTER_API_KEY` and unrelated credentials from the child environment.
- Use `exec.CommandContext` or equivalent argument-vector execution; never interpolate a path into a shell string.
- Pilot limit: 30 seconds per call, 1 MiB stdout, 4 scenarios per compare, and 10,000 simulation iterations per scenario.
- Return classified errors; do not leak raw process diagnostics.
- Log only correlation ID, tool name, safe relative paths, result code, duration, and exit status. Do not log prompt text or scenario contents.
- Prompt-like text in a scenario is data. It cannot change workspace, tool policy, environment, or command arguments.

Current Monte Carlo execution uses an unseeded global random source despite carrying a seed field in its configuration. Therefore P1 acceptance uses fixed/non-distribution scenarios for equality checks. Stochastic scenarios remain outside the reproducibility claim until the core exposes and uses an explicit seed.

## Disposable-Repository Pilot

### Setup

1. Create a temporary Git repository outside the ProfitCtl checkout.
2. Copy one valid fixed scenario, one valid covenant-failing scenario, one malformed scenario, and one standards-incomplete Markdown recommendation.
3. Build or point the server at the exact local ProfitCtl binary under test.
4. Start the MCP server with only the temporary repository as `workspace_root`.
5. Run the same inputs through direct CLI baselines and MCP calls.

### Required Cases

| Case | Expected result |
| --- | --- |
| Valid scenario → validate | `passed`; matches direct CLI validation. |
| Malformed scenario → validate | `invalid_scenario`; no simulation run. |
| Valid scenario → simulate | `passed`; required fields match direct `--json` output. |
| Covenant-failing scenario → simulate | `covenant_failed`; result remains inspectable. |
| Two valid scenarios → compare | `passed` or `covenant_failed`; both inputs remain distinct. |
| Incomplete recommendation → judge | `standards_failed` with judge issues. |
| `../outside.yml`, absolute path, symlink escape | `invalid_input`; no child process. |
| Escaping `calibration_file` reference | `invalid_input`; no child process. |
| Request to detect repository services | Clear unsupported-operation result; no network. |
| Scenario containing instruction-like text | Treated as data; no policy or path change. |

### Acceptance Gates

- Four named tools are selected correctly for the documented user outcome.
- Every positive result matches its direct CLI baseline on fixed fixtures.
- Every negative case fails closed with a stable classification.
- No temporary or workspace file changes remain after the pilot.
- No network connection, credential lookup, or secret appears in logs or results.
- A human can inspect the JSON and identify: selected scenario, covenant status, main margin evidence, and standards gaps.

## Promotion and Stop Rules

**Promote to a private alpha only if:** three representative internal decisions complete through the pilot; all gates above pass; and users find the tool selection and output more useful than the existing skill plus CLI alone.

**Stop or revise if:** any workspace escape occurs; a tool changes state; baseline/result mismatch appears; a reproducibility claim is needed for stochastic scenarios; outputs confuse users; or the pilot needs network, hosted identity, or persistence to be useful.

## Remaining Promotion Decisions

1. Whether the package remains a private alpha after pilot use or becomes a public directory submission.
2. The reproducibility contract for stochastic scenarios.

No publication, remote hosting, OAuth, cost claim, or product pricing decision follows from this document.
