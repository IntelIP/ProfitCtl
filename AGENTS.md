# AGENTS.md

## Tooling Policy

- Greptile is deprecated for this repository.
- Prefer CodeRabbit for local Codex review loops.
- Keep the CodeRabbit usage-based add-on disabled unless the user explicitly approves overage billing.
- Use native GitHub checks and review comments with `$check-pr` for PR readiness.

## Review Workflow

1. Run `coderabbit --agent` or `cr --agent` from this repo when the user asks for a local review or pre-PR quality pass.
2. Fix `Critical` and `Warning` findings first.
3. Re-run CodeRabbit once after material fixes.
4. Use `$check-pr` after local review when the task is PR readiness.

## Review Focus

- Prioritize pricing correctness, simulation integrity, CLI output regressions, release safety, and configuration drift.
- Treat benchmark scenarios and docs as secondary to executable behavior unless the user explicitly asks for doc review.
