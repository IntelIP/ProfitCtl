# Work Plan: profitctl MVP

## Context

### Original Request

Build an open-core CLI tool called `profitctl` that simulates software unit economics and enforces profitability constraints through explicit YAML configuration. The tool should help developers understand fixed COGS, variable COGS, cost with scale, and fail CI builds when profit covenants are breached.

**Core Requirements**:
- CLI-first developer tool
- Explicit YAML configuration for costs, pricing, and covenants
- Scale simulations (1 → 10k+ users) with exponential growth
- p95/p99 stress testing
- Hybrid covenant rules (expression + threshold-based)
- Multiple output formats (CLI, JSON, markdown)
- CI-friendly exit codes
- Local-first execution with no external API dependencies

**User Background**: Strong Python experience, some TypeScript web apps with AI support. Willing to learn Go but needs guidance on syntax and structure.

### Interview Summary

**Key Discussions**:
- **Language Choice**: Go (chosen for single binary distribution, battle-tested CLI ecosystem, learning opportunity)
- **Cost Modeling**: All cost types supported (compute, storage, API calls, LLM tokens, bandwidth, database queries)
- **Cost Layers**: Separation of infrastructure vs application costs, higher vs lower on stack
- **Scale Simulation**: Exponential growth patterns
- **Covenant Rules**: Hybrid approach combining expression-based and threshold-based rules
- **Outputs**: All formats - exit codes, JSON reports, markdown summaries for PR comments
- **Validation**: Separate `validate` command for YAML syntax + covenant rule checking
- **MVP Minimum**: Users must understand fixed COGS, variable COGS, and cost with scale

**Research Findings**:
- **CLI Framework**: Cobra (battle-tested, used by kubectl, infracost, Docker CLI)
- **YAML Libraries**: gopkg.in/yaml.v3 for parsing, go-playground/validator for struct validation
- **Cost Modeling**: Fixed/Variable cost structures with normal/exponential distributions
- **p95/p99**: Monte Carlo simulation with configurable iterations (recommend 10,000)
- **Pure Functions**: All calculations as pure functions with immutable inputs
- **Go Learning**: Python/TS developer needs guidance on Go syntax, error handling, goroutines

### Metis Review

**Identified Gaps (Resolved)**:
- Cost types scope: Expanded to support all cost types with extensible abstraction
- Covenant syntax: Decided on hybrid approach (expression + threshold)
- Output formats: All three confirmed (exit codes, JSON, markdown)
- Validation command: Confirmed separate command for syntax + covenant checking
- Cost layers: Added infrastructure vs application separation

**Guardrails Confirmed**:
- No external API calls in MVP (local calculations only)
- No persistent state (stateless simulations)
- No database dependencies
- Deterministic output (same input = same output)
- No AI/ML inference for calculations

---

## Work Objectives

### Core Objective

Build a deterministic, developer-centric CLI tool that simulates software unit economics, provides visibility into fixed/variable COGS across scale, and enforces profitability constraints in CI pipelines.

### Concrete Deliverables

1. **CLI Binary**: `profitctl` executable with init, simulate, validate commands
2. **YAML Configuration**: `profit.schema.yaml` and `examples/profit.yml`
3. **Cost Engine**: Pure Go functions for fixed/variable cost calculations with layer separation
4. **Simulation Engine**: Exponential scale + p95/p99 Monte Carlo simulation
5. **Covenant Validator**: Hybrid rule engine with expression + threshold support
6. **Output Formatters**: CLI tables, JSON reports, markdown summaries
7. **CI Integration**: Exit codes (0=pass, 1=fail, 2=error)
8. **Documentation**: Go learning guide embedded in work plan

### Definition of Done

- [ ] `profitctl init` creates valid profit.yml from repo scan
- [ ] `profitctl simulate` runs end-to-end with exponential scale
- [ ] `profitctl validate` checks YAML syntax + covenant rules
- [ ] Outputs show fixed COGS, variable COGS, cost with scale
- [ ] Outputs margin + p95 results
- [ ] Covenant breach fails build (exit code 1)
- [ ] JSON report generated (--json flag)
- [ ] Markdown summary generated (--markdown flag)
- [ ] Config is explicit, editable, and validatable
- [ ] No external dependencies required
- [ ] Single binary distribution

### Must Have

- Go CLI with Cobra framework
- YAML parsing and validation
- Fixed cost modeling (monthly/yearly/daily)
- Variable cost modeling with distributions (normal, exponential, uniform)
- Exponential scale simulation
- p95/p99 stress testing
- Threshold-based covenants (min_margin, max_cost_per_user)
- Expression-based covenants (p95_margin > 20%)
- Human-readable CLI output with tables
- JSON output for CI artifacts
- Exit codes for CI integration
- Example profit.yml configuration
- Comprehensive Go learning resources

### Must NOT Have (Guardrails)

- No external API calls (no cloud provider integrations)
- No persistent database or state
- No AI/ML inference
- No dashboards or web UI
- No auto-pricing changes
- No competitor scraping
- No live billing integrations
- No runtime telemetry (open core feature)

---

## Verification Strategy

### Test Infrastructure Decision

**User wants tests**: YES (TDD approach recommended for pure functions)
**Infrastructure**: Will set up Go test framework (testing package + testify)
**QA approach**: TDD - RED-GREEN-REFACTOR for all calculation functions

### Test Setup (Task 0)

- [ ] 0. Setup Test Infrastructure
  - Install: `go mod init github.com/profitctl/profitctl`
  - Add dependencies: `go get github.com/spf13/cobra`, `gopkg.in/yaml.v3`, `github.com/stretchr/testify`
  - Create: `internal/cost/cost_test.go` with example test
  - Verify: `go test ./...` → shows test output

### TDD Workflow for Each Task

Each TODO follows RED-GREEN-REFACTOR:

1. **RED**: Write failing test first
   - Test file: `[module]_test.go`
   - Test command: `go test ./[module]/... -v`
   - Expected: FAIL (test exists, implementation doesn't)

2. **GREEN**: Implement minimum code to pass
   - Command: `go test ./[module]/... -v`
   - Expected: PASS

3. **REFACTOR**: Clean up while keeping green
   - Command: `go test ./... -v`
   - Expected: PASS (all tests still pass)

### Manual Execution Verification (Always Include)

Even with tests, manual verification for CLI integration:

**For CLI commands:**
- [ ] Using interactive_bash (tmux session):
  - Command: `go build -o profitctl . && ./profitctl [command]`
  - Expected output: Command-specific output displayed
  - Exit code: 0 for success, 1 for errors

**For output formats:**
- [ ] CLI output: `./profitctl simulate` → formatted table output
- [ ] JSON output: `./profitctl simulate --json` → valid JSON
- [ ] Markdown output: `./profitctl simulate --markdown` → markdown table
- [ ] Validate command: `./profitctl validate` → "Config is valid" or errors

**Evidence Required:**
- [ ] Command output captured (copy-paste actual terminal output)
- [ ] JSON output validated (parse without errors)
- [ ] Exit codes verified (0, 1, 2)

---

## Task Flow

```
Task 0: Setup Test Infrastructure
    ↓
Task 1: Go Environment + Cobra Scaffold
    ↓
Task 2: YAML Schema + Validation
    ↓
Task 3: Core Types + Cost Structures
    ↓
Task 4: Fixed Cost Engine
    ↓
Task 5: Variable Cost Engine
    ↓
Task 6: Cost Layer Abstraction
    ↓
Task 7: Scale Simulation Engine
    ↓
Task 8: p95/p99 Stress Testing
    ↓
Task 9: Pricing Calculator
    ↓
Task 10: Margin Calculator
    ↓
Task 11: Covenant Engine (Hybrid)
    ↓
Task 12: CLI Output Formatter
    ↓
Task 13: JSON Output Formatter
    ↓
Task 14: Markdown Output Formatter
    ↓
Task 15: Init Command (Repo Scanner)
    ↓
Task 16: Simulate Command
    ↓
Task 17: Validate Command
    ↓
Task 18: CI Integration (Exit Codes)
    ↓
Task 19: Example profit.yml
    ↓
Task 20: Go Learning Guide
```

## Parallelization

| Group | Tasks | Reason |
|-------|-------|--------|
| A | 4, 5 | Independent cost calculations |
| B | 12, 13, 14 | Output formatters can be developed independently |
| C | 9, 10 | Pricing and margin can be developed separately |

---

## TODOs

- [ ] 0. Setup Test Infrastructure

  **What to do**:
  - Initialize Go module
  - Add Cobra, YAML, and testing dependencies
  - Create basic test structure
  - Configure go.mod for module

  **Must NOT do**:
  - No production code yet (pure test setup)

  **Parallelizable**: NO (foundation for everything)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/spf13/cobra` - CLI framework documentation
  - `gopkg.in/yaml.v3` - YAML parsing examples
  - `github.com/stretchr/testify` - Testing patterns

  **Documentation References**:
  - `tour.golang.org` - Official Go tutorial
  - `golang.org/doc/install` - Go installation guide

  **Acceptance Criteria**:
  - [ ] Test file created: `internal/cost/cost_test.go`
  - [ ] `go mod init github.com/profitctl/profitctl`
  - [ ] Dependencies installed via `go get`
  - [ ] `go test ./...` → shows test framework working
  - [ ] `go build -o profitctl .` → creates binary

  **Manual Execution Verification**:
  - [ ] `go version` → shows Go version
  - [ ] `go mod tidy` → creates/go.mod
  - [ ] `go build -o profitctl .` → profitctl.exe created

  **Commit**: YES
  - Message: `chore: initialize Go module and test infrastructure`
  - Files: `go.mod`, `go.sum`, `internal/cost/cost_test.go`

---

- [ ] 1. Go Environment + Cobra Scaffold

  **What to do**:
  - Create project structure (cmd/, internal/, pkg/, schema/)
  - Implement root command with Cobra
  - Create stub commands (init, simulate, validate)
  - Set up flag handling
  - Add help and version commands

  **Must NOT do**:
  - No business logic implementation yet

  **Parallelizable**: NO (foundation for CLI)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/spf13/cobra` - Command structure examples
  - `github.com/spf13/viper` - Configuration management
  - `github.com/infracost/infracost` - Similar CLI structure (open source reference)

  **Go Learning Context**:
  - Go packages: Each directory is a package, `init()` runs on import
  - Cobra commands: `NewCobraCommand()` pattern, AddCommand() composition
  - Error handling: Explicit `if err != nil` checks required

  **Acceptance Criteria**:
  - [ ] Root command created: `cmd/root.go`
  - [ ] Init command: `cmd/init.go`
  - [ ] Simulate command: `cmd/simulate.go`
  - [ ] Validate command: `cmd/validate.go`
  - [ ] `go build -o profitctl . && ./profitctl --help` → shows all commands
  - [ ] `./profitctl init --help` → shows init help
  - [ ] `./profitctl simulate --help` → shows simulate help
  - [ ] `./profitctl validate --help` → shows validate help

  **Manual Execution Verification**:
  - [ ] Command: `go build -o profitctl . && ./profitctl`
  - Output: CLI help with all 3 commands displayed
  - Exit code: 0

  **Commit**: YES
  - Message: `feat: add Cobra CLI scaffold with init/simulate/validate commands`
  - Files: `cmd/root.go`, `cmd/init.go`, `cmd/simulate.go`, `cmd/validate.go`, `main.go`

---

- [ ] 2. YAML Schema + Validation

  **What to do**:
  - Create `schema/profit.schema.yaml` (JSON Schema format)
  - Define struct tags for YAML binding
  - Implement YAML parsing with gopkg.in/yaml.v3
  - Add struct validation with go-playground/validator
  - Create clear error messages for validation failures

  **Must NOT do**:
  - No simulation logic yet (just parsing)

  **Parallelizable**: NO (foundation for all data flow)

  **References**:

  **Pattern References** (existing code to follow):
  - `gopkg.in/yaml.v3` - Unmarshal with struct tags
  - `github.com/go-playground/validator` - Validation tags like `validate:"required,min=1"`
  - `pkg/igorp24/dataloader` - Config loading patterns

  **Schema Structure**:
  ```yaml
  fixed_costs:
    - name: string
      amount: number
      period: monthly | yearly | daily
      layer: infrastructure | application | service
  variable_costs:
    - name: string
      cost_per_unit: number
      units_per_user: number | distribution
      distribution: normal | exponential | uniform
      layer: infrastructure | application | service
  pricing:
    plans:
      - name: string
        price: number
        limits:
          users: number
  covenants:
    - type: threshold | expression
      field: string  # for threshold
      operator: gt | lt | gte | lte | eq
      value: number
      expression: string  # for expression
      message: string
  ```

  **Go Learning Context**:
  - Struct tags: `yaml:"name,omitempty" validate:"required"`
  - Validation: `validate:"required,min=0"` for numbers
  - Error handling: Collect all errors, return as single message

  **Acceptance Criteria**:
  - [ ] Schema file: `schema/profit.schema.yaml`
  - [ ] Config types: `pkg/types/config.go`
  - [ ] Parser: `internal/config/parser.go`
  - [ ] Validator: `internal/config/validator.go`
  - [ ] Valid config parses without error
  - [ ] Invalid config returns clear error messages
  - [ ] Test: `go test ./internal/config/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Create test config: `examples/profit.yml`
  - [ ] Run: `./profitctl validate -f examples/profit.yml`
  - [ ] Output: "Config is valid" or list of errors

  **Commit**: YES
  - Message: `feat: add YAML schema and validation`
  - Files: `schema/profit.schema.yaml`, `pkg/types/config.go`, `internal/config/parser.go`, `internal/config/validator.go`

---

- [ ] 3. Core Types + Cost Structures

  **What to do**:
  - Define cost layer types (infrastructure, application, service)
  - Define cost period types (monthly, yearly, daily)
  - Define distribution types (normal, exponential, uniform)
  - Create pure types for all cost calculations
  - Ensure immutability (value receivers where possible)

  **Must NOT do**:
  - No calculation logic yet (just types)

  **Parallelizable**: NO (foundation for cost engine)

  **References**:

  **Pattern References** (existing code to follow):
  - `golang.org/x/exp/slices` - Immutable operations
  - `github.com/google/go-cmp` - Value comparison for testing

  **Type Definitions**:
  ```go
  type CostLayer string
  const (
      LayerInfrastructure CostLayer = "infrastructure"
      LayerApplication    CostLayer = "application"
      LayerService        CostLayer = "service"
  )

  type CostPeriod string
  const (
      PeriodMonthly CostPeriod = "monthly"
      PeriodYearly  CostPeriod = "yearly"
      PeriodDaily   CostPeriod = "daily"
  )

  type DistributionType string
  const (
      DistNormal      DistributionType = "normal"
      DistExponential DistributionType = "exponential"
      DistUniform     DistributionType = "uniform"
  )

  type FixedCost struct {
      Name   string     `yaml:"name" validate:"required"`
      Amount float64    `yaml:"amount" validate:"required,min=0"`
      Period CostPeriod `yaml:"period" validate:"required,oneof=monthly yearly daily"`
      Layer  CostLayer  `yaml:"layer" validate:"required,oneof=infrastructure application service"`
  }

  type VariableCost struct {
      Name             string          `yaml:"name" validate:"required"`
      CostPerUnit      float64         `yaml:"cost_per_unit" validate:"required,min=0"`
      UnitsPerUser     float64         `yaml:"units_per_user" validate:"required,min=0"`
      DistributionType DistributionType `yaml:"distribution" validate:"required,oneof=normal exponential uniform"`
      DistributionParams map[string]float64 // mean, stddev, min, max
      Layer            CostLayer       `yaml:"layer" validate:"required,oneof=infrastructure application service"`
  }

  type CostLayerBreakdown struct {
      Infrastructure float64 `json:"infrastructure"`
      Application    float64 `json:"application"`
      Service        float64 `json:"service"`
  }
  ```

  **Go Learning Context**:
  - Structs: Value types, passed by value unless using pointer
  - Constants: `const` blocks for related constants
  - Maps: Must initialize before use (`make(map[string]float64)`)
  - JSON/YAML tags: Match field names, `omitempty` for optional

  **Acceptance Criteria**:
  - [ ] Type definitions: `pkg/types/cost.go`
  - [ ] All cost types defined
  - [ ] All distribution types defined
  - [ ] CostLayerBreakdown for layer separation
  - [ ] Test: `go test ./pkg/types/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Command: `go build -o profitctl .`
  - Output: No compile errors for types
  - Exit code: 0

  **Commit**: YES
  - Message: `feat: add core types for costs, layers, and distributions`
  - Files: `pkg/types/cost.go`, `pkg/types/distribution.go`

---

- [ ] 4. Fixed Cost Engine

  **What to do**:
  - Implement `CalculateFixedCosts(fixedCosts []FixedCost, months int)` pure function
  - Handle period conversion (monthly/yearly/daily)
  - Aggregate by layer (infrastructure, application, service)
  - Return total fixed costs and layer breakdown
  - Write comprehensive tests

  **Must NOT do**:
  - No variable cost calculations

  **Parallelizable**: YES (independent of variable costs)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/shopspring/decimal` - Precise financial calculations (recommend using)
  - `pkg/decimal` - Wrapper for float64 to avoid precision issues

  **Formula**:
  ```
  monthly_cost = yearly_cost / 12
  daily_cost = monthly_cost / 30
  ```

  **Go Learning Context**:
  - Floating point: Be aware of precision issues with money
  - Methods: `func (c FixedCost) MonthlyAmount() float64`
  - Testing: Table-driven tests for various inputs

  **Acceptance Criteria**:
  - [ ] Fixed cost engine: `internal/cost/fixed.go`
  - [ ] Tests: `internal/cost/fixed_test.go`
  - [ ] $1000/yearly → $83.33/monthly
  - [ ] $1000/yearly for 6 months → $5000
  - [ ] Layer aggregation: costs separated by layer
  - [ ] Test: `go test ./internal/cost/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Create test config with fixed costs
  - [ ] Run: `./profitctl simulate -f test.yml`
  - [ ] Verify fixed costs appear in output
  - [ ] Verify layer breakdown matches config

  **Commit**: YES
  - Message: `feat: implement fixed cost calculation engine`
  - Files: `internal/cost/fixed.go`, `internal/cost/fixed_test.go`

---

- [ ] 5. Variable Cost Engine

  **What to do**:
  - Implement `CalculateVariableCosts(variableCosts []VariableCost, users int)` pure function
  - Implement distribution generators (normal, exponential, uniform)
  - Handle per-user calculations with variability
  - Aggregate by layer
  - Write comprehensive tests including distribution verification

  **Must NOT do**:
  - No scale simulation yet

  **Parallelizable**: YES (independent of fixed costs)

  **References**:

  **Pattern References** (existing code to follow):
  - `golang.org/x/exp/rand` - Random number generation for distributions
  - `github.com/gonum/stat` - Statistical distributions (alternative)

  **Formulas**:
  ```
  normal: mean + stddev * rand.Normal()
  exponential: -ln(rand) / rate
  uniform: min + (max - min) * rand.Float64()
  
  variable_cost = cost_per_unit * units_per_user * users
  ```

  **Go Learning Context**:
  - Packages: `math/rand` for random, `math` for math functions
  - Seed: Set seed once for reproducibility (`rand.Seed(42)`)
  - Closures: Useful for generating random values

  **Acceptance Criteria**:
  - [ ] Variable cost engine: `internal/cost/variable.go`
  - [ ] Distribution implementations: `internal/cost/distribution.go`
  - [ ] Tests: `internal/cost/variable_test.go`
  - [ ] Normal distribution: mean ± 2 stddev covers ~95%
  - [ ] Exponential distribution: decreasing frequency
  - [ ] Uniform distribution: values between min and max
  - [ ] Layer aggregation: costs separated by layer
  - [ ] Test: `go test ./internal/cost/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Create test config with variable costs
  - [ ] Run simulation with 100 users
  - [ ] Verify variable costs scale with user count
  - [ ] Verify layer breakdown appears in output

  **Commit**: YES
  - Message: `feat: implement variable cost calculation with distributions`
  - Files: `internal/cost/variable.go`, `internal/cost/distribution.go`, `internal/cost/variable_test.go`

---

- [ ] 6. Cost Layer Abstraction

  **What to do**:
  - Create `CostEngine` interface with `Calculate(inputs) Result`
  - Implement `NewCostEngine(config Config) CostEngine`
  - Combine fixed + variable costs into unified calculation
  - Return layer breakdown for each cost type
  - Enable extensibility for future cost sources

  **Must NOT do**:
  - No simulation logic yet

  **Parallelizable**: NO (depends on 4 and 5)

  **References**:

  **Pattern References** (existing code to follow):
  - `io.Reader` / `io.Writer` - Standard Go interface patterns
  - `github.com/infracost/infracost` - Cost source abstraction

  **Interface Design**:
  ```go
  type CostEngine interface {
      CalculateFixedCosts(users int, months int) FixedCostResult
      CalculateVariableCosts(users int) VariableCostResult
      CalculateTotalCosts(users int, months int) TotalCostResult
  }

  type FixedCostResult struct {
      Total        float64
      Monthly      float64
      ByLayer      CostLayerBreakdown
      ByCost       []FixedCostLineItem
  }

  type VariableCostResult struct {
      Total      float64
      ByLayer    CostLayerBreakdown
      ByCost     []VariableCostLineItem
      Distribution map[string][]float64 // For reporting
  }

  type TotalCostResult struct {
      FixedCosts    FixedCostResult
      VariableCosts VariableCostResult
      GrandTotal    float64
      GrandByLayer  CostLayerBreakdown
  }
  ```

  **Go Learning Context**:
  - Interfaces: Implicitly satisfied (no `implements` keyword)
  - Composition: Interfaces can embed other interfaces
  - Pointers: Use `*CostEngine` if engine needs configuration

  **Acceptance Criteria**:
  - [ ] Cost engine interface: `internal/cost/engine.go`
  - [ ] Engine implementation: `internal/cost/engine_impl.go`
  - [ ] Total cost calculation: fixed + variable combined
  - [ ] Layer breakdown for total costs
  - [ ] Extensible design for future cost sources
  - [ ] Test: `go test ./internal/cost/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl simulate -f examples/profit.yml`
  - [ ] Verify fixed and variable costs both appear
  - [ ] Verify layer breakdown shows infrastructure, application, service

  **Commit**: YES
  - Message: `feat: implement unified cost engine with layer abstraction`
  - Files: `internal/cost/engine.go`, `internal/cost/engine_impl.go`

---

- [ ] 7. Scale Simulation Engine

  **What to do**:
  - Implement `RunScaleSimulation(engine CostEngine, config Config)` pure function
  - Model exponential growth: `users = baseUsers * (growthFactor ^ step)`
  - Generate simulation points (e.g., 100, 500, 1000, 5000, 10000 users)
  - Calculate costs at each scale point
  - Return timeline of costs with growth trajectory

  **Must NOT do**:
  - No p95/p99 yet (separate task)

  **Parallelizable**: NO (depends on cost engine)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/gonum/optimize` - Growth curve fitting (alternative)

  **Exponential Growth Model**:
  ```
  base_users = 100 (configurable)
  growth_factor = 1.5 (configurable, 50% growth per step)
  steps = [0, 1, 2, 3, 4, 5]
  
  users_at_step(n) = base_users * (growth_factor ^ n)
  ```

  **Go Learning Context**:
  - Loops: `for i := 0; i < n; i++` (no while, use for)
  - Slices: `make([]Result, 0, n)` for pre-allocated slices
  - Append: `results = append(results, result)` for dynamic growth

  **Acceptance Criteria**:
  - [ ] Scale simulation engine: `internal/simulation/scale.go`
  - [ ] Growth trajectory calculation
  - [ ] Multiple scale points (100, 500, 1000, 5000, 10000)
  - [ ] Configurable base users and growth factor
  - [ ] Cost projection at each scale point
  - [ ] Test: `go test ./internal/simulation/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl simulate -f examples/profit.yml`
  - [ ] Verify multiple scale points in output
  - [ ] Verify exponential growth pattern in user counts
  - [ ] Verify costs scale appropriately

  **Commit**: YES
  - Message: `feat: implement exponential scale simulation engine`
  - Files: `internal/simulation/scale.go`, `internal/simulation/scale_test.go`

---

- [ ] 8. p95/p99 Stress Testing

  **What to do**:
  - Implement `RunMonteCarlo(engine CostEngine, config Config, iterations int)` pure function
  - Generate random usage scenarios based on variable cost distributions
  - Calculate costs for each iteration (10,000 iterations recommended)
  - Sort results to find p95 and p99 percentiles
  - Return worst-case scenarios for stress testing

  **Must NOT do**:
  - No covenant validation yet

  **Parallelizable**: NO (depends on variable cost distributions)

  **References**:

  **Pattern References** (existing code to follow):
  - `golang.org/x/exp/rand` - Random number generation
  - `golang.org/x/sync/errgroup` - Parallel iterations (optional optimization)

  **Monte Carlo Algorithm**:
  ```
  iterations = 10000
  results = []
  
  for i in 0..iterations:
      scenario = GenerateRandomScenario(variable_costs)
      cost = CalculateScenarioCost(scenario)
      results = append(results, cost)
  
  sort(results)
  p95 = results[int(0.95 * len(results))]
  p99 = results[int(0.99 * len(results))]
  ```

  **Go Learning Context**:
  - Sorting: `sort.Float64s(results)` for float64 slices
  - Indexing: `results[idx]` after sorting
  - Performance: 10k iterations is fast in Go, no optimization needed yet

  **Acceptance Criteria**:
  - [ ] Monte Carlo engine: `internal/simulation/montecarlo.go`
  - [ ] Configurable iteration count
  - [ ] Random scenario generation based on distributions
  - [ ] Percentile calculation (p95, p99)
  - [ ] Worst-case user cost calculation
  - [ ] Test: `go test ./internal/simulation/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl simulate -f examples/profit.yml`
  - [ ] Verify p95 and p99 appear in output
  - [ ] Verify p95 < p99 (p99 should be higher/worse)
  - [ ] Verify results are deterministic (same seed)

  **Commit**: YES
  - Message: `feat: implement p95/p99 Monte Carlo stress testing`
  - Files: `internal/simulation/montecarlo.go`, `internal/simulation/montecarlo_test.go`

---

- [ ] 9. Pricing Calculator

  **What to do**:
  - Implement `CalculateRevenue(pricing PricingConfig, users int)` pure function
  - Handle multiple pricing plans with user limits
  - Calculate revenue based on plan tiers
  - Handle overage scenarios if applicable
  - Return revenue by plan and total

  **Must NOT do**:
  - No margin calculation yet

  **Parallelizable**: YES (independent of cost engine)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/stripe/stripe-go` - Pricing model patterns

  **Pricing Model**:
  ```
  plans:
    - name: basic
      price: 10
      limits:
        users: 1000
    - name: pro
      price: 29
      limits:
        users: 5000
    - name: enterprise
      price: 99
      limits:
        users: 10000

  If users = 2500:
  - First 1000 users at $10 = $10,000
  - Next 1500 users at $29 = $43,500
  - Total revenue = $53,500
  ```

  **Go Learning Context**:
  - Slices: `for i, plan := range plans` for iteration
  - Conditionals: `if users > limit { ... }`
  - Float64 formatting: `fmt.Sprintf("%.2f", value)` for currency

  **Acceptance Criteria**:
  - [ ] Pricing calculator: `internal/pricing/calculator.go`
  - [ ] Multiple plan support
  - [ ] User limit handling
  - [ ] Revenue aggregation
  - [ ] Test: `go test ./internal/pricing/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Create config with pricing plans
  - [ ] Run: `./profitctl simulate -f examples/profit.yml`
  - [ ] Verify pricing appears in output
  - [ ] Verify revenue scales with users

  **Commit**: YES
  - Message: `feat: implement multi-tier pricing calculator`
  - Files: `internal/pricing/calculator.go`, `internal/pricing/calculator_test.go`

---

- [ ] 10. Margin Calculator

  **What to do**:
  - Implement `CalculateMargins(revenue float64, costs CostLayerBreakdown)` pure function
  - Calculate gross margin percentage
  - Calculate margin by layer (infrastructure margin, application margin, etc.)
  - Calculate cost per user metrics
  - Format margin results for output

  **Must NOT do**:
  - No covenant validation yet

  **Parallelizable**: YES (independent of pricing)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/decimalgo/decimal` - Financial precision

  **Margin Formulas**:
  ```
  gross_margin = (revenue - total_costs) / revenue * 100
  margin_percentage = gross_margin / 100
  
  cost_per_user = total_costs / users
  infrastructure_margin = infrastructure_costs / total_costs * 100
  ```

  **Go Learning Context**:
  - Division: `if revenue == 0 { return 0 }` to avoid divide by zero
  - Float formatting: `fmt.Sprintf("%.1f%%", margin)` for percentage
  - Returning multiple values: `func CalculateMargins(...) (float64, float64, error)`

  **Acceptance Criteria**:
  - [ ] Margin calculator: `internal/pricing/margin.go`
  - [ ] Gross margin calculation
  - [ ] Layer margin breakdown
  - [ ] Cost per user metrics
  - [ ] Test: `go test ./internal/pricing/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl simulate -f examples/profit.yml`
  - [ ] Verify margin percentage appears in output
  - [ ] Verify cost per user appears in output
  - [ ] Verify layer margins match expected values

  **Commit**: YES
  - Message: `feat: implement margin and cost-per-user calculations`
  - Files: `internal/pricing/margin.go`, `internal/pricing/margin_test.go`

---

- [ ] 11. Covenant Engine (Hybrid)

  **What to do**:
  - Implement `ValidateCovenants(covenants []Covenant, results SimulationResult)` pure function
  - Implement threshold-based rules (min_margin, max_cost_increase, etc.)
  - Implement expression-based rules (p95_margin > 20%, etc.)
  - Return validation result with pass/fail and messages
  - Handle both numeric and boolean covenant types

  **Must NOT do**:
  - No CLI integration yet

  **Parallelizable**: NO (depends on simulation results)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/expr-lang/expr` - Safe expression evaluation (optional)
  - Custom simple expression parser for MVP (recommended)

  **Threshold Rules**:
  ```go
  type ThresholdCovenant struct {
      Field    string  // "margin", "cost_per_user", "p95_margin"
      Operator string  // "gt", "lt", "gte", "lte", "eq"
      Value    float64
      Message  string
  }

  // Examples:
  // { field: "margin", operator: "gte", value: 20, message: "Gross margin must be >= 20%" }
  // { field: "cost_per_user", operator: "lte", value: 5, message: "Cost per user must be <= $5" }
  ```

  **Expression Rules**:
  ```go
  type ExpressionCovenant struct {
      Expression string // "p95_margin > 20"
      Message    string
  }

  // Parser evaluates simple expressions:
  // - Operators: >, <, >=, <=, ==, !=
  // - Fields: margin, p95_margin, cost_per_user, p95_cost_per_user
  // - Logical: &&, || (optional for v2)
  ```

  **Go Learning Context**:
  - String parsing: `strings.Split()` for basic tokenization
  - Reflection: `reflect.ValueOf()` for field access (use sparingly)
  - Error handling: Collect all failures, return together

  **Acceptance Criteria**:
  - [ ] Covenant engine: `internal/covenant/engine.go`
  - [ ] Threshold rule validation
  - [ ] Expression rule validation (simple parser)
  - [ ] Validation result with pass/fail and messages
  - [ ] Test: `go test ./internal/covenant/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Create config with covenant that should pass
  - [ ] Run: `./profitctl simulate -f examples/profit.yml`
  - [ ] Verify: "All covenants passed" in output
  - [ ] Create config with covenant that should fail
  - [ ] Verify: Covenant breach message appears
  - [ ] Verify: Exit code 1

  **Commit**: YES
  - Message: `feat: implement hybrid covenant validation engine`
  - Files: `internal/covenant/engine.go`, `internal/covenant/engine_test.go`

---

- [ ] 12. CLI Output Formatter

  **What to do**:
  - Implement `FormatCLIResult(result SimulationResult)` function
  - Create formatted tables using tablewriter or similar
  - Display fixed COGS, variable COGS, cost with scale
  - Display margin, p95, p99 results
  - Display covenant validation results
  - Use colors for important information (pass/fail)

  **Must NOT do**:
  - No JSON or markdown yet

  **Parallelizable**: NO (depends on simulation results)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/olekukonko/tablewriter` - Table formatting
  - `github.com/fatih/color` - Terminal colors

  **Output Format**:
  ```
  Scenario: 1,000 users
  ────────────────────
  Mean margin: +64%
  p95 margin: +18%
  Worst-case user: -$2.30/day

  Fixed COGS: $1,500/month
    Infrastructure: $500
    Application: $700
    Service: $300

  Variable COGS: $0.01/user
  Cost per user: $1.50

  Covenant Status: PASSED ✓
  ```

  **Go Learning Context**:
  - Formatting: `fmt.Printf()` with format verbs
  - Tables: `tablewriter.NewWriter(os.Stdout)`
  - Colors: `color.New(color.FgGreen).Println()`

  **Acceptance Criteria**:
  - [ ] CLI formatter: `internal/output/cli.go`
  - [ ] Formatted tables for results
  - [ ] Color-coded status (green=pass, red=fail)
  - [ ] Clear section separation
  - [ ] Test: `go test ./internal/output/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl simulate -f examples/profit.yml`
  - [ ] Verify: Formatted table output
  - [ ] Verify: Colors appear (if terminal supports)
  - [ ] Verify: All sections present

  **Commit**: YES
  - Message: `feat: implement human-readable CLI output formatting`
  - Files: `internal/output/cli.go`, `internal/output/cli_test.go`

---

- [ ] 13. JSON Output Formatter

  **What to do**:
  - Implement `FormatJSONResult(result SimulationResult)` function
  - Create structured JSON output for CI artifacts
  - Include all calculation results (costs, margins, p95/p99)
  - Include covenant validation results
  - Use proper JSON formatting (indent, field names)

  **Must NOT do**:
  - No CLI or markdown yet

  **Parallelizable**: YES (independent of CLI formatter)

  **References**:

  **Pattern References** (existing code to follow):
  - `encoding/json` - Standard Go JSON library
  - `json.MarshalIndent()` for formatted output

  **JSON Structure**:
  ```json
  {
    "scenario": {
      "users": 1000,
      "growth_factor": 1.5
    },
    "costs": {
      "fixed": {
        "total": 1500,
        "monthly": 1500,
        "by_layer": {
          "infrastructure": 500,
          "application": 700,
          "service": 300
        }
      },
      "variable": {
        "total": 500,
        "per_user": 0.5,
        "by_layer": {...}
      },
      "total": 2000
    },
    "revenue": {
      "total": 10000,
      "by_plan": [...]
    },
    "margin": {
      "gross": 80,
      "by_layer": {...},
      "cost_per_user": 2
    },
    "stress_test": {
      "p95": {
        "margin": 64,
        "cost_per_user": 3.50
      },
      "p99": {
        "margin": 58,
        "cost_per_user": 4.20
      }
    },
    "covenants": {
      "passed": true,
      "violations": []
    }
  }
  ```

  **Go Learning Context**:
  - JSON tags: `json:"field_name"` on struct fields
  - Omitempty: `json:"field_name,omitempty"` to skip zero values
  - Marshal: `json.MarshalIndent(result, "", "  ")`

  **Acceptance Criteria**:
  - [ ] JSON formatter: `internal/output/json.go`
  - [ ] Valid JSON output (parseable by jq)
  - [ ] All fields included
  - [ ] Proper indentation
  - [ ] Test: `go test ./internal/output/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl simulate -f examples/profit.yml --json`
  - [ ] Verify: Valid JSON output
  - [ ] Verify: `jq . < output.json` works
  - [ ] Verify: All expected fields present

  **Commit**: YES
  - Message: `feat: implement JSON output for CI artifacts`
  - Files: `internal/output/json.go`, `internal/output/json_test.go`

---

- [ ] 14. Markdown Output Formatter

  **What to do**:
  - Implement `FormatMarkdownResult(result SimulationResult)` function
  - Create markdown table for PR comments
  - Include summary statistics (margin, p95, covenants)
  - Include cost breakdown by layer
  - Use markdown formatting (tables, headers, emojis)

  **Must NOT do**:
  - No full report (just summary for PR comments)

  **Parallelizable**: YES (independent of other formatters)

  **References**:

  **Pattern References** (existing code to follow):
  - GitHub Actions PR comment format
  - Markdown table syntax (`| header |`, `|---|`, `| value |`)

  **Markdown Output**:
  ```markdown
  ## profitctl Results

  | Metric | Value |
  |--------|-------|
  | Users | 1,000 |
  | Margin | 64% |
  | p95 Margin | 18% |
  | Cost per User | $1.50 |
  | Covenant Status | ✅ PASSED |

  ### Cost Breakdown

  | Layer | Fixed | Variable |
  |-------|-------|----------|
  | Infrastructure | $500 | $200 |
  | Application | $700 | $150 |
  | Service | $300 | $100 |
  ```

  **Go Learning Context**:
  - Strings: `fmt.Sprintf()` for building strings
  - Tables: String concatenation with `|`, `-`, `:`
  - Formatting: Consistent spacing and alignment

  **Acceptance Criteria**:
  - [ ] Markdown formatter: `internal/output/markdown.go`
  - [ ] Valid markdown table format
  - [ ] All summary metrics included
  - [ ] Cost breakdown by layer
  - [ ] Test: `go test ./internal/output/... -v` → PASS

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl simulate -f examples/profit.yml --markdown`
  - Verify: Valid markdown output
  - Verify: Tables render correctly in GitHub

  **Commit**: YES
  - Message: `feat: implement markdown output for PR comments`
  - Files: `internal/output/markdown.go`, `internal/output/markdown_test.go`

---

- [ ] 15. Init Command (Repo Scanner)

  **What to do**:
  - Implement `profitctl init` command
  - Create repo scanner to detect technology stack
  - Generate initial `profit.yml` from detected stack
  - Include sensible defaults for common stacks
  - Allow user to edit generated config

  **Must NOT do**:
  - No simulation in init command

  **Parallelizable**: NO (CLI integration)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/github/linguist` - Language detection
  - `github.com/infracost/infracost` - Stack detection patterns

  **File Detection**:
  ```
  go.mod → Go
  package.json → Node.js
  requirements.txt → Python
  Cargo.toml → Rust
  Dockerfile → Container
  terraform/*.tf → Infrastructure
  next.config.js → Next.js
  astro.config.mjs → Astro
  ```

  **Go Learning Context**:
  - File existence: `os.Stat(path)` returns FileInfo
  - Path joining: `filepath.Join(dir, "go.mod")`
  - Reading: `ioutil.ReadFile()` (deprecated, use `os.ReadFile()`)

  **Acceptance Criteria**:
  - [ ] Init command: `cmd/init.go`
  - [ ] Repo scanner: `internal/scanner/scanner.go`
  - [ ] Template generator: `internal/scanner/template.go`
  - [ ] Detects common stacks (Go, Node, Python, etc.)
  - [ ] Generates valid profit.yml
  - [ ] Run: `./profitctl init` → creates profit.yml
  - [ ] Run: `./profitctl validate profit.yml` → valid

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl init`
  - [ ] Verify: profit.yml created
  - [ ] Verify: Detected stack mentioned in output
  - [ ] Verify: Config is valid

  **Commit**: YES
  - Message: `feat: implement init command with repo scanning`
  - Files: `cmd/init.go`, `internal/scanner/scanner.go`, `internal/scanner/template.go`

---

- [ ] 16. Simulate Command

  **What to do**:
  - Implement `profitctl simulate` command
  - Parse profit.yml configuration
  - Run cost engine calculations
  - Run scale simulation
  - Run p95/p99 stress testing
  - Validate covenants
  - Format and output results based on flags

  **Must NOT do**:
  - No validation logic (separate command)

  **Parallelizable**: NO (CLI integration)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/spf13/cobra` - Command flags and execution
  - `github.com/spf13/viper` - Config file loading

  **Command Structure**:
  ```
  profitctl simulate [flags]
    -f, --file string   Config file (default "profit.yml")
    --json              Output as JSON
    --markdown          Output as markdown
    --quiet             Minimal output
    -v, --verbose       Verbose output
  ```

  **Go Learning Context**:
  - Flags: `cmd.Flags().StringVarP(&configFile, "file", "f", "profit.yml")`
  - Cobra execution: `Run: func(cmd *cobra.Command, args []string) {...}`
  - Error handling: Return errors from command function

  **Acceptance Criteria**:
  - [ ] Simulate command: `cmd/simulate.go`
  - [ ] File flag parsing (-f, --file)
  - [ ] Output format flags (--json, --markdown)
  - [ ] Calls cost engine, simulation, covenant engine
  - [ ] Formats output based on flags
  - [ ] Run: `./profitctl simulate` → human output
  - [ ] Run: `./profitctl simulate --json` → JSON output
  - [ ] Run: `./profitctl simulate --markdown` → markdown

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl simulate -f examples/profit.yml`
  - [ ] Verify: Simulation runs end-to-end
  - [ ] Verify: Output shows all metrics
  - [ ] Verify: Exit code 0

  **Commit**: YES
  - Message: `feat: implement simulate command with all engines`
  - Files: `cmd/simulate.go`

---

- [ ] 17. Validate Command

  **What to do**:
  - Implement `profitctl validate` command
  - Parse profit.yml configuration
  - Validate YAML syntax
  - Validate against schema
  - Validate covenant rules syntax
  - Output validation results

  **Must NOT do**:
  - No simulation in validate command

  **Parallelizable**: NO (CLI integration)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/santhosh-tekuri/jsonschema` - Schema validation

  **Command Structure**:
  ```
  profitctl validate [flags]
    -f, --file string   Config file (default "profit.yml")
    --strict            Strict validation (fail on warnings)
  ```

  **Go Learning Context**:
  - Multiple errors: `fmt.Errorf("validation failed:\n%s", strings.Join(errors, "\n"))`
  - Exit: `os.Exit(1)` for command-level errors

  **Acceptance Criteria**:
  - [ ] Validate command: `cmd/validate.go`
  - [ ] YAML syntax validation
  - [ ] Schema validation
  - [ ] Covenant rule validation
  - [ ] Clear error messages
  - [ ] Exit code 0 for valid, 2 for errors
  - [ ] Run: `./profitctl validate` → "Config is valid"
  - [ ] Run: `./profitctl validate invalid.yml` → error list

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl validate -f examples/profit.yml`
  - [ ] Verify: "Config is valid" output
  - [ ] Run: `./profitctl validate broken.yml`
  - [ ] Verify: Clear error messages

  **Commit**: YES
  - Message: `feat: implement validate command for config checking`
  - Files: `cmd/validate.go`

---

- [ ] 18. CI Integration (Exit Codes)

  **What to do**:
  - Implement proper exit codes for all commands
  - Exit 0: Success (simulate passed, validate passed)
  - Exit 1: Failure (covenant breach, validation failed)
  - Exit 2: Error (file not found, parsing error, etc.)
  - Document exit codes in help and README
  - Add GitHub Actions example workflow

  **Must NOT do**:
  - No special CI logic beyond exit codes

  **Parallelizable**: NO (CLI integration)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/infracost/infracost` - CI integration patterns
  - GitHub Actions documentation for exit codes

  **Exit Code Specification**:
  ```
  0   - Success (all checks passed)
  1   - Covenant failure (profit targets not met)
  2   - Error (file not found, parsing error, etc.)
  ```

  **GitHub Actions Example**:
  ```yaml
  - name: Run profitctl
    run: ./profitctl simulate --exit-code-on-breach 1
    continue-on-error: false
  ```

  **Go Learning Context**:
  - Exit: `os.Exit(0)` or `os.Exit(1)`
  - Deferred functions: `defer func() { ... }()` for cleanup
  - Error wrapping: `fmt.Errorf("failed: %w", err)` for context

  **Acceptance Criteria**:
  - [ ] Exit code 0: Success
  - [ ] Exit code 1: Covenant breach
  - [ ] Exit code 2: Errors
  - [ ] GitHub Actions workflow example: `.github/workflows/profit.yml`
  - [ ] Documentation in README
  - [ ] Test: `./profitctl simulate` → exit 0
  - [ ] Test: `./profitctl validate broken.yml` → exit 2

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl simulate -f examples/profit.yml && echo "Exit: $?"`
  - [ ] Verify: Exit code 0
  - [ ] Run: `./profitctl validate missing.yml; echo "Exit: $?"`
  - [ ] Verify: Exit code 2

  **Commit**: YES
  - Message: `feat: implement CI-friendly exit codes`
  - Files: `cmd/simulate.go`, `cmd/validate.go`, `.github/workflows/profit.yml`

---

- [ ] 19. Example profit.yml

  **What to do**:
  - Create comprehensive example `examples/profit.yml`
  - Include fixed costs (infrastructure, application, service layers)
  - Include variable costs with realistic distributions
  - Include pricing plans with limits
  - Include covenants (threshold and expression-based)
  - Document all fields with comments

  **Must NOT do**:
  - No production secrets or real pricing

  **Parallelizable**: NO (documentation)

  **References**:

  **Pattern References** (existing code to follow):
  - `github.com/infracost/infracost` - Example configurations

  **Example Structure**:
  ```yaml
  # profitctl Example Configuration
  # This demonstrates all features of the MVP

  # Fixed Costs (monthly)
  fixed_costs:
    - name: "AWS EC2"
      amount: 500
      period: monthly
      layer: infrastructure

    - name: "Database (RDS)"
      amount: 200
      period: monthly
      layer: infrastructure

    - name: "Auth Service (Auth0)"
      amount: 100
      period: monthly
      layer: service

    - name: "Dev Team Tools"
      amount: 300
      period: monthly
      layer: application

  # Variable Costs (per user, per month)
  variable_costs:
    - name: "API Calls"
      cost_per_unit: 0.0001
      units_per_user: 10000
      distribution: normal
      mean: 10000
      stddev: 2000
      layer: application

    - name: "Storage (S3)"
      cost_per_unit: 0.023
      units_per_user: 5  # GB
      distribution: uniform
      min: 2
      max: 10
      layer: infrastructure

    - name: "LLM Tokens"
      cost_per_unit: 0.01
      units_per_user: 5000
      distribution: exponential
      layer: service

  # Pricing Plans
  pricing:
    plans:
      - name: "Basic"
        price: 10
        limits:
          users: 1000

      - name: "Pro"
        price: 29
        limits:
          users: 5000

      - name: "Enterprise"
        price: 99
        limits:
          users: 10000

  # Profit Covenants (hybrid: threshold + expression)
  covenants:
    # Threshold-based covenant
    - type: threshold
      field: margin
      operator: gte
      value: 20
      message: "Gross margin must be at least 20%"

    # Expression-based covenant
    - type: expression
      expression: "p95_margin > 15"
      message: "p95 margin must be greater than 15%"

    # Cost per user threshold
    - type: threshold
      field: cost_per_user
      operator: lte
      value: 5
      message: "Cost per user must not exceed $5"

  # Simulation Settings
  simulation:
    base_users: 100
    growth_factor: 1.5  # 50% growth per step
    iterations: 10000   # Monte Carlo iterations
  ```

  **Go Learning Context**:
  - Comments: `//` for Go, `#` for YAML
  - Documentation: Comments on struct fields become docs
  - Examples: Show realistic values, not just placeholders

  **Acceptance Criteria**:
  - [ ] Example file: `examples/profit.yml`
  - [ ] All features demonstrated
  - [ ] All fields documented
  - [ ] Valid YAML syntax
  - [ ] Passes validation
  - [ ] Runs successfully in simulation

  **Manual Execution Verification**:
  - [ ] Run: `./profitctl validate -f examples/profit.yml`
  - [ ] Verify: "Config is valid"
  - [ ] Run: `./profitctl simulate -f examples/profit.yml`
  - [ ] Verify: Simulation runs successfully

  **Commit**: YES
  - Message: `docs: add comprehensive example profit.yml`
  - Files: `examples/profit.yml`

---

- [ ] 20. Go Learning Guide

  **What to do**:
  - Create comprehensive Go learning guide for Python/TS developers
  - Include syntax comparisons (Python/TS → Go)
  - Cover key concepts (goroutines, interfaces, error handling)
  - Provide pattern examples for profitctl code
  - Link to official resources (Tour of Go, Effective Go)
  - Include common pitfalls and how to avoid them

  **Must NOT do**:
  - No duplicate of official Go docs (link instead)

  **Parallelizable**: NO (documentation)

  **References**:

  **Pattern References** (existing code to follow):
  - `tour.golang.org` - Official interactive tutorial
  - `gobyexample.com` - Practical code examples
  - `golang.org/doc/effective_go` - Best practices

  **Guide Structure**:
  ```markdown
  # Go Learning Guide for profitctl

  ## Why Go?

  - Single binary distribution
  - Excellent for CLI tools
  - Strong concurrency model
  - Battle-tested ecosystem

  ## Syntax Comparison

  ### Functions

  Python:
  ```python
  def calculate_margin(revenue, costs):
      return (revenue - costs) / revenue
  ```

  TypeScript:
  ```typescript
  function calculateMargin(revenue: number, costs: number): number {
      return (revenue - costs) / revenue;
  }
  ```

  Go:
  ```go
  func calculateMargin(revenue, costs float64) float64 {
      return (revenue - costs) / revenue
  }
  ```

  ### Structs vs Classes vs Interfaces

  ### Error Handling (no exceptions)

  ### Packages and Imports

  ### Concurrency (goroutines)

  ## Common Pitfalls

  1. Forgetting `if err != nil`
  2. Nil slices vs empty slices
  3. Maps must be initialized
  4. Goroutine closures

  ## profitctl Code Patterns

  ### Pure Functions

  ### Interface Design

  ### Testing Patterns
  ```

  **Go Learning Context**:
  - Learning curve: Expect 1-2 weeks for basics, 1 month for comfort
  - Practice: The best way to learn is by building
  - Community: r/golang, Gophers Slack

  **Acceptance Criteria**:
  - [ ] Learning guide: `docs/GO_LEARNING_GUIDE.md`
  - [ ] Syntax comparisons included
  - [ ] Key concepts explained
  - [ ] Common pitfalls documented
  - [ ] Links to official resources
  - [ ] profitctl code examples included

  **Manual Execution Verification**:
  - [ ] Guide is readable and clear
  - [ ] Code examples are correct
  - [ ] Links work
  - [ ] Helpful for Python/TS developer

  **Commit**: YES
  - Message: `docs: add Go learning guide for Python/TS developers`
  - Files: `docs/GO_LEARNING_GUIDE.md`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 0 | `chore: initialize Go module and test infrastructure` | `go.mod`, `go.sum`, test files | `go test ./...` |
| 1 | `feat: add Cobra CLI scaffold with init/simulate/validate commands` | `cmd/`, `main.go` | `./profitctl --help` |
| 2 | `feat: add YAML schema and validation` | `schema/`, `internal/config/` | `./profitctl validate` |
| 3 | `feat: add core types for costs, layers, and distributions` | `pkg/types/` | `go build` |
| 4 | `feat: implement fixed cost calculation engine` | `internal/cost/fixed.go` | Tests pass |
| 5 | `feat: implement variable cost calculation with distributions` | `internal/cost/variable.go` | Tests pass |
| 6 | `feat: implement unified cost engine with layer abstraction` | `internal/cost/engine.go` | Tests pass |
| 7 | `feat: implement exponential scale simulation engine` | `internal/simulation/scale.go` | Tests pass |
| 8 | `feat: implement p95/p99 Monte Carlo stress testing` | `internal/simulation/montecarlo.go` | Tests pass |
| 9 | `feat: implement multi-tier pricing calculator` | `internal/pricing/calculator.go` | Tests pass |
| 10 | `feat: implement margin and cost-per-user calculations` | `internal/pricing/margin.go` | Tests pass |
| 11 | `feat: implement hybrid covenant validation engine` | `internal/covenant/engine.go` | Tests pass |
| 12 | `feat: implement human-readable CLI output formatting` | `internal/output/cli.go` | Manual verify |
| 13 | `feat: implement JSON output for CI artifacts` | `internal/output/json.go` | `jq` validation |
| 14 | `feat: implement markdown output for PR comments` | `internal/output/markdown.go` | Manual verify |
| 15 | `feat: implement init command with repo scanning` | `cmd/init.go`, `internal/scanner/` | `./profitctl init` |
| 16 | `feat: implement simulate command with all engines` | `cmd/simulate.go` | End-to-end test |
| 17 | `feat: implement validate command for config checking` | `cmd/validate.go` | `./profitctl validate` |
| 18 | `feat: implement CI-friendly exit codes` | CLI commands, workflows | Exit codes verified |
| 19 | `docs: add comprehensive example profit.yml` | `examples/profit.yml` | Valid config |
| 20 | `docs: add Go learning guide for Python/TS developers` | `docs/GO_LEARNING_GUIDE.md` | Readable guide |

---

## Success Criteria

### Verification Commands

```bash
# Setup
go mod init github.com/profitctl/profitctl
go get github.com/spf13/cobra gopkg.in/yaml.v3 github.com/stretchr/testify

# Build
go build -o profitctl .

# CLI Tests
./profitctl --help                           # All commands visible
./profitctl init                             # Creates profit.yml
./profitctl validate -f examples/profit.yml  # "Config is valid"
./profitctl simulate -f examples/profit.yml  # Full simulation output
./profitctl simulate -f examples/profit.yml --json  # JSON output
./profitctl simulate -f examples/profit.yml --markdown  # Markdown output

# Exit Code Tests
./profitctl validate examples/profit.yml; echo "Exit: $?"  # 0
./profitctl simulate examples/profit.yml; echo "Exit: $?"  # 0

# Test Suite
go test ./...  # All tests pass
```

### Final Checklist

- [ ] All "Must Have" present
  - [ ] CLI with Cobra framework
  - [ ] YAML parsing and validation
  - [ ] Fixed cost modeling
  - [ ] Variable cost modeling with distributions
  - [ ] Exponential scale simulation
  - [ ] p95/p99 stress testing
  - [ ] Threshold-based covenants
  - [ ] Expression-based covenants
  - [ ] Human-readable CLI output
  - [ ] JSON output for CI
  - [ ] Markdown output for PR comments
  - [ ] Exit codes (0, 1, 2)
  - [ ] Example profit.yml
  - [ ] Go learning guide

- [ ] All "Must NOT Have" absent
  - [ ] No external API calls
  - [ ] No persistent state
  - [ ] No AI/ML inference
  - [ ] No dashboards
  - [ ] No auto-pricing changes

- [ ] All tests pass
- [ ] All commands work end-to-end
- [ ] Config is explicit and editable
- [ ] No external dependencies required
- [ ] Single binary distribution works
