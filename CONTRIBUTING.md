# Contributing

## Development Setup

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
go test ./...
```

## Standards

- Run `go test ./...` before opening a PR.
- When changing the economics engine or Monte Carlo path, run `go test ./internal/simulation -run='^$' -bench=. -count=1` and include the result if it changes materially.
- Keep changes scoped and documented.
- Prefer explicit error handling and stable CLI behavior.
- Keep comments focused on rationale/constraints, not roadmap placeholders.
- Use the issue and PR templates in `.github/` so review context is consistent.
- For release-facing changes, include a smoke test command and the expected result in the PR body.

## Pull Requests

- Use descriptive titles.
- Include validation commands and results.
- Note any behavior, output, or contract changes.
- Link the related Linear issue when one exists.
- Keep release or public-facing changes small enough to review in one pass.
- If the change affects `simulate`, `compare`, `calibrate`, install, or release publishing, call that out explicitly and add a regression test or fixture where practical.
- Wait for CI and maintainer review before tagging an external release.

## Maintainer Workflow

- Do not push directly to `main` for normal feature or release-facing work.
- Open a PR, let CI run, and get at least one maintainer review before merge.
- Treat `CODEOWNERS` as the required review path for contributor-facing docs, release plumbing, and policy changes.
- Cut public releases from merged `main` only.
- If a PR changes install, release, or benchmark behavior, verify the exact user-facing command path before merging.

## CI/CD

- PR and main checks run in GitHub Actions.
- Release publishing remains an explicit maintainer action for semver tags.
- External releases are cut from `main` only after the release checklist is complete.
- CODEOWNERS should be treated as the default review path for release, policy, and contributor-facing changes.
