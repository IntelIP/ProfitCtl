package simulation

import (
	"testing"

	"github.com/IntelIP/ProfitCtl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMonteCarloConfig(t *testing.T) {
	config := NewMonteCarloConfig(10000)

	assert.Equal(t, 10000, config.Iterations)
	assert.Greater(t, config.Seed, int64(0), "Seed should be positive")
}

func TestNewMonteCarloConfigDefaultIterations(t *testing.T) {
	config := NewMonteCarloConfig(0)

	assert.Equal(t, 10000, config.Iterations, "Should default to 10000 iterations")
}

func TestRunMonteCarlo(t *testing.T) {
	fixedCosts := []types.FixedCost{
		{
			Name:   "Server",
			Amount: 1000,
			Period: types.PeriodMonthly,
			Layer:  types.LayerInfrastructure,
		},
	}

	mean := 1000.0
	stddev := 100.0
	variableCosts := []types.VariableCost{
		{
			Name:         "API Calls",
			CostPerUnit:  0.01,
			UnitsPerUser: 100,
			Distribution: types.DistNormal,
			Mean:         &mean,
			StdDev:       &stddev,
			Layer:        types.LayerApplication,
		},
	}

	users := 100
	months := 1
	iterations := 1000

	result := RunMonteCarlo(fixedCosts, variableCosts, users, months, NewMonteCarloConfig(iterations))

	assert.Equal(t, iterations, result.Iterations, "Should have correct iterations")
	assert.Len(t, result.Samples, iterations, "Should have correct number of samples")
	assert.Greater(t, result.Mean, 1000.0, "Mean should include fixed costs")
	assert.Greater(t, result.P95, result.Mean, "P95 should be >= mean")
	assert.Greater(t, result.P99, result.P95, "P99 should be >= P95")
	assert.Greater(t, result.StdDev, 0.0, "StdDev should be positive")
}

func TestRunMonteCarloPercentiles(t *testing.T) {
	fixedCosts := []types.FixedCost{}
	variableCosts := []types.VariableCost{}

	users := 100
	months := 1
	iterations := 10000

	result := RunMonteCarlo(fixedCosts, variableCosts, users, months, NewMonteCarloConfig(iterations))

	assert.Equal(t, iterations, result.Iterations)

	p95Index := int(0.95 * float64(iterations))
	p99Index := int(0.99 * float64(iterations))

	assert.Equal(t, result.Samples[p95Index], result.P95, "P95 should match 95th percentile")
	assert.Equal(t, result.Samples[p99Index], result.P99, "P99 should match 99th percentile")
}

func TestCalculatePercentile(t *testing.T) {
	tests := []struct {
		name     string
		samples  []float64
		p        float64
		expected float64
	}{
		{
			name:     "95th percentile",
			samples:  []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			p:        0.95,
			expected: 10,
		},
		{
			name:     "50th percentile (median)",
			samples:  []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			p:        0.5,
			expected: 6,
		},
		{
			name:     "10th percentile",
			samples:  []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			p:        0.1,
			expected: 2,
		},
		{
			name:     "Empty samples",
			samples:  []float64{},
			p:        0.95,
			expected: 0,
		},
		{
			name:     "Single sample",
			samples:  []float64{42},
			p:        0.95,
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePercentile(tt.samples, tt.p)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateMean(t *testing.T) {
	tests := []struct {
		name     string
		samples  []float64
		expected float64
	}{
		{
			name:     "Simple average",
			samples:  []float64{1, 2, 3, 4, 5},
			expected: 3,
		},
		{
			name:     "Zero values",
			samples:  []float64{0, 0, 0},
			expected: 0,
		},
		{
			name:     "Empty samples",
			samples:  []float64{},
			expected: 0,
		},
		{
			name:     "Single value",
			samples:  []float64{42},
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMean(tt.samples)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name     string
		samples  []float64
		expected float64
	}{
		{
			name:     "Odd number of samples",
			samples:  []float64{1, 2, 3, 4, 5},
			expected: 3,
		},
		{
			name:     "Even number of samples",
			samples:  []float64{1, 2, 3, 4},
			expected: 2.5,
		},
		{
			name:     "Empty samples",
			samples:  []float64{},
			expected: 0,
		},
		{
			name:     "Single sample",
			samples:  []float64{42},
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMedian(tt.samples)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateStdDev(t *testing.T) {
	tests := []struct {
		name     string
		samples  []float64
		expected float64
	}{
		{
			name:     "No variance",
			samples:  []float64{5, 5, 5, 5, 5},
			expected: 0,
		},
		{
			name:     "Some variance",
			samples:  []float64{1, 2, 3, 4, 5},
			expected: 1.414,
		},
		{
			name:     "Empty samples",
			samples:  []float64{},
			expected: 0,
		},
		{
			name:     "Single sample",
			samples:  []float64{42},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateStdDev(tt.samples)
			assert.InDelta(t, tt.expected, result, 0.01, "Standard deviation should match")
		})
	}
}

func TestRunStressTest(t *testing.T) {
	fixedCosts := []types.FixedCost{
		{
			Name:   "Server",
			Amount: 5000,
			Period: types.PeriodMonthly,
			Layer:  types.LayerInfrastructure,
		},
	}

	mean := 50.0
	stddev := 10.0
	variableCosts := []types.VariableCost{
		{
			Name:         "API Calls",
			CostPerUnit:  0.001,
			UnitsPerUser: 1000,
			Distribution: types.DistNormal,
			Mean:         &mean,
			StdDev:       &stddev,
			Layer:        types.LayerApplication,
		},
	}

	users := 1000
	months := 12
	iterations := 1000

	result := RunStressTest(fixedCosts, variableCosts, users, months, iterations)

	assert.Equal(t, users, result.Users)
	assert.Equal(t, months, result.Months)
	assert.Equal(t, iterations, result.MonteCarlo.Iterations)
	assert.Greater(t, result.MeanCostPerUser, 0.0, "Mean cost per user should be positive")
	assert.GreaterOrEqual(t, result.P95CostPerUser, result.MeanCostPerUser, "P95 cost per user should be >= mean")
	assert.GreaterOrEqual(t, result.P99CostPerUser, result.P95CostPerUser, "P99 cost per user should be >= P95")
	assert.Greater(t, result.WorstCaseTotalCost, 0.0, "Worst case total cost should be positive")
}

func TestRunStressTestPerUserMetrics(t *testing.T) {
	fixedCosts := []types.FixedCost{
		{
			Name:   "Fixed Cost",
			Amount: 10000,
			Period: types.PeriodMonthly,
			Layer:  types.LayerInfrastructure,
		},
	}

	min := 80.0
	max := 120.0
	variableCosts := []types.VariableCost{
		{
			Name:         "Variable Cost",
			CostPerUnit:  0.01,
			UnitsPerUser: 100,
			Distribution: types.DistUniform,
			Min:          &min,
			Max:          &max,
			Layer:        types.LayerApplication,
		},
	}

	users := 100
	months := 1
	iterations := 500

	result := RunStressTest(fixedCosts, variableCosts, users, months, iterations)

	expectedMeanPerUser := result.MonteCarlo.Mean / float64(users)
	expectedP95PerUser := result.MonteCarlo.P95 / float64(users)
	expectedP99PerUser := result.MonteCarlo.P99 / float64(users)

	assert.InDelta(t, expectedMeanPerUser, result.MeanCostPerUser, 0.01, "Mean cost per user")
	assert.InDelta(t, expectedP95PerUser, result.P95CostPerUser, 0.01, "P95 cost per user")
	assert.InDelta(t, expectedP99PerUser, result.P99CostPerUser, 0.01, "P99 cost per user")
	assert.InDelta(t, result.MonteCarlo.P99, result.WorstCaseTotalCost, 0.01, "Worst case total cost")
}

func TestRunStressTest_ExampleStyleInputsStayWithinReasonableRange(t *testing.T) {
	mean := 10000.0
	stddev := 2000.0
	variableCosts := []types.VariableCost{
		{
			Name:         "API Calls",
			CostPerUnit:  0.0001,
			UnitsPerUser: 10000,
			Distribution: types.DistNormal,
			Mean:         &mean,
			StdDev:       &stddev,
			Layer:        types.LayerApplication,
		},
	}

	result := RunStressTest(nil, variableCosts, 50, 1, 1000)

	assert.InDelta(t, 1.0, result.MeanCostPerUser, 0.3, "mean cost per user should reflect sampled API call volume without compounding units_per_user twice")
	assert.Less(t, result.P95CostPerUser, 2.5, "p95 cost per user should remain in a plausible range for the example config")
}

func TestRoundToPrecision(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		precision int
		expected  float64
	}{
		{
			name:      "Round to 2 decimal places",
			value:     3.14159,
			precision: 2,
			expected:  3.14,
		},
		{
			name:      "Round to 1 decimal place",
			value:     2.6789,
			precision: 1,
			expected:  2.7,
		},
		{
			name:      "Round to 0 decimal places",
			value:     9.9,
			precision: 0,
			expected:  10,
		},
		{
			name:      "Round down",
			value:     1.234,
			precision: 2,
			expected:  1.23,
		},
		{
			name:      "Round up",
			value:     1.235,
			precision: 2,
			expected:  1.24,
		},
		{
			name:      "Zero value",
			value:     0,
			precision: 2,
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roundToPrecision(tt.value, tt.precision)
			assert.Equal(t, tt.expected, result)
		})
	}
}
