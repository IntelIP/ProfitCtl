# Design reviews in Codex and Claude Code

A proposal to move work across a service boundary can change both the bill and the failure modes. The native prompt hook supplies a comparison to the model before it answers. The stop hook can request one correction when the model introduces a consequential architecture change in its own answer. A routine edit passes through silently.

The local agent evaluates a trusted profile with the Go calculator. Both clients can also call `profitctl_review_design` with one self-contained proposal. The calculation runs offline. Model-backed investigations continue to use the selected model only when a client explicitly starts or resumes them. The service and the native hooks can remain enabled between sessions.

## Choose the designs

Start with the same business outcome and cost scope for both options. Model fan-out, duplicate attempts, billed duration, data movement, fixed costs and vendor pricing bands as relevant. Record how the high monthly volume differs from the short peak. The expected-month figure includes fixed charges. The peak-window figure gives additional usage costs for that burst. Peak arrivals, service times and available capacity determine whether the workers can clear the queue before its deadline.

For the illustrative Condere proposal the alternatives are one assessment job and a queue with a separately scheduled stage worker. Both preserve the network-blocked repository sandbox. A queue message can carry an opaque assessment reference. Give the producer permission to enqueue and the worker permission to consume. Authorize the referenced assessment at the consumer. Persist an idempotency key with the stage result so a redelivery cannot repeat a paid operation. Set a retry budget, a dead-letter path and an alert for the oldest waiting job. Validate restart recovery, confidentiality and owner acceptance with representative workloads before a production decision. These apply Security+ confidentiality, integrity and availability principles to actual design choices. [CompTIA Security+](https://www.comptia.org/en-us/certifications/security/)

Replay the example from the repository root. All performance and overhead assumptions are identified in the profile.

```sh
profitctl system design examples/design/condere-queue.json
```

The example uses Google Cloud Run Jobs' published on-demand CPU and memory rates in Iowa (us-central1). After the billing account consumes its shared free allowance, a 2-vCPU, 2-GiB job costs $0.000040 per active second before discounts. The queue's $15 monthly fee and $0.001 handoff fee are assumptions, as are all durations, volumes and worker limits. The comparison covers the compute and queue charges that differ between these two options. [Google Cloud Run pricing](https://cloud.google.com/run/pricing)

| Modeled workload | One job | Split with queue |
| --- | ---: | ---: |
| 200 attempts in a month | $8.16 | $20.48 |
| 2,000 attempts in a month | $81.60 | $69.80 |
| 60 arrivals in one hour, added usage | $2.448 | $1.644 |

The queue first saves money at 1,120 whole attempts per month. Four additional billed minutes in the finalizer move that threshold to 3,948. At 2,000 attempts, the queue then costs $89.00 and the single job $81.60. At 60 arrivals per hour, the slower finalizer needs 11 workers on average against 10 slots assumed available, so sustained arrivals would build a backlog. Measure billed worker time, retries and completion latency before changing Condere.

## Activate on macOS

Use the frozen Go executable and Python environment from the same build. The example below uses the checkout environment from the [standalone agent setup](standalone-investigations.md). For an installation that survives checkout updates, substitute the immutable release paths. Set `workspace` to the existing private investigation workspace and `project_root` to the Condere checkout used by your clients.

```sh
runtime="$HOME/Library/Application Support/ProfitCtl/design"
python_bin="$PWD/agent/.venv/bin/python"
agent_bin="$PWD/agent/.venv/bin/profitctl-agent"
profitctl_bin="$PWD/build/client-bin/profitctl"
install -d -m 700 "$runtime" "$runtime/profiles" "$runtime/backups"
install -m 600 examples/design/condere-queue.json "$runtime/profiles/condere-queue.json"
"$python_bin" scripts/design-launch-agent.py install --runtime "$runtime" \
  --workspace "$workspace" --profitctl-binary "$profitctl_bin"
```

Create an owner-controlled `routes.json` in that runtime directory. Use absolute project roots. Match both the workload and the proposed change. Set `infer_project` only for directories dedicated to that project. New projects can send their own inline proposals or add an operator-reviewed profile and route.

```sh
"$python_bin" - "$runtime/routes.json" "$project_root" <<'PY'
import json, os, sys
from pathlib import Path
os.umask(0o077)
route = {"profile": "condere-queue", "project": "Condere",
         "roots": [str(Path(sys.argv[2]).resolve(strict=True))], "infer_project": True,
         "workloads": ["report assembly", "assessment finalizer", "assessment finalization"],
         "topics": ["queue", "queues", "message broker"]}
with open(sys.argv[1], "x") as stream:
    json.dump({"version": 1, "routes": [route]}, stream)
PY
"$python_bin" -m profitctl_agent.design_hook_config install \
  --codex-config "$HOME/.codex/hooks.json" --claude-config "$HOME/.claude/settings.json" \
  --python "$python_bin" --socket "$runtime/agent.sock" --routes "$runtime/routes.json" \
  --backup-dir "$runtime/backups"
```

Register or update each client's `profitctl-agent` entry to launch `"$agent_bin" --workspace "$workspace" --profitctl-binary "$profitctl_bin" mcp --socket "$runtime/agent.sock"`. The bridge reaches the running process through a private Unix socket. Keep the supported model setting in the service, and use the existing workspace so investigations retain their history. In Codex, review the two new ProfitCtl commands in the native hook trust interface. Open a fresh client session to load the settings. [Codex hooks](https://developers.openai.com/codex/hooks/) · [Claude Code hooks](https://code.claude.com/docs/en/hooks)

## Operate and verify

In the approved Condere checkout, submit the following ordinary design request to each client:

> Should Condere split the report assembly stage from the assessment job and put it behind a queue with a separate worker? Recommend a design and explain its operational consequences.

Check that the answer gives the two designs at expected volume and at the stated peak, the volume where the cost ranking crosses, and the sensitivity that reverses the recommendation. The controls protecting the repository and preventing duplicate paid work still matter when costs favor the queue. If a proposal concerns a different workload, an inline review needs that workload's data.

```sh
"$python_bin" scripts/design-launch-agent.py status --runtime "$runtime"
```

The service is private to the local OS account. macOS launchd starts it at user sign-in and restarts it on failure. Closing both clients leaves it running. Sleep, shutdown and logout suspend or end this user service. The native prompt and stop hooks fire on client events. [Apple launchd lifecycle](https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html)

To remove the integration, run the same hook configuration command with `uninstall`, restore the previous client MCP launch arguments, and run `"$python_bin" scripts/design-launch-agent.py uninstall --runtime "$runtime"`. Investigation records remain available through the standalone CLI.
