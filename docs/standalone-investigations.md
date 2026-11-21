# Start in Codex and resume in Claude Code

The standalone `profitctl-agent` runs LangChain's Deep Agents SDK. It uses ProfitCtl's Go calculator and saved decisions. A run ID identifies the investigation and its checkpoint. The case ID identifies the business decision and its revisions. Both clients must launch the agent against the same private workspace.

## Install from this checkout

The local runtime requires a POSIX host such as macOS or Linux, Python 3.12 or 3.13, `uv`, and the Go version in `go.mod`. Run the following commands at the repository root.

```sh
uv sync --frozen --project agent
agent_bin="$PWD/agent/.venv/bin/profitctl-agent"
mkdir -p build/client-bin
go build -o build/client-bin/profitctl .
profitctl_bin="$PWD/build/client-bin/profitctl"
"$agent_bin" --help
```

For an existing installation, set `workspace` to the absolute path of the private directory containing your saved cases and pinned inputs. For a fresh trial, create an isolated workspace and seed its first revision from a public, illustrative Condere case.

```sh
workspace="$(mktemp -d)"
mkdir -p "$workspace/examples/system"
cp examples/system/accepted-outcomes.json "$workspace/examples/system/"
"$profitctl_bin" decision create condere-research-depth --workspace "$workspace" \
  --input examples/decisions/condere-research-depth.json
```

Choose the inference mode explicitly for the agent process.

| `PROFITCTL_AGENT_MODEL` | Use |
| --- | --- |
| `scripted` | Rehearse the workflow with a deterministic model and real calculations, decisions and checkpoints. |
| `claude-subscription` | The official local Claude CLI returns structured model output in restricted mode. ProfitCtl executes the tools and owns the checkpoint. Preflight supports the default macOS Keychain profile and requires a first-party Claude Pro or Max subscription with paid extra usage disabled and available included quota. |
| `openai:gpt-…` | Select an approved first-party OpenAI model ID and supply its API credential and budget through the operator's runtime environment. |

The commands below use `scripted` so the first run can be completed offline. For a model-driven run on the local Claude subscription, select `claude-subscription` after disabling extra usage in your Claude account. Preflight checks the subscription before inference. Live inference receives the case context and calculator results. Codex and Claude Code can both start and resume these investigations. Configure them with the same model mode and workspace.

If Claude rate-limits the account check, ProfitCtl can retry that read once after a provider-requested delay of up to 10 seconds. For a longer delay or another HTTP 429, follow the wait time in the error and retry the same run ID. A missing or unusable retry time prompts a manual retry after at least one minute. Every model call requires a fresh successful check.

## Connect both clients

```sh
model=scripted
codex mcp add profitctl-agent -- /usr/bin/env -i "HOME=$HOME" "PATH=$PATH" LANG=C "PROFITCTL_AGENT_MODEL=$model" \
  "$agent_bin" --workspace "$workspace" --profitctl-binary "$profitctl_bin" mcp
claude mcp add --scope local profitctl-agent -- /usr/bin/env -i "HOME=$HOME" "PATH=$PATH" LANG=C "PROFITCTL_AGENT_MODEL=$model" \
  "$agent_bin" --workspace "$workspace" --profitctl-binary "$profitctl_bin" mcp
```

Run Claude's registration in the project where you will open its chat. Start fresh client sessions to load the server. The existing `profitctl` server can continue to serve direct calculation and decision tools. The `profitctl-agent` server supplies three additional operations.

| Operation | What to provide |
| --- | --- |
| `profitctl_start_investigation` | A new stable `run_id`, existing `case_id`, and plain-language `question`. Reuse the same run ID and inputs if the client loses the reply. |
| `profitctl_investigation_status` | The exact `run_id` returned to the operator. This reads the durable checkpoint without invoking a model. |
| `profitctl_resume_investigation` | The same `run_id`, its `base_revision` as `expected_revision`, and an `action` of `supply` or `decline`. A supplied observation includes its source, kind, and updates. |

## Run the handoff

1. In Codex, ask **“Start a ProfitCtl investigation named `condere-depth-01` for the saved case `condere-research-depth`. Should every assessment use deeper research? Show the baseline and tell me what you need from me.”**
2. Keep the run ID from the reply. When the agent asks for evidence, close the Codex session and its stdio server. Your new Claude process will reopen the checkpoint from disk.
3. In Claude Code, ask **“Open ProfitCtl investigation `condere-depth-01`. Show the original calculation and the pending evidence request. Wait for my answer.”**
4. Supply a sourced observation for the matching cohort, including the exact option name, or explicitly decline. Ask Claude to resume the same run and show its result, outstanding uncertainties and decision revision.

The CLI reaches the same workspace.

```sh
"$agent_bin" --workspace "$workspace" --profitctl-binary "$profitctl_bin" status condere-depth-01
```

For a supplied response, pass `resume RUN_ID --expected-revision N --action supply --evidence-file evidence.json`, replacing `N` with the reported `base_revision`. An explicit decline uses `--action decline` without evidence. The MCP tool accepts the JSON object directly. Obtain counts, source identity, cost basis and charge IDs from the operator and the matching baseline. The example below uses an explicit assumption for the public trial.

```json
{
  "kind": "assumed",
  "source": "Sensitivity assumption for the public ten-assessment fixture",
  "updates": [{"option": "Research every request", "accepted": 8}]
}
```

For the trial, eight assumed acceptances at $56 total model $7 per accepted outcome. A real decision needs measured acceptance and attributable costs for its own cohort.

If the connection drops, ask either client for the same run's status. An interrupted run shows the owner response already saved, the pending step, and whether publication has begun. Ask the client to use the returned `retry_tool_call`. It contains the saved arguments, so the owner can continue by run ID alone.

For model-backed steps longer than Codex's default 60 seconds, set the server's `tool_timeout_sec` to the allowed run duration in the `[mcp_servers.profitctl-agent]` section of your Codex configuration. When the saved case has changed independently, inspect the newer case before starting a new investigation with another ID. [Codex MCP configuration](https://developers.openai.com/codex/mcp/)
