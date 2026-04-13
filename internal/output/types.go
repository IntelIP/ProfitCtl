package output

import (
	"github.com/IntelIP/ProfitCtl/internal/cost"
	"github.com/IntelIP/ProfitCtl/internal/covenant"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/IntelIP/ProfitCtl/internal/simulation"
)

type FullEconomicsSummary struct {
	DeliveryCost        float64
	ProductizationCost  float64
	AdoptionCost        float64
	TotalCost           float64
	DeliveryCostPerUser float64
	TotalCostPerUser    float64
	DeliveryMargin      float64
	TotalMargin         float64
}

// SimulationResult contains all results from a simulation run
type SimulationResult struct {
	// Scenario information
	Users         int
	BillableUsers int
	Months        int
	GrowthFactor  float64

	// Costs
	FixedCosts    cost.FixedCostResult
	VariableCosts cost.VariableCostResult
	TotalCosts    cost.TotalCostResult

	// Revenue and Pricing
	Revenue     pricing.RevenueResult
	PaymentFees pricing.PaymentFeeResult
	Calibration pricing.CalibrationResult

	// Margins
	Margin          pricing.MarginResult
	OperatingMargin pricing.MarginResult
	FullEconomics   FullEconomicsSummary

	// Scale simulation
	ScaleResult simulation.ScaleSimulationResult

	// Stress test (Monte Carlo)
	StressResult simulation.StressTestResult

	// Covenant validation
	Covenants covenant.ValidationResult
}
