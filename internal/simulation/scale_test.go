package simulation

import (
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/internal/cost"
	"github.com/IntelIP/ProfitCtl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCalculateUsers(t *testing.T) {
	tests := []struct {
		name         string
		baseUsers    int
		growthFactor float64
		step         int
		expected     int
	}{
		{
			name:         "Base case",
			baseUsers:    100,
			growthFactor: 1.5,
			step:         0,
			expected:     100,
		},
		{
			name:         "One growth step",
			baseUsers:    100,
			growthFactor: 1.5,
			step:         1,
			expected:     150,
		},
		{
			name:         "Two growth steps",
			baseUsers:    100,
			growthFactor: 1.5,
			step:         2,
			expected:     225,
		},
		{
			name:         "Higher growth factor",
			baseUsers:    100,
			growthFactor: 2.0,
			step:         3,
			expected:     800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateUsers(tt.baseUsers, tt.growthFactor, tt.step)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewScaleConfig(t *testing.T) {
	cfg := &config.SimulationConfig{
		BaseUsers:    100,
		GrowthFactor: 1.5,
		Iterations:   10000,
	}

	scaleConfig := NewScaleConfig(cfg)

	assert.Equal(t, 100, scaleConfig.BaseUsers)
	assert.Equal(t, 1.5, scaleConfig.GrowthFactor)
	assert.Equal(t, 6, scaleConfig.Steps, "Default steps should be 6")
}

func TestRunScaleSimulation(t *testing.T) {
	fixedCosts := []types.FixedCost{
		{
			Name:   "Server",
			Amount: 500,
			Period: types.PeriodMonthly,
			Layer:  types.LayerInfrastructure,
		},
		{
			Name:   "Database",
			Amount: 200,
			Period: types.PeriodMonthly,
			Layer:  types.LayerApplication,
		},
	}

	variableCosts := []types.VariableCost{
		{
			Name:         "API Calls",
			CostPerUnit:  0.0001,
			UnitsPerUser: 10000,
			Distribution: types.DistUniform,
			Min:          newFloat64(8000),
			Max:          newFloat64(12000),
			Layer:        types.LayerApplication,
		},
	}

	scaleConfig := ScaleConfig{
		BaseUsers:    100,
		GrowthFactor: 1.5,
		Steps:        3,
	}

	months := 6

	result := RunScaleSimulation(fixedCosts, variableCosts, scaleConfig, months)

	assert.Len(t, result.Scenarios, 3, "Should have 3 scenarios")
	assert.Equal(t, scaleConfig, result.Config)

	expectedUsers := []int{100, 150, 225}
	for i, scenario := range result.Scenarios {
		assert.Equal(t, expectedUsers[i], scenario.Users, "Scenario %d users", i)
		assert.Equal(t, i, scenario.Step, "Scenario %d step", i)
		assert.Greater(t, scenario.FixedCosts.Total, 0.0, "Scenario %d fixed costs", i)
		assert.Greater(t, scenario.VariableCost.Total, 0.0, "Scenario %d variable costs", i)
		assert.Equal(t, scenario.FixedCosts.Total+scenario.VariableCost.Total, scenario.TotalCosts.GrandTotal, "Scenario %d grand total", i)
	}
}

func TestRunScaleSimulationExponentialGrowth(t *testing.T) {
	fixedCosts := []types.FixedCost{
		{
			Name:   "Server",
			Amount: 1000,
			Period: types.PeriodMonthly,
			Layer:  types.LayerInfrastructure,
		},
	}

	variableCosts := []types.VariableCost{
		{
			Name:         "Storage",
			CostPerUnit:  0.023,
			UnitsPerUser: 5,
			Distribution: types.DistUniform,
			Min:          newFloat64(4),
			Max:          newFloat64(6),
			Layer:        types.LayerInfrastructure,
		},
	}

	scaleConfig := ScaleConfig{
		BaseUsers:    100,
		GrowthFactor: 2.0,
		Steps:        5,
	}

	months := 1

	result := RunScaleSimulation(fixedCosts, variableCosts, scaleConfig, months)

	expectedUsers := []int{100, 200, 400, 800, 1600}
	for i, scenario := range result.Scenarios {
		assert.Equal(t, expectedUsers[i], scenario.Users, "Scenario %d users", i)

		if i > 0 {
			prevScenario := result.Scenarios[i-1]
			assert.Greater(t, scenario.TotalCosts.Variable, prevScenario.TotalCosts.Variable, "Variable costs should increase")
			expectedRatio := float64(expectedUsers[i]) / float64(expectedUsers[i-1])
			actualRatio := scenario.TotalCosts.Variable / prevScenario.TotalCosts.Variable
			assert.InDelta(t, expectedRatio, actualRatio, 0.1, "Variable costs should grow proportionally to users")
		}
	}
}

func TestCalculateTotalCosts(t *testing.T) {
	fixedResult := cost.FixedCostResult{
		Total:   1500,
		Monthly: 250,
		ByLayer: types.CostLayerBreakdown{
			Infrastructure: 500,
			Application:    700,
			Service:        300,
		},
	}

	variableResult := cost.VariableCostResult{
		Total:   500,
		PerUser: 0.5,
		ByLayer: types.CostLayerBreakdown{
			Infrastructure: 200,
			Application:    150,
			Service:        150,
		},
	}

	totalCosts := calculateTotalCosts(fixedResult, variableResult)

	assert.Equal(t, float64(1500), totalCosts.Fixed, "Fixed costs")
	assert.Equal(t, float64(500), totalCosts.Variable, "Variable costs")
	assert.Equal(t, float64(2000), totalCosts.GrandTotal, "Grand total")
	assert.Equal(t, 700.0, totalCosts.ByLayer.Infrastructure, "Infrastructure total")
	assert.Equal(t, 850.0, totalCosts.ByLayer.Application, "Application total")
	assert.Equal(t, 450.0, totalCosts.ByLayer.Service, "Service total")
}

func TestRunScaleSimulationLayerBreakdown(t *testing.T) {
	fixedCosts := []types.FixedCost{
		{
			Name:   "Infra Fixed",
			Amount: 500,
			Period: types.PeriodMonthly,
			Layer:  types.LayerInfrastructure,
		},
		{
			Name:   "App Fixed",
			Amount: 300,
			Period: types.PeriodMonthly,
			Layer:  types.LayerApplication,
		},
		{
			Name:   "Service Fixed",
			Amount: 200,
			Period: types.PeriodMonthly,
			Layer:  types.LayerService,
		},
	}

	variableCosts := []types.VariableCost{
		{
			Name:         "Infra Var",
			CostPerUnit:  0.01,
			UnitsPerUser: 100,
			Distribution: types.DistUniform,
			Min:          newFloat64(90),
			Max:          newFloat64(110),
			Layer:        types.LayerInfrastructure,
		},
		{
			Name:         "App Var",
			CostPerUnit:  0.005,
			UnitsPerUser: 200,
			Distribution: types.DistUniform,
			Min:          newFloat64(190),
			Max:          newFloat64(210),
			Layer:        types.LayerApplication,
		},
		{
			Name:         "Service Var",
			CostPerUnit:  0.002,
			UnitsPerUser: 50,
			Distribution: types.DistUniform,
			Min:          newFloat64(45),
			Max:          newFloat64(55),
			Layer:        types.LayerService,
		},
	}

	scaleConfig := ScaleConfig{
		BaseUsers:    100,
		GrowthFactor: 1.5,
		Steps:        2,
	}

	months := 1

	result := RunScaleSimulation(fixedCosts, variableCosts, scaleConfig, months)

	for _, scenario := range result.Scenarios {
		expectedInfra := 500 + 0.01*100*float64(scenario.Users)
		expectedApp := 300 + 0.005*200*float64(scenario.Users)
		expectedService := 200 + 0.002*50*float64(scenario.Users)

		assert.InDelta(t, expectedInfra, scenario.TotalCosts.ByLayer.Infrastructure, 1.0, "Infrastructure total")
		assert.InDelta(t, expectedApp, scenario.TotalCosts.ByLayer.Application, 1.0, "Application total")
		assert.InDelta(t, expectedService, scenario.TotalCosts.ByLayer.Service, 1.0, "Service total")
	}
}

func newFloat64(v float64) *float64 {
	return &v
}
