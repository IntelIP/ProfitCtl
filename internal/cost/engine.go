package cost

import (
	"math"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/IntelIP/ProfitCtl/pkg/types"
)

// CostEngine provides a unified interface for calculating fixed and variable costs
type CostEngine interface {
	CalculateFixedCosts(users int, months int) FixedCostResult
	CalculateVariableCosts(users int) VariableCostResult
	CalculateTotalCosts(users int, months int) TotalCostResult
}

// costEngineImpl implements the CostEngine interface
type costEngineImpl struct {
	fixedCosts    []types.FixedCost
	variableCosts []types.VariableCost
	pricing       *config.PricingConfig
}

// NewCostEngine creates a new CostEngine from a config
func NewCostEngine(cfg *config.Config) CostEngine {
	return &costEngineImpl{
		fixedCosts:    cfg.FixedCosts,
		variableCosts: cfg.VariableCosts,
		pricing:       cfg.Pricing,
	}
}

// CalculateFixedCosts calculates fixed costs for the given number of users and months
func (e *costEngineImpl) CalculateFixedCosts(users int, months int) FixedCostResult {
	return CalculateFixedCosts(e.fixedCosts, months)
}

// CalculateVariableCosts calculates variable costs for the given number of users
func (e *costEngineImpl) CalculateVariableCosts(users int) VariableCostResult {
	paidUsers := users
	if e.pricing != nil {
		paidUsers = pricing.CalculateRevenue(e.pricing, users).PaidUsers
	}

	return CalculateVariableCostsWithScope(e.variableCosts, users, paidUsers)
}

// TotalCostResult represents the combined fixed and variable costs
type TotalCostResult struct {
	FixedCosts            FixedCostResult
	VariableCosts         VariableCostResult
	GrandTotal            float64
	GrandByLayer          types.CostLayerBreakdown
	GrandByEconomicsLayer types.EconomicsLayerBreakdown
}

// CalculateTotalCosts calculates total costs (fixed + variable) for the given number of users and months
func (e *costEngineImpl) CalculateTotalCosts(users int, months int) TotalCostResult {
	fixedResult := e.CalculateFixedCosts(users, months)
	variableResult := e.CalculateVariableCosts(users)

	grandTotal := fixedResult.Total + variableResult.Total
	grandByLayer := types.CostLayerBreakdown{
		Infrastructure: fixedResult.ByLayer.Infrastructure + variableResult.ByLayer.Infrastructure,
		Application:    fixedResult.ByLayer.Application + variableResult.ByLayer.Application,
		Service:        fixedResult.ByLayer.Service + variableResult.ByLayer.Service,
	}
	grandByEconomicsLayer := types.EconomicsLayerBreakdown{
		Delivery:       fixedResult.ByEconomicsLayer.Delivery + variableResult.ByEconomicsLayer.Delivery,
		Productization: fixedResult.ByEconomicsLayer.Productization + variableResult.ByEconomicsLayer.Productization,
		Adoption:       fixedResult.ByEconomicsLayer.Adoption + variableResult.ByEconomicsLayer.Adoption,
	}

	// Round all values to 2 decimal places for financial precision
	grandTotal = math.Round(grandTotal*100) / 100
	grandByLayer.Infrastructure = math.Round(grandByLayer.Infrastructure*100) / 100
	grandByLayer.Application = math.Round(grandByLayer.Application*100) / 100
	grandByLayer.Service = math.Round(grandByLayer.Service*100) / 100
	grandByEconomicsLayer.Round()

	return TotalCostResult{
		FixedCosts:            fixedResult,
		VariableCosts:         variableResult,
		GrandTotal:            grandTotal,
		GrandByLayer:          grandByLayer,
		GrandByEconomicsLayer: grandByEconomicsLayer,
	}
}
