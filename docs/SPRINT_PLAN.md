# ProfitCtl Versioned Sprint Plan

Updated: 2026-10-07

Status: Proposed delivery backlog. No implementation, dates, staffing, spending, or release publication is committed by this document.

## Product goal

Help finance, product, and engineering choose changes that improve product economics while preserving the quality, speed, and reliability customers need.

Build from the current calculator into a system that connects **engineering change → workload behavior → operating cost → customer outcome → financial result**. Start with one Condere assessment workflow. Add autonomous investigation only after the calculations and experiments are trustworthy.

This plan turns the [Product Development Roadmap](ROADMAP.md) into ten proposed sprints, twenty user stories, and seventy tasks. Each sprint targets a distinct release. The [Open-Core Roadmap](OPEN_CORE_ROADMAP.md) continues to govern distribution, adoption, and commercial packaging.

## Current state and release boundary

| Evidence | Current finding | Planning consequence |
| --- | --- | --- |
| [Latest stable release](https://github.com/IntelIP/ProfitCtl/releases/tag/v0.3.0) | Confirmed: `v0.3.0`, published 2026-08-05 | This is the stable starting point. |
| [Current prerelease](https://github.com/IntelIP/ProfitCtl/releases/tag/v0.4.0-rc.1) | Confirmed: evaluator prerelease, published 2026-09-13 | Finish this release before adding another capability layer. |
| Current source and local CLI inspection | `system inspect`, `evaluate`, `compare`, and `receipts` exist. Explicit assumptions drive calculations; unknown costs remain unknown. | Extend the existing Go engine and interfaces. A rewrite is outside scope. |
| Production example inspected during roadmap work | Condere example returned `incomplete`, with 19 missing inputs and null total cost and revenue. | Reconcile real inputs before making a financial recommendation. This count is a snapshot, not a permanent test oracle. |
| Synthetic scenario checks | Baseline `$191.50`; lower model price with unchanged calls `$141.50`; lower price with more calls `$276.50`. | Use these as calculation regressions. They do not prove real savings. |
| [Historical modeling backlog](profitctl-bug-backlog-from-intelip-modeling.md) | Contains older bug and calibration statuses that have not all been revalidated here. | S01 reproduces relevant cases before creating fixes. Do not reopen resolved bugs from an old note. |
| AgentShift, reconnected 2026-10-07 | PCTL-19 (normalized ledger) and PCTL-29 (system economics) are recorded Done. PCTL-21 and its investigation tasks contain active work. | Extend the existing contracts and calculations; preserve active work. Tracker completion alone does not establish a new release. |

**Planned v1.0.0 boundary:** one organization can take one representative workflow from sourced economics through a controlled experiment, an agent-assisted recommendation, a recorded human decision, and comparison with subsequent actual results. AWS, Azure, and local hosting are modeled alternatives; claims of measured deployment performance require observations from those environments.

The release does not promise an accounting system, general causal inference, autonomous production changes, automatic cloud migration, every enterprise connector, or an unrestricted agent swarm. Wider hosted service and commercial commitments remain separate decisions.

## Delivery assumptions

- **Cadence:** propose two-week planning windows. Ten windows would be a twenty-week planning envelope only if capacity and dependencies support it; this is not a delivery forecast. Set dates at sprint kickoff and reforecast from observed throughput.
- **Capacity:** propose one implementation owner with agent assistance, a tech lead as final integrator, and available product/finance reviewers. Names and availability are unassigned. Hudson Aikins is the proposed product decision owner; that does not assign engineering work.
- **Work in progress:** at most three started stories, including review, validation, demo, and blocked started work. This plan normally pulls one story at a time. Backlog creation does not start work.
- **Task size:** each task below has one bounded result for one focused session. If it cannot fit, split it before starting and retain its parent acceptance criteria. A story may span several sessions.
- **Planning state:** all stories start in `Refine`, with unmet dependencies shown below. No story is `Ready` until its owner, fixtures, decisions, and integration order are established.
- **Release policy:** every version is a proposed target. A failed gate keeps the target open; it does not create a release or silently move unfinished work to the next sprint. Use another prerelease when appropriate.
- **Scope control:** the two stories and release task are the sprint scope ceiling. At kickoff, commit only what fits available capacity. Later sprints may be split and renumbered before commitment.

## Release sequence

| Sprint | Target version | Roadmap stage | New result available to users | Depends on |
| --- | --- | --- | --- | --- |
| [S01](#s01) | **v0.4.0** | Existing foundation | A dependable local economics release with honest limits | Existing prerelease |
| [S02](#s02) | **v0.5.0** | 1. Measured economics | Usage and invoices reconcile into an inspectable cost ledger | S01 |
| [S03](#s03) | **v0.6.0** | 1. Measured economics | One workflow has a sourced cost per accepted outcome and broader economics | S02 |
| [S04](#s04) | **v0.7.0** | 2. Change impact | Proposed changes become traceable financial scenarios | S03 |
| [S05](#s05) | **v0.8.0** | 3. Tested alternatives | A baseline and candidate can be tested in bounded sandboxes | S04 |
| [S06](#s06) | **v0.9.0** | 3. Tested alternatives | Teams compare hosting and scaling options on equivalent requirements | S05 |
| [S07](#s07) | **v0.10.0** | 4. Coordinated investigations | One agent can work through a saved case and resume after interruption | S06 |
| [S08](#s08) | **v0.11.0** | 4. Coordinated investigations | Bounded specialists can investigate in parallel and combine evidence | S07 |
| [S09](#s09) | **v0.12.0** | 5. Continuous improvement | Forecasts are checked against actuals and material drift is surfaced | S08 |
| [S10](#s10) | **v1.0.0** | Controlled enterprise pilot | A team can complete and review the full decision cycle | S09 |

## Contract inherited by every story and task

**Authority and ownership.** Each story belongs to this plan, its named sprint, and the linked roadmap stage. The story's owner is unassigned in the stated role; its target is the sprint, not an invented date. Tasks inherit the story's purpose, acceptance source, owned surface, exclusions, risk, and stop conditions. Human-decision tasks require a named human at kickoff. Technical execution may be assigned to an agent only after its bounds are clear.

**Dependencies.** Within each story, task 1 precedes task 2, which precedes task 3. Additional story dependencies are explicit. A successor consumes validated predecessor outputs; it must not invent them. Parallel development requires independent surfaces, isolated branches/worktrees, and one named integrator. No branches or worktrees are created by this plan. When implementation is authorized, inspect registered worktrees, use one story per branch/PR where practical, and integrate in dependency order.

**Invariants.** Preserve existing gross-margin semantics, source provenance, units, billing periods, and explicit unknowns. Add delivery, human effort, productization, adoption, and migration costs as separate named views without double counting. An estimate, a simulation, a sandbox measurement, and a billed actual must remain distinguishable. An agent cannot approve its own consequential action. Neither context separation nor a Git worktree is a security sandbox.

**Testing.** Each story requires the direct tests stated below, relevant regression tests, and a repeatable user demo. Add behavior to the existing validation profiles rather than creating a parallel proof system. For product changes, use `validate-product` and `tabellio-validate gate` with [the committed validation manifest](../tabellio.validation.json). Bind results to the exact candidate commit and input/fixture versions. Required static, schema, semantic, workflow, operational, and security checks must pass. Unit tests alone do not establish the business outcome.

**Cost boundary.** Repository validation stays offline and at zero provider spend, as required by [AGENTS.md](../AGENTS.md). Paid experiments, live credentials, production data, deployment, billing changes, publication, and merge require their own explicit authority. A live pilot, if authorized, runs separately from repository validation under a recorded budget. Unknown required spend or missing live evidence blocks the associated claim; an offline replay cannot stand in for it.

**Evidence.** Preserve the candidate identity, fixture or source digest, commands, expected and actual outcomes, `passed`/`failed`/`blocked` results, output artifact, duration, cost/resource record, human acceptance where required, and cleanup result. Keep generated validation evidence in the established CI/artifact location, not committed to source. Use sanitized records and stable artifact links. All product tests in this plan are currently **not run**; this planning document is not release evidence.

**Stop conditions.** Stop a task when its declared result and direct check exist. Stop and mark blocked when required source data, authority, ownership, budget, or test oracle is missing. Return unresolved product intent to refinement. A failed check remains failed until fixed and rerun on the resulting candidate. Do not widen the scope to unrelated improvements.

**Release task.** Each sprint has a distinct `Snn-R` manual task owned by the unassigned release lead. It depends on both stories. Freeze the integrated candidate, run required checks, perform the sprint demo, and prepare versioned release notes covering capability, limitations, compatibility, and rollback. Record whether publication is authorized and, only after actual publication, the release URL and artifact identity. Close the sprint by classifying every committed item as done, blocked, or cancelled; record one learning and the next decision. Release preparation and publication are separate states.

**Success thresholds.** Arithmetic fixtures have explicit expected answers below. Business targets for quality, latency, forecast error, human effort, and acceptable return are decisions, not invented numbers. Record their values and measurement method before the relevant experiment or acceptance run. Missing required thresholds block that evaluation.

<a id="s01"></a>
## S01 — v0.4.0: Make the existing foundation dependable

**Goal:** an evaluator can install ProfitCtl, compare scenarios, and see exactly where evidence is missing.

**Start:** inspect the current prerelease, source, and existing release requirements. **Terminal result:** a validated v0.4.0 candidate and repeatable installation/demo record. **Scope boundary:** existing local CLI and MCP companion; no new agent runtime or live billing connector. **Learning outcome:** understand what the current deterministic engine can and cannot establish.

<a id="s01-a"></a>
### S01-A: Trust the existing economics calculations

**Story:** As a product owner, I need current calculations and unsupported inputs to behave consistently, so that I can use ProfitCtl without mistaking a modeling gap for a financial conclusion.

**Why now:** later predictions inherit any error in the foundation. **Success:** existing supported examples reproduce; relevant historical defects are classified against current behavior; incomplete inputs remain visible.

**Work type:** implementation. **Owner:** unassigned engine maintainer. **Target:** S01. **Risk:** high financial correctness. **Blocked by:** no prior sprint; kickoff ownership and fixtures still required.

**Sources and surface:** [system economics design](design/system-economics-v1.md), [system examples](../examples/system/README.md), and [historical modeling backlog](profitctl-bug-backlog-from-intelip-modeling.md). Own existing calculation/parser/CLI regression paths in `internal/systemmodel`, `internal/pricing`, `internal/config`, and `cmd`; exclude new pricing products and unrelated cleanup.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S01-A1 | research: reproduce monthly-margin, unlimited-plan, and currently supported pricing examples; record fixed/open/out-of-scope findings | Each finding has current command/input/output evidence; an old status alone cannot create a fix. |
| S01-A2 | implementation: repair only reproduced release-blocking calculation or parser inconsistencies | Independent expected arithmetic and valid/invalid input cases pass. If no defect reproduces, retain the existing behavior and record that result. |
| S01-A3 | implementation: bind scenario comparison and unknown-input regressions into existing validation | Synthetic totals reproduce `$191.50`, `$141.50`, and `$276.50`; removing a required input produces `incomplete`, not zero or fabricated savings. |

**Acceptance/demo:** given a known baseline, an evaluator sees the effect of a cheaper model both with unchanged calls and with increased calls. Given an unknown rate, the CLI names the missing input and does not present a complete total. Demo artifact: command transcript plus machine-readable outputs. Existing pricing fixtures remain compatible; unsupported cases fail with actionable explanations.

<a id="s01-b"></a>
### S01-B: Install and use the candidate as an evaluator

**Story:** As an evaluator, I need a working install and one reproducible economics walkthrough, so that I can assess the product without maintaining its source tree.

**Why now:** a useful release must work through its delivery surface. **Success:** supported packages exercise the same core behavior; the MCP companion is included where promised; limitations match observed behavior.

**Work type:** manual-task. **Owner:** unassigned release maintainer. **Target:** S01. **Risk:** medium delivery/compatibility. **Blocked by:** S01-A.

**Sources and surface:** [current prerelease](https://github.com/IntelIP/ProfitCtl/releases/tag/v0.4.0-rc.1), [installation guide](INSTALL.md), and existing release workflows. Own packaging checks and evaluator instructions; exclude deploying the website or publishing packages during preparation.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S01-B1 | manual-task: reconcile current package contents and supported-platform claims with the release workflow | Every advertised target has an executable smoke-test path; unavailable target evidence stays blocked. |
| S01-B2 | implementation: complete a clean-install CLI/MCP walkthrough using local fixture files | Installed version, `system evaluate`, `compare`, receipt import, and an allowed MCP read work; an out-of-root path is rejected. |
| S01-B3 | manual-task: prepare the evaluator guide and compatibility notes from observed results | A second operator can follow the guide without private setup; remove an input and confirm the documented failure path. |

**Sprint testing objective:** prove installation and honest calculation behavior on the integrated candidate, including package integrity and existing validation profiles. **Sprint demo:** install → calculate known example → compare → inspect incomplete Condere inputs. **Release task S01-R:** prepare **v0.4.0**; do not call the prerelease stable before publication.

<a id="s02"></a>
## S02 — v0.5.0: Reconcile observed usage and cost

**Goal:** finance and engineering can inspect where a workflow's recorded costs came from and what remains unmatched.

**Start:** S01 accepted; a sanitized receipt and invoice fixture contract exists. **Terminal result:** a repeatable ledger and reconciliation report. **Scope boundary:** file imports first; no live billing credentials or accounting integrations. **Learning outcome:** turn agent files into sourced business evidence.

<a id="s02-a"></a>
### S02-A: Collect usage without losing its meaning

**Story:** As an engineering manager, I need receipts normalized with their original identity and context, so that retries, failed work, and duplicated exports do not distort workflow costs.

**Why now:** actual economics needs reliable observations. **Success:** records retain source, units, period, currency, and status; import is repeatable; unknown charges remain explicit.

**Work type:** implementation. **Owner:** unassigned data/engine maintainer. **Target:** S02. **Risk:** high accounting correctness. **Blocked by:** S01-B.

**Sources and surface:** [roadmap Stage 1](ROADMAP.md#stage-1-measured-economics), existing `internal/systemmodel/receipts.go`, and the PCTL-19 cost/ledger foundation. Extend existing receipt normalization and import schemas; exclude a second ledger implementation, full FOCUS ingestion, and arbitrary provider adapters in this slice.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S02-A1 | prototype: map Condere receipts onto existing cost/ledger contracts; specify only missing source/run/attempt identity, period, units, currency, charge status, and outcome fields | Reuse PCTL-15's calibration spike. Contract fixtures accept supported receipts and reject malformed units, unsupported currency conversion, or missing identity. |
| S02-A2 | implementation: adapt the existing sanitized Condere export and deduplicate repeated records | Import the same export twice: charge count and known subtotal remain unchanged; conflicting content under the same identity is reported. |
| S02-A3 | implementation: preserve failed, cancelled, retried, and unknown attempts in summaries | Fixtures retain all incurred charges; missing charge is unknown rather than free; unsupported records have an actionable rejection report. |

**Acceptance/demo:** import a file, import it again, then inspect a failed attempt and an unknown charge. Expected: stable totals and visible provenance. Artifact: normalized records and import/rejection summary. No raw prompts, secrets, or unnecessary customer content enter the ledger.

<a id="s02-b"></a>
### S02-B: Explain the difference between receipts and bills

**Story:** As a finance reviewer, I need observed usage reconciled with matching cost evidence, so that I can distinguish allocated cost, shared cost, credits, and missing information.

**Why now:** provider-reported charges and invoice totals can differ. **Success:** matched periods reconcile; residuals remain visible; allocation rules are named and reproducible.

**Work type:** implementation. **Owner:** unassigned economics maintainer with finance reviewer. **Target:** S02. **Risk:** high financial interpretation. **Blocked by:** S02-A.

**Sources and surface:** [cost intelligence design](cost-intelligence-system-design.md) and [cost model standards](cost-model-standards.md). Own reconciliation calculations and CLI report; exclude undocumented allocation or automatic foreign-exchange assumptions.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S02-B1 | human-decision: finance/data owners specify matching periods, currency, credits/tax treatment, and shared-cost allocation for the pilot | A reviewer can classify every fixture line as matched, allocated, excluded with reason, or unresolved. |
| S02-B2 | implementation: calculate known subtotals, residuals, and missing-source requests | Synthetic charges `$4 + $3 + $1.50` total `$8.50`; a `$10` bill leaves `$1.50` visible. Mixed periods do not silently reconcile. |
| S02-B3 | implementation: export a reproducible reconciliation report with source links and allocation rules | Repeating the same inputs produces the same financial result; removing one required source marks affected results incomplete. |

**Sprint testing objective:** deterministic imports, duplicate/conflict handling, missing charges, credits, mismatched periods, and independent reconciliation arithmetic. **Sprint demo:** finance traces the `$1.50` residual to a named gap rather than approving a complete-cost claim. **Release task S02-R:** prepare **v0.5.0** with the file formats and allocation limits documented.

<a id="s03"></a>
## S03 — v0.6.0: Establish one workflow's real economics

**Goal:** the team knows what an accepted Condere assessment costs and which unresolved input most affects the decision.

**Start:** S02 accepted; data owner and finance reviewer available. **Terminal result:** a sourced pilot baseline and one bounded improvement hypothesis. **Scope boundary:** one representative workflow and observation period; no portfolio-wide benefit claim. **Learning outcome:** connect technical completion to a customer-accepted outcome.

<a id="s03-a"></a>
### S03-A: Measure cost per accepted outcome

**Story:** As a product and finance team, we need quality acceptance, revenue, and human effort alongside delivery cost, so that low-cost failed outputs do not look like a good business.

**Why now:** technical success is not customer value. **Success:** accepted outcomes have an explicit definition; failed work remains in costs; delivery margin and broader economics stay separate.

**Work type:** implementation. **Owner:** unassigned economics maintainer; product/finance reviewers required. **Target:** S03. **Risk:** high metric interpretation. **Blocked by:** S02-B.

**Sources and surface:** [full economics layers](full-economics-cost-layers.md) and [roadmap success measures](ROADMAP.md#success-measures). Own outcome joins and separate economics views; exclude revenue recognition advice or silently changing the existing gross-margin definition.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S03-A1 | human-decision: define accepted assessment, quality rubric, latency requirement, observation period, revenue treatment, and labor categories | Product and finance can independently classify a reviewed sample; all required thresholds have values and measurement rules before evaluation. |
| S03-A2 | implementation: join outcomes and attributable labor/revenue to the reconciled ledger | Synthetic 10 attempts, 6 accepted, `$30` delivery cost including failures, and `$12` labor yield `$5` delivery cost and `$7` including labor per accepted outcome. No cost appears twice. |
| S03-A3 | implementation: present named margin/cost views and denominator failure cases | Synthetic `$90` revenue and `$30` delivery cost give `$60` delivery gross profit; zero accepted outcomes gives undefined unit cost, and missing revenue gives unknown margin. |

**Acceptance/demo:** review the synthetic sample, exclude a low-quality outcome, and observe the changed denominator without losing its incurred cost. Artifact: outcome rubric, linked judgments, and economics report. Examples are arithmetic fixtures, not Condere measurements.

<a id="s03-b"></a>
### S03-B: Reconstruct the first Condere baseline

**Story:** As an engineering manager, I need one sourced baseline and its dominant cost drivers, so that I can choose a worthwhile experiment instead of optimizing by intuition.

**Why now:** the current production example has material unknowns. **Success:** material inputs are sourced or explicitly unresolved; a reviewer can reproduce the baseline; one candidate has a stated mechanism and testable hypothesis.

**Work type:** research. **Owner:** unassigned pilot lead with Condere data owner. **Target:** S03. **Risk:** high evidence quality and data access. **Blocked by:** S03-A and access to sanitized matching records.

**Sources and surface:** [first execution slice](ROADMAP.md#first-execution-slice-condere-baseline) and current Condere example. Own pilot dataset manifest, scenario, and findings; exclude production access or new paid assessments under this task's planning authority.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S03-B1 | manual-task: collect existing sanitized receipts, matching bills, accepted outcomes, and human effort for the agreed period | Every source has an owner, identity, period, and allowed use; absent records produce a specific gap, not an estimate presented as actual. |
| S03-B2 | research: reconcile the baseline and resolve the highest-impact missing inputs | A second reviewer reproduces it; receipts-only totals are not called total product cost; sample volume is not called monthly demand. |
| S03-B3 | research: rank measured cost drivers and specify one candidate change | Candidate identifies changed behavior, expected direction, quality constraints, unknowns, and the evidence that would reject it. No numeric savings promise is invented. |

**Sprint testing objective:** reproduce the accepted-outcome arithmetic, handle zero/missing denominators, then reconcile the real pilot sample. A synthetic-only report cannot pass the real-baseline acceptance. **Sprint demo:** finance traces cost per accepted assessment to receipts, labor, and judgments; engineering names one measured improvement candidate. **Release task S03-R:** prepare **v0.6.0**; missing material pilot evidence leaves the sprint outcome blocked.

<a id="s04"></a>
## S04 — v0.7.0: Explain the impact of proposed changes

**Goal:** product and engineering can compare a proposed change with the baseline and see which assumptions drive the result.

**Start:** S03 accepted. **Terminal result:** a reproducible decision report for the pilot candidate. **Scope boundary:** declared relationships and sensitivity analysis; no automatic causal claims from a code diff. **Learning outcome:** separate what code suggests from what an experiment must measure.

<a id="s04-a"></a>
### S04-A: Link a change to its economic assumptions

**Story:** As an engineering lead, I need code and configuration changes linked to workload assumptions, so that reviewers can see why a forecast changed.

**Why now:** a lower unit price can be offset by more calls, retries, or slower execution. **Success:** baseline/candidate identities are pinned; every material changed input has a source or hypothesis; unsupported effects remain unknown.

**Work type:** implementation. **Owner:** unassigned scenario maintainer. **Target:** S04. **Risk:** high unsupported inference. **Blocked by:** S03-B.

**Sources and surface:** [system economics design](design/system-economics-v1.md), `internal/systemmodel/compare.go`, and the S03 pilot candidate. Own scenario versions, explicit workload relationships, and comparison output; exclude writing candidate application code automatically.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S04-A1 | implementation: define a change record linking source revisions, deployment assumptions, workload, pricing snapshot, and input evidence | A report can reconstruct both scenarios from immutable identities; missing baseline/source identity rejects comparison as complete. |
| S04-A2 | implementation: reuse PCTL-29 driver propagation and add only missing call, retry, cache, batch, runtime, storage, or transfer relationships required by the candidate | Changing one declared factor changes its dependent costs; unrelated costs stay fixed; invalid cycles or incompatible units fail clearly. |
| S04-A3 | implementation: show a trace from changed inputs to each material financial difference | Price-only and price-plus-call-growth examples lead to different conclusions; unmeasured retry behavior is labeled a hypothesis. |

**Acceptance/demo:** start from the S03 candidate, inspect its changed inputs, and trace the resulting cost difference. Artifact: versioned scenarios and explanation. A code diff may suggest an input to investigate; it cannot establish that input's measured value.

<a id="s04-b"></a>
### S04-B: Compare financial options under uncertainty

**Story:** As a product owner, I need demand ranges, implementation effort, and payback beside recurring cost, so that I can choose whether a change is worth testing.

**Why now:** a cheaper monthly bill may not repay migration or operational effort. **Success:** scenarios use the same horizon and scope; assumptions that reverse the decision are visible; recommendations can remain inconclusive.

**Work type:** implementation. **Owner:** unassigned economics maintainer with product reviewer. **Target:** S04. **Risk:** high decision quality. **Blocked by:** S04-A.

**Sources and surface:** [full economics layers](full-economics-cost-layers.md), S03 acceptance rules, and S04-A change records. Own scenario/sensitivity calculations and report; exclude automatic pricing changes or treating speculative revenue uplift as observed.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S04-B1 | human-decision: define demand ranges, comparison horizon, implementation/migration effort, and decision constraints | Finance and product approve the input meanings; actuals and assumptions have different labels. |
| S04-B2 | implementation: calculate sensitivity, break-even, and separate one-time/recurring costs | Synthetic `$600` one-time cost and `$100` monthly benefit imply six months simple payback; zero/negative benefit has no positive payback. |
| S04-B3 | implementation: produce one structured decision report for the existing CLI/MCP and cost-aware skill workflow, with alternatives, uncertainty, and next evidence needed | The same inputs give consistent results across consumers; an assumption that reverses the ranking is exposed; incomplete options are not ranked as definitely best. |

**Sprint testing objective:** independent arithmetic, dependency propagation, consistent horizons, and reversal/inconclusive cases. **Sprint demo:** change one uncertain input and explain why the preferred option changes. **Release task S04-R:** prepare **v0.7.0** with a bounded experiment specification as the handoff to S05.

<a id="s05"></a>
## S05 — v0.8.0: Test alternatives in controlled sandboxes

**Goal:** engineering can test a candidate without exposing the host or losing experiment evidence.

**Start:** S04 accepted; workload, quality rubric, sandbox limits, and experiment authority specified. **Terminal result:** a reproducible paired comparison with inspectable artifacts and cleanup evidence. **Scope boundary:** one execution backend and one candidate change; no automatic deployment. **Learning outcome:** distinguish agent context, filesystem access, worktrees, and real execution isolation.

<a id="s05-a"></a>
### S05-A: Give workers a bounded place to execute

**Story:** As an engineering operator, I need each experiment to have scoped files, credentials, network access, and resource limits, so that an agent can execute useful work within an enforceable boundary.

**Why now:** tool instructions alone cannot enforce access or spending limits. **Success:** workers receive only allowed inputs; violations are denied; outputs survive termination outside the disposable environment.

**Work type:** implementation. **Owner:** unassigned runtime maintainer. **Target:** S05. **Risk:** high host/data/resource exposure. **Blocked by:** S04-B and a selected supported sandbox backend.

**Sources and surface:** [roadmap Stage 3](ROADMAP.md#stage-3-tested-alternatives). Own experiment launcher, filesystem contract, and artifact export; exclude broad host mounts, inherited operator credentials, or a homegrown virtualization platform.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S05-A1 | research: choose one existing sandbox backend and define readonly inputs, scratch space, retained outputs, network policy, and resource limits | A short capability check demonstrates enforcement; document remaining trust assumptions. A worktree or default container alone is insufficient evidence of isolation. |
| S05-A2 | implementation: launch experiments with scoped identity, timeout, cancellation, and durable artifact export | Allowed work produces a digest-addressed result; out-of-scope files/network/credentials are inaccessible; configured compute/time limits terminate work. |
| S05-A3 | implementation: handle interruption and cleanup of the entire worker process tree | Kill a running fixture: retained outputs remain traceable, child processes end within the configured grace period, and disposable resources are removed. |

**Acceptance/demo:** a worker reads an allowed fixture, writes an output, attempts a denied access, and is cancelled. Artifact: execution record, denied-operation evidence, retained output, and cleanup result. Test the effective runtime boundary, not merely its configuration file.

<a id="s05-b"></a>
### S05-B: Compare a baseline and candidate fairly

**Story:** As a product and engineering reviewer, we need matched experiments with quality and human correction effort, so that a cost reduction is not purchased through worse customer outcomes.

**Why now:** the scenario's uncertain behavior now needs observation. **Success:** runs are comparable; reported differences include uncertainty; quality failures prevent a positive recommendation.

**Work type:** implementation. **Owner:** unassigned experiment maintainer with product reviewer. **Target:** S05. **Risk:** high misleading comparison. **Blocked by:** S05-A and S03 quality decisions.

**Sources and surface:** S04 experiment specification and S03 outcome rubric. Own paired-run protocol, resource measurements, and comparison report; exclude multi-variable optimization or live provider calls in the validation suite.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S05-B1 | human-decision: freeze the workload corpus, baseline/candidate identities, repetitions, run order, warm/cold conditions, quality/latency criteria, and allowed budget | The protocol identifies what would reject the candidate before results exist; workload and quality versions are pinned. |
| S05-B2 | implementation: run matched local experiments and collect latency, resource use, failures, retries, outcome quality, human effort, and experiment cost | A known local change preserves expected outputs; a deliberately degraded candidate fails quality despite lower resource use. Paid providers remain disabled in validation. |
| S05-B3 | implementation: report observed differences, variability, exclusions, and reproduction instructions | Another operator can reproduce the comparison; different workload/environment versions are rejected or clearly separated. No simulated model response is called a live quality result. |

**Sprint testing objective:** boundary enforcement plus matched baseline/candidate behavior. Repeat runs according to the frozen protocol; report an inconclusive result when variability or missing observations prevent a decision. **Sprint demo:** show both a valid comparison and a cheap candidate rejected for quality. **Release task S05-R:** prepare **v0.8.0**; separately record whether the real Condere candidate requires an authorized live experiment.

<a id="s06"></a>
## S06 — v0.9.0: Compare hosting and scaling choices

**Goal:** the team can compare AWS, Azure, and local hosting against the same workload and customer requirements.

**Start:** S05 accepted; representative workload and measured resource profile available. **Terminal result:** a comparable hosting decision report with explicit feasibility limits. **Scope boundary:** one workload/deployment shape per environment, versioned price inputs, no provisioning or migration. **Learning outcome:** distinguish a price table from a capacity and operating-cost model.

<a id="s06-a"></a>
### S06-A: Compare total operating cost on equivalent terms

**Story:** As a finance and infrastructure reviewer, I need equivalent deployment options with their full cost scope, so that a low compute price does not hide transfer, labor, or migration cost.

**Why now:** provider labels and nominal CPU counts are not equivalent service outcomes. **Success:** options share workload and service requirements; prices have dates/regions/units; omitted cost categories remain visible.

**Work type:** implementation. **Owner:** unassigned infrastructure/economics maintainer. **Target:** S06. **Risk:** high misleading equivalence. **Blocked by:** S05-B.

**Sources and surface:** [roadmap integration strategy](ROADMAP.md#integration-strategy), S05 measurements, and existing cost standards. Own versioned comparison inputs and calculations; exclude a universal cloud catalog or purchase recommendations based only on list prices.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S06-A1 | research: specify matched AWS, Azure, and local options for the pilot's architecture and requirements | Each option names region/location, hardware/service characteristics, availability assumptions, data movement, and unsupported differences. |
| S06-A2 | implementation: load dated price snapshots and local cost inputs for compute, storage, transfer, minimum charges, support/labor, and migration | Independent fixtures exercise billing increments, currency/period checks, discounts, and missing rates; local hardware amortization and power assumptions are explicit. |
| S06-A3 | implementation: produce recurring, one-time, and broader-cost comparisons at the same demand levels | A cheaper compute option with higher egress/operations cost can rank worse; stale or incomplete inputs prevent a confident recommendation. |

**Acceptance/demo:** trace an option's total to its price sources and operating assumptions. Change egress or labor and see the result change. Artifact: three comparable scenarios plus an equivalence/limitations table. A modeled option is never labeled a measured deployment.

<a id="s06-b"></a>
### S06-B: Check whether a cheaper option can handle growth

**Story:** As an engineering manager, I need observed throughput and latency limits beside cost forecasts, so that I do not select an option that becomes uneconomical or unreliable at load.

**Why now:** scaling is not a linear multiplication of today's bill. **Success:** experiments expose capacity limits; demand assumptions are separate from measured service capability; infeasible options cannot win the comparison.

**Work type:** implementation. **Owner:** unassigned performance maintainer. **Target:** S06. **Risk:** high reliability extrapolation. **Blocked by:** S06-A and S05 workload protocol.

**Sources and surface:** S05 experiments, S03 service requirements, and S06 deployment scenarios. Own load profiles, observed capacity curves, and feasibility checks; exclude building an autoscaler or inferring a new provider's performance without calibration.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S06-B1 | implementation: define steady and burst arrival profiles independent of response time, with queue/error/latency measurements | The load fixture continues scheduled arrivals under slowdown; backpressure and dropped work are visible. |
| S06-B2 | research: measure the supported environment at increasing load and annotate modeled environments with calibration gaps | Capture sustainable capacity and failure behavior; do not transfer local CPU results directly to an untested cloud service. |
| S06-B3 | implementation: combine capacity and economic constraints into option feasibility | Synthetic demand of 100 requests/second rejects an option limited to 80, even if cheaper; a 120-capacity option still must satisfy the declared latency/quality constraints. |

**Sprint testing objective:** provider arithmetic, comparable scope, steady/burst load behavior, and infeasible/unknown-capacity cases. **Sprint demo:** increase demand until an option fails the service requirement and inspect the revised recommendation. **Release task S06-R:** prepare **v0.9.0**, distinguishing observed capacity from modeled cloud alternatives.

<a id="s07"></a>
## S07 — v0.10.0: Complete a saved investigation with one agent

**Goal:** one agent can gather evidence and prepare an economics decision without losing progress or crossing a human decision boundary.

**Start:** S06 accepted; orchestration choice and case data policy decided in S07-A1. **Terminal result:** one saved investigation resumes after interruption and reaches a reviewable decision. **Scope boundary:** one coordinator and one sequential worker path; no parallel swarm yet. **Learning outcome:** build a harness around reliable tools, with files, saved state, and explicit human decisions.

<a id="s07-a"></a>
### S07-A: Keep a durable decision case

**Story:** As a program manager, I need the goal, evidence, work state, assumptions, and decisions saved together, so that an investigation can continue after an agent or operator stops.

**Why now:** a chat transcript alone does not provide reliable workflow recovery. **Success:** case state is versioned; completed work is not duplicated on resume; assumptions and artifacts remain inspectable.

**Work type:** implementation. **Owner:** unassigned workflow maintainer. **Target:** S07. **Risk:** high recovery/consistency. **Blocked by:** S06-B.

**Sources and surface:** [roadmap Stage 4](ROADMAP.md#stage-4-coordinated-agent-investigations), [Deep Agents course](https://academy.langchain.com/courses/foundation-introduction-to-deepagents), and [Walking Labs harness course](https://walkinglabs.github.io/learn-harness-engineering/en/). Own a thin orchestration adapter, case state, artifact references, and recovery tests; preserve the Go economics engine.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S07-A1 | research: compare the existing CLI/MCP path with Deep Agents/LangGraph, Agno, and an OpenAI agent harness against one offline recovery/approval contract | Record capabilities, integration cost, persistence, isolation, licensing, and operating burden; select one orchestration owner. Symphony informs task/workspace management, not financial calculation. |
| S07-A2 | implementation: persist case identity, goal, constraints, task state, input/output references, assumptions, and decisions outside worker sandboxes | Reopen the case in a new process: state and artifact identities match; stale or corrupt checkpoints fail visibly. |
| S07-A3 | implementation: add checkpoint recovery, idempotent result recording, cancellation, and bounded retries | Interrupt after result persistence but before acknowledgement: resume records one result and avoids repeating completed side effects. |

**Acceptance/demo:** create a case, complete a calculation, stop the process, and resume it. Artifact: case state history and recovered report. Separate reusable business evidence from temporary agent scratch files; retention/deletion behavior is explicit.

<a id="s07-b"></a>
### S07-B: Let one agent prepare a decision within limits

**Story:** As a product owner, I need an agent to investigate a bounded question and pause for material decisions, so that I can delegate evidence gathering while retaining control of tradeoffs and commitments.

**Why now:** sequential behavior provides the baseline for evaluating later parallelism. **Success:** the agent uses real tools and cites evidence; limits are enforced outside the model; approval applies only to the reviewed action and inputs.

**Work type:** implementation. **Owner:** unassigned agent integration maintainer. **Target:** S07. **Risk:** high authority and output correctness. **Blocked by:** S07-A.

**Sources and surface:** S03 rubric, S04 decision report, S05 sandbox, and S07-A case contract. Own sequential coordinator/tool policy and approval records; exclude shared operator credentials or automatic consequential actions.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S07-B1 | implementation: expose bounded inspect, reconcile, compare, and experiment actions through the chosen adapter | A scripted worker completes a real local calculation; malformed requests and out-of-scope tools are denied; source text cannot grant tool authority. |
| S07-B2 | implementation: enforce step/time/cost limits and approvals bound to case, action, input revision, scope, and expiry | Denial stops execution; editing a material input invalidates old approval; simulated budget exhaustion stops before the next action. Validation uses no paid model. |
| S07-B3 | manual-task: evaluate the sequential case with task-specific judgments and measured supervision | Record unsupported claims, useful decisions, approval/correction/rescue minutes, elapsed time, tool failures, and all agent/experiment costs. A replay is labeled as such; live efficacy needs a separately authorized run. |

**Sprint testing objective:** recovery, duplicate prevention, tool correctness, bounded execution, stale approvals, and evidence-grounded reporting. **Sprint demo:** resume the pilot investigation, pause on a material assumption, then continue only after a valid decision. **Release task S07-R:** prepare **v0.10.0**; retain the sequential workflow as the comparison baseline for S08.

<a id="s08"></a>
## S08 — v0.11.0: Coordinate bounded parallel specialists

**Goal:** independent specialists can shorten a complex investigation without conflicting work, hidden cost, or reduced decision quality.

**Start:** S07 accepted; a sequential benchmark and independent task boundaries exist. **Terminal result:** a parallel investigation survives worker failures and produces one coherent report, with measured comparison against the simpler workflow. **Scope boundary:** a coordinator with a small bounded worker pool; no recursive unlimited spawning. **Learning outcome:** agent swarms require scheduling, ownership, communication, and integration—not just more prompts.

<a id="s08-a"></a>
### S08-A: Assign and combine specialist work safely

**Story:** As an investigation owner, I need economics, infrastructure, and experiment specialists to work on independent questions and share evidence, so that complex decisions can progress in parallel.

**Why now:** the single-worker contract now provides reliable boundaries. **Success:** each task has one owner; messages refer to evidence; the coordinator resolves dependencies and combines results without granting extra authority.

**Work type:** implementation. **Owner:** unassigned orchestration maintainer. **Target:** S08. **Risk:** high concurrency and isolation. **Blocked by:** S07-B.

**Sources and surface:** S07 case/tool contracts and S05 sandbox policy. Own the dependency scheduler, task ownership/leases, typed messages, and result integration; exclude unrestricted peer filesystem writes or nested schedulers.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S08-A1 | implementation: assign tasks by dependencies, lease, worker scope, and concurrency budget | Fixture cap of three workers is never exceeded; blocked tasks do not start; an expired lease can be reassigned without accepting two authoritative results. |
| S08-A2 | implementation: pass evidence/request/disagreement messages through the coordinator with case/task/source identities | Independent workers communicate; stale results and cross-case artifacts are rejected; conflicting edits/results return to the designated integrator. |
| S08-A3 | implementation: integrate specialist outputs with explicit uncertainty and final calculation through the trusted engine | One report links every material claim to evidence; agreement among agents does not turn an unsupported assumption into a fact. |

**Acceptance/demo:** economics and infrastructure workers investigate independently while an experiment worker supplies measurements. The coordinator combines their evidence and escalates a disagreement. Artifact: task graph, messages, costs, and integrated report. Each worker's effective filesystem, network, and credential scope is checked separately.

<a id="s08-b"></a>
### S08-B: Prove parallel work earns its overhead

**Story:** As a product owner, I need parallel investigations compared with a sequential baseline, so that autonomy improves useful work rather than merely increasing model and coordination spend.

**Why now:** more agents can cost more and make failures harder to detect. **Success:** recovery and limits hold under failure; quality remains acceptable; time, human effort, and cost tradeoffs are measured.

**Work type:** research. **Owner:** unassigned evaluation lead with product reviewer. **Target:** S08. **Risk:** high misleading efficacy claim. **Blocked by:** S08-A.

**Sources and surface:** S07 sequential benchmark and S03/S07 evaluation rubrics. Own failure experiments and comparative evaluation; exclude changing acceptance thresholds after seeing results.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S08-B1 | implementation: inject worker crash, duplicate delivery, delayed/stale result, concurrent write conflict, and budget exhaustion | Case recovers or stops explicitly; no duplicate accepted result, unauthorized write, orphan worker, or unrecorded cost is permitted. |
| S08-B2 | human-decision: freeze matched cases and minimum useful improvement in elapsed time or human effort, with quality and total-cost limits | The protocol uses the same workload, tools, evidence access, and resource accounting for sequential and parallel runs. |
| S08-B3 | research: run the comparison and recommend sequential default or a narrowly enabled parallel mode | Report coordination, failed work, experiment, storage, and human costs. Enable the claimed improvement only if predefined criteria pass; otherwise refine the parallel feature. |

**Sprint testing objective:** fault recovery and measurable benefit over the sequential baseline. Offline fixtures prove mechanics; actual agent benefit requires an authorized representative evaluation. **Sprint demo:** interrupt one worker, show recovery, then compare cost/quality/human effort with the sequential case. **Release task S08-R:** prepare **v0.11.0** with the tested concurrency and workload limits stated.

<a id="s09"></a>
## S09 — v0.12.0: Learn from actual business results

**Goal:** the team can see where forecasts were wrong and whether an accepted change produced a worthwhile result.

**Start:** S08 accepted; later-period actuals available or a named data dependency recorded. **Terminal result:** one forecast-versus-actual report and a human-reviewed follow-up decision. **Scope boundary:** imports/manual refresh first; recurring execution must have separately approved triggers and budgets. **Learning outcome:** distinguish modeled savings, observed changes, and attributable business impact.

<a id="s09-a"></a>
### S09-A: Compare predictions with later observations

**Story:** As a finance reviewer, I need forecasts compared with matching subsequent usage, bills, and outcomes, so that I can assess the model before relying on another recommendation.

**Why now:** a useful model must remain accountable after the decision. **Success:** the original forecast is preserved; comparison uses matched scope/periods; prediction error and unresolved differences are explicit.

**Work type:** implementation. **Owner:** unassigned economics/data maintainer. **Target:** S09. **Risk:** high forecast validity. **Blocked by:** S08-B and later-period matching data.

**Sources and surface:** [roadmap Stage 5](ROADMAP.md#stage-5-continuous-economic-improvement) and S02 reconciliation contracts. Own forecast snapshots, actuals joins, and error reports; exclude silently recalibrating the original forecast to fit the outcome.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S09-A1 | implementation: freeze forecast inputs and reserve later observations for evaluation | Forecast and calibration-period identities are recorded; future data cannot leak into the earlier forecast record. |
| S09-A2 | implementation: reuse accepted PCTL-21 variance calculations to compare saved forecasts with later cost, resource use, and accepted outcomes under the same scope | Synthetic forecast `$100` and actual `$120` produce a `$20` gap; denominator and direction of any percentage error are named, with explicit handling of zero actuals. |
| S09-A3 | research: review a held-out pilot period and propose a versioned assumption update | Original and revised assumptions remain available; changed demand, rates, workload mix, and acceptance rates are distinguished. Missing data blocks accuracy claims. |

**Acceptance/demo:** show a forecast, import later actuals, and explain an error without rewriting history. Artifact: forecast/error report and a proposed model version with reviewer rationale.

<a id="s09-b"></a>
### S09-B: Reopen decisions only when the change matters

**Story:** As a program manager, I need material drift and net benefit surfaced with a clear next decision, so that the team can improve economics without constant low-value investigation.

**Why now:** monitoring and agent work create their own costs. **Success:** material changes create reviewable work; unchanged data creates no duplicate investigation; impact accounts for implementation and supervision.

**Work type:** implementation. **Owner:** unassigned workflow/economics maintainer. **Target:** S09. **Risk:** medium noise; high impact attribution. **Blocked by:** S09-A.

**Sources and surface:** S04 decision report, S07 case history, and S09 forecast comparison. Own drift rules, review queue, and net-benefit report; exclude automatically changing product pricing, cloud resources, or production code.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S09-B1 | human-decision: define material drift, responsible reviewer, approved refresh trigger, and any recurrence/spending limits | Rules have named owners and testable thresholds; no external notifications or recurring paid jobs are enabled by default. |
| S09-B2 | implementation: produce deduplicated follow-up cases from material new evidence | Same import twice creates one case; unchanged inputs remain quiet; stale/missing sources create a data-quality issue rather than a savings claim. |
| S09-B3 | implementation: report gross operating difference and net impact including implementation, migration, agent, experiment, and ongoing human costs | Synthetic savings `$100/month`, added recurring expense `$30/month`, and `$280` implementation cost imply `$70/month` net recurring benefit and four months simple payback. Observational impact is labeled as such. |

**Sprint testing objective:** temporal separation, reproducible error calculations, quiet unchanged-state behavior, duplicate prevention, and net-benefit arithmetic. **Sprint demo:** import later data, surface one material issue, review its evidence, and accept or reject a follow-up. **Release task S09-R:** prepare **v0.12.0**; forecast accuracy and realized benefit remain unverified until supported by the pilot's actuals.

<a id="s10"></a>
## S10 — v1.0.0: Release the controlled team workflow

**Goal:** finance, product, and engineering can use the full workflow with clear access, accountability, support, and measured limitations.

**Start:** S09 accepted; pilot participants, data rights, support surface, and release criteria agreed. **Terminal result:** the scoped end-to-end pilot is accepted and the exact validated release is prepared for explicitly authorized publication. **Scope boundary:** one supported execution environment and organization pilot; broader service rollout and compliance certification remain separate. **Learning outcome:** turn a working investigation into a repeatable team practice.

<a id="s10-a"></a>
### S10-A: Protect and review team decision records

**Story:** As an enterprise team lead, I need scoped access, attributable approvals, and controlled evidence retention, so that finance and engineering can collaborate without exposing another team's data or losing decision history.

**Why now:** personal agent experiments do not establish team operating controls. **Success:** the supported interface enforces team and action boundaries; decisions identify the responsible person; retention/export is testable.

**Work type:** implementation. **Owner:** unassigned platform maintainer with security/data owner. **Target:** S10. **Risk:** high access and data handling. **Blocked by:** S09-B.

**Sources and surface:** S07 approvals/state and S08 worker isolation. Own the chosen interface's access checks, evidence export, and retention behavior; reuse existing identity services when present. Exclude building a new identity provider or claiming broad enterprise certification.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S10-A1 | human-decision: define finance/product/engineering roles, workspace boundaries, retention, supported deployment surface, and support owner | A role/action matrix identifies who may inspect, propose, approve, execute, and delete; an absent owner blocks readiness. |
| S10-A2 | implementation: enforce permissions and record attributable decisions across the supported interface and worker tools | Two workspace fixtures prove cross-workspace reads/writes and credential reuse are denied; a worker cannot impersonate an approver. |
| S10-A3 | implementation: provide versioned export, retention/deletion handling, and recovery instructions | Export/reopen preserves permitted history; expired approvals cannot be replayed; deletion removes scoped evidence while preserving only explicitly allowed audit metadata. |

**Acceptance/demo:** finance reviews a report, product approves its exact decision, and engineering sees the permitted experiment scope. An unauthorized user is denied. Artifact: role tests, decision history, and retention/recovery results.

<a id="s10-b"></a>
### S10-B: Demonstrate the full product decision cycle

**Story:** As the product sponsor, I need one representative workflow completed and independently reviewed from baseline to actual outcome, so that I can decide whether ProfitCtl is ready for broader use.

**Why now:** component tests cannot establish the complete team outcome. **Success:** the pilot is reproducible; required quality/economics/effort criteria are met or explicitly fail; installation, recovery, and support match the released scope.

**Work type:** manual-task. **Owner:** unassigned release/pilot lead; product and finance acceptance required. **Target:** S10. **Risk:** high release and benefit claims. **Blocked by:** S10-A and required pilot evidence/authority.

**Sources and surface:** all prior sprint acceptance contracts, [validation manifest](../tabellio.validation.json), and [Open-Core Roadmap](OPEN_CORE_ROADMAP.md). Own pilot acceptance, compatibility, release documentation, and support handoff; exclude outreach, pricing commitments, hosted expansion, or automatic publication.

| Task | Type and bounded result | Direct check |
| --- | --- | --- |
| S10-B1 | manual-task: freeze v1 scope and run baseline → scenario → experiment → agent recommendation → human decision → later-actual comparison | Finance can reproduce the economics; product can defend the quality judgment; engineering can reproduce the experiment. Required live evidence is separately authorized and cannot be substituted with a test response. |
| S10-B2 | manual-task: verify supported upgrade/install, saved-case compatibility, failure recovery, limits, and rollback | A saved supported pre-v1 case reopens or migrates through an explicit versioned path; a failed migration preserves the original data and offers recovery. |
| S10-B3 | manual-task: obtain pilot acceptance and prepare release/support notes with measured limitations | Record decision quality, forecast error, human effort, full operating cost, and observed impact against pre-agreed criteria; unsupported benefit claims are removed. |

**Sprint testing objective:** exact-candidate validation plus a representative team journey, access failures, interruption recovery, compatibility, and real evidence review. **Sprint demo:** one complete decision case, one rejected alternative, one recovered interruption, and the subsequent actual-result report. **Release task S10-R:** prepare **v1.0.0**; declare shipment only after an authorized release exists, and adoption only after observed use.

## Critical dependencies and first pull

The release sequence is intentionally serial at its acceptance boundaries. Independent preparatory work may happen earlier only when it consumes no unresolved assumption and does not bypass the WIP limit.

| Dependency | Needed for | Resolution owner |
| --- | --- | --- |
| Current defect/release classification | S01 | Engine and release owners, unassigned |
| Sanitized matching receipts, bills, outcomes, and labor | S02–S03 | Condere data owner and finance reviewer, unassigned |
| Accepted-outcome rubric and service thresholds | S03 onward | Product decision owner with reviewers |
| Sandbox backend and enforceable limits | S05 | Runtime owner, unassigned |
| Provider equivalence and capacity evidence | S06 | Infrastructure/performance owners, unassigned |
| Orchestration choice and persisted case contract | S07 | Tech lead/final integrator, unassigned |
| Matched sequential/parallel evaluation and any paid-run authority | S08 | Product decision owner and evaluation lead |
| Later-period actuals | S09–S10 | Finance/data owner, unassigned |
| Team roles, retention, supported surface, and pilot acceptance | S10 | Product, security/data, and support owners |

**First pull:** assign the S01 owner and perform S01-A1. Use its result to commit the actual v0.4.0 scope. The first substantive new capability after that foundation is S02's receipt-to-bill reconciliation. Do not start a swarm while the baseline is still unknown.

## Tracker handoff

**AgentShift synchronization: Confirmed on 2026-10-07.** The connection is restored. The canonical ProfitCtl project is `dc3a0fc2-fd0e-4707-8a6d-3cdc33aa7dc7`. All 100 mapped records were read back: 10 release/sprint groups, 20 stories, and 70 tasks. This created 97 records and reused PCTL-15 for S02-A1, PCTL-26 for S04-B, and PCTL-23 for S09-B, preserving their prior context and acceptance requirements.

The connector exposes issues, parent relationships, and blocker relationships, but no sprint/cycle or release fields. Each proposed sprint/version therefore has a parent issue. This is an explicit grouping convention, not a native scheduled sprint. All 100 mapped records remain Backlog, with refinement requirements in their descriptions. No dates, new assignments, or implementation starts were introduced.

The `S01`–`S10`, `Snn-A/B`, and task identifiers above are **local planning references**. The [verified AgentShift backlog map](AGENTSHIFT_SPRINT_BACKLOG.md) records their actual PCTL IDs, UUIDs, parents, and native links. Existing identifiers, including PCTL-17, are preserved.

Ninety-one native links were verified: 86 blockers and 5 related references. Four blocker requirements could not be confirmed through the connector's native readback and are recorded in verified dependent descriptions:

- PCTL-23 → PCTL-124 (S10-A).
- PCTL-23 → PCTL-122 (S09-R).
- PCTL-21 → PCTL-23 (S09-B).
- PCTL-21 → PCTL-115 (S09-A).

Check those written prerequisites, parent readiness, and active WIP before pulling work. The connector limitation does not remove a dependency. Existing active work and all 32 records outside the reuse set retain their inspected title, state, priority, and parent.

First refinement: S01-A / **PCTL-37**, beginning with S01-A1 / **PCTL-38**. This planning synchronization does not represent implementation, publication, or release completion.
