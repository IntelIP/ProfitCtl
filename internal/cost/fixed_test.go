package cost

import (
	"testing"

	"github.com/IntelIP/ProfitCtl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCalculateMonthlyAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		period   types.CostPeriod
		expected float64
	}{
		{"monthly to monthly", 1000, types.PeriodMonthly, 1000},
		{"yearly to monthly", 12000, types.PeriodYearly, 1000},
		{"daily to monthly", 100, types.PeriodDaily, 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := types.FixedCost{
				Amount: tt.amount,
				Period: tt.period,
			}

			result := CalculateMonthlyAmount(cost)
			assert.Equal(t, tt.expected, result, "Monthly amount calculation")
		})
	}
}

func TestCalculateFixedCosts(t *testing.T) {
	tests := []struct {
		name     string
		costs    []types.FixedCost
		months   int
		expected float64
	}{
		{"monthly to monthly", []types.FixedCost{
			{Name: "Server", Amount: 1000, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
		}, 6, 6000},
		{"yearly to monthly", []types.FixedCost{
			{Name: "Server", Amount: 12000, Period: types.PeriodYearly, Layer: types.LayerInfrastructure},
		}, 6, 6000},
		{"daily to monthly", []types.FixedCost{
			{Name: "Server", Amount: 100, Period: types.PeriodDaily, Layer: types.LayerInfrastructure},
		}, 6, 18000},
		{"mixed periods", []types.FixedCost{
			{Name: "Server", Amount: 12000, Period: types.PeriodYearly, Layer: types.LayerInfrastructure},
			{Name: "Database", Amount: 600, Period: types.PeriodMonthly, Layer: types.LayerApplication},
		}, 6, 9600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateFixedCosts(tt.costs, tt.months)
			assert.Equal(t, tt.expected, result.Total, "Fixed costs calculation")
		})
	}
}

func TestFixedCostLayerAggregation(t *testing.T) {
	costs := []types.FixedCost{
		{Name: "Server", Amount: 500, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
		{Name: "Database", Amount: 700, Period: types.PeriodMonthly, Layer: types.LayerApplication},
		{Name: "Auth0", Amount: 300, Period: types.PeriodMonthly, Layer: types.LayerService},
	}

	result := CalculateFixedCosts(costs, 1)

	assert.Equal(t, 1500.0, result.Total, "Total fixed cost")
	assert.Equal(t, 500.0, result.ByLayer.Infrastructure, "Infrastructure layer cost")
	assert.Equal(t, 700.0, result.ByLayer.Application, "Application layer cost")
	assert.Equal(t, 300.0, result.ByLayer.Service, "Service layer cost")
}

func TestFixedCostEconomicsLayerAggregation(t *testing.T) {
	costs := []types.FixedCost{
		{Name: "Runtime Infra", Amount: 500, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
		{Name: "AI Dev Tooling", Amount: 300, Period: types.PeriodMonthly, Layer: types.LayerService, EconomicsLayer: types.EconomicsLayerProductization},
		{Name: "Design Partner Support", Amount: 200, Period: types.PeriodMonthly, Layer: types.LayerApplication, EconomicsLayer: types.EconomicsLayerAdoption},
	}

	result := CalculateFixedCosts(costs, 1)

	assert.Equal(t, 500.0, result.ByEconomicsLayer.Delivery, "Delivery economics layer cost")
	assert.Equal(t, 300.0, result.ByEconomicsLayer.Productization, "Productization economics layer cost")
	assert.Equal(t, 200.0, result.ByEconomicsLayer.Adoption, "Adoption economics layer cost")
}
