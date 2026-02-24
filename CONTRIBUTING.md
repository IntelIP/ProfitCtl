# Contributing

## Development Setup

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
go test ./...
```

## Standards

- Run `go test ./...` before opening a PR.
- Keep changes scoped and documented.
- Prefer explicit error handling and stable CLI behavior.
- Keep comments focused on rationale/constraints, not roadmap placeholders.

## Pull Requests

- Use descriptive titles.
- Include validation commands and results.
- Note any behavior, output, or contract changes.

## CI/CD

- PR and main checks run in Woodpecker.
- Release publish/deploy runs on semver tags.
