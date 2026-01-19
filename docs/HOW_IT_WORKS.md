# How profitctl Works

## Overview

`profitctl` is a CLI tool that simulates unit economics for software products by calculating costs, revenue, margins, and validating profitability constraints (covenants). It helps developers understand how their product scales financially and can fail CI builds when profitability targets aren't met.

## Architecture

The tool is built with a modular architecture in Go using the Cobra CLI framework:

```
profitctl
├── cmd/              # CLI commands (simulate, init, validate)
├── internal/
│   ├── config/      # YAML parsing and validation
│   ├── cost/        # Fixed & variable cost calculations
│   ├── pricing/     # Revenue and margin calculations
│   ├── simulation/  # Scale and Monte Carlo simulations
│   ├── covenant/    # Profitability constraint validation
│   └── output/      # Formatting (CLI, JSON, Markdown)
└── pkg/types/       # Shared type definitions
```

## Execution Flow

### Step-by-Step Process

When you run `profitctl simulate`, here's what happens:

```
1. Parse Configuration (profit.yml)
   ↓
2. Calculate Fixed Costs (monthly)
   ↓
3. Calculate Variable Costs (per user)
   ↓
4. Run Scale Simulation (growth scenarios)
   ↓
5. Run Stress Test (Monte Carlo p95/p99)
   ↓
6. Calculate Revenue (from pricing plans)
   ↓
7. Calculate Margins (gross, per-user, by layer)
   ↓
8. Validate Covenants (profitability rules)
   ↓
9. Format Output (CLI/JSON/Markdown)
   ↓
10. Exit with Code (0=pass, 1=breach, 2=error)
```

Let's break down each step:

## 1. Configuration Parsing

**Location**: `internal/config/parser.go`

The tool reads a YAML configuration file (`profit.yml` by default) that defines:

- **Fixed Costs**: Monthly/yearly costs that don't vary with users (e.g., AWS hosting, SaaS subscriptions)
- **Variable Costs**: Per-user costs that scale (e.g., API calls, storage, compute)
- **Pricing Plans**: Tiered pricing with user limits (e.g., Basic $10, Pro $29)
- **Covenants**: Profitability rules (e.g., "margin must be ≥ 20%")
- **Simulation Settings**: Base users, growth factor, Monte Carlo iterations

**Example**:
```yaml
fixed_costs:
  - name: AWS EC2
    amount: 500
    period: monthly
    layer: infrastructure

variable_costs:
  - name: API Calls
    cost_per_unit: 0.0001
    units_per_user: 10000
    distribution: normal
    mean: 10000
    stddev: 2000

simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000
```

## 2. Fixed Cost Calculation

**Location**: `internal/cost/fixed.go`

Fixed costs are converted to monthly amounts and aggregated by layer:

- **Infrastructure Layer**: Hosting, databases, CDN
- **Application Layer**: Monitoring, logging, tools
- **Service Layer**: Third-party APIs, auth services

**How it works**:
```go
// Convert all periods to monthly
daily * 30 = monthly
monthly * 1 = monthly
yearly / 12 = monthly

// Aggregate by layer
total_infrastructure = sum of infrastructure costs
total_application = sum of application costs
total_service = sum of service costs
total_monthly = infrastructure + application + service
```

**Example**:
- AWS EC2: $500/month (infrastructure)
- Auth0: $100/month (service)
- Dev Tools: $300/month (application)
- **Total Fixed**: $900/month

## 3. Variable Cost Calculation

**Location**: `internal/cost/variable.go`

Variable costs are calculated per user with statistical distributions:

- **Normal Distribution**: Bell curve (e.g., API usage patterns)
- **Uniform Distribution**: Flat range (e.g., storage requirements)
- **Exponential Distribution**: Heavy tail (e.g., outlier user behavior)

**How it works**:
```go
// Deterministic calculation
variable_cost_per_user = sum of (cost_per_unit * units_per_user)

// Stochastic calculation (for stress testing)
For each variable cost:
  Generate random units based on distribution
  Calculate: cost_per_unit * random_units
  Sum all variable costs for that user
```

**Example**:
- API Calls: $0.0001/unit * 10,000 units = $1.00/user (normal distribution)
- Storage: $0.023/GB * 5 GB = $0.115/user (uniform: 2-10 GB)
- LLM Tokens: $0.01/token * 5,000 tokens = $50/user (exponential)
- **Total Variable**: ~$51.12/user

## 4. Scale Simulation

**Location**: `internal/simulation/scale.go`

Models how costs change as users grow exponentially:

**How it works**:
```go
scenarios = []
current_users = base_users

for each step:
  fixed_cost = calculate_monthly_fixed(total_months)
  variable_cost = calculate_variable(current_users)
  total_cost = fixed_cost + variable_cost
  scenarios.append({users: current_users, cost: total_cost})
  current_users *= growth_factor  // Exponential growth
```

**Example**:
```
Step 0: 100 users  → $6,000 cost
Step 1: 150 users  → $8,700 cost (1.5x growth)
Step 2: 225 users  → $12,475 cost (1.5² growth)
Step 3: 337 users  → $17,214 cost (1.5³ growth)
...
```

This shows how costs scale with user growth.

## 5. Monte Carlo Stress Testing

**Location**: `internal/simulation/montecarlo.go`

Runs thousands of simulations with random variable costs to find worst-case scenarios:

**How it works**:
```go
results = []
for i = 0 to iterations (e.g., 10,000):
  total_cost = fixed_cost
  for each user:
    variable_cost = random_variable_cost()  // Based on distribution
    total_cost += variable_cost
  cost_per_user = total_cost / num_users
  results.append(cost_per_user)

// Find percentiles
p95_cost_per_user = 95th percentile of results
p99_cost_per_user = 99th percentile of results
```

**Why it matters**: Some users might cost much more than average (e.g., heavy API users, large storage requirements). Stress testing finds these outliers.

**Example**:
- Mean cost per user: $51.12
- p95 cost per user: $68,701 (worst 5% of scenarios)
- p99 cost per user: $120,000+ (worst 1% of scenarios)

## 6. Revenue Calculation

**Location**: `internal/pricing/calculator.go`

Calculates revenue from tiered pricing plans:

**How it works**:
```go
revenue = 0
remaining_users = total_users

for each plan (sorted by price):
  users_in_plan = min(remaining_users, plan.limits.users)
  revenue += plan.price * users_in_plan
  remaining_users -= users_in_plan

// If users exceed all plans, use highest tier
if remaining_users > 0:
  revenue += highest_plan.price * remaining_users
```

**Example**:
```
100 users total:
- 100 users in Basic ($10) = $1,000 revenue

10,000 users total:
- 1,000 users in Basic ($10) = $10,000
- 5,000 users in Pro ($29) = $145,000
- 4,000 users in Enterprise ($99) = $396,000
Total: $551,000 revenue
```

## 7. Margin Calculation

**Location**: `internal/pricing/margin.go`

Calculates profitability metrics:

**Metrics**:
- **Gross Margin**: `((revenue - total_cost) / revenue) * 100`
- **Cost Per User**: `total_cost / num_users`
- **Layer Margins**: Breakdown by infrastructure, application, service

**Example**:
```
Revenue: $1,000
Total Cost: $6,000 (fixed $900 + variable $5,100)
Gross Margin: (($1,000 - $6,000) / $1,000) * 100 = -500%
Cost Per User: $6,000 / 100 = $60/user

Layer Breakdown:
- Infrastructure: $700 fixed + $1,150 variable = $1,850 (31%)
- Application: $300 fixed + $100 variable = $400 (7%)
- Service: $100 fixed + $5,000 variable = $5,100 (85%)
```

## 8. Covenant Validation

**Location**: `internal/covenant/engine.go`

Validates profitability constraints defined in the config:

**Supported Fields**:
- `margin`: Gross margin percentage
- `cost_per_user`: Average cost per user
- `p95_margin`: Margin at 95th percentile (stress test)
- `p95_cost_per_user`: Cost per user at 95th percentile
- `p99_margin`: Margin at 99th percentile
- `p99_cost_per_user`: Cost per user at 99th percentile

**Supported Operators**:
- `gte` (≥): Greater than or equal
- `lte` (≤): Less than or equal
- `gt` (>): Greater than
- `lt` (<): Less than
- `eq` (=): Equal to

**How it works**:
```go
violations = []
for each covenant:
  actual_value = get_value_from_results(covenant.field)
  if !compare(actual_value, covenant.operator, covenant.value):
    violations.append(covenant.message)

passed = len(violations) == 0
```

**Example**:
```yaml
covenants:
  - field: margin
    operator: gte
    value: 20
    message: Gross margin must be at least 20%

  - field: cost_per_user
    operator: lte
    value: 5
    message: Cost per user must not exceed $5
```

If margin = -500% and cost_per_user = $60:
- ❌ Margin violation: -500% < 20%
- ❌ Cost violation: $60 > $5
- **Result**: FAILED (2 violations)

## 9. Output Formatting

**Location**: `internal/output/`

Three output formats:

### CLI Format (Default)
Human-readable tables and summaries:
```
=== profitctl Simulation Results ===

Scenario: 100 users
───────────────────────────────
Mean margin: -500.00%
Fixed COGS: $900.00/month
Variable COGS: $51.12/user
Cost per user: $60.00

Covenant Status: ❌ FAILED

Violations:
  • Gross margin must be at least 20%
    Expected: >= 20.00, Actual: -500.00
  • Cost per user must not exceed $5
    Expected: <= 5.00, Actual: 60.00
```

### JSON Format (`--json`)
Structured data for CI/CD and automation:
```json
{
  "scenario": {"users": 100, "months": 12},
  "costs": {
    "fixed": {"total": 10800, "monthly": 900},
    "variable": {"total": 5112, "per_user": 51.12}
  },
  "margin": {"gross": -500.0, "cost_per_user": 60.0},
  "covenants": {
    "passed": false,
    "violations": [...]
  }
}
```

### Markdown Format (`--markdown`)
GitHub-compatible tables for PR comments:
```markdown
## profitctl Results

| Metric | Value |
|--------|-------|
| Margin | -500.0% |
| Cost per User | $60.00 |
| Covenant Status | ❌ FAILED |
```

## 10. Exit Codes

**CI/CD Integration**:
- **Exit Code 0**: Success, all covenants passed
- **Exit Code 1**: Covenant breach (fails CI build)
- **Exit Code 2**: Error (config parsing, file not found, etc.)

This allows CI pipelines to fail builds when profitability targets aren't met:
```yaml
# GitHub Actions example
- name: Check Profitability
  run: profitctl simulate -f profit.yml
  # Build fails if exit code != 0
```

## Cost Engine Architecture

The cost engine (`internal/cost/engine.go`) unifies fixed and variable cost calculations:

```go
type CostEngine interface {
    CalculateFixedCosts(months int) FixedCostResult
    CalculateVariableCosts(users int) VariableCostResult
    CalculateTotalCosts(users int, months int) TotalCostResult
}
```

**Total Cost Calculation**:
```go
total = fixed_costs(12 months) + variable_costs(num_users)
total_by_layer = fixed_by_layer + variable_by_layer
```

This provides a clean interface for cost calculations used by simulations and margin calculations.

## Statistical Distributions

Variable costs use statistical distributions to model real-world variability:

### Normal Distribution
**Use case**: Most users cluster around average, few outliers
```yaml
distribution: normal
mean: 10000      # Average units per user
stddev: 2000     # Standard deviation (spread)
```

### Uniform Distribution
**Use case**: Users evenly distributed across a range
```yaml
distribution: uniform
min: 2           # Minimum units
max: 10          # Maximum units
```

### Exponential Distribution
**Use case**: Heavy tail - many small users, few very large users
```yaml
distribution: exponential
rate: 0.001      # Rate parameter (lower = more variance)
```

## Complete Example Workflow

Let's trace through a complete example:

### Input: `profit.yml`
```yaml
fixed_costs:
  - name: AWS
    amount: 500
    period: monthly

variable_costs:
  - name: API Calls
    cost_per_unit: 0.0001
    units_per_user: 10000

pricing:
  plans:
    - name: Basic
      price: 10

simulation:
  base_users: 100

covenants:
  - field: margin
    operator: gte
    value: 20
```

### Execution Steps:

1. **Parse config**: Load YAML → structured config
2. **Fixed costs**: $500/month * 12 months = $6,000
3. **Variable costs**: $0.0001 * 10,000 = $1/user * 100 users = $100
4. **Scale simulation**: 100 → 150 → 225 users (with costs)
5. **Stress test**: Run 10,000 Monte Carlo iterations → find p95/p99
6. **Revenue**: 100 users * $10 = $1,000
7. **Margins**: 
   - Total cost: $6,000 + $100 = $6,100
   - Gross margin: (($1,000 - $6,100) / $1,000) * 100 = -510%
8. **Covenants**: -510% < 20% → ❌ FAILED
9. **Output**: Format results (CLI/JSON/Markdown)
10. **Exit**: Exit code 1 (covenant breach)

### Output:
```
Covenant Status: ❌ FAILED
Violations:
  • Gross margin must be at least 20%
    Expected: >= 20.00, Actual: -510.00
```

## Key Design Principles

1. **Pure Functions**: All calculations are pure (no side effects), making testing easy
2. **Layer Separation**: Costs separated into infrastructure/application/service layers
3. **Statistical Modeling**: Variable costs use distributions to model real-world variance
4. **Stress Testing**: Monte Carlo simulation finds worst-case scenarios
5. **CI-Friendly**: Exit codes enable automated profitability checks
6. **Multiple Formats**: CLI for humans, JSON for automation, Markdown for PRs

## Testing

The tool includes comprehensive tests:
- **Unit Tests**: Each component tested in isolation
- **Integration Tests**: End-to-end CLI execution tests
- **Test Fixtures**: Various config scenarios (valid, minimal, covenant breach, etc.)

Run tests:
```bash
go test ./...
```

## Summary

`profitctl` works by:

1. Reading a YAML configuration defining costs, pricing, and covenants
2. Calculating fixed and variable costs with statistical modeling
3. Running scale simulations to see how costs grow
4. Running stress tests to find worst-case scenarios
5. Calculating revenue and margins
6. Validating profitability constraints (covenants)
7. Formatting results in multiple ways
8. Exiting with codes suitable for CI/CD integration

This gives developers visibility into unit economics and enforces profitability constraints in their development workflow.