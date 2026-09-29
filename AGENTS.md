# AGENTS.md

## Tooling Policy

- Use Codex plus native GitHub checks and review comments for local review loops.
- Do not use paid review bots or usage-based review add-ons unless the user explicitly approves the spend.
- Use native GitHub checks and review comments with `$check-pr` for PR readiness.

## Review Workflow

1. Inspect the PR or local diff with Codex using GitHub metadata, diff, checks, local tests, and changed files.
2. Lead with findings ordered by severity, then classify each item as `actionable`, `informational`, or `already addressed`.
3. Fix actionable findings first.
4. Re-run affected checks after material fixes.
5. Update Plane with status, blockers, and validation evidence.
6. Use `$check-pr` after local review when the task is PR readiness.

## Review Focus

- Prioritize pricing correctness, simulation integrity, CLI output regressions, release safety, and configuration drift.
- Treat benchmark scenarios and docs as secondary to executable behavior unless the user explicitly asks for doc review.

## Product validation gate

- Unit tests are structural evidence only. Before review or merge readiness, run `tabellio-validate gate` with the committed `tabellio.validation.json` against the exact candidate commit.
- Required schema, semantic, workflow, operational, and security evidence must pass. `blocked` is not `passed`; any new commit invalidates old evidence.
- Pricing correctness, simulation integrity, maintained reference cost standards, CLI output, structured LLM parsing, and zero-cost validation are explicit boundaries.
- Never call OpenRouter, deploy, publish, or mutate billing during validation. Upload generated evidence from CI; do not commit it.
- Track rollout and failures in Plane item `PCTL-17`.

## Code cleanup

- Run `bash scripts/check-dead-code.sh` after removing code and before review.
- For cross-file changes, run `graphify update .`, inspect callers and dependencies, and verify findings in current source and framework registrations.
