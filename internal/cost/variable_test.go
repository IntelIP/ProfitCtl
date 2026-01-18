package cost

import (
	"math"
	"testing"

	"github.com/profitctl/profitctl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewDistributionGeneratorNormal(t *testing.T) {
	mean := 100.0
	stddev := 15.0
	vc := types.VariableCost{
		Name:         "Test Cost",
		CostPerUnit:  0.01,
		UnitsPerUser: 1000,
		Distribution: types.DistNormal,
		Mean:         &mean,
		StdDev:       &stddev,
		Layer:        types.LayerApplication,
	}

	gen, err := NewDistributionGenerator(vc)
	assert.NoError(t, err, "Should create normal distribution generator")
	assert.NotNil(t, gen, "Generator should not be nil")

	samples := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		samples[i] = gen()
	}

	sampleMean := calculateMean(samples)
	sampleStdDev := calculateStdDev(samples)

	assert.InDelta(t, mean, sampleMean, stddev, "Sample mean should be within 1 stddev of target")
	assert.InDelta(t, stddev, sampleStdDev, 5.0, "Sample stddev should be close to target")
}

func TestNewDistributionGeneratorUniform(t *testing.T) {
	min := 50.0
	max := 150.0
	vc := types.VariableCost{
		Name:         "Test Cost",
		CostPerUnit:  0.01,
		UnitsPerUser: 1000,
		Distribution: types.DistUniform,
		Min:          &min,
		Max:          &max,
		Layer:        types.LayerApplication,
	}

	gen, err := NewDistributionGenerator(vc)
	assert.NoError(t, err, "Should create uniform distribution generator")
	assert.NotNil(t, gen, "Generator should not be nil")

	samples := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		samples[i] = gen()
	}

	minSample := samples[0]
	maxSample := samples[0]
	for _, s := range samples {
		if s < minSample {
			minSample = s
		}
		if s > maxSample {
			maxSample = s
		}
	}

	assert.GreaterOrEqual(t, minSample, min, "Sample min should be >= distribution min")
	assert.LessOrEqual(t, maxSample, max, "Sample max should be <= distribution max")
}

func TestNewDistributionGeneratorExponential(t *testing.T) {
	rate := 0.01
	vc := types.VariableCost{
		Name:         "Test Cost",
		CostPerUnit:  0.01,
		UnitsPerUser: 1000,
		Distribution: types.DistExponential,
		Rate:         &rate,
		Layer:        types.LayerApplication,
	}

	gen, err := NewDistributionGenerator(vc)
	assert.NoError(t, err, "Should create exponential distribution generator")
	assert.NotNil(t, gen, "Generator should not be nil")

	samples := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		samples[i] = gen()
	}

	assert.Greater(t, samples[0], 0.0, "Exponential samples should be positive")

	for _, s := range samples {
		assert.Greater(t, s, 0.0, "All exponential samples should be positive")
	}
}

func TestNewDistributionGeneratorErrors(t *testing.T) {
	tests := []struct {
		name    string
		vc      types.VariableCost
		wantErr bool
	}{
		{
			name: "Normal without mean",
			vc: types.VariableCost{
				Name:         "Test",
				Distribution: types.DistNormal,
				StdDev:       newFloat64(10.0),
			},
			wantErr: true,
		},
		{
			name: "Normal without stddev",
			vc: types.VariableCost{
				Name:         "Test",
				Distribution: types.DistNormal,
				Mean:         newFloat64(100.0),
			},
			wantErr: true,
		},
		{
			name: "Uniform without min",
			vc: types.VariableCost{
				Name:         "Test",
				Distribution: types.DistUniform,
				Max:          newFloat64(100.0),
			},
			wantErr: true,
		},
		{
			name: "Uniform without max",
			vc: types.VariableCost{
				Name:         "Test",
				Distribution: types.DistUniform,
				Min:          newFloat64(10.0),
			},
			wantErr: true,
		},
		{
			name: "Exponential without rate",
			vc: types.VariableCost{
				Name:         "Test",
				Distribution: types.DistExponential,
			},
			wantErr: true,
		},
		{
			name: "Unknown distribution",
			vc: types.VariableCost{
				Name:         "Test",
				Distribution: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDistributionGenerator(tt.vc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateVariableCosts(t *testing.T) {
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

	users := 1000

	result := CalculateVariableCosts(variableCosts, users)

	expectedAPICost := 0.0001 * 10000 * float64(users)
	expectedStorageCost := 0.023 * 5 * float64(users)
	expectedTotal := expectedAPICost + expectedStorageCost

	assert.Equal(t, expectedTotal, result.Total, "Total variable cost")
	assert.InDelta(t, expectedTotal/float64(users), result.PerUser, 0.01, "Cost per user")
	assert.Equal(t, expectedAPICost, result.ByLayer.Application, "Application layer cost")
	assert.Equal(t, expectedStorageCost, result.ByLayer.Infrastructure, "Infrastructure layer cost")
	assert.Len(t, result.ByCost, 2, "Should have 2 cost line items")
}

func TestCalculateVariableCostsZeroUsers(t *testing.T) {
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
	}

	result := CalculateVariableCosts(variableCosts, 0)

	assert.Equal(t, 0.0, result.Total, "Total should be zero for zero users")
	assert.Equal(t, 0.0, result.PerUser, "Per user cost should be zero for zero users")
	assert.Empty(t, result.ByCost, "Should have no cost line items")
}

func TestCalculateVariableCostWithVariability(t *testing.T) {
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

	result := CalculateVariableCostWithVariability(variableCosts, users)

	assert.Greater(t, result.Total, 0.0, "Total should be positive")
	assert.Greater(t, result.PerUser, 0.0, "Per user cost should be positive")
	assert.Len(t, result.ByCost, 1, "Should have 1 cost line item")
	assert.Len(t, result.Samples, 1, "Should have 1 sample")
	assert.InDelta(t, result.Total, result.Samples[0], 0.01, "Sample should equal total cost within tolerance")
}

func TestVariableCostLayerAggregation(t *testing.T) {
	variableCosts := []types.VariableCost{
		{
			Name:         "Compute",
			CostPerUnit:  0.01,
			UnitsPerUser: 1000,
			Distribution: types.DistUniform,
			Min:          newFloat64(900),
			Max:          newFloat64(1100),
			Layer:        types.LayerInfrastructure,
		},
		{
			Name:         "Database",
			CostPerUnit:  0.005,
			UnitsPerUser: 500,
			Distribution: types.DistNormal,
			Mean:         newFloat64(500),
			StdDev:       newFloat64(50),
			Layer:        types.LayerApplication,
		},
		{
			Name:         "Auth",
			CostPerUnit:  0.001,
			UnitsPerUser: 100,
			Distribution: types.DistExponential,
			Rate:         newFloat64(0.01),
			Layer:        types.LayerService,
		},
	}

	users := 1000

	result := CalculateVariableCosts(variableCosts, users)

	expectedInfra := 0.01 * 1000 * float64(users)
	expectedApp := 0.005 * 500 * float64(users)
	expectedService := 0.001 * 100 * float64(users)

	assert.Equal(t, expectedInfra, result.ByLayer.Infrastructure, "Infrastructure layer")
	assert.Equal(t, expectedApp, result.ByLayer.Application, "Application layer")
	assert.Equal(t, expectedService, result.ByLayer.Service, "Service layer")
}

func newFloat64(v float64) *float64 {
	return &v
}

func calculateMean(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	sum := 0.0
	for _, s := range samples {
		sum += s
	}
	return sum / float64(len(samples))
}

func calculateStdDev(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	mean := calculateMean(samples)
	sumSquared := 0.0
	for _, s := range samples {
		diff := s - mean
		sumSquared += diff * diff
	}
	return math.Sqrt(sumSquared / float64(len(samples)))
}
