package pricing

import (
	"testing"

	"github.com/profitctl/profitctl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCalculateMargins_ZeroRevenue(t *testing.T) {
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 500,
		Application:    300,
		Service:        200,
	}

	result := CalculateMargins(0, 1000, layerCosts, 100)

	assert.Equal(t, 0.0, result.GrossMargin, "Gross margin should be zero for zero revenue")
	assert.Equal(t, 0.0, result.MarginPercentage, "Margin percentage should be zero")
	assert.Equal(t, 10.0, result.CostPerUser, "Cost per user should be calculated")
}

func TestCalculateMargins_ZeroCosts(t *testing.T) {
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 0,
		Application:    0,
		Service:        0,
	}

	result := CalculateMargins(1000, 0, layerCosts, 100)

	assert.Equal(t, 100.0, result.GrossMargin, "Gross margin should be 100% for zero costs")
	assert.Equal(t, 1.0, result.MarginPercentage, "Margin percentage should be 1.0")
	assert.Equal(t, 0.0, result.CostPerUser, "Cost per user should be zero")
}

func TestCalculateMargins_ZeroUsers(t *testing.T) {
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 500,
		Application:    300,
		Service:        200,
	}

	result := CalculateMargins(1000, 1000, layerCosts, 0)

	assert.Equal(t, 0.0, result.CostPerUser, "Cost per user should be zero for zero users")
}

func TestCalculateMargins_BasicCalculation(t *testing.T) {
	// Revenue: $1000, Costs: $800, Users: 100
	// Margin: (1000 - 800) / 1000 * 100 = 20%
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 400,
		Application:    300,
		Service:        100,
	}

	result := CalculateMargins(1000, 800, layerCosts, 100)

	assert.Equal(t, 20.0, result.GrossMargin, "Gross margin should be 20%")
	assert.InDelta(t, 0.2, result.MarginPercentage, 0.001, "Margin percentage should be 0.2")
	assert.Equal(t, 8.0, result.CostPerUser, "Cost per user should be $8")
}

func TestCalculateMargins_NegativeMargin(t *testing.T) {
	// Revenue: $500, Costs: $800, Users: 100
	// Margin: (500 - 800) / 500 * 100 = -60%
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 400,
		Application:    300,
		Service:        100,
	}

	result := CalculateMargins(500, 800, layerCosts, 100)

	assert.Equal(t, -60.0, result.GrossMargin, "Gross margin should be -60%")
	assert.InDelta(t, -0.6, result.MarginPercentage, 0.001, "Margin percentage should be -0.6")
	assert.Equal(t, 8.0, result.CostPerUser, "Cost per user should be $8")
}

func TestCalculateMargins_LayerCostPercentages(t *testing.T) {
	// Total costs: $1000
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 500, // 50%
		Application:    300, // 30%
		Service:        200, // 20%
	}

	result := CalculateMargins(1000, 1000, layerCosts, 100)

	assert.Equal(t, 50.0, result.LayerCostPercent.Infrastructure, "Infrastructure should be 50%")
	assert.Equal(t, 30.0, result.LayerCostPercent.Application, "Application should be 30%")
	assert.Equal(t, 20.0, result.LayerCostPercent.Service, "Service should be 20%")
}

func TestCalculateMargins_LayerCostPercentages_ZeroTotal(t *testing.T) {
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 0,
		Application:    0,
		Service:        0,
	}

	result := CalculateMargins(1000, 0, layerCosts, 100)

	assert.Equal(t, 0.0, result.LayerCostPercent.Infrastructure, "Infrastructure percentage should be zero")
	assert.Equal(t, 0.0, result.LayerCostPercent.Application, "Application percentage should be zero")
	assert.Equal(t, 0.0, result.LayerCostPercent.Service, "Service percentage should be zero")
}

func TestCalculateMargins_LayerMargins(t *testing.T) {
	// Revenue: $1000
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 500, // Margin: $500
		Application:    300, // Margin: $700
		Service:        200, // Margin: $800
	}

	result := CalculateMargins(1000, 1000, layerCosts, 100)

	assert.Equal(t, 500.0, result.LayerMargins.Infrastructure, "Infrastructure margin contribution")
	assert.Equal(t, 700.0, result.LayerMargins.Application, "Application margin contribution")
	assert.Equal(t, 800.0, result.LayerMargins.Service, "Service margin contribution")
}

func TestCalculateMargins_LayerMargins_ZeroRevenue(t *testing.T) {
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 500,
		Application:    300,
		Service:        200,
	}

	result := CalculateMargins(0, 1000, layerCosts, 100)

	assert.Equal(t, 0.0, result.LayerMargins.Infrastructure, "Layer margins should be zero for zero revenue")
	assert.Equal(t, 0.0, result.LayerMargins.Application, "Layer margins should be zero for zero revenue")
	assert.Equal(t, 0.0, result.LayerMargins.Service, "Layer margins should be zero for zero revenue")
}

func TestCalculateMargins_100PercentMargin(t *testing.T) {
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 0,
		Application:    0,
		Service:        0,
	}

	result := CalculateMargins(1000, 0, layerCosts, 100)

	assert.Equal(t, 100.0, result.GrossMargin, "Gross margin should be 100%")
	assert.Equal(t, 1.0, result.MarginPercentage, "Margin percentage should be 1.0")
}

func TestCalculateMargins_PrecisionRounding(t *testing.T) {
	// Revenue: $1000, Costs: $666.666, Users: 100
	// Margin: (1000 - 666.666) / 1000 * 100 = 33.3334%
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 333.333,
		Application:    222.222,
		Service:        111.111,
	}

	result := CalculateMargins(1000, 666.666, layerCosts, 100)

	// Should round to 2 decimal places
	assert.Equal(t, 33.33, result.GrossMargin, "Gross margin should be rounded to 2 decimals")
	assert.Equal(t, 6.67, result.CostPerUser, "Cost per user should be rounded to 2 decimals")
}

func TestCalculateMargins_AllLayersEqual(t *testing.T) {
	layerCosts := types.CostLayerBreakdown{
		Infrastructure: 333.33,
		Application:    333.33,
		Service:        333.34, // Rounding difference
	}

	result := CalculateMargins(1000, 1000, layerCosts, 100)

	// All layers should be approximately 33.33%
	assert.InDelta(t, 33.33, result.LayerCostPercent.Infrastructure, 0.1, "Infrastructure should be ~33.33%")
	assert.InDelta(t, 33.33, result.LayerCostPercent.Application, 0.1, "Application should be ~33.33%")
	assert.InDelta(t, 33.34, result.LayerCostPercent.Service, 0.1, "Service should be ~33.34%")
}