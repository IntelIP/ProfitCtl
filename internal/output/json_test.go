package output

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/covenant"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/stretchr/testify/assert"
)

func TestFormatJSONResult_ValidJSON(t *testing.T) {
	result := createMockSimulationResult()

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err, "Should produce valid JSON")

	// Try to unmarshal to verify it's valid JSON
	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err, "JSON should be parseable")
}

func TestFormatJSONResult_AllFields(t *testing.T) {
	result := createMockSimulationResult()

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	// Verify scenario
	assert.Equal(t, 1000, jsonResult.Scenario.Users)
	assert.Equal(t, 12, jsonResult.Scenario.Months)
	assert.Equal(t, 1.5, jsonResult.Scenario.GrowthFactor)

	// Verify costs
	assert.Equal(t, 6000.0, jsonResult.Costs.Fixed.Total)
	assert.Equal(t, 500.0, jsonResult.Costs.Fixed.Monthly)
	assert.Equal(t, 200.0, jsonResult.Costs.Fixed.ByLayer.Infrastructure)
	assert.Equal(t, 200.0, jsonResult.Costs.Fixed.ByLayer.Application)
	assert.Equal(t, 100.0, jsonResult.Costs.Fixed.ByLayer.Service)

	assert.Equal(t, 1000.0, jsonResult.Costs.Variable.Total)
	assert.Equal(t, 1.0, jsonResult.Costs.Variable.PerUser)
	assert.Equal(t, 400.0, jsonResult.Costs.Variable.ByLayer.Infrastructure)
	assert.Equal(t, 400.0, jsonResult.Costs.Variable.ByLayer.Application)
	assert.Equal(t, 200.0, jsonResult.Costs.Variable.ByLayer.Service)

	assert.Equal(t, 7000.0, jsonResult.Costs.Total)
	assert.Equal(t, 6200.0, jsonResult.FullEconomics.Costs.Delivery)
	assert.Equal(t, 500.0, jsonResult.FullEconomics.Costs.Productization)
	assert.Equal(t, 300.0, jsonResult.FullEconomics.Costs.Adoption)
	assert.Equal(t, 7000.0, jsonResult.FullEconomics.Costs.Total)
	assert.Equal(t, 6.2, jsonResult.FullEconomics.Costs.DeliveryPerUser)
	assert.Equal(t, 7.0, jsonResult.FullEconomics.Costs.TotalPerUser)

	// Verify revenue
	assert.Equal(t, "tiered", jsonResult.Revenue.Mode)
	assert.Equal(t, 10000.0, jsonResult.Revenue.Total)
	assert.Len(t, jsonResult.Revenue.ByPlan, 1)
	assert.Equal(t, "basic", jsonResult.Revenue.ByPlan[0].PlanName)
	assert.Equal(t, 10.0, jsonResult.Revenue.ByPlan[0].Price)
	assert.Equal(t, 1000, jsonResult.Revenue.ByPlan[0].Users)
	assert.Equal(t, 10000.0, jsonResult.Revenue.ByPlan[0].Revenue)

	// Verify margin
	assert.Equal(t, 30.0, jsonResult.Margin.Gross)
	assert.Equal(t, 7.0, jsonResult.Margin.CostPerUser)
	assert.Equal(t, 38.0, jsonResult.FullEconomics.Margin.Delivery)
	assert.Equal(t, 30.0, jsonResult.FullEconomics.Margin.Full)

	// Verify stress test
	assert.Equal(t, 7.0, jsonResult.StressTest.Mean.CostPerUser)
	assert.Equal(t, 8.5, jsonResult.StressTest.P95.CostPerUser)
	assert.Equal(t, 9.0, jsonResult.StressTest.P99.CostPerUser)
	assert.Equal(t, 9000.0, jsonResult.StressTest.WorstCaseTotalCost)

	// Verify covenants
	assert.True(t, jsonResult.Covenants.Passed)
	assert.Empty(t, jsonResult.Covenants.Violations)
}

func TestFormatJSONResult_CovenantViolations(t *testing.T) {
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

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	assert.False(t, jsonResult.Covenants.Passed)
	assert.Len(t, jsonResult.Covenants.Violations, 2)
	assert.Equal(t, "margin", jsonResult.Covenants.Violations[0].Field)
	assert.Equal(t, "cost_per_user", jsonResult.Covenants.Violations[1].Field)
	assert.Equal(t, "Gross margin must be >= 20%", jsonResult.Covenants.Violations[0].Message)
}

func TestFormatJSONResult_NoRevenue(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.Total = 0
	result.Revenue.ByPlan = []pricing.PlanRevenue{}

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	// Revenue should be 0, but JSON should still be valid
	assert.Equal(t, 0.0, jsonResult.Revenue.Total)
	assert.Empty(t, jsonResult.Revenue.ByPlan)
}

func TestFormatJSONResult_NoMargin(t *testing.T) {
	result := createMockSimulationResult()
	result.Margin.GrossMargin = 0
	result.Margin.CostPerUser = 0

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	// Should handle zero margin gracefully
	assert.Equal(t, 0.0, jsonResult.Margin.Gross)
	assert.Equal(t, 0.0, jsonResult.Margin.CostPerUser)
}

func TestFormatJSONResult_MultiplePlans(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.ByPlan = []pricing.PlanRevenue{
		{PlanName: "basic", Price: 10, Users: 1000, Revenue: 10000},
		{PlanName: "pro", Price: 29, Users: 500, Revenue: 14500},
	}

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	assert.Len(t, jsonResult.Revenue.ByPlan, 2)
	assert.Equal(t, "basic", jsonResult.Revenue.ByPlan[0].PlanName)
	assert.Equal(t, "pro", jsonResult.Revenue.ByPlan[1].PlanName)
}

func TestFormatJSONResult_P95P99Margins(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.Total = 10000
	result.StressResult.P95CostPerUser = 8.5
	result.StressResult.P99CostPerUser = 9.0
	result.Users = 1000

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	// P95 margin = (10000 - 8500) / 10000 * 100 = 15%
	p95TotalCost := 8.5 * 1000
	expectedP95Margin := ((10000.0 - p95TotalCost) / 10000.0) * 100
	assert.InDelta(t, expectedP95Margin, jsonResult.StressTest.P95.Margin, 0.01)

	// P99 margin = (10000 - 9000) / 10000 * 100 = 10%
	p99TotalCost := 9.0 * 1000
	expectedP99Margin := ((10000.0 - p99TotalCost) / 10000.0) * 100
	assert.InDelta(t, expectedP99Margin, jsonResult.StressTest.P99.Margin, 0.01)
}

func TestFormatJSONResult_MixModeIncludesShares(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.Mode = "mix"
	result.Revenue.ByPlan = []pricing.PlanRevenue{
		{PlanName: "free", Price: 0, Share: floatPtr(0.7), Users: 700, Revenue: 0},
		{PlanName: "pro", Price: 29, Share: floatPtr(0.3), Users: 300, Revenue: 8700},
	}
	result.Revenue.Total = 8700

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	assert.Equal(t, "mix", jsonResult.Revenue.Mode)
	assert.Len(t, jsonResult.Revenue.ByPlan, 2)
	assert.NotNil(t, jsonResult.Revenue.ByPlan[0].Share)
	assert.Equal(t, 0.7, *jsonResult.Revenue.ByPlan[0].Share)
	assert.NotNil(t, jsonResult.Revenue.ByPlan[1].Share)
	assert.Equal(t, 0.3, *jsonResult.Revenue.ByPlan[1].Share)
}

func TestFormatJSONResult_HybridModeIncludesBreakdown(t *testing.T) {
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

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	assert.Equal(t, "hybrid", jsonResult.Revenue.Mode)
	assert.Equal(t, 6500.0, jsonResult.Revenue.Total)
	assert.Equal(t, 2000.0, jsonResult.Revenue.RecurringTotal)
	assert.Equal(t, 5000.0, jsonResult.Revenue.OneTimeTotal)
	assert.Equal(t, 250.0, jsonResult.Revenue.MinimumUplift)
	assert.Len(t, jsonResult.Revenue.Components, 2)
	assert.Equal(t, "base_platform_fee", jsonResult.Revenue.Components[0].Name)
	assert.Equal(t, 68.85, jsonResult.Margin.OperatingGross)
	assert.Equal(t, 12.46, jsonResult.Margin.OperatingCostPerUser)
}

func TestFormatJSONResult_PaymentFees(t *testing.T) {
	result := createMockSimulationResult()
	result.PaymentFees = pricing.PaymentFeeResult{
		Processor:             "stripe",
		Currency:              "usd",
		MonthlyAmount:         29,
		AnnualAmortizedAmount: 3,
		OperatingAmount:       20,
		OneTimeAmount:         12,
		FreeUserAmount:        0,
		PaidUserAmount:        32,
		FreeUsers:             700,
		PaidMonthlyUsers:      225,
		PaidAnnualUsers:       75,
		PercentageAmount:      29,
		FixedAmount:           3,
		Total:                 32,
	}

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	assert.Equal(t, "stripe", jsonResult.PaymentFees.Processor)
	assert.Equal(t, "usd", jsonResult.PaymentFees.Currency)
	assert.Equal(t, 29.0, jsonResult.PaymentFees.Monthly)
	assert.Equal(t, 3.0, jsonResult.PaymentFees.AnnualAmortized)
	assert.Equal(t, 225.0, jsonResult.PaymentFees.PaidMonthlyUsers)
	assert.Equal(t, 75.0, jsonResult.PaymentFees.PaidAnnualUsers)
	assert.Equal(t, 29.0, jsonResult.PaymentFees.PercentageAmount)
	assert.Equal(t, 3.0, jsonResult.PaymentFees.FixedAmount)
	assert.Equal(t, 20.0, jsonResult.PaymentFees.OperatingTotal)
	assert.Equal(t, 12.0, jsonResult.PaymentFees.OneTimeAmount)
	assert.Equal(t, 32.0, jsonResult.PaymentFees.Total)
}

func TestFormatJSONResult_Calibration(t *testing.T) {
	result := createMockSimulationResult()
	result.Calibration = pricing.CalibrationResult{
		Period:             "2026-03",
		Source:             "stripe_export",
		RevenueActual:      1000,
		RevenueModeled:     1100,
		RevenueDelta:       100,
		PaymentFeesActual:  30,
		PaymentFeesModeled: 32,
		PaymentFeesDelta:   2,
		FreeUsersActual:    700,
		FreeUsersModeled:   680,
		FreeUsersDelta:     -20,
		PaidMonthlyActual:  225,
		PaidMonthlyModeled: 220,
		PaidMonthlyDelta:   -5,
		PaidAnnualActual:   75,
		PaidAnnualModeled:  80,
		PaidAnnualDelta:    5,
		PlanMixDeltas: map[string]float64{
			"pro": 0.05,
		},
	}

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(jsonBytes, &jsonResult)
	assert.NoError(t, err)

	assert.Equal(t, "2026-03", jsonResult.Calibration.Period)
	assert.Equal(t, "stripe_export", jsonResult.Calibration.Source)
	assert.Equal(t, 100.0, jsonResult.Calibration.RevenueDelta)
	assert.Equal(t, 2.0, jsonResult.Calibration.PaymentFeesDelta)
	assert.Equal(t, -20.0, jsonResult.Calibration.FreeUsersDelta)
	assert.Equal(t, 0.05, jsonResult.Calibration.PlanMixDeltas["pro"])
}

func TestFormatJSONResult_Indentation(t *testing.T) {
	result := createMockSimulationResult()

	jsonBytes, err := FormatJSONResult(result)
	assert.NoError(t, err)

	// Verify indentation (should have 2 spaces per level)
	jsonStr := string(jsonBytes)

	// Should contain newlines and spaces for indentation
	assert.Contains(t, jsonStr, "\n")
	assert.Contains(t, jsonStr, "  \"scenario\"") // 2-space indent
}

func TestPrintJSONResult(t *testing.T) {
	result := createMockSimulationResult()

	err := PrintJSONResult(result)
	assert.NoError(t, err)
}

func TestWriteJSONResult(t *testing.T) {
	result := createMockSimulationResult()

	tmpFile := "/tmp/profitctl_test.json"
	defer func() {
		// Cleanup
		_ = os.Remove(tmpFile)
	}()

	err := WriteJSONResult(result, tmpFile)
	assert.NoError(t, err)

	// Read file back and verify it's valid JSON
	data, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)

	var jsonResult JSONResult
	err = json.Unmarshal(data, &jsonResult)
	assert.NoError(t, err)
	assert.Equal(t, 1000, jsonResult.Scenario.Users)
}
