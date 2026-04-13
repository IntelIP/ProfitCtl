package output

import (
	"bytes"
	"os"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/cost"
	"github.com/IntelIP/ProfitCtl/internal/covenant"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/IntelIP/ProfitCtl/internal/simulation"
	"github.com/IntelIP/ProfitCtl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func createMockSimulationResult() SimulationResult {
	return SimulationResult{
		Users:        1000,
		Months:       12,
		GrowthFactor: 1.5,
		FixedCosts: cost.FixedCostResult{
			Total:   6000,
			Monthly: 500,
			ByLayer: types.CostLayerBreakdown{
				Infrastructure: 200,
				Application:    200,
				Service:        100,
			},
		},
		VariableCosts: cost.VariableCostResult{
			Total:   1000,
			PerUser: 1.0,
			ByLayer: types.CostLayerBreakdown{
				Infrastructure: 400,
				Application:    400,
				Service:        200,
			},
		},
		TotalCosts: cost.TotalCostResult{
			GrandTotal: 7000,
			GrandByLayer: types.CostLayerBreakdown{
				Infrastructure: 600,
				Application:    600,
				Service:        300,
			},
		},
		FullEconomics: FullEconomicsSummary{
			DeliveryCost:        6200,
			ProductizationCost:  500,
			AdoptionCost:        300,
			TotalCost:           7000,
			DeliveryCostPerUser: 6.2,
			TotalCostPerUser:    7.0,
			DeliveryMargin:      38.0,
			TotalMargin:         30.0,
		},
		Revenue: pricing.RevenueResult{
			Mode:           "tiered",
			Total:          10000,
			RecurringTotal: 10000,
			ByPlan: []pricing.PlanRevenue{
				{PlanName: "basic", Price: 10, Users: 1000, Revenue: 10000},
			},
		},
		Margin: pricing.MarginResult{
			GrossMargin: 30.0,
			CostPerUser: 7.0,
			LayerMargins: pricing.LayerMarginBreakdown{
				Infrastructure: 9400,
				Application:    9400,
				Service:        9700,
			},
		},
		OperatingMargin: pricing.MarginResult{
			GrossMargin: 30.0,
			CostPerUser: 7.0,
			LayerMargins: pricing.LayerMarginBreakdown{
				Infrastructure: 9400,
				Application:    9400,
				Service:        9700,
			},
		},
		ScaleResult: simulation.ScaleSimulationResult{
			Scenarios: []simulation.ScaleScenario{
				{Users: 1000, Step: 0, TotalCosts: simulation.TotalCosts{GrandTotal: 7000}},
			},
		},
		StressResult: simulation.StressTestResult{
			Users:              1000,
			Months:             12,
			MeanCostPerUser:    7.0,
			P95CostPerUser:     8.5,
			P99CostPerUser:     9.0,
			WorstCaseTotalCost: 9000,
			MonteCarlo: simulation.MonteCarloResult{
				Iterations: 10000,
				P95:        8500,
			},
		},
		Covenants: covenant.ValidationResult{
			Passed:     true,
			Violations: []covenant.Violation{},
		},
	}
}

func TestFormatCLIResult_AllSections(t *testing.T) {
	result := createMockSimulationResult()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify all sections are present
	assert.Contains(t, output, "=== profitctl Simulation Results ===")
	assert.Contains(t, output, "Scenario: 1000 users")
	assert.Contains(t, output, "Mean margin: 30.00%")
	assert.Contains(t, output, "Fixed COGS: $500.00/month")
	assert.Contains(t, output, "Infrastructure: $200.00")
	assert.Contains(t, output, "Application: $200.00")
	assert.Contains(t, output, "Service: $100.00")
	assert.Contains(t, output, "Variable COGS: $1.0000/user")
	assert.Contains(t, output, "Cost per user: $7.00")
	assert.Contains(t, output, "Full Economics:")
	assert.Contains(t, output, "Delivery cost: $6200.00")
	assert.Contains(t, output, "Full margin: 30.00%")
	assert.Contains(t, output, "✅ PASSED")
}

func TestFormatCLIResult_CovenantViolations(t *testing.T) {
	result := createMockSimulationResult()
	result.Covenants.Passed = false
	result.Covenants.Violations = []covenant.Violation{
		{
			Field:    "margin",
			Operator: "gte",
			Value:    20,
			Actual:   15,
			Message:  "Gross margin must be >= 20%",
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "❌ FAILED")
	assert.Contains(t, output, "Violations:")
	assert.Contains(t, output, "Gross margin must be >= 20%")
	assert.Contains(t, output, "Expected: >= 20.00, Actual: 15.00")
}

func TestFormatCLIResult_VerboseMode(t *testing.T) {
	result := createMockSimulationResult()
	result.ScaleResult.Scenarios = []simulation.ScaleScenario{
		{Users: 100, Step: 0, TotalCosts: simulation.TotalCosts{GrandTotal: 700}},
		{Users: 150, Step: 1, TotalCosts: simulation.TotalCosts{GrandTotal: 1050}},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, true) // verbose mode

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Scale Simulation:")
	assert.Contains(t, output, "Step 0: 100 users")
	assert.Contains(t, output, "Step 1: 150 users")
	assert.Contains(t, output, "Stress Test Details (10000 iterations):")
	assert.Contains(t, output, "Mean: $7.00")
	assert.Contains(t, output, "P95: $8.50")
	assert.Contains(t, output, "P99: $9.00")
}

func TestFormatCLIResult_NoVerboseMode(t *testing.T) {
	result := createMockSimulationResult()
	result.ScaleResult.Scenarios = []simulation.ScaleScenario{
		{Users: 100, Step: 0, TotalCosts: simulation.TotalCosts{GrandTotal: 700}},
		{Users: 150, Step: 1, TotalCosts: simulation.TotalCosts{GrandTotal: 1050}},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false) // not verbose

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should not contain verbose details
	assert.NotContains(t, output, "Scale Simulation:")
	assert.NotContains(t, output, "Stress Test Details")
}

func TestFormatCLIResult_ZeroRevenue(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.Total = 0
	result.Revenue.ByPlan = []pricing.PlanRevenue{}
	result.Margin.GrossMargin = 0

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should not crash and should handle zero revenue gracefully
	assert.Contains(t, output, "=== profitctl Simulation Results ===")
	assert.Contains(t, output, "Fixed COGS: $500.00/month")
}

func TestFormatCLIResult_NoP95Margin(t *testing.T) {
	result := createMockSimulationResult()
	result.StressResult.MonteCarlo.P95 = 0
	result.StressResult.P95CostPerUser = 0

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should not crash when p95 is not available
	assert.Contains(t, output, "=== profitctl Simulation Results ===")
	assert.NotContains(t, output, "p95 margin:")
}

func TestFormatCLIResult_ZeroVariableCosts(t *testing.T) {
	result := createMockSimulationResult()
	result.VariableCosts.Total = 0
	result.VariableCosts.PerUser = 0
	result.VariableCosts.ByLayer = types.CostLayerBreakdown{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Variable COGS: $0.00/user")
}

func TestFormatCLIResult_MixModeVerbose(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.Mode = "mix"
	result.Revenue.ByPlan = []pricing.PlanRevenue{
		{PlanName: "free", Price: 0, Share: floatPtr(0.7), Users: 700, Revenue: 0},
		{PlanName: "pro", Price: 29, Share: floatPtr(0.3), Users: 300, Revenue: 8700},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, true)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Pricing mode: mix")
	assert.Contains(t, output, "Revenue Mix:")
	assert.Contains(t, output, "free: 70.0% -> 700 users")
	assert.Contains(t, output, "pro: 30.0% -> 300 users")
}

func TestFormatCLIResult_HybridRevenueBreakdown(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue = pricing.RevenueResult{
		Mode:           "hybrid",
		Total:          6500,
		RecurringTotal: 2000,
		OneTimeTotal:   5000,
		MinimumUplift:  250,
		Components: []pricing.RevenueComponent{
			{Name: "base_platform_fee", Amount: 500},
			{Name: "pilot_setup_fee", Amount: 5000},
		},
	}
	result.OperatingMargin = pricing.MarginResult{GrossMargin: 68.85, CostPerUser: 12.46}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Pricing mode: hybrid")
	assert.Contains(t, output, "Booked margin: 30.00%")
	assert.Contains(t, output, "Operating margin: 68.85%")
	assert.Contains(t, output, "Booked cost per user: $7.00")
	assert.Contains(t, output, "Operating cost per user: $12.46")
	assert.Contains(t, output, "Revenue: $6500.00/month")
	assert.Contains(t, output, "Recurring: $2000.00")
	assert.Contains(t, output, "One-time: $5000.00")
	assert.Contains(t, output, "Minimum uplift: $250.00")
	assert.Contains(t, output, "base_platform_fee: $500.00")
}

func TestFormatCLIResult_PaymentFees(t *testing.T) {
	result := createMockSimulationResult()
	result.PaymentFees = pricing.PaymentFeeResult{
		MonthlyAmount:         29,
		AnnualAmortizedAmount: 3,
		OperatingAmount:       20,
		OneTimeAmount:         12,
		PercentageAmount:      29,
		FixedAmount:           3,
		PaidUserAmount:        32,
		Total:                 32,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Payment Fees: $32.00")
	assert.Contains(t, output, "Monthly fees: $29.00")
	assert.Contains(t, output, "Annual amortized fees: $3.00")
	assert.Contains(t, output, "Operating fees: $20.00")
	assert.Contains(t, output, "One-time fees: $12.00")
	assert.Contains(t, output, "Revenue percentage fee: $29.00")
	assert.Contains(t, output, "Fixed transaction fees: $3.00")
	assert.Contains(t, output, "Paid user fees: $32.00")
}

func TestFormatCLIResult_Calibration(t *testing.T) {
	result := createMockSimulationResult()
	result.Calibration = pricing.CalibrationResult{
		Period:           "2026-03",
		Source:           "stripe_export",
		RevenueDelta:     100,
		PaymentFeesDelta: 2,
		FreeUsersDelta:   -20,
		PaidMonthlyDelta: -5,
		PaidAnnualDelta:  5,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Calibration:")
	assert.Contains(t, output, "Period: 2026-03")
	assert.Contains(t, output, "Source: stripe_export")
	assert.Contains(t, output, "Revenue delta: $100.00")
	assert.Contains(t, output, "Payment fee delta: $2.00")
}

func floatPtr(v float64) *float64 {
	return &v
}

func TestFormatCLIResult_MultipleViolations(t *testing.T) {
	result := createMockSimulationResult()
	result.Covenants.Passed = false
	result.Covenants.Violations = []covenant.Violation{
		{
			Field:    "margin",
			Operator: "gte",
			Value:    20,
			Actual:   15,
			Message:  "Gross margin must be >= 20%",
		},
		{
			Field:    "cost_per_user",
			Operator: "lte",
			Value:    5,
			Actual:   7,
			Message:  "Cost per user must be <= $5",
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	FormatCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "❌ FAILED")
	// Should show both violations
	assert.Contains(t, output, "Gross margin must be >= 20%")
	assert.Contains(t, output, "Cost per user must be <= $5")
}

func TestPrintCLIResult(t *testing.T) {
	result := createMockSimulationResult()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintCLIResult(result, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should produce the same output as FormatCLIResult
	assert.Contains(t, output, "=== profitctl Simulation Results ===")
}

func TestPrintCLIResultToFile(t *testing.T) {
	result := createMockSimulationResult()

	// Create temporary file
	tmpFile := "/tmp/profitctl_test_output.txt"
	defer os.Remove(tmpFile)

	err := PrintCLIResultToFile(result, false, tmpFile)
	assert.NoError(t, err)

	// Read file back
	data, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)

	output := string(data)
	assert.Contains(t, output, "=== profitctl Simulation Results ===")
	assert.Contains(t, output, "Scenario: 1000 users")
	assert.Contains(t, output, "Cost per user: $7.00")
}
func TestFormatCLIQuietResult(t *testing.T) {
	result := createMockSimulationResult()
	quiet := FormatCLIQuietResult(result)
	assert.Contains(t, quiet, "PASSED")
	assert.Contains(t, quiet, "margin=")
	assert.Contains(t, quiet, "violations=0")

	result.Covenants.Passed = false
	result.Covenants.Violations = []covenant.Violation{{Message: "breach"}}
	quiet = FormatCLIQuietResult(result)
	assert.Contains(t, quiet, "FAILED")
	assert.Contains(t, quiet, "violations=1")
}
