# Make and revisit an economic decision

Use ProfitCtl when pricing, architecture, or agent behavior could change a product decision. Codex or Claude Code inspects the project and formulates alternatives. ProfitCtl calculates the comparison and keeps a revision the other client can retrieve. These local decision tools require a build that includes `profitctl decision`. See [client setup](INSTALL.md#codex-and-claude-code).

## Start with a question

> Compare selective research with deeper research on every assessment. Identify the workload and revision, include failed attempts and review effort, and tell me what acceptance rate would justify the additional spend. Save the case as research-depth. Identify the measurement we need if the evidence is incomplete.

Compare equivalent workloads and quality requirements. Identify the environment, period, included costs and acceptance rule. A completed job, a reviewer-accepted result and a paying customer measure different things.

Use authorized, sanitized observations in the workspace. The host can use the scenario tools and the opt-in `profitctl_compare_outcomes` tool. Record a choice when the owner actually makes it.

## Calculate a threshold

```sh
profitctl system outcomes examples/system/accepted-outcomes.json
```

In this **synthetic example**, ten baseline attempts cost $30 for delivery and $12 for review. Six met the acceptance rule. The cost was $7 per accepted outcome. The candidate costs $56 for ten attempts. It would need eight accepted outcomes to match the baseline, assuming its cost per attempt holds. Candidate acceptance is unmeasured, so its unit cost is null. The CLI returns exit code 1 with the partial comparison.

Use one line per economic charge. An invoice and a provider estimate for the same usage are two sources for one charge. Mark the selected basis as `invoice`, `reported`, or `modeled`. Include failed work and identify each source. Use `unknown` with a null amount for missing charges. The result retains the known subtotal. Interpret break-even rates alongside the customer-facing quality and latency constraints.

## Save, revise and reopen

From this checkout, try a private workspace and two fictional airport-cache revisions:

```sh
workspace="$(mktemp -d)"
profitctl decision create airport-cache-demo --workspace "$workspace" --input examples/decisions/forecast.json
profitctl decision revise airport-cache-demo --workspace "$workspace" --expected 1 --input examples/decisions/observed.json
profitctl decision list --workspace "$workspace" --query airport
profitctl decision read airport-cache-demo --workspace "$workspace" --revision 1
profitctl decision read airport-cache-demo --workspace "$workspace"
```

Point both decision-enabled MCP clients at the same absolute workspace. In a fresh chat in the other client, ask:

> Find our airport cache decision. What did revision 1 predict, what does revision 2 report, and which costs are estimates?

The answer should cite both revisions and identify the synthetic example. A different edit against an outdated revision returns a conflict. Read the latest revision before reconciling the changes. The CLI uses exit code 4 for a conflict.

## Decisions from existing projects

| Decision | Relevant inputs | What to carry into the next review |
| --- | --- | --- |
| Condere research depth | Current assessment revision, research on failed and successful runs, reviewer acceptance, source quality, labor and research charges | The acceptance improvement that would pay for additional research, or the missing measurement that prevents an estimate |
| Jumpseat cache policy | Transport requests and cache hits across restarts, endpoint freshness requirements and briefing usability | Predicted versus observed requests, age of cached data, and the difference between list-price exposure and billing |
| IntelIP launch pricing | Seats, plan mix, fees, support reserve and revenue for each scenario | Which assumption reverses the modeled recommendation and which actual cohort would test it |

The [IntelIP case](../examples/decisions/intelip-pricing.json) references public scenario files and their hashes. Run from the repository root:

```sh
profitctl compare benchmark_scenarios/intelip_ops_conservative.yml benchmark_scenarios/intelip_rollout_paid_pilot.yml benchmark_scenarios/intelip_rollout_tight_free.yml benchmark_scenarios/intelip_rollout_workspace_minimum_proxy.yml --json
profitctl decision create intelip-launch --workspace . --input examples/decisions/intelip-pricing.json
```

The comparison returns exit code 1 because the conservative scenario breaches its covenant. Compare the recommendation with actual seats, paid invoices, and support effort once a comparable cohort exists. For matched forecast and actual cost observations, `profitctl diff` and `profitctl explain` show the variance and its unexplained remainder. See [variance inputs](cost-variance.md).

## Protect the record

- Choose a private workspace and share it only with clients that should see the same decisions. Both clients act with the OS account's access. Use sanitized evidence references appropriate to the host model. Keep secrets, signed URLs, transcripts and raw invoices outside cases.
- The default MCP server provides read-only scenario tools. Enable decision writes explicitly on local stdio with `--decisions`. Recording a case grants no authority to spend, publish, or change infrastructure.
- For a recorded calculation, saving verifies input files against their supplied SHA-256 digests. This identifies the bytes used. A business interpretation or authorship still requires its own evidence.
- Include the workspace in the owner's backup scope. Test recovery by reading a known revision from a restored workspace. Git exclusion protects against accidental publication. Apply OS access controls and disk encryption appropriate to the data.
