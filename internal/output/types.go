package output

import (
	"github.com/IntelIP/ProfitCtl/internal/cost"
	"github.com/IntelIP/ProfitCtl/internal/covenant"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/IntelIP/ProfitCtl/internal/simulation"
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
	Revenue     pricing.RevenueResult
	PaymentFees pricing.PaymentFeeResult
	Calibration pricing.CalibrationResult

	// Margins
	Margin pricing.MarginResult

	// Scale simulation
	ScaleResult simulation.ScaleSimulationResult

	// Stress test (Monte Carlo)
	StressResult simulation.StressTestResult

	// Covenant validation
	Covenants covenant.ValidationResult
}
