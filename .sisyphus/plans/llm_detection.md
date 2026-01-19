# Work Plan: LLM-Powered Auto-Detection for profitctl

## Context

### Objective

Add intelligent codebase analysis to profitctl using LLM to automatically detect services, dependencies, and usage patterns. Generate `profit.yml` configuration files automatically by analyzing repository code, infrastructure files, and dependencies.

**Requirements**:
- Support both Ollama (local LLM) and OpenRouter (cloud LLM) providers
- Privacy option: Local analysis with Ollama (no code leaves machine)
- Convenience option: Cloud analysis with OpenRouter (no local setup)
- Provider abstraction for easy extension
- Auto-detect: services, databases, dependencies, usage patterns, cost estimates

## Architecture

```
profitctl detect
├── Code Collection → Gather relevant files (go.mod, package.json, Dockerfile, *.tf)
├── LLM Provider → Ollama (HTTP API) or OpenRouter (SDK)
├── LLM Analysis → Send code context with structured prompts
├── Response Parsing → Extract JSON detection results
└── Config Generation → Convert to profit.yml format
```

**Provider Abstraction**: `LLMProvider` interface allows switching providers without changing analyzer code.

## Implementation Tasks

### Core Components

1. **Provider Interface** (`internal/scanner/llm/types.go`, `provider.go`)
   - `LLMProvider` interface with `Chat()`, `IsAvailable()`, `ListModels()`
   - Provider factory `NewProvider(ollama|openrouter)`

2. **Ollama Provider** (`internal/scanner/llm/ollama.go`)
   - HTTP client for `http://localhost:11434`
   - Default model: `mistral:8b`
   - Health check via `/api/tags`

3. **OpenRouter Provider** (`internal/scanner/llm/openrouter.go`)
   - reVrost/go-openrouter SDK wrapper
   - API key from env/flag
   - Default model: cost-effective (claude-3-haiku)

4. **Code Collector** (`internal/scanner/collector.go`)
   - Scan for: go.mod, package.json, requirements.txt, Dockerfile, *.tf
   - Truncate large files
   - Return `CodeContext` structs

5. **LLM Analysis** (`internal/scanner/llm/analyzer.go`, `prompts.go`, `parser.go`)
   - Build structured prompts (system + user messages)
   - Request JSON output with schema
   - Parse response (extract JSON from markdown if needed)

6. **Config Generator** (`internal/scanner/llm/generator.go`)
   - Convert `AnalysisResponse` → `config.Config`
   - Map services to fixed/variable costs
   - Add default pricing/covenants if missing

7. **Detect Command** (`cmd/detect.go`)
   - Flags: `--provider` (ollama|openrouter), `--model`, `--output`, `--verbose`
   - Environment variables: `OLLAMA_HOST`, `OLLAMA_MODEL`, `OPENROUTER_API_KEY`

## Files to Create/Modify

**New Files**:
- `cmd/detect.go` - Detect command
- `internal/scanner/collector.go` - Code collection
- `internal/scanner/llm/types.go` - Types and interface
- `internal/scanner/llm/provider.go` - Provider factory
- `internal/scanner/llm/ollama.go` - Ollama implementation
- `internal/scanner/llm/openrouter.go` - OpenRouter implementation
- `internal/scanner/llm/analyzer.go` - Analysis orchestration
- `internal/scanner/llm/prompts.go` - Prompt templates
- `internal/scanner/llm/parser.go` - Response parsing
- `internal/scanner/llm/generator.go` - Config generation

**Modified Files**:
- `cmd/commands.go` - Add detectCmd
- `go.mod` - Add `github.com/reVrost/go-openrouter`
- `docs/HOW_IT_WORKS.md` - Add LLM detection section
- `docs/DETECT.md` - Usage guide (new)

## Dependencies

- `github.com/reVrost/go-openrouter` - OpenRouter SDK (new)
- Go stdlib: `net/http`, `encoding/json`, `context` - For Ollama (no new deps)

## Usage

```bash
# Ollama (local, default)
profitctl detect
profitctl detect --provider ollama --model mistral:8b

# OpenRouter (cloud)
profitctl detect --provider openrouter --model anthropic/claude-3-haiku

# Common
profitctl detect --output my-profit.yml --verbose
```

## Prerequisites

**Ollama**:
- Install: `brew install ollama` or `curl -fsSL https://ollama.com/install.sh | sh`
- Pull model: `ollama pull mistral:8b`
- Start: `ollama serve`

**OpenRouter**:
- Sign up at https://openrouter.ai
- Get API key, set `OPENROUTER_API_KEY` env var

## Success Criteria

- [ ] `profitctl detect` command works with both providers
- [ ] Code collector gathers relevant files correctly
- [ ] LLM analysis detects services, dependencies, usage patterns
- [ ] Generated profit.yml is valid and parseable
- [ ] Provider abstraction allows switching easily
- [ ] Error handling for missing Ollama/OpenRouter
- [ ] Documentation includes setup for both providers

## Considerations

**Privacy**: Ollama keeps code local. OpenRouter sends code to API.

**Cost**: Ollama free (compute only). OpenRouter per-token pricing.

**Performance**: Ollama depends on hardware. OpenRouter consistent.

**Default Models**: Ollama `mistral:8b`, OpenRouter `claude-3-haiku` (cost-effective).
