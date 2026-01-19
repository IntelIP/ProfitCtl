# Monte Carlo Simulation Implementation Summary

## Overview

This document summarizes the implementation of Tasks 5, 7, and 8:
- **Task 5**: Variable Cost Engine with Statistical Distributions
- **Task 7**: Scale Simulation Engine with Exponential Growth
- **Task 8**: p95/p99 Stress Testing with Monte Carlo Simulation

---

## Task 5: Variable Cost Engine

### Implementation: `internal/cost/variable.go`

#### Key Features

1. **Statistical Distribution Generators**
   - **Normal Distribution**: `mean + stddev * rand.NormFloat64()`
   - **Uniform Distribution**: `min + (max - min) * rand.Float64()`
   - **Exponential Distribution**: `-ln(1 - rand.Float64()) / rate`

2. **Distribution Generator Factory**
   ```go
   func NewDistributionGenerator(vc VariableCost) (DistributionGenerator, error)
   ```
   - Validates distribution parameters
   - Returns typed generator function
   - Handles all three distribution types

3. **Variable Cost Calculation**
   - `CalculateVariableCosts()`: Deterministic calculation without variability
   - `CalculateVariableCostWithVariability()`: Stochastic calculation with distribution sampling
   - Per-user cost calculations
   - Layer aggregation (infrastructure, application, service)

4. **Data Structures**
   - `VariableCostResult`: Total, per-user, layer breakdown, samples
   - `VariableCostLineItem`: Individual cost line items with distribution info

### Testing: `internal/cost/variable_test.go`

- Normal distribution validation (mean/stddev)
- Uniform distribution validation (min/max)
- Exponential distribution validation (rate)
- Error handling for missing parameters
- Layer aggregation tests
- Zero user edge case

---

## Task 7: Scale Simulation Engine

### Implementation: `internal/simulation/scale.go`

#### Key Features

1. **Exponential Growth Model**
   ```go
   users_at_step(n) = base_users * (growth_factor ^ n)
   ```

2. **Scale Configuration**
   - Base users (default: 100)
   - Growth factor (default: 1.5, 50% per step)
   - Steps (default: 6)

3. **Scale Simulation Results**
   - Multiple scale points: 100, 150, 225, 338, 506, 760 users (with default config)
   - Fixed costs at each scale point
   - Variable costs at each scale point
   - Total costs (fixed + variable)
   - Layer breakdowns at each scale point

4. **Data Structures**
   - `ScaleSimulationResult`: Array of scenarios
   - `ScaleScenario`: Single simulation point with user count, costs
   - `TotalCosts`: Combined fixed + variable with layer breakdown
   - `ScaleConfig`: Simulation parameters

### Testing: `internal/simulation/scale_test.go`

- Exponential growth calculation tests
- Scale configuration tests
- Full simulation end-to-end tests
- Cost growth proportionality tests
- Layer aggregation tests

---

## Task 8: p95/p99 Stress Testing

### Implementation: `internal/simulation/montecarlo.go`

#### Key Features

1. **Monte Carlo Simulation**
   - Configurable iterations (default: 10,000)
   - Configurable seed for reproducibility
   - Random scenario generation using distribution generators
   - Fixed costs remain constant
   - Variable costs sampled from distributions

2. **Percentile Calculations**
   ```go
   p95 = samples[int(0.95 * len(samples))]
   p99 = samples[int(0.99 * len(samples))]
   ```

3. **Statistical Metrics**
   - **Mean**: Average cost across all iterations
   - **Median**: Middle value of sorted samples
   - **StdDev**: Standard deviation (measure of variability)
   - **P95**: 95th percentile (worst 5% of cases)
   - **P99**: 99th percentile (worst 1% of cases)

4. **Stress Test Results**
   - Per-user metrics (mean, p95, p99 cost per user)
   - Worst-case total cost (p99)
   - Monte Carlo simulation results with samples

5. **Data Structures**
   - `MonteCarloResult`: Iterations, samples, percentiles, statistics
   - `MonteCarloConfig`: Iterations, seed
   - `StressTestResult`: User metrics, worst-case scenarios

### Testing: `internal/simulation/montecarlo_test.go`

- Percentile calculation tests (p95, p99, median)
- Mean, median, stdDev calculation tests
- Monte Carlo simulation validation
- Per-user metric calculations
- Stress test end-to-end validation
- Precision rounding tests

---

## Key Implementation Details

### Random Number Generation

- Uses `math/rand` package (as specified)
- Standard library: `rand.NormFloat64()`, `rand.Float64()`
- Seeded with `time.Now().UnixNano()` for randomness

### Precision Handling

- Financial calculations rounded to 2 decimal places
- Uses `math.Round(value * 100) / 100` pattern
- Prevents floating-point precision issues

### Error Handling

- Distribution parameter validation
- Clear error messages for missing parameters
- Graceful fallback to base cost if distribution fails

### Layer Separation

- Costs separated by: infrastructure, application, service
- Aggregation at each calculation level
- Supports cost analysis by stack layer

---

## Test Coverage

All implementations include comprehensive test coverage:

- **Variable Cost Engine**: 10 test cases
- **Scale Simulation Engine**: 6 test cases
- **Monte Carlo Simulation**: 11 test cases

Total: **27 test cases**, all passing

---

## Usage Example

```go
import (
    "github.com/profitctl/profitctl/internal/cost"
    "github.com/profitctl/profitctl/internal/simulation"
    "github.com/profitctl/profitctl/pkg/types"
)

// Define variable costs with distributions
variableCosts := []types.VariableCost{
    {
        Name:         "API Calls",
        CostPerUnit:  0.0001,
        UnitsPerUser: 10000,
        Distribution: types.DistNormal,
        Mean:         newFloat64(10000),
        StdDev:       newFloat64(2000),
        Layer:        types.LayerApplication,
    },
    {
        Name:         "Storage",
        CostPerUnit:  0.023,
        UnitsPerUser: 5,
        Distribution: types.DistUniform,
        Min:          newFloat64(2),
        Max:          newFloat64(10),
        Layer:        types.LayerInfrastructure,
    },
}

// Run scale simulation
scaleConfig := simulation.NewScaleConfig(&config.SimulationConfig{
    BaseUsers:    100,
    GrowthFactor:  1.5,
    Iterations:   10000,
})
scaleResult := simulation.RunScaleSimulation(fixedCosts, variableCosts, scaleConfig, 6)

// Run stress test (Monte Carlo)
stressResult := simulation.RunStressTest(fixedCosts, variableCosts, 1000, 12, 10000)

fmt.Printf("P95 Cost Per User: $%.2f\n", stressResult.P95CostPerUser)
fmt.Printf("P99 Cost Per User: $%.2f\n", stressResult.P99CostPerUser)
fmt.Printf("Worst Case Total Cost: $%.2f\n", stressResult.WorstCaseTotalCost)
```

---

## Files Created

1. `internal/cost/variable.go` - Variable cost engine with distributions
2. `internal/cost/variable_test.go` - Comprehensive tests
3. `internal/simulation/scale.go` - Exponential scale simulation
4. `internal/simulation/scale_test.go` - Scale simulation tests
5. `internal/simulation/montecarlo.go` - Monte Carlo p95/p99 engine
6. `internal/simulation/montecarlo_test.go` - Monte Carlo tests

---

## Requirements Met

✅ **Normal distribution with mean/stddev**
- Implemented in `NewDistributionGenerator()`
- Uses `rand.NormFloat64()`

✅ **Uniform distribution with min/max**
- Implemented in `NewDistributionGenerator()`
- Uses `rand.Float64()`

✅ **Exponential distribution with rate**
- Implemented in `NewDistributionGenerator()`
- Uses inverse CDF transformation

✅ **Monte Carlo with 10,000 iterations for p95/p99**
- Default iterations: 10,000
- Configurable via `MonteCarloConfig`
- Percentile calculation validates p95 < p99

✅ **Use math/rand for random number generation**
- All random operations use `math/rand` package
- No external dependencies

✅ **Return the implementation code**
- All source code provided
- Complete with tests
- Production-ready
