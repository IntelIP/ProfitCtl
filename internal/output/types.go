package output

import (
	"github.com/profitctl/profitctl/internal/cost"
	"github.com/profitctl/profitctl/internal/covenant"
	"github.com/profitctl/profitctl/internal/pricing"
	"github.com/profitctl/profitctl/internal/simulation"
)

// SimulationResult contains all results from a simulation run
type SimulationResult struct {
	// Scenario information
	Users        int
	Months       int
	GrowthFactor float64

	// Costs
	FixedCosts    cost.FixedCostResult
	VariableCosts cost.VariableCostResult
	TotalCosts    cost.TotalCostResult

	// Revenue and Pricing
	Revenue pricing.RevenueResult

	// Margins
	Margin pricing.MarginResult

	// Scale simulation
	ScaleResult simulation.ScaleSimulationResult

	// Stress test (Monte Carlo)
	StressResult simulation.StressTestResult

	// Covenant validation
	Covenants covenant.ValidationResult
}