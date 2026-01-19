# profitctl Architecture Overview

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Interface                         │
│                      (cmd/simulate.go)                       │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                   Configuration Parser                       │
│                   (internal/config/)                         │
│                                                              │
│  Reads profit.yml → YAML → Config struct                    │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                      Cost Engine                             │
│                   (internal/cost/)                           │
│                                                              │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │ Fixed Costs  │         │Variable Costs│                 │
│  │              │         │              │                 │
│  │ - Monthly    │         │ - Per User   │                 │
│  │ - Yearly     │         │ - Distributions│              │
│  │ - By Layer   │         │ - By Layer   │                 │
│  └──────────────┘         └──────────────┘                 │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                  Simulation Engines                          │
│                (internal/simulation/)                        │
│                                                              │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │ Scale        │         │Monte Carlo   │                 │
│  │ Simulation   │         │Stress Test   │                 │
│  │              │         │              │                 │
│  │ Growth:      │         │ Iterations:  │                 │
│  │ 100 → 150    │         │ 10,000 runs  │                 │
│  │ 150 → 225    │         │ Find p95/p99 │                 │
│  └──────────────┘         └──────────────┘                 │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                   Pricing Calculator                         │
│                  (internal/pricing/)                         │
│                                                              │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │ Revenue      │         │  Margins     │                 │
│  │              │         │              │                 │
│  │ - Tiered     │         │ - Gross      │                 │
│  │ - By Plan    │         │ - Per User   │                 │
│  │              │         │ - By Layer   │                 │
│  └──────────────┘         └──────────────┘                 │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                  Covenant Validator                          │
│                 (internal/covenant/)                         │
│                                                              │
│  Validates profitability constraints:                       │
│  - Margin thresholds (≥, ≤)                                 │
│  - Cost per user caps                                       │
│  - p95/p99 stress test thresholds                           │
│                                                              │
│  Returns: Pass/Fail + Violations                            │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                    Output Formatters                         │
│                   (internal/output/)                         │
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │
│  │   CLI    │  │   JSON   │  │ Markdown │                 │
│  │  Tables  │  │  Structured│ │ GitHub   │                 │
│  │          │  │           │  │  PRs     │                 │
│  └──────────┘  └──────────┘  └──────────┘                 │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
                  Exit Code (0/1/2)
```

## Data Flow

### Input
```yaml
profit.yml
├── fixed_costs: [list of monthly/yearly costs]
├── variable_costs: [list of per-user costs with distributions]
├── pricing: [tiered pricing plans]
├── covenants: [profitability rules]
└── simulation: [base users, growth factor, iterations]
```

### Processing Pipeline

```
1. Parse Config
   └─> Config struct

2. Calculate Fixed Costs (12 months)
   └─> FixedCostResult {
         Total: $10,800
         Monthly: $900
         ByLayer: {Infrastructure: $600, Application: $300, Service: $100}
       }

3. Calculate Variable Costs (100 users)
   └─> VariableCostResult {
         Total: $5,112
         PerUser: $51.12
         ByLayer: {Infrastructure: $1,150, Application: $100, Service: $5,000}
       }

4. Run Scale Simulation
   └─> ScaleResult {
         Scenarios: [
           {Users: 100, Fixed: $10,800, Variable: $5,112, Total: $15,912},
           {Users: 150, Fixed: $10,800, Variable: $7,668, Total: $18,468},
           ...
         ]
       }

5. Run Monte Carlo Stress Test (10,000 iterations)
   └─> StressResult {
         MeanCostPerUser: $51.12
         P95CostPerUser: $68,701
         P99CostPerUser: $120,000
       }

6. Calculate Revenue (100 users, tiered pricing)
   └─> RevenueResult {
         Total: $1,000
         ByPlan: [{Plan: "Basic", Users: 100, Revenue: $1,000}]
       }

7. Calculate Margins
   └─> MarginResult {
         GrossMargin: -500%
         CostPerUser: $60.00
         LayerMargins: {...}
         LayerCostPercent: {Infrastructure: 31%, Application: 7%, Service: 85%}
       }

8. Validate Covenants
   └─> CovenantValidation {
         Passed: false
         Violations: [
           "Gross margin must be at least 20% (Expected: >= 20.00, Actual: -500.00)",
           "Cost per user must not exceed $5 (Expected: <= 5.00, Actual: 60.00)"
         ]
       }
```

### Output
```
CLI Format:
  - Human-readable tables
  - Scenario summary
  - Covenant status
  - Violations list

JSON Format:
  - Structured data
  - CI/CD integration
  - Automated parsing

Markdown Format:
  - GitHub-compatible tables
  - PR comments
  - Documentation
```

## Component Responsibilities

### cmd/simulate.go
- **Orchestrates** the entire simulation pipeline
- **Coordinates** all engines (cost, pricing, simulation, covenant)
- **Handles** CLI flags and output formatting
- **Sets** exit codes for CI/CD integration

### internal/config/
- **Parses** YAML configuration files
- **Validates** configuration structure
- **Converts** YAML to Go structs

### internal/cost/
- **Calculates** fixed costs (monthly/yearly)
- **Calculates** variable costs (per-user with distributions)
- **Aggregates** costs by layer (infrastructure/application/service)
- **Provides** unified `CostEngine` interface

### internal/simulation/
- **Scale Simulation**: Models exponential user growth
- **Monte Carlo**: Runs thousands of random simulations
- **Stress Testing**: Finds worst-case scenarios (p95/p99)

### internal/pricing/
- **Revenue Calculation**: Tiered pricing with user limits
- **Margin Calculation**: Gross margin, cost per user, layer breakdowns

### internal/covenant/
- **Validates** profitability constraints
- **Checks** thresholds (margin, cost per user, p95/p99)
- **Reports** violations with messages

### internal/output/
- **CLI Formatter**: Human-readable tables and summaries
- **JSON Formatter**: Structured data for automation
- **Markdown Formatter**: GitHub-compatible tables

## Key Design Patterns

### 1. Pure Functions
All calculations are pure functions (no side effects):
```go
func CalculateMargins(revenue, costs float64, ...) MarginResult {
    // Pure calculation, no side effects
    return result
}
```

### 2. Interface Abstraction
Engines use interfaces for modularity:
```go
type CostEngine interface {
    CalculateFixedCosts(months int) FixedCostResult
    CalculateVariableCosts(users int) VariableCostResult
    CalculateTotalCosts(users, months int) TotalCostResult
}
```

### 3. Statistical Modeling
Variable costs use distributions for real-world variance:
```go
// Normal distribution for API usage
distribution: normal
mean: 10000
stddev: 2000

// Exponential distribution for heavy-tail scenarios
distribution: exponential
rate: 0.001
```

### 4. Layer Separation
Costs separated into three layers:
- **Infrastructure**: Hosting, databases, CDN
- **Application**: Monitoring, logging, tools
- **Service**: Third-party APIs, auth services

### 5. CI/CD Integration
Exit codes enable automated profitability checks:
```go
if !covenantValidation.Passed {
    os.Exit(1)  // Fail CI build
}
```

## Testing Strategy

### Unit Tests
- Each component tested in isolation
- Pure functions tested with various inputs
- Edge cases covered (zero users, empty config, etc.)

### Integration Tests
- End-to-end CLI execution
- Real config files with various scenarios
- Exit code verification

### Test Fixtures
- `valid_config.yml`: Comprehensive valid configuration
- `minimal_config.yml`: Minimal required fields
- `covenant_breach_config.yml`: Configuration designed to fail covenants
- `no_pricing_config.yml`: Configuration without pricing

## Extension Points

The architecture supports future enhancements:

1. **Expression-based Covenants**: Currently threshold-based, can add expression evaluation
2. **Additional Distributions**: Can add more statistical distributions for variable costs
3. **Custom Output Formats**: Can add CSV, XML, etc.
4. **Real-time Monitoring**: Can extend to live cost tracking
5. **Multi-currency**: Can add currency conversion

## Performance Considerations

- **Monte Carlo Iterations**: Configurable (default 10,000)
- **Scale Simulation**: Grows with growth factor (typically 5-10 steps)
- **Pure Functions**: No side effects, easy to parallelize
- **Memory**: All calculations in-memory, suitable for typical scenarios (< 1M users)

## Summary

`profitctl` uses a modular, layered architecture with:
- **Separation of Concerns**: Each component has a single responsibility
- **Pure Functions**: Easy to test and reason about
- **Statistical Modeling**: Real-world cost variability
- **CI/CD Integration**: Exit codes for automated checks
- **Multiple Output Formats**: CLI, JSON, Markdown for different use cases