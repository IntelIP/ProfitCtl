# LLM Detection Sprint - Label Application Guide

## Available Labels
```
Testing Related:
- testing (e4ae9b97-0564-45db-a9c6-af76e26a85d6) - Purple
- testing-layer (4265bf21-414e-402c-8c21-0f9d27054d7f) - Gray

Component Labels:
- backend (c822c5d1-fa83-479f-80e3-4f153e2f31f7) - Purple
- infra (aed9f379-3fe0-4cc6-aae3-6739129bf950) - Purple
- api (e5adac2e-6334-42f0-852f-4bcc322e9e33) - Gray
- core-logic (eb3a6a2b-9fb5-45a9-add8-4cea8ea0f242) - Gray
- agents (4d7a78e8-b704-4051-815d-844c4c3ee1d3) - Gray

Quality Labels:
- tech-debt (7c389069-73fb-45dc-a5fe-6fa057a1f4d5) - Blue
- documentation (not in list - can use backend) - Purple
```

## Label Assignment by Issue

### PR #1: Foundation & Setup (Label: backend + testing-layer)

**Multiple labels per issue (use Linear's multi-select)**

- **INT-17**: backend, infra
- **INT-18**: backend, core-logic
- **INT-19**: backend, core-logic
- **INT-20**: backend, testing-layer
- **INT-21**: backend, testing
- **INT-22**: backend, testing, testing-layer
- **INT-23**: backend, testing, testing-layer

### PR #2: Core Implementation (Label: backend + api)

- **INT-24**: backend, api, core-logic
- **INT-25**: backend, infra
- **INT-26**: backend, core-logic
- **INT-27**: backend, testing, testing-layer
- **INT-28**: backend, testing, testing-layer
- **INT-29**: backend, testing, testing-layer

### PR #3: Analysis & Generation (Label: backend + testing-layer)

- **INT-30**: backend, core-logic
- **INT-31**: backend, testing, testing-layer
- **INT-32**: backend, core-logic
- **INT-33**: backend, testing, testing-layer
- **INT-34**: backend, testing
- **INT-37**: backend, testing, testing-layer

### PR #4: Documentation & Integration (Label: backend)

- **INT-35**: backend (documentation)
- **INT-36**: backend (examples)
- **INT-38**: backend, testing, testing-layer

## Total Labels Per Issue Count

- **testing**: 10 issues (all PR review + test issues)
- **testing-layer**: 8 issues (testing infrastructure + integration tests)
- **backend**: 22 issues (ALL issues - this is a CLI/backend feature)
- **infra**: 2 issues (dependency and setup)
- **api**: 1 issue (OpenRouter provider)
- **core-logic**: 6 issues (types, prompts, generator, parser)

## How to Apply in Linear

1. Go to: https://linear.app/intelip/team/INT/all
2. Filter: INT-17 through INT-38
3. Select multiple issues
4. Use "Label" button to add labels
5. Apply labels in batches by category

## Benefits

With labels applied, you can now:
- Filter by "testing" to see all test-related work
- Filter by "backend" to see all profitctl CLI work
- Filter by "core-logic" to see algorithm/parsing work
- Filter by combinations (e.g., backend + testing)
- Get visual color coding in the board view
- Track work by component type

## Sprint Organization Summary

**Sprint**: Sprint 2026-Week.3 (Jan 18-24, 2026)
**Issues**: 22 issues (INT-17 to INT-38)
**Total Points**: ~35 points
**Date Range**: January 18-24, 2026
**Current State**: Backlog (move to Todo for sprint start)
**Project**: IntelIP – Backend / Agents
**Recommended Action**: Apply labels then start sprint
