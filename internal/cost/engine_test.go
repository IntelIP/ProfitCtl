package cost

import (
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewCostEngine(t *testing.T) {
	cfg := &config.Config{
		FixedCosts: []types.FixedCost{
			{Name: "Server", Amount: 500, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
		},
		VariableCosts: []types.VariableCost{
			{
				Name:         "API Calls",
				CostPerUnit:  0.0001,
				UnitsPerUser: 10000,
				Distribution: types.DistNormal,
				Mean:         newFloat64Ptr(10000),
				StdDev:       newFloat64Ptr(2000),
				Layer:        types.LayerApplication,
			},
		},
	}

	engine := NewCostEngine(cfg)
	assert.NotNil(t, engine, "Engine should not be nil")
}

func TestCostEngine_CalculateFixedCosts(t *testing.T) {
	cfg := &config.Config{
		FixedCosts: []types.FixedCost{
			{Name: "Server", Amount: 500, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
			{Name: "Database", Amount: 300, Period: types.PeriodMonthly, Layer: types.LayerApplication},
		},
		VariableCosts: []types.VariableCost{},
	}

	engine := NewCostEngine(cfg)
	result := engine.CalculateFixedCosts(1000, 6)

	// 6 months * (500 + 300) = 4800
	expectedTotal := 4800.0
	expectedMonthly := 800.0

	assert.Equal(t, expectedTotal, result.Total, "Total fixed cost for 6 months")
	assert.Equal(t, expectedMonthly, result.Monthly, "Monthly fixed cost")
	assert.Equal(t, 500.0, result.ByLayer.Infrastructure, "Infrastructure layer")
	assert.Equal(t, 300.0, result.ByLayer.Application, "Application layer")
	assert.Equal(t, 0.0, result.ByLayer.Service, "Service layer")
}

func TestCostEngine_CalculateVariableCosts(t *testing.T) {
	mean := 10000.0
	stddev := 2000.0
	cfg := &config.Config{
		FixedCosts: []types.FixedCost{},
		VariableCosts: []types.VariableCost{
			{
				Name:         "API Calls",
				CostPerUnit:  0.0001,
				UnitsPerUser: 10000,
				Distribution: types.DistNormal,
				Mean:         &mean,
				StdDev:       &stddev,
				Layer:        types.LayerApplication,
			},
			{
				Name:         "Storage",
				CostPerUnit:  0.023,
				UnitsPerUser: 5,
				Distribution: types.DistUniform,
				Min:          newFloat64Ptr(2),
				Max:          newFloat64Ptr(10),
				Layer:        types.LayerInfrastructure,
			},
		},
	}

	engine := NewCostEngine(cfg)
	result := engine.CalculateVariableCosts(1000)

	expectedAPICost := 0.0001 * 10000 * 1000
	expectedStorageCost := 0.023 * 5 * 1000
	expectedTotal := expectedAPICost + expectedStorageCost

	assert.Equal(t, expectedTotal, result.Total, "Total variable cost")
	assert.InDelta(t, expectedTotal/1000.0, result.PerUser, 0.01, "Cost per user")
	assert.Equal(t, expectedAPICost, result.ByLayer.Application, "Application layer")
	assert.Equal(t, expectedStorageCost, result.ByLayer.Infrastructure, "Infrastructure layer")
	assert.Equal(t, 0.0, result.ByLayer.Service, "Service layer")
}

func TestCostEngine_CalculateTotalCosts(t *testing.T) {
	mean := 10000.0
	stddev := 2000.0
	cfg := &config.Config{
		FixedCosts: []types.FixedCost{
			{Name: "Server", Amount: 500, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
			{Name: "Database", Amount: 300, Period: types.PeriodMonthly, Layer: types.LayerApplication},
			{Name: "Auth0", Amount: 200, Period: types.PeriodMonthly, Layer: types.LayerService},
		},
		VariableCosts: []types.VariableCost{
			{
				Name:         "API Calls",
				CostPerUnit:  0.0001,
				UnitsPerUser: 10000,
				Distribution: types.DistNormal,
				Mean:         &mean,
				StdDev:       &stddev,
				Layer:        types.LayerApplication,
			},
			{
				Name:         "Storage",
				CostPerUnit:  0.023,
				UnitsPerUser: 5,
				Distribution: types.DistUniform,
				Min:          newFloat64Ptr(2),
				Max:          newFloat64Ptr(10),
				Layer:        types.LayerInfrastructure,
			},
		},
	}

	engine := NewCostEngine(cfg)
	users := 1000
	months := 6

	result := engine.CalculateTotalCosts(users, months)

	// Fixed costs: 6 months * (500 + 300 + 200) = 6000
	expectedFixedTotal := 6000.0
	// Variable costs: (0.0001 * 10000 * 1000) + (0.023 * 5 * 1000) = 1000 + 115 = 1115
	expectedVariableTotal := 1115.0
	expectedGrandTotal := expectedFixedTotal + expectedVariableTotal

	assert.Equal(t, expectedFixedTotal, result.FixedCosts.Total, "Fixed costs total")
	assert.Equal(t, expectedVariableTotal, result.VariableCosts.Total, "Variable costs total")
	assert.Equal(t, expectedGrandTotal, result.GrandTotal, "Grand total cost")

	// Verify layer aggregation
	// Fixed: Infrastructure=500, Application=300, Service=200
	// Variable: Infrastructure=115, Application=1000, Service=0
	expectedInfra := 500.0 + 115.0
	expectedApp := 300.0 + 1000.0
	expectedService := 200.0 + 0.0

	assert.Equal(t, expectedInfra, result.GrandByLayer.Infrastructure, "Infrastructure layer total")
	assert.Equal(t, expectedApp, result.GrandByLayer.Application, "Application layer total")
	assert.Equal(t, expectedService, result.GrandByLayer.Service, "Service layer total")
}

func TestCostEngine_CalculateTotalCosts_ZeroUsers(t *testing.T) {
	cfg := &config.Config{
		FixedCosts: []types.FixedCost{
			{Name: "Server", Amount: 500, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
		},
		VariableCosts: []types.VariableCost{
			{
				Name:         "API Calls",
				CostPerUnit:  0.0001,
				UnitsPerUser: 10000,
				Distribution: types.DistNormal,
				Mean:         newFloat64Ptr(10000),
				StdDev:       newFloat64Ptr(2000),
				Layer:        types.LayerApplication,
			},
		},
	}

	engine := NewCostEngine(cfg)
	result := engine.CalculateTotalCosts(0, 6)

	// Fixed costs should still be calculated (6 * 500 = 3000)
	// Variable costs should be zero for 0 users
	assert.Equal(t, 3000.0, result.FixedCosts.Total, "Fixed costs should be calculated even for 0 users")
	assert.Equal(t, 0.0, result.VariableCosts.Total, "Variable costs should be zero for 0 users")
	assert.Equal(t, 3000.0, result.GrandTotal, "Grand total should equal fixed costs")
}

func TestCostEngine_CalculateTotalCosts_AllLayers(t *testing.T) {
	mean := 100.0
	rate := 0.01
	cfg := &config.Config{
		FixedCosts: []types.FixedCost{
			{Name: "EC2", Amount: 100, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
			{Name: "App Code", Amount: 50, Period: types.PeriodMonthly, Layer: types.LayerApplication},
			{Name: "Auth Service", Amount: 25, Period: types.PeriodMonthly, Layer: types.LayerService},
		},
		VariableCosts: []types.VariableCost{
			{
				Name:         "Compute",
				CostPerUnit:  0.01,
				UnitsPerUser: 100,
				Distribution: types.DistNormal,
				Mean:         &mean,
				StdDev:       newFloat64Ptr(10),
				Layer:        types.LayerInfrastructure,
			},
			{
				Name:         "API",
				CostPerUnit:  0.005,
				UnitsPerUser: 50,
				Distribution: types.DistUniform,
				Min:          newFloat64Ptr(40),
				Max:          newFloat64Ptr(60),
				Layer:        types.LayerApplication,
			},
			{
				Name:         "Tokens",
				CostPerUnit:  0.001,
				UnitsPerUser: 10,
				Distribution: types.DistExponential,
				Rate:         &rate,
				Layer:        types.LayerService,
			},
		},
	}

	engine := NewCostEngine(cfg)
	result := engine.CalculateTotalCosts(100, 1) // 100 users, 1 month

	// Verify all layers have costs
	assert.Greater(t, result.GrandByLayer.Infrastructure, 0.0, "Infrastructure layer should have costs")
	assert.Greater(t, result.GrandByLayer.Application, 0.0, "Application layer should have costs")
	assert.Greater(t, result.GrandByLayer.Service, 0.0, "Service layer should have costs")

	// Verify grand total matches sum of all layers
	expectedGrandTotal := result.GrandByLayer.Infrastructure +
		result.GrandByLayer.Application +
		result.GrandByLayer.Service
	assert.InDelta(t, expectedGrandTotal, result.GrandTotal, 0.01, "Grand total should match sum of layers")
}

func TestCostEngine_EmptyConfig(t *testing.T) {
	cfg := &config.Config{
		FixedCosts:    []types.FixedCost{},
		VariableCosts: []types.VariableCost{},
	}

	engine := NewCostEngine(cfg)

	fixedResult := engine.CalculateFixedCosts(1000, 6)
	assert.Equal(t, 0.0, fixedResult.Total, "Fixed costs should be zero for empty config")

	variableResult := engine.CalculateVariableCosts(1000)
	assert.Equal(t, 0.0, variableResult.Total, "Variable costs should be zero for empty config")

	totalResult := engine.CalculateTotalCosts(1000, 6)
	assert.Equal(t, 0.0, totalResult.GrandTotal, "Grand total should be zero for empty config")
}

func newFloat64Ptr(v float64) *float64 {
	return &v
}
