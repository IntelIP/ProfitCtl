package cost

import (
	"math"

	"github.com/IntelIP/ProfitCtl/pkg/types"
)

// CalculateMonthlyAmount converts any cost period to monthly amount
func CalculateMonthlyAmount(cost types.FixedCost) float64 {
	switch cost.Period {
	case types.PeriodMonthly:
		return cost.Amount
	case types.PeriodYearly:
		return cost.Amount / 12
	case types.PeriodDaily:
		return cost.Amount * 30
	default:
		return cost.Amount
	}
}

// FixedCostResult contains fixed cost calculation results
type FixedCostResult struct {
	Total   float64
	Monthly float64
	ByLayer types.CostLayerBreakdown
}

// CalculateFixedCosts calculates total fixed costs and layer breakdown
func CalculateFixedCosts(costs []types.FixedCost, months int) FixedCostResult {
	result := FixedCostResult{
		ByLayer: types.CostLayerBreakdown{
			Infrastructure: 0,
			Application:    0,
			Service:        0,
		},
	}

	for _, cost := range costs {
		monthlyAmount := CalculateMonthlyAmount(cost)
		result.Total += monthlyAmount * float64(months)
		result.Monthly += monthlyAmount

		switch cost.Layer {
		case types.LayerInfrastructure:
			result.ByLayer.Infrastructure += monthlyAmount
		case types.LayerApplication:
			result.ByLayer.Application += monthlyAmount
		case types.LayerService:
			result.ByLayer.Service += monthlyAmount
		}
	}

	result.Total = math.Round(result.Total*100) / 100
	result.Monthly = math.Round(result.Monthly*100) / 100
	result.ByLayer.Infrastructure = math.Round(result.ByLayer.Infrastructure*100) / 100
	result.ByLayer.Application = math.Round(result.ByLayer.Application*100) / 100
	result.ByLayer.Service = math.Round(result.ByLayer.Service*100) / 100

	return result
}
