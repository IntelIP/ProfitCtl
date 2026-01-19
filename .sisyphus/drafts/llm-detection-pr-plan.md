# Draft: LLM Detection PR Plan

## Requirements (Confirmed)

### Testing Strategy
- **Mock LLM providers**: Use Go interfaces + mocks (testify/mock or custom)
- **Golden file validation**: Save expected profit.yml outputs for test cases
- **Comprehensive edge cases**: Test malformed JSON, missing files, API errors
- **Testing infrastructure**: Set up from scratch with guidance

### Error Handling Requirements
- **OpenRouter malformed JSON**: Graceful error with clear message
- **No relevant files found**: Return empty config or helpful message
- **API errors**: Proper error wrapping with context

### File Size & Scope
- **Focus**: Only config files (requirements.txt, package.json, go.mod, Dockerfile, *.tf)
- **Ignore**: Source code files (won't send entire codebase to LLM)
- **Size limits**: Only applied to collected config files

### Implementation Order
1. **OpenRouter first** - Get working end-to-end
2. **Then Ollama** - Switch provider after core works
3. **Incremental PRs** - One small feature at a time, review continuously

### PR Requirements
- Everything reviewed continuously
- Testing infrastructure built out with guidance needed
- Comprehensive edge cases covered

## Technical Decisions

### Provider Priority
- Start with: OpenRouter (cloud API, easier to test)
- Add later: Ollama (local, requires model download)

### Testing Approach
- Use `testify/mock` for LLM provider mocking
- Golden files in `test/fixtures/golden/`
- Test cases: happy path, malformed JSON, missing files, API errors

### V0 Scope (Immediate)
- OpenRouter provider only
- Config file collection (requirements.txt, package.json, go.mod, Dockerfile, *.tf)
- Basic service detection
- Simple profit.yml generation
- Error handling for API failures

### Deferred to Later (v0.0.2+)
- Ollama provider
- Advanced service detection nuances
- Interactive mode
- Caching
- Comprehensive infra cost estimation
- Dependency graph visualization

## Open Questions for Final Plan

1. **Golden file organization**: Should we have one per test case or one per scenario type?
2. **Test fixtures**: Do you have example repositories we can use for integration testing?
3. **CI setup**: Should we add GitHub Actions workflow for running tests in PRs?
4. **Documentation**: Should docs/DETECT.md be in the same PR as implementation or separate?

## Next Steps

Once you confirm these decisions are correct, I'll:
1. Consult Metis to catch any missed questions/gaps
2. Create the work plan broken into incremental PRs
3. Set up testing infrastructure as the first PR
4. Build from there with continuous review

Does this capture everything correctly? Any adjustments before I create the final work plan?