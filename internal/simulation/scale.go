package simulation

import (
	"math"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/internal/cost"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/IntelIP/ProfitCtl/pkg/types"
)

type ScaleSimulationResult struct {
	Scenarios []ScaleScenario
	Config    ScaleConfig
}

type ScaleScenario struct {
	Users        int
	Step         int
	FixedCosts   cost.FixedCostResult
	VariableCost cost.VariableCostResult
	TotalCosts   TotalCosts
}

type TotalCosts struct {
	Fixed            float64
	Variable         float64
	GrandTotal       float64
	ByLayer          types.CostLayerBreakdown
	ByEconomicsLayer types.EconomicsLayerBreakdown
}

type ScaleConfig struct {
	BaseUsers    int
	GrowthFactor float64
	Steps        int
}

func NewScaleConfig(cfg *config.SimulationConfig) ScaleConfig {
	return ScaleConfig{
		BaseUsers:    cfg.BaseUsers,
		GrowthFactor: cfg.GrowthFactor,
		Steps:        6,
	}
}

func RunScaleSimulation(fixedCosts []types.FixedCost, variableCosts []types.VariableCost, pricingCfg *config.PricingConfig, config ScaleConfig, months int) ScaleSimulationResult {
	result := ScaleSimulationResult{
		Scenarios: make([]ScaleScenario, 0, config.Steps),
		Config:    config,
	}

	for step := 0; step < config.Steps; step++ {
		users := calculateUsers(config.BaseUsers, config.GrowthFactor, step)

		scenario := ScaleScenario{
			Users: users,
			Step:  step,
		}

		scenario.FixedCosts = cost.CalculateFixedCosts(fixedCosts, months)
		paidUsers := users
		if pricingCfg != nil {
			paidUsers = pricing.CalculateRevenue(pricingCfg, users).PaidUsers
		}
		scenario.VariableCost = cost.CalculateVariableCostsWithScope(variableCosts, users, paidUsers)

		scenario.TotalCosts = calculateTotalCosts(scenario.FixedCosts, scenario.VariableCost)

		result.Scenarios = append(result.Scenarios, scenario)
	}

	return result
}

func calculateUsers(baseUsers int, growthFactor float64, step int) int {
	users := float64(baseUsers) * math.Pow(growthFactor, float64(step))
	return int(math.Round(users))
}

func calculateTotalCosts(fixedCosts cost.FixedCostResult, variableCost cost.VariableCostResult) TotalCosts {
	return TotalCosts{
		Fixed:      fixedCosts.Total,
		Variable:   variableCost.Total,
		GrandTotal: fixedCosts.Total + variableCost.Total,
		ByLayer: types.CostLayerBreakdown{
			Infrastructure: fixedCosts.ByLayer.Infrastructure + variableCost.ByLayer.Infrastructure,
			Application:    fixedCosts.ByLayer.Application + variableCost.ByLayer.Application,
			Service:        fixedCosts.ByLayer.Service + variableCost.ByLayer.Service,
		},
		ByEconomicsLayer: types.EconomicsLayerBreakdown{
			Delivery:       fixedCosts.ByEconomicsLayer.Delivery + variableCost.ByEconomicsLayer.Delivery,
			Productization: fixedCosts.ByEconomicsLayer.Productization + variableCost.ByEconomicsLayer.Productization,
			Adoption:       fixedCosts.ByEconomicsLayer.Adoption + variableCost.ByEconomicsLayer.Adoption,
		},
	}
}
