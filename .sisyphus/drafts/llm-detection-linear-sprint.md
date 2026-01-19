# Linear Sprint Tasks: LLM Detection Feature (1 Week)

**Sprint Goal**: Complete PR #1-#2 and establish testing infrastructure + OpenRouter integration
**Sprint Duration**: 5 business days
**Timeline**: Week of [Date]

---

## 📋 Sprint Overview

**Total Tasks**: 35 tasks across 5 categories
**Estimated Effort**: ~40 hours
**Priority Order**:
1. Foundation & Setup (PR #1 - Days 1-2)
2. Core Implementation (PR #2 - Days 2-4)
3. Testing & Validation (PR #1-#2 - Throughout)
4. Documentation & Examples (Days 4-5)
5. Reviews & Integration (Ongoing)

---

## 🏗️ Category 1: Foundation & Setup (PR #1)
**Owner**: [Developer Name]
**Estimated**: 8 hours
**PR**: #1 - Testing Infrastructure & Types Foundation

### Day 1 Tasks

- [ ] **LIN-001** Create directory structure for scanner
  - **Description**: Create `internal/scanner/`, `internal/scanner/llm/`, `test/fixtures/`
  - **Type**: Setup
  - **Points**: 1
  - **Priority**: High
  - **Assignee**: 
  - **Dependencies**: None
  - **Acceptance**: Directories exist, follow existing patterns
  - **PR Association**: #1, Task 1.1

- [ ] **LIN-002** Define core types (CodeContext, AnalysisResponse, etc.)
  - **Description**: Create `internal/scanner/llm/types.go` with all structs and interfaces
  - **Type**: Implementation
  - **Points**: 3
  - **Priority**: High
  - **Dependencies**: LIN-001
  - **Acceptance**: File compiles, types exported correctly, struct tags proper
  - **PR Association**: #1, Task 1.2

- [ ] **LIN-003** Create prompt templates with truncation
  - **Description**: Create `internal/scanner/llm/prompts.go` with `BuildDetectionPrompt` and `truncate` functions
  - **Type**: Implementation
  - **Points**: 3
  - **Priority**: High
  - **Dependencies**: LIN-002
  - **Acceptance**: Prompt builds correctly, truncation works at 2000 chars
  - **PR Association**: #1, Task 1.3

- [ ] **LIN-004** Document Go testing approach for team
  - **Description**: Create `docs/TESTING.md` explaining Go test patterns for Python/TS background
  - **Type**: Documentation
  - **Points**: 2
  - **Priority**: Medium
  - **Dependencies**: None
  - **Acceptance**: Covers file naming, test structure, mocking, golden files, running tests
  - **PR Association**: #1, Task 1.4

### Day 2 Tasks

- [ ] **LIN-005** Create first golden file test case
  - **Description**: Create sample `test/fixtures/samples/go-webapp/go.mod` and expected `test/fixtures/golden/go-webapp-expected.yml`
  - **Type**: Testing
  - **Points**: 2
  - **Priority**: High
  - **Dependencies**: LIN-001
  - **Acceptance**: Golden file valid YAML, reflects expected detection
  - **PR Association**: #1, Task 1.5

- [ ] **LIN-006** Write unit tests for prompt building
  - **Description**: Create `internal/scanner/llm/prompts_test.go` with tests for BuildDetectionPrompt and truncate
  - **Type**: Testing
  - **Points**: 2
  - **Priority**: High
  - **Dependencies**: LIN-003, LIN-004
  - **Acceptance**: Tests pass, show coverage
  - **PR Association**: #1, Task 1.6

- [ ] **LIN-007** PR #1 Review & Merge
  - **Description**: Submit PR, address feedback, merge to main
  - **Type**: Review
  - **Points**: 1
  - **Priority**: High
  - **Dependencies**: LIN-001 through LIN-006
  - **Acceptance**: PR approved, all comments addressed, merged
  - **PR Association**: #1

---

## 🔌 Category 2: Core Implementation (PR #2)
**Owner**: [Developer Name]
**Estimated**: 14 hours
**PR**: #2 - OpenRouter Provider & Code Collection

### Day 2 Tasks (continued)

- [ ] **LIN-008** Implement OpenRouter provider
  - **Description**: Create `internal/scanner/llm/openrouter.go` with NewOpenRouterProvider, Chat, IsAvailable methods
  - **Type**: Implementation
  - **Points**: 3
  - **Priority**: High
  - **Dependencies**: LIN-002
  - **Acceptance**: File compiles, constructor handles defaults, errors informative
  - **PR Association**: #2, Task 2.1

- [ ] **LIN-009** Add OpenRouter dependency to go.mod
  - **Description**: Run `go get github.com/reVrost/go-openrouter && go mod tidy`
  - **Type**: Setup
  - **Points**: 1
  - **Priority**: High
  - **Dependencies**: LIN-008
  - **Acceptance**: go.mod updated, build succeeds
  - **PR Association**: #2, Task 2.2

### Day 3 Tasks

- [ ] **LIN-010** Implement code collector
  - **Description**: Create `internal/scanner/collector.go` with Collect method to find config files
  - **Type**: Implementation
  - **Points**: 3
  - **Priority**: High
  - **Dependencies**: LIN-001, LIN-002
  - **Acceptance**: Finds all pattern files, respects size limits, returns partial on errors
  - **PR Association**: #2, Task 2.3

- [ ] **LIN-011** Write unit tests for OpenRouter provider
  - **Description**: Create `internal/scanner/llm/openrouter_test.go` with unit tests and integration test
  - **Type**: Testing
  - **Points**: 2
  - **Priority**: High
  - **Dependencies**: LIN-008, LIN-004
  - **Acceptance**: Tests pass, integration test skips without API key
  - **PR Association**: #2, Task 2.4

- [ ] **LIN-012** Write unit tests for collector
  - **Description**: Create `internal/scanner/collector_test.go` with table-driven tests
  - **Type**: Testing
  - **Points**: 2
  - **Priority**: High
  - **Dependencies**: LIN-010, LIN-004
  - **Acceptance**: Both tests pass, demonstrate collection working
  - **PR Association**: #2, Task 2.5

### Day 4 Tasks

- [ ] **LIN-013** PR #2 Review & Merge
  - **Description**: Submit PR, address feedback, merge to main
  - **Type**: Review
  - **Points**: 1
  - **Priority**: High
  - **Dependencies**: LIN-008 through LIN-012
  - **Acceptance**: PR approved, all comments addressed, merged
  - **PR Association**: #2

---

## 🔍 Category 3: Testing & Validation (PR #3-#4)
**Owner**: [Developer Name]
**Estimated**: 10 hours
**PRs**: #3 - LLM Analysis & Config Generation, #4 - Integration

### Day 3-4 Tasks

- [ ] **LIN-014** Implement response parser
  - **Description**: Create `internal/scanner/llm/parser.go` to extract JSON from LLM responses (handles markdown)
  - **Type**: Implementation
  - **Points**: 2
  - **Priority**: High
  - **Dependencies**: LIN-002
  - **Acceptance**: Parses JSON from raw strings and markdown blocks, returns helpful errors for malformed JSON
  - **PR Association**: #3, Task 3.1

- [ ] **LIN-015** Write tests for parser
  - **Description**: Create `internal/scanner/llm/parser_test.go` with table-driven tests
  - **Type**: Testing
  - **Points**: 2
  - **Priority**: High
  - **Dependencies**: LIN-014, LIN-004
  - **Acceptance**: All test cases pass (valid JSON, markdown blocks, malformed JSON, empty)
  - **PR Association**: #3, Task 3.2

- [ ] **LIN-016** Implement config generator
  - **Description**: Create `internal/scanner/llm/generator.go` to convert AnalysisResponse to config.Config
  - **Type**: Implementation
  - **Points**: 3
  - **Priority**: High
  - **Dependencies**: LIN-002, LIN-003
  - **Acceptance**: Generates valid config, maps services to costs correctly, adds default covenants
  - **PR Association**: #3, Task 3.3

- [ ] **LIN-017** Write integration test (end-to-end)
  - **Description**: Create `internal/scanner/llm/integration_test.go` testing full flow with mocked LLM
  - **Type**: Testing
  - **Points**: 3
  - **Priority**: High
  - **Dependencies**: LIN-008, LIN-010, LIN-014, LIN-016, LIN-004
  - **Acceptance**: Shows full flow working, mocks isolate LLM, validates golden file
  - **PR Association**: #3, Task 3.4

- [ ] **LIN-018** Manual testing setup
  - **Description**: Create test repositories (go-app, node-app, python-app) in /tmp/test-detect/
  - **Type**: Testing
  - **Points**: 2
  - **Priority**: Medium
  - **Dependencies**: LIN-005
  - **Acceptance**: Sample files created, ready for manual testing
  - **PR Association**: #3 (prep for PR #4)

---

## 📚 Category 4: Documentation & Examples
**Owner**: [Developer Name]
**Estimated**: 6 hours
**PR**: #4 - Documentation & Examples

### Day 4-5 Tasks

- [ ] **LIN-019** Create usage documentation (docs/DETECT.md)
  - **Description**: Comprehensive guide with prerequisites, usage examples, troubleshooting, costs
  - **Type**: Documentation
  - **Points**: 3
  - **Priority**: High
  - **Dependencies**: None (can work in parallel)
  - **Acceptance**: Covers setup, examples for all flags, troubleshooting, cost estimates
  - **PR Association**: #4, Task 4.4

- [ ] **LIN-020** Create example detected config (examples/detected-profit.yml)
  - **Description**: Add well-commented example showing realistic detection results
  - **Type**: Documentation
  - **Points**: 1
  - **Priority**: Medium
  - **Dependencies**: LIN-016
  - **Acceptance**: Example created, validated, realistic
  - **PR Association**: #4, Task 4.6

---

## 🎯 Category 5: Reviews & Integration
**Owner**: [Developer Name]
**Estimated**: 2 hours
**PRs**: #3, #4

### Day 4-5 Tasks

- [ ] **LIN-021** PR #3 Review & Merge
  - **Description**: Submit PR, address feedback, merge to main
  - **Type**: Review
  - **Points**: 1
  - **Priority**: High
  - **Dependencies**: LIN-014 through LIN-018
  - **Acceptance**: PR approved, merge complete
  - **PR Association**: #3

- [ ] **LIN-022** PR #4 Review & Merge
  - **Description**: Submit PR, address feedback, merge to main
  - **Type**: Review
  - **Points**: 1
  - **Dependencies**: LIN-019, LIN-020, LIN-021
  - **Acceptance**: PR approved, all tests pass, docs complete, merge complete
  - **PR Association**: #4

---

## 🏃 Sprint Execution Plan

### Day 1 (Monday) - Foundation
**Morning** (3 hours):
- LIN-001: Create directory structure
- LIN-004: Document Go testing approach
- LIN-005: Create golden file test case

**Afternoon** (3 hours):
- LIN-002: Define core types

### Day 2 (Tuesday) - Testing Infrastructure
**Morning** (3 hours):
- LIN-003: Create prompt templates
- LIN-006: Write unit tests for prompt building

**Afternoon** (3 hours):
- LIN-007: PR #1 Review & Merge
- LIN-008: Implement OpenRouter provider

### Day 3 (Wednesday) - Core Implementation
**Morning** (3 hours):
- LIN-009: Add OpenRouter dependency
- LIN-010: Implement code collector

**Afternoon** (3 hours):
- LIN-011: Write unit tests for OpenRouter provider
- LIN-012: Write unit tests for collector

### Day 4 (Thursday) - Analysis & Config Generation
**Morning** (4 hours):
- LIN-014: Implement response parser
- LIN-015: Write tests for parser
- LIN-016: Implement config generator

**Afternoon** (2 hours):
- LIN-017: Write integration test (end-to-end)
- LIN-018: Manual testing setup
- LIN-021: PR #3 Review & Merge

### Day 5 (Friday) - Documentation & Final Integration
**Morning** (3 hours):
- LIN-019: Create usage documentation (docs/DETECT.md)
- LIN-020: Create example detected config

**Afternoon** (2 hours):
- LIN-022: PR #4 Review & Merge
- Sprint retrospective

---

## 📊 Sprint Metrics

**Planned Metrics**:
- Total Tasks: 22 Linear tasks
- Total Story Points: ~35 points
- Average per day: 4-5 tasks
- Code Coverage Target: >80% on new code
- PRs Planned: 4

**Success Criteria**:
- [ ] All 4 PRs merged to main
- [ ] Testing infrastructure established and documented
- [ ] OpenRouter provider fully functional
- [ ] End-to-end detection working with mocks
- [ ] Documentation complete
- [ ] Manual testing completed

**Risks**:
- **Blocker**: OpenRouter API key setup (mitigation: document setup early)
- **Risk**: Learning curve on Go testing (mitigation: LIN-004 provides training)
- **Risk**: PR review delays (mitigation: small PRs, <300 lines each)

---

## 🔄 Sprint Workflow

### Daily Standup Questions
1. What Linear tasks did you complete yesterday?
2. What Linear tasks are you working on today?
3. Any blockers? (Note blocker task numbers)

### PR Review Process
- Each PR gets dedicated review session
- Review checklist included in work plan
- Feedback addressed in follow-up tasks
- Approval before merge

### Testing Strategy
- **Before merge**: Run full test suite
- **Coverage check**: Ensure >80% coverage on new files
- **Manual validation**: Test with sample repos
- **Documentation**: Update if behavior changes

---

## 📦 Deliverables by End of Sprint

### Code Deliverables
1. **PR #1 Merged**: Testing infrastructure + types
2. **PR #2 Merged**: OpenRouter provider + code collection
3. **PR #3 Merged**: LLM analysis + config generation (with tests)
4. **PR #4 Merged**: Detect command + documentation

### Testing Deliverables
- Unit tests for all components
- Integration test with mocked LLM
- Manual testing results documented
- Golden files for regression testing
- Test coverage reports

### Documentation Deliverables
- `docs/TESTING.md` - Testing approach guide
- `docs/DETECT.md` - User guide for detect command
- `examples/detected-profit.yml` - Example output
- Inline code comments
- PR descriptions

### Product Deliverables
- Working `profitctl detect` command
- OpenRouter integration complete
- Service detection from config files
- Automated config generation
- Error handling for edge cases

---

## 🔮 Next Sprint Preview (v0.0.2)

**Planned Work** (not part of this sprint):
- Ollama provider implementation
- Advanced service detection nuances
- Interactive mode (accept/reject detections)
- Caching layer for LLM responses
- Comprehensive infrastructure cost estimation
- Dependency graph visualization
- CI/CD pipeline integration

---

## 📝 Notes

### Sprint Planning Notes
- All times are estimates; adjust based on actual velocity
- Focus on completing PRs #1-#2 first (foundation critical)
- Documentation can be done in parallel if blocked
- Manual testing should use sample repos before real projects
- API key setup is prerequisite for integration tests

### For Linear Integration
- Create a Sprint in Linear called "LLM Detection v0.0.1"
- Import all LIN-xxx tasks
- Set priority: High for LIN-001 through LIN-013
- Set priority: Medium for LIN-014 through LIN-018
- Set priority: Low for documentation (can shift if needed)
- Link PRs to respective tasks
- Track burndown daily

### Git Workflow
```bash
# Feature branches
feature/llm-detection-pr1-testing  # PR #1
git checkout -b feature/llm-detection-pr1-testing

# Merge strategy
# PR #1 → main
# PR #2 → main (rebased on main after PR #1 merge)
# PR #3 → main (rebased on main after PR #2 merge)
# PR #4 → main (rebased on main after PR #3 merge)
```