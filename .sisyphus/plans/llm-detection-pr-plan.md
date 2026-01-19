# Work Plan: LLM-Powered Auto-Detection for profitctl (Incremental PRs)

## Context

### Original Request
Implement intelligent codebase analysis using LLM to automatically detect services, dependencies, and usage patterns for automatic `profit.yml` generation.

### Interview Summary
**Testing Strategy**: 
- Mock LLM providers using Go interfaces + testify/mock
- Golden file validation for profit.yml outputs
- Comprehensive edge cases (malformed JSON, missing files, API errors)
- Testing infrastructure built from scratch with guidance

**Implementation Order**:
1. Start with OpenRouter (cloud API, easier to test)
2. Add Ollama after OpenRouter works end-to-end
3. Incremental PRs - one small feature at a time with continuous review

**Error Handling**:
- Malformed JSON from OpenRouter: graceful error
- No relevant files: return empty config with helpful message
- API errors: proper error wrapping with context

**File Scope**:
- Only config files: requirements.txt, package.json, go.mod, Dockerfile, *.tf
- Ignore source code files
- Size limits applied to config files only

**V0 Scope (Immediate)**:
- OpenRouter provider only
- Config file collection
- Basic service detection
- Simple profit.yml generation
- Error handling for API failures

**Deferred (v0.0.2+)**:
- Ollama provider
- Advanced service detection nuances
- Interactive mode, caching, infra cost estimation

### Test Infrastructure Assessment

**Current State**: Testing infrastructure EXISTS (go test, testify)
**User Preference**: Tests after implementation (not TDD initially)
**Framework**: Go standard testing + testify

**QA Strategy**: Automated tests with mocking + manual verification

---

## PR Strategy: Incremental Development

We'll create 4 sequential PRs, each building on the previous:

1. **PR #1**: Testing Infrastructure & Types Foundation
2. **PR #2**: OpenRouter Provider & Code Collection
3. **PR #3**: LLM Analysis & Config Generation
4. **PR #4**: Detect Command Integration & Documentation

---

## PR #1: Testing Infrastructure & Types Foundation

**Goal**: Establish testing patterns and core types before implementation.

### Tasks

- [ ] 1.1 Create directory structure
  ```
  internal/scanner/
  ├── collector.go
  ├── llm/
  │   ├── types.go
  │   ├── provider.go (interface)
  │   └── prompts.go
  test/
  └── fixtures/
      ├── golden/
      └── samples/
  ```
  
  **What to do**:
  - Create directories: `internal/scanner/`, `internal/scanner/llm/`, `test/fixtures/`
  - These will house our detection code and test data
  
  **References**:
  - Existing structure: `internal/config/`, `internal/cost/` - Follow same patterns
  
  **Acceptance**:
  - Directories exist
  - `ls -la internal/scanner/` shows llm subdirectory

- [ ] 1.2 Define core types (`internal/scanner/llm/types.go`)
  
  **What to do**:
  ```go
  package llm
  
  type CodeContext struct {
      Files map[string]string // path -> content
  }
  
  type AnalysisRequest struct {
      Context CodeContext
      Prompt  string
  }
  
  type AnalysisResponse struct {
      Services      []DetectedService
      Dependencies  []DetectedDependency
      Patterns      []UsagePattern
  }
  
  type DetectedService struct {
      Name        string
      Type        string // "database", "cache", "api", "storage"
      Provider    string // "aws", "gcp", "custom"
      FixedCost   *float64
      VariableCost *float64
  }
  
  type LLMProvider interface {
      Chat(ctx context.Context, messages []Message) (string, error)
      IsAvailable(ctx context.Context) bool
  }
  
  type Message struct {
      Role    string
      Content string
  }
  ```
  
  **References**:
  - `cmd/simulate.go:16-20` - See flag variable patterns
  - `internal/config/parser.go:11-51` - Config struct patterns
  - `pkg/types/structures.go:4-34` - Struct field patterns
  
  **Acceptance**:
  - File compiles: `go build ./internal/scanner/llm/types.go`
  - Types are exported properly (capitalized names)
  - Struct tags match existing patterns (yaml, json as needed)

- [ ] 1.3 Create prompt templates (`internal/scanner/llm/prompts.go`)
  
  **What to do**:
  ```go
  package llm
  
  const SystemPrompt = `You are a service detection assistant. Analyze the provided configuration files and identify services, dependencies, and usage patterns. Respond ONLY with valid JSON matching the schema.`
  
  func BuildDetectionPrompt(context CodeContext) string {
      // Concatenate file contents with paths
      var sb strings.Builder
      sb.WriteString("Analyze these configuration files:\n\n")
      for path, content := range context.Files {
          sb.WriteString(fmt.Sprintf("## %s\n\n", path))
          sb.WriteString("```\n")
          sb.WriteString(truncate(content, 2000))
          sb.WriteString("\n```\n\n")
      }
      sb.WriteString("\nRespond with JSON:\n")
      sb.WriteString(`{"services": [{"name": "", "type": "", "provider": "", "fixed_cost": null, "variable_cost": null}], "dependencies": [{"name": "", "type": ""}], "patterns": []}`)
      return sb.String()
  }
  
  func truncate(s string, maxLen int) string {
      if len(s) > maxLen {
          return s[:maxLen] + "... (truncated)"
      }
      return s
  }
  ```
  
  **References**:
  - No direct reference - this is new functionality
  - Test truncation logic manually
  
  **Acceptance**:
  - `BuildDetectionPrompt` creates well-formatted markdown
  - Truncation works (test with 3000 char string, verify output < 2000)
  - Log output to verify format: `fmt.Println(BuildDetectionPrompt(testContext))`

- [ ] 1.4 Set up test infrastructure (explain Go testing to user)
  
  **What to do**: Document testing approach for future PR reviewers
  
  ```markdown
  # Testing Approach for LLM Detection
  
  ## Go Testing Basics (for Python/TypeScript background)
  
  ### File Naming
  - Test files: `*_test.go` (same package as code)
  - Golden files: `test/fixtures/golden/*.yml`
  - Sample configs: `test/fixtures/samples/*/*`
  
  ### Test Structure
  ```go
  func TestFunctionName(t *testing.T) {
      // Arrange
      input := ...
      
      // Act
      result := FunctionName(input)
      
      // Assert
      assert.Equal(t, expected, result)
  }
  ```
  
  ### Mocking LLM Provider
  We'll create a mock implementation of LLMProvider interface for testing:
  ```go
  type MockLLMProvider struct {
      mock.Mock
  }
  
  func (m *MockLLMProvider) Chat(ctx context.Context, msgs []Message) (string, error) {
      args := m.Called(ctx, msgs)
      return args.String(0), args.Error(1)
  }
  ```
  
  ### Golden File Testing
  ```go
  func TestGeneratedConfig(t *testing.T) {
      // Generate config
      got := generator.Generate(...)
      
      // Read golden file
      want := readGoldenFile(t, "test/fixtures/golden/simple.yml")
      
      // Compare
      assert.Equal(t, want, got)
  }
  ```
  
  ### Running Tests
  ```bash
  # All tests
  go test ./...
  
  # Specific package
  go test ./internal/scanner/llm/...
  
  # With coverage
  go test -cover ./...
  
  # Verbose
  go test -v ./internal/scanner/llm/...
  ```
  
  ### Test Categories
  1. **Unit tests**: Mock everything external (LLM, file system)
  2. **Integration tests**: Use real file system, mock LLM
  3. **Golden tests**: Compare output against committed golden files
  ```
  
  **Acceptance**:
  - File created: `docs/TESTING.md` (or add to existing docs)
  - Content covers all basics listed above
  - User understands how to write and run tests

- [ ] 1.5 Create first golden file test case
  
  **What to do**:
  - Create sample: `test/fixtures/samples/go-webapp/go.mod`
  ```go
  module example.com/webapp
  
  go 1.21
  
  require (
      github.com/gin-gonic/gin v1.9.1
      github.com/lib/pq v1.10.9
  )
  ```
  
  - Create expected: `test/fixtures/golden/go-webapp-expected.yml`
  ```yaml
  # This is what LLM should generate
  project:
    name: webapp
  
  fixed_costs:
    - name: Database
      amount: 15.00
      period: monthly
      layer: infrastructure
  
  variable_costs:
    - name: API Requests
      cost_per_unit: 0.001
      units_per_user: 100.0
      layer: service
  
  # ... rest of config
  ```
  
  **References**:
  - `examples/valid_profit.yml` - Use as template for golden file
  
  **Acceptance**:
  - Golden file is valid YAML (can be parsed)
  - Reflects expected detection from a Go web app go.mod
  - Sample file exists for testing

- [ ] 1.6 Write unit tests for prompt building
  
  **What to do** (`internal/scanner/llm/prompts_test.go`):
  ```go
  package llm
  
  import (
      "strings"
      "testing"
      "github.com/stretchr/testify/assert"
  )
  
  func TestBuildDetectionPrompt(t *testing.T) {
      context := CodeContext{
          Files: map[string]string{
              "go.mod": "module example.com/test\n\ngo 1.21\n",
              "requirements.txt": "flask==2.3.3\nredis==4.6.0",
          },
      }
      
      prompt := BuildDetectionPrompt(context)
      
      // Should contain file headers
      assert.Contains(t, prompt, "## go.mod")
      assert.Contains(t, prompt, "## requirements.txt")
      assert.Contains(t, prompt, "module example.com/test")
      assert.Contains(t, prompt, "flask==2.3.3")
  }
  
  func TestTruncate(t *testing.T) {
      longStr := strings.Repeat("a", 3000)
      truncated := truncate(longStr, 2000)
      
      assert.Len(t, truncated, 2000+len("... (truncated)"))
      assert.Contains(t, truncated, "... (truncated)")
  }
  ```
  
  **References**:
  - See `internal/config/parser_test.go` - Example test structure
  - Use `go test -v ./internal/scanner/llm/...` to run
  
  **Acceptance**:
  - Tests pass: `go test ./internal/scanner/llm/...`
  - Coverage shows lines covered
  - Tests are readable and well-commented

### PR #1 Review Checklist

- [ ] Directory structure created
- [ ] Core types defined and compile
- [ ] Prompt templates implemented
- [ ] Testing strategy documented
- [ ] Golden files created
- [ ] Unit tests written and pass
- [ ] Code reviewed via PR

---

## PR #2: OpenRouter Provider & Code Collection

**Goal**: Implement LLM provider and file collection logic.

### Tasks

- [ ] 2.1 Implement OpenRouter provider (`internal/scanner/llm/openrouter.go`)
  
  **What to do**:
  ```go
  package llm
  
  import (
      "context"
      "fmt"
      "os"
      "github.com/reVrost/go-openrouter"
  )
  
  type OpenRouterProvider struct {
      client *openrouter.Client
      model  string
  }
  
  func NewOpenRouterProvider(apiKey string, model string) *OpenRouterProvider {
      client := openrouter.NewClient(apiKey)
      if model == "" {
          model = "anthropic/claude-3-haiku" // Cost-effective default
      }
      return &OpenRouterProvider{
          client: client,
          model:  model,
      }
  }
  
  func (p *OpenRouterProvider) Chat(ctx context.Context, messages []Message) (string, error) {
      // Convert our Message type to openrouter.Message
      orMessages := make([]openrouter.Message, len(messages))
      for i, msg := range messages {
          orMessages[i] = openrouter.Message{
              Role:    msg.Role,
              Content: msg.Content,
          }
      }
      
      resp, err := p.client.ChatCompletion(ctx, &openrouter.ChatCompletionRequest{
          Model:    p.model,
          Messages: orMessages,
      })
      if err != nil {
          return "", fmt.Errorf("OpenRouter API error: %w", err)
      }
      
      if len(resp.Choices) == 0 {
          return "", fmt.Errorf("no response from LLM")
      }
      
      return resp.Choices[0].Message.Content, nil
  }
  
  func (p *OpenRouterProvider) IsAvailable(ctx context.Context) bool {
      // Simple health check - we could make a cheap API call
      return p.client != nil
  }
  ```
  
  **References**:
  - `go-openrouter` docs: Will need to install first
  - Similar pattern to existing clients in the codebase
  
  **Acceptance**:
  - File compiles: `go build ./internal/scanner/llm/openrouter.go`
  - Constructor handles empty model with default
  - Error messages are informative

- [ ] 2.2 Add OpenRouter dependency (`go.mod`)
  
  **What to do**:
  ```bash
  go get github.com/reVrost/go-openrouter
  go mod tidy
  ```
  
  **References**:
  - Check existing `go.mod` - see line 6 for dependency format
  
  **Acceptance**:
  - `go.mod` updated with new dependency
  - `go.sum` updated with checksums
  - `go build ./...` succeeds

- [ ] 2.3 Implement code collector (`internal/scanner/collector.go`)
  
  **What to do**:
  ```go
  package scanner
  
  import (
      "io/fs"
      "os"
      "path/filepath"
      "strings"
      "go.uber.org/multierr" // We may want this for collecting multiple errors
  )
  
  type Collector struct {
      // Config for what to collect
      MaxFileSize int64
  }
  
  func NewCollector() *Collector {
      return &Collector{
          MaxFileSize: 1024 * 1024, // 1MB limit for config files
      }
  }
  
  func (c *Collector) Collect(root string) (llm.CodeContext, error) {
      context := llm.CodeContext{
          Files: make(map[string]string),
      }
      
      patterns := []string{
          "go.mod",
          "package.json",
          "requirements.txt",
          "Pipfile",
          "Dockerfile",
          "docker-compose.yml",
          "*.tf", // Terraform files
          "*.yaml", // Kubernetes manifests
          "*.yml",
      }
      
      var errs error
      
      for _, pattern := range patterns {
          matches, err := filepath.Glob(filepath.Join(root, pattern))
          if err != nil {
              errs = multierr.Append(errs, err)
              continue
          }
          
          for _, match := range matches {
              info, err := os.Stat(match)
              if err != nil {
                  errs = multierr.Append(errs, err)
                  continue
              }
              
              if info.Size() > c.MaxFileSize {
                  continue // Skip large files
              }
              
              content, err := os.ReadFile(match)
              if err != nil {
                  errs = multierr.Append(errs, err)
                  continue
              }
              
              // Store relative path
              relPath, _ := filepath.Rel(root, match)
              context.Files[relPath] = string(content)
          }
      }
      
      return context, errs // Return partial results even with errors
  }
  ```
  
  **References**:
  - `config.ParseConfig` in `internal/config/parser.go:54` - file reading pattern
  - Check for other file collection patterns in codebase
  
  **Acceptance**:
  - Collects all matching files in a test directory
  - Respects MaxFileSize limit
  - Returns partial results on errors
  - Test with actual project: `collector.Collect(".")` should find go.mod

- [ ] 2.4 Write unit tests for OpenRouter provider
  
  **What to do** (`internal/scanner/llm/openrouter_test.go`):
  ```go
  package llm
  
  import (
      "context"
      "testing"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/mock"
  )
  
  func TestNewOpenRouterProvider(t *testing.T) {
      provider := NewOpenRouterProvider("test-key", "")
      assert.NotNil(t, provider)
      assert.Equal(t, "anthropic/claude-3-haiku", provider.model)
  }
  
  // Integration test (requires API key)
  func TestOpenRouterProviderIntegration(t *testing.T) {
      if os.Getenv("OPENROUTER_API_KEY") == "" {
          t.Skip("OPENROUTER_API_KEY not set")
      }
      
      provider := NewOpenRouterProvider(os.Getenv("OPENROUTER_API_KEY"), "")
      ctx := context.Background()
      
      messages := []Message{
          {Role: "user", Content: "Say hello"},
      }
      
      resp, err := provider.Chat(ctx, messages)
      assert.NoError(t, err)
      assert.NotEmpty(t, resp)
  }
  ```
  
  **References**:
  - No existing OpenRouter tests in codebase - this is new
  - Follow pattern from `cmd/simulate_test.go`
  
  **Acceptance**:
  - Tests compile and pass
  - Integration test skipped without API key
  - `go test ./internal/scanner/llm/... -v` shows results

- [ ] 2.5 Write unit tests for collector
  
  **What to do** (`internal/scanner/collector_test.go`):
  ```go
  package scanner
  
  import (
      "os"
      "path/filepath"
      "testing"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )
  
  func TestCollector_Collect(t *testing.T) {
      // Create temp test directory
      tmpDir := t.TempDir()
      
      // Create test files
      testFiles := map[string]string{
          "go.mod": "module example.com/test\ngo 1.21",
          "package.json": `{"name": "test", "version": "1.0.0"}`,
          "requirements.txt": "flask==2.3.3\nredis==4.6.0",
          "Dockerfile": "FROM node:18\nWORKDIR /app",
          "main.go": "package main\n\nfunc main() {}", // Should be ignored
      }
      
      for name, content := range testFiles {
          path := filepath.Join(tmpDir, name)
          err := os.WriteFile(path, []byte(content), 0644)
          require.NoError(t, err)
      }
      
      collector := NewCollector()
      context, err := collector.Collect(tmpDir)
      
      assert.NoError(t, err)
      assert.Len(t, context.Files, 4) // main.go should not be collected
      assert.Contains(t, context.Files, "go.mod")
      assert.Contains(t, context.Files, "package.json")
      assert.NotContains(t, context.Files, "main.go")
  }
  
  func TestCollector_SkipLargeFiles(t *testing.T) {
      tmpDir := t.TempDir()
      
      // Create large file (>1MB)
      largeContent := make([]byte, 2*1024*1024)
      err := os.WriteFile(filepath.Join(tmpDir, "large.tf"), largeContent, 0644)
      require.NoError(t, err)
      
      collector := NewCollector()
      context, err := collector.Collect(tmpDir)
      
      assert.NoError(t, err)
      assert.NotContains(t, context.Files, "large.tf")
  }
  ```
  
  **References**:
  - `testing.T.TempDir()` pattern from standard Go testing
  - `require.NoError` pattern from testify
  
  **Acceptance**:
  - Both tests pass
  - Tests demonstrate collection works correctly
  - Size limit test proves files are skipped

### PR #2 Review Checklist

- [ ] OpenRouter provider implemented and compiles
- [ ] Dependency added to go.mod
- [ ] Collector finds config files correctly
- [ ] Unit tests for OpenRouter provider
- [ ] Unit tests for collector
- [ ] Integration test documents API key requirement
- [ ] Code reviewed via PR

---

## PR #3: LLM Analysis & Config Generation

**Goal**: Connect provider with analysis and generate profit.yml.

### Tasks

- [ ] 3.1 Implement response parser (`internal/scanner/llm/parser.go`)
  
  **What to do**:
  ```go
  package llm
  
  import (
      "encoding/json"
      "fmt"
      "regexp"
      "strings"
  )
  
  // ParseResponse extracts JSON from LLM response (handles markdown code blocks)
  func ParseResponse(raw string) (*AnalysisResponse, error) {
      // Remove markdown code block markers if present
      jsonStr := raw
      
      // Try to extract JSON from ```json ... ``` blocks
      re := regexp.MustCompile("`{3}(?:json)?\n(.*?)\n`{3}")
      matches := re.FindStringSubmatch(raw)
      if len(matches) > 1 {
          jsonStr = matches[1]
      }
      
      // Trim whitespace
      jsonStr = strings.TrimSpace(jsonStr)
      
      var response AnalysisResponse
      if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
          return nil, fmt.Errorf("failed to parse LLM response JSON: %w\nResponse: %s", err, jsonStr)
      }
      
      return &response, nil
  }
  ```
  
  **References**:
  - Standard Go JSON parsing patterns
  - No existing parser in codebase - this is new
  
  **Acceptance**:
  - Parses JSON from: `{"services": [{"name": "test"}]}`
  - Parses JSON from markdown: "```json\n{...}\n```"
  - Returns error for malformed JSON with helpful message
  - Test manually with sample LLM response

- [ ] 3.2 Write tests for parser
  
  **What to do** (`internal/scanner/llm/parser_test.go`):
  ```go
  package llm
  
  import (
      "testing"
      "github.com/stretchr/testify/assert"
  )
  
  func TestParseResponse(t *testing.T) {
      tests := []struct {
          name    string
          input   string
          wantErr bool
      }{
          {
              name: "valid JSON",
              input: `{"services": [{"name": "test", "type": "database"}]}`,
              wantErr: false,
          },
          {
              name: "JSON in markdown block",
              input: "```json\n{\"services\": [{\"name\":\"test\"}]}\n```",
              wantErr: false,
          },
          {
              name: "malformed JSON",
              input: `{"services": [}`,
              wantErr: true,
          },
          {
              name: "empty response",
              input: "",
              wantErr: true,
          },
      }
      
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              result, err := ParseResponse(tt.input)
              if tt.wantErr {
                  assert.Error(t, err)
              } else {
                  assert.NoError(t, err)
                  assert.NotNil(t, result)
              }
          })
      }
  }
  ```
  
  **References**:
  - Standard table-driven test pattern in Go
  - See `internal/config/parser_test.go` for examples
  
  **Acceptance**:
  - All test cases pass
  - Malformed JSON shows helpful error
  - Empty string handled correctly

- [ ] 3.3 Implement config generator (`internal/scanner/llm/generator.go`)
  
  **What to do**:
  ```go
  package llm
  
  import (
      "github.com/profitctl/profitctl/internal/config"
      "github.com/profitctl/profitctl/pkg/types"
  )
  
  type Generator struct{}
  
  func NewGenerator() *Generator {
      return &Generator{}
  }
  
  // Generate converts AnalysisResponse to config.Config
  func (g *Generator) Generate(analysis *AnalysisResponse) (*config.Config, error) {
      cfg := &config.Config{
          Project: &config.ProjectInfo{
              Name: "detected-project", // Could extract from context
          },
          FixedCosts:    []types.FixedCost{},
          VariableCosts: []types.VariableCost{},
          Pricing:       nil, // User will add this manually
          Covenants:     []config.Covenant{},
          Simulation:    nil, // User will add this manually
      }
      
      // Convert detected services to costs
      for _, svc := range analysis.Services {
          if svc.FixedCost != nil && *svc.FixedCost > 0 {
              cfg.FixedCosts = append(cfg.FixedCosts, types.FixedCost{
                  Name:   svc.Name,
                  Amount: *svc.FixedCost,
                  Period: types.PeriodMonthly,
                  Layer:  types.LayerInfrastructure, // Could map from svc.Type
              })
          }
          
          if svc.VariableCost != nil && *svc.VariableCost > 0 {
              cfg.VariableCosts = append(cfg.VariableCosts, types.VariableCost{
                  Name:         svc.Name,
                  CostPerUnit:  *svc.VariableCost,
                  UnitsPerUser: 1.0, // Default 1 unit per user
                  Distribution: types.DistNormal,
                  Layer:        types.LayerService,
              })
          }
      }
      
      // Add default covenants for detected services
      if len(cfg.FixedCosts) > 0 || len(cfg.VariableCosts) > 0 {
          cfg.Covenants = append(cfg.Covenants, config.Covenant{
              Type:    "threshold",
              Field:   "margin",
              Operator: "gt",
              Value:   0.0,
              Message: "Margin must be positive for profitability",
          })
      }
      
      return cfg, nil
  }
  ```
  
  **References**:
  - `internal/config/parser.go:11-18` - Config struct construction
  - `pkg/types/structures.go:4-26` - FixedCost and VariableCost construction
  - `internal/config/parser.go:39-44` - Covenant construction
  
  **Acceptance**:
  - Generator produces valid config.Config
  - All required fields populated
  - Map FixedCost/VariableCost from detected services
  - Run `go test ./internal/config/...` to validate generated config parses

- [ ] 3.4 Write integration test: Collector → LLM → Parser → Generator
  
  **What to do** (`internal/scanner/llm/integration_test.go`):
  ```go
  package llm
  
  import (
      "context"
      "testing"
      "github.com/profitctl/profitctl/internal/scanner"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/mock"
  )
  
  // MockLLMProvider for testing
type MockLLMProvider struct {
      mock.Mock
  }

  func (m *MockLLMProvider) Chat(ctx context.Context, messages []Message) (string, error) {
      args := m.Called(ctx, messages)
      return args.String(0), args.Error(1)
  }

  func (m *MockLLMProvider) IsAvailable(ctx context.Context) bool {
      args := m.Called(ctx)
      return args.Bool(0)
  }

  func TestEndToEndDetection(t *testing.T) {
      // 1. Collect files
      collector := scanner.NewCollector()
      context, err := collector.Collect("./test/fixtures/samples/go-webapp")
      assert.NoError(t, err)
      
      // 2. Mock LLM response
      mockProvider := &MockLLMProvider{}
      mockProvider.On("Chat", mock.Anything, mock.Anything).Return(
          `{"services": [{"name": "PostgreSQL", "type": "database", "provider": "aws", "fixed_cost": 15.0}]}`,
          nil,
      )
      
      // 3. Call LLM
      prompt := BuildDetectionPrompt(context)
      rawResponse, err := mockProvider.Chat(context.Background(), []Message{
          {Role: "system", Content: SystemPrompt},
          {Role: "user", Content: prompt},
      })
      assert.NoError(t, err)
      
      // 4. Parse response
      analysis, err := ParseResponse(rawResponse)
      assert.NoError(t, err)
      assert.Len(t, analysis.Services, 1)
      assert.Equal(t, "PostgreSQL", analysis.Services[0].Name)
      
      // 5. Generate config
      generator := NewGenerator()
      cfg, err := generator.Generate(analysis)
      assert.NoError(t, err)
      assert.NotNil(t, cfg)
      assert.Len(t, cfg.FixedCosts, 1)
      assert.Equal(t, 15.0, cfg.FixedCosts[0].Amount)
  }
  ```
  
  **References**:
  - testify/mock documentation
  - `internal/simulation/montecarlo_test.go` - complex test example
  
  **Acceptance**:
  - Test demonstrates full flow
  - Mock properly isolates LLM
  - Verify golden file output matches expected
  - `go test -v ./internal/scanner/llm/...` shows test passing

### PR #3 Review Checklist

- [ ] Response parser extracts JSON from markdown
- [ ] Parser handles malformed JSON gracefully
- [ ] Config generator creates valid config structs
- [ ] Services map correctly to FixedCost/VariableCost
- [ ] Integration test demonstrates full flow
- [ ] Error cases covered
- [ ] Code reviewed via PR

---

## PR #4: Detect Command Integration & Documentation

**Goal**: Wire everything together and create the CLI command.

### Tasks

- [ ] 4.1 Implement Analyzer orchestrator (`internal/scanner/llm/analyzer.go`)
  
  **What to do**:
  ```go
  package llm
  
  import (
      "context"
      "fmt"
      "gopkg.in/yaml.v3"
      "os"
      "github.com/profitctl/profitctl/internal/scanner"
  )
  
  type Analyzer struct {
      provider LLMProvider
      collector *scanner.Collector
      generator *Generator
  }
  
  func NewAnalyzer(provider LLMProvider) *Analyzer {
      return &Analyzer{
          provider: provider,
          collector: scanner.NewCollector(),
          generator: NewGenerator(),
      }
  }
  
  func (a *Analyzer) Analyze(ctx context.Context, rootPath string) (*config.Config, error) {
      // 1. Collect files
      fmt.Fprintf(os.Stderr, "Collecting configuration files from %s...\n", rootPath)
      context, err := a.collector.Collect(rootPath)
      if err != nil {
          fmt.Fprintf(os.Stderr, "Warning: some errors during collection: %v\n", err)
      }
      
      if len(context.Files) == 0 {
          return nil, fmt.Errorf("no configuration files found. Looking for: go.mod, package.json, requirements.txt, Dockerfile, *.tf")
      }
      
      // 2. Build and send prompt
      prompt := BuildDetectionPrompt(context)
      fmt.Fprintf(os.Stderr, "Sending %d bytes to LLM...\n", len(prompt))
      
      rawResponse, err := a.provider.Chat(ctx, []Message{
          {Role: "system", Content: SystemPrompt},
          {Role: "user", Content: prompt},
      })
      if err != nil {
          return nil, fmt.Errorf("LLM analysis failed: %w", err)
      }
      
      // 3. Parse response
      analysis, err := ParseResponse(rawResponse)
      if err != nil {
          return nil, fmt.Errorf("failed to parse LLM response: %w", err)
      }
      
      fmt.Fprintf(os.Stderr, "Detected %d services\n", len(analysis.Services))
      
      // 4. Generate config
      cfg, err := a.generator.Generate(analysis)
      if err != nil {
          return nil, fmt.Errorf("failed to generate config: %w", err)
      }
      
      return cfg, nil
  }
  
  // Save writes config to file
  func Save(cfg *config.Config, outputPath string) error {
      data, err := yaml.Marshal(cfg)
      if err != nil {
          return fmt.Errorf("failed to marshal config: %w", err)
      }
      
      return os.WriteFile(outputPath, data, 0644)
  }
  ```
  
  **References**:
  - `internal/config/parser.go:54-66` - File reading/writing patterns
  - `internal/output/*.go` - Output formatting patterns
  
  **Acceptance**:
  - Analyzer orchestrates all steps
  - Shows progress to stderr (verbose mode)
  - Handles empty file collection gracefully
  - Error messages explain what failed

- [ ] 4.2 Create detect command (`cmd/detect.go`)
  
  **What to do**:
  ```go
  package cmd
  
  import (
      "context"
      "fmt"
      "os"
      "github.com/profitctl/profitctl/internal/scanner/llm"
      "github.com/spf13/cobra"
  )
  
  var (
      detectProvider string
      detectModel    string
      detectOutput   string
      detectVerbose  bool
  )
  
  var detectCmd = &cobra.Command{
      Use:   "detect",
      Short: "Auto-detect services and generate profit.yml",
      Long: `Analyze codebase configuration files to automatically detect services,
dependencies, and usage patterns. Generates a profit.yml configuration file
using LLM analysis (OpenRouter or Ollama).`,
      RunE: runDetect,
  }
  
  func init() {
      detectCmd.Flags().StringVar(&detectProvider, "provider", "openrouter", "LLM provider (openrouter)")
      detectCmd.Flags().StringVar(&detectModel, "model", "", "Specific model to use")
      detectCmd.Flags().StringVarP(&detectOutput, "output", "o", "profit.yml", "Output file path")
      detectCmd.Flags().BoolVarP(&detectVerbose, "verbose", "v", false, "Verbose output")
  }
  
  func runDetect(cmd *cobra.Command, args []string) error {
      // Get API key from env var
      apiKey := os.Getenv("OPENROUTER_API_KEY")
      if apiKey == "" {
          return fmt.Errorf("OPENROUTER_API_KEY environment variable not set. Get one from https://openrouter.ai")
      }
      
      // Determine model
      model := detectModel
      if model == "" {
          // Default to cost-effective model
          model = "anthropic/claude-3-haiku"
      }
      
      // Create provider
      provider := llm.NewOpenRouterProvider(apiKey, model)
      if !provider.IsAvailable(context.Background()) {
          return fmt.Errorf("OpenRouter provider not available - check API key and connectivity")
      }
      
      // Find root path (current directory or first arg)
      rootPath := "."
      if len(args) > 0 {
          rootPath = args[0]
      }
      
      if detectVerbose {
          fmt.Printf("Detecting from: %s\n", rootPath)
          fmt.Printf("Using model: %s\n", model)
      }
      
      // Run analysis
      analyzer := llm.NewAnalyzer(provider)
      cfg, err := analyzer.Analyze(context.Background(), rootPath)
      if err != nil {
          return err
      }
      
      // Save config
      if err := llm.Save(cfg, detectOutput); err != nil {
          return fmt.Errorf("failed to save config: %w", err)
      }
      
      fmt.Printf("Generated %s\n", detectOutput)
      
      if detectVerbose {
          fmt.Printf("\nDetected services: %d\n", len(cfg.FixedCosts)+len(cfg.VariableCosts))
          for _, fc := range cfg.FixedCosts {
              fmt.Printf("  - %s: $%.2f/%s\n", fc.Name, fc.Amount, fc.Period)
          }
          for _, vc := range cfg.VariableCosts {
              fmt.Printf("  - %s: $%.4f per user\n", vc.Name, vc.CostPerUnit)
          }
      }
      
      return nil
  }
  ```
  
  **References**:
  - `cmd/simulate.go:38-140` - Command structure pattern
  - `cmd/simulate.go:16-20` - Flag variable pattern
  - `cmd/root.go:21-24` - Adding commands to root
  - `cmd/commands.go:37-44` - validateCmd pattern
  
  **Acceptance**:
  - Command shows in help: `profitctl detect --help`
  - Flags work: `--provider`, `--model`, `--output`, `--verbose`
  - Requires OPENROUTER_API_KEY env var
  - Shows helpful errors if key missing
  - Verbose mode shows progress

- [ ] 4.3 Add detect command to commands (`cmd/commands.go`)
  
  **What to do**:
  ```go
  // In AddCommands function, add:
  detectCmd := &cobra.Command{
      Use:   "detect",
      Short: "Auto-detect services and generate profit.yml",
      Run: func(cmd *cobra.Command, args []string) {
          fmt.Println("Detect command not yet implemented - see PR #4")
      },
  }
  
  // And add to rootCmd.AddCommand()
  rootCmd.AddCommand(detectCmd, simulateCmd, validateCmd)
  ```
  
  **References**:
  - `cmd/commands.go:46` - rootCmd.AddCommand pattern
  
  **Acceptance**:
  - `profitctl --help` shows detect command
  - Command added to root command structure

- [ ] 4.4 Create usage documentation (`docs/DETECT.md`)
  
  **What to do**:
  ```markdown
  # profitctl detect
  
  Auto-detect services and generate profit.yml using LLM analysis.
  
  ## Prerequisites
  
  ### OpenRouter Setup
  
  1. Sign up at https://openrouter.ai
  2. Get your API key from the dashboard
  3. Set environment variable:
     ```bash
     export OPENROUTER_API_KEY="your-key-here"
     ```
  
  ## Usage
  
  ```bash
  # Detect from current directory
  profitctl detect
  
  # Detect from specific directory
  profitctl detect /path/to/project
  
  # Specify custom output file
  profitctl detect -o my-profit.yml
  
  # Use specific model
  profitctl detect --model anthropic/claude-3-sonnet
  
  # Verbose output
  profitctl detect -v
  ```
  
  ## Configuration Files Detected
  
  The detector looks for:
  - `go.mod` - Go dependencies
  - `package.json` - Node.js dependencies
  - `requirements.txt`, `Pipfile` - Python dependencies
  - `Dockerfile`, `docker-compose.yml` - Container configuration
  - `*.tf` - Terraform infrastructure (AWS, GCP, Azure)
  - `*.yaml`, `*.yml` - Kubernetes manifests
  
  ## Example
  
  For a Go web app with PostgreSQL:
  
  ```bash
  $ profitctl detect -v
  Detecting from: .
  Using model: anthropic/claude-3-haiku
  Collecting configuration files from ...
  Sending 1847 bytes to LLM...
  Detected 2 services
  Generated profit.yml
  
  Detected services:
    - PostgreSQL: $15.00/monthly
    - Redis Cache: $0.0001 per user
  ```
  
  Generated `profit.yml`:
  ```yaml
  project:
    name: my-webapp
  
  fixed_costs:
    - name: PostgreSQL
      amount: 15.00
      period: monthly
      layer: infrastructure
  
  variable_costs:
    - name: Redis Cache
      cost_per_unit: 0.0001
      units_per_user: 1
      layer: service
  
  covenants:
    - type: threshold
      field: margin
      operator: gt
      value: 0
      message: Margin must be positive for profitability
  ```
  
  ## Troubleshooting
  
  ### "OPENROUTER_API_KEY not set"
  Set the environment variable with your API key.
  
  ### "no configuration files found"
  Make sure you're running from a directory with supported config files (go.mod, package.json, etc.).
  
  ### "LLM analysis failed"
  Check your internet connection and API key validity.
  
  ## Costs
  
  OpenRouter charges per token. Using claude-3-haiku is ~$0.25 per analysis. Heavier models (claude-3-sonnet) cost more but may give better results.
  ```
  
  **References**:
  - `docs/HOW_IT_WORKS.md` - Follow formatting style
  - `examples/valid_profit.yml` - Example config format
  
  **Acceptance**:
  - Complete documentation with examples
  - Troubleshooting section covers common errors
  - Cost estimates included
  - Matches existing doc formatting

- [ ] 4.5 Manual testing
  
  **What to do**: Test with real repositories
  
  Create test repositories:
  ```bash
  mkdir -p /tmp/test-detect/{go-app,node-app,python-app}
  ```
  
  **Test 1: Go app** (`/tmp/test-detect/go-app/go.mod`):
  ```go
  module test-app
  
  go 1.21
  
  require (
      github.com/gin-gonic/gin v1.9.1
      github.com/lib/pq v1.10.9
  )
  ```
  
  ```bash
  cd /tmp/test-detect/go-app
  profitctl detect -v
  # Verify: Should detect web server and database
  cat profit.yml
  ```
  
  **Test 2: Node app** (`/tmp/test-detect/node-app/package.json`):
  ```json
  {
    "name": "api-server",
    "dependencies": {
      "express": "^4.18.2",
      "redis": "^4.6.0"
    }
  }
  ```
  
  ```bash
  cd /tmp/test-detect/node-app
  profitctl detect -v
  # Verify: Should detect Express and Redis
  ```
  
  **Test 3: No files** (`/tmp/test-detect/empty`):
  ```bash
  cd /tmp/test-detect/empty
  profitctl detect
  # Expected: "no configuration files found" error
  ```
  
  **Test 4: Large files (should be skipped)**:
  ```bash
  cd /tmp/test-detect/large-file
  dd if=/dev/zero of=big.tf bs=1M count=3  # 3MB file
  profitctl detect -v
  # Should skip large.tf
  ```
  
  **Acceptance**:
  - All 4 tests performed
  - Generated profit.yml files are valid (can be parsed)
  - Error cases show helpful messages
  - Document results for PR review

- [ ] 4.6 Add example to examples/ (`examples/detected-profit.yml`)
  
  **What to do**: Save one good example
  ```yaml
  # Example generated by profitctl detect for a Go web app
  # Source: github.com/example/webapi
  
  project:
    name: webapi
  
  fixed_costs:
    - name: PostgreSQL Database
      amount: 15.00
      period: monthly
      layer: infrastructure
    
    - name: Redis Cache
      amount: 10.00
      period: monthly
      layer: infrastructure
  
  variable_costs:
    - name: API Requests
      cost_per_unit: 0.0001
      units_per_user: 100
      layer: service
  
  pricing:
    plans:
      - name: Starter
        price: 9.99
        limits:
          users: 100
      - name: Professional
        price: 49.99
        limits:
          users: 1000
  
  simulation:
    base_users: 100
    growth_factor: 2.0
    iterations: 1000
  
  covenants:
    - type: threshold
      field: margin
      operator: gt
      value: 0
      message: Must maintain positive margins
  ```
  
  **Acceptance**:
  - Example file created
  - Well-commented
  - Represents realistic detection results

### PR #4 Review Checklist

- [ ] Detect command works end-to-end
- [ ] Help text comprehensive
- [ ] Documentation complete (docs/DETECT.md)
- [ ] Manual testing completed and documented
- [ ] Example config added
- [ ] Error messages helpful
- [ ] Code reviewed via PR

---

## Testing Strategy Summary

### Test Coverage Goals
- **Unit tests**: All components (provider, collector, parser, generator) >80%
- **Integration tests**: End-to-end flow with mocks
- **Manual tests**: Real repositories with OpenRouter

### Test Files Created
```
test/fixtures/
├── golden/
│   ├── go-webapp-expected.yml
│   └── node-api-expected.yml
└── samples/
    ├── go-webapp/
    │   └── go.mod
    ├── node-api/
    │   └── package.json
    └── python-api/
        └── requirements.txt

internal/scanner/llm/
├── openrouter_test.go
├── prompts_test.go
├── parser_test.go
└── integration_test.go

internal/scanner/
└── collector_test.go
```

### Running Tests
```bash
# All unit tests
go test ./internal/scanner/...

# With coverage
go test -cover ./internal/scanner/...

# Integration test (requires OPENROUTER_API_KEY)
OPENROUTER_API_KEY=xxx go test -v ./internal/scanner/llm/... -run Integration

# Manual testing
profitctl detect -v ./test/fixtures/samples/go-webapp
```

---

## Success Criteria

### Overall Deliverables
- [ ] All 4 PRs reviewed and merged
- [ ] Testing infrastructure established
- [ ] OpenRouter provider working end-to-end
- [ ] User can run `profitctl detect` successfully
- [ ] Documentation complete
- [ ] Example configs provided

### Technical Success
- [ ] Generated profit.yml is valid and parseable by `profitctl simulate`
- [ ] Error handling covers all edge cases (malformed JSON, missing files, API errors)
- [ ] Mock-based tests allow testing without API calls
- [ ] Integration tests verify real API works (when API key available)

### User Experience Success
- [ ] Clear error messages guide user to solutions
- [ ] Verbose mode shows progress
- [ ] Documentation explains setup and usage
- [ ] Troubleshooting section covers common issues
- [ ] Cost estimates help users understand usage

---

## Future Enhancements (v0.0.2+)

### Ollama Provider
- Add Ollama provider after OpenRouter stable
- Configuration via `OLLAMA_HOST`, `OLLAMA_MODEL`
- Local analysis keeps code private

### Advanced Features
- Interactive mode: Accept/reject detections
- Caching: Avoid re-analyzing unchanged code
- Comprehensive infra cost estimation
- Dependency graph visualization
- Support for more config file types
- CI/CD integration (automatic detection on PR)

---

## Notes for Reviewers

### Testing Infrastructure
The testing infrastructure is built incrementally across PR #1. Each PR adds tests for the components it introduces. Golden files provide regression testing for config generation.

### Mock Strategy
We use `testify/mock` for LLM provider mocking. This allows us to test the entire flow without API calls, making tests fast and reliable.

### Documentation
User-facing documentation is added in PR #4. Internal documentation (TESTING.md) is added in PR #1 to help contributors understand the testing approach.

### Backwards Compatibility
This feature is purely additive. No existing commands are modified. The new `detect` command is self-contained.

---

## Appendix: Golden File Management

### When to Update Golden Files
Golden files should be updated when:
1. Detection logic changes intentionally
2. Prompt templates are modified
3. Config generation logic changes

### How to Update
```bash
# Run with update flag (we'll implement this later)
profitctl detect --update-golden ./test/fixtures/samples/go-webapp
```

### Review Process
Golden file changes should be reviewed carefully in PRs to ensure the changes are expected and correct.