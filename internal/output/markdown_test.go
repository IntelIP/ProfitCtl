package output

import (
	"os"
	"strings"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/covenant"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/stretchr/testify/assert"
)

func TestFormatMarkdownResult_BasicStructure(t *testing.T) {
	result := createMockSimulationResult()

	markdown := FormatMarkdownResult(result)

	// Verify header
	assert.Contains(t, markdown, "## profitctl Results")

	// Verify table structure
	assert.Contains(t, markdown, "| Metric | Value |")
	assert.Contains(t, markdown, "|--------|-------|")

	// Verify basic metrics
	assert.Contains(t, markdown, "| Users | 1000 |")
	assert.Contains(t, markdown, "| Pricing Mode | tiered |")
	assert.Contains(t, markdown, "| Margin | 30.0% |")
	assert.Contains(t, markdown, "| Cost per User | $7.00 |")
	assert.Contains(t, markdown, "| Covenant Status | ✅ PASSED |")
}

func TestFormatMarkdownResult_CovenantViolations(t *testing.T) {
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

	markdown := FormatMarkdownResult(result)

	assert.Contains(t, markdown, "| Covenant Status | ❌ FAILED |")
	assert.Contains(t, markdown, "### Covenant Violations")
	assert.Contains(t, markdown, "- **margin**: Gross margin must be >= 20%")
	assert.Contains(t, markdown, "Expected: >= 20.00, Actual: 15.00")
}

func TestFormatMarkdownResult_CostBreakdown(t *testing.T) {
	result := createMockSimulationResult()

	markdown := FormatMarkdownResult(result)

	// Verify cost breakdown section
	assert.Contains(t, markdown, "### Cost Breakdown")
	assert.Contains(t, markdown, "| Layer | Fixed | Variable |")
	assert.Contains(t, markdown, "|-------|-------|----------|")
	assert.Contains(t, markdown, "| Infrastructure | $200.00 | $400.00 |")
	assert.Contains(t, markdown, "| Application | $200.00 | $400.00 |")
	assert.Contains(t, markdown, "| Service | $100.00 | $200.00 |")
}

func TestFormatMarkdownResult_StressTest(t *testing.T) {
	result := createMockSimulationResult()

	markdown := FormatMarkdownResult(result)

	// Verify stress test section
	assert.Contains(t, markdown, "### Stress Test Results")
	assert.Contains(t, markdown, "| Mean Cost per User | $7.00 |")
	assert.Contains(t, markdown, "| P95 Cost per User | $8.50 |")
	assert.Contains(t, markdown, "| P99 Cost per User | $9.00 |")
	assert.Contains(t, markdown, "| Worst Case Total Cost | $9000.00 |")
}

func TestFormatMarkdownResult_NoStressTest(t *testing.T) {
	result := createMockSimulationResult()
	result.StressResult.P95CostPerUser = 0

	markdown := FormatMarkdownResult(result)

	// Should not contain stress test section if P95 is zero
	assert.NotContains(t, markdown, "### Stress Test Results")
}

func TestFormatMarkdownResult_P95Margin(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.Total = 10000
	result.StressResult.P95CostPerUser = 8.5
	result.Users = 1000

	markdown := FormatMarkdownResult(result)

	// P95 margin should be calculated: (10000 - 8500) / 10000 * 100 = 15%
	assert.Contains(t, markdown, "| p95 Margin | 15.0% |")
}

func TestFormatMarkdownResult_NoRevenue(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.Total = 0
	result.Margin.GrossMargin = 0

	markdown := FormatMarkdownResult(result)

	// Should not crash with zero revenue
	assert.Contains(t, markdown, "## profitctl Results")
	// Should not contain margin if zero
	assert.NotContains(t, markdown, "| Margin |")
}

func TestFormatMarkdownResult_MultipleViolations(t *testing.T) {
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

	markdown := FormatMarkdownResult(result)

	// Should contain both violations
	assert.Contains(t, markdown, "- **margin**: Gross margin must be >= 20%")
	assert.Contains(t, markdown, "- **cost_per_user**: Cost per user must be <= $5")
}

func TestFormatMarkdownResult_ValidMarkdownFormat(t *testing.T) {
	result := createMockSimulationResult()

	markdown := FormatMarkdownResult(result)

	// Verify table syntax is valid markdown
	lines := strings.Split(markdown, "\n")

	// Check for table header rows
	hasTableHeader := false
	hasTableSeparator := false
	for _, line := range lines {
		if strings.Contains(line, "|") && strings.Contains(line, "Metric") {
			hasTableHeader = true
		}
		if strings.Contains(line, "|") && strings.Contains(line, "---") {
			hasTableSeparator = true
		}
	}

	assert.True(t, hasTableHeader, "Should have table header")
	assert.True(t, hasTableSeparator, "Should have table separator row")
}

func TestPrintMarkdownResult(t *testing.T) {
	result := createMockSimulationResult()

	// Test that PrintMarkdownResult doesn't crash
	PrintMarkdownResult(result)
	// No assertion needed - just verify no panic
}

func TestWriteMarkdownResult(t *testing.T) {
	result := createMockSimulationResult()

	tmpFile := "/tmp/profitctl_test.md"
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	err := WriteMarkdownResult(result, tmpFile)
	assert.NoError(t, err)

	// Read file back
	data, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)

	markdown := string(data)
	assert.Contains(t, markdown, "## profitctl Results")
	assert.Contains(t, markdown, "| Users | 1000 |")
}

func TestFormatMarkdownResult_EmptyPlans(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.ByPlan = []pricing.PlanRevenue{}

	markdown := FormatMarkdownResult(result)

	// Should handle empty plans gracefully
	assert.Contains(t, markdown, "## profitctl Results")
}

func TestFormatMarkdownResult_NoP95MarginWithoutRevenue(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.Total = 0
	result.StressResult.P95CostPerUser = 8.5

	markdown := FormatMarkdownResult(result)

	// Should not show p95 margin if revenue is zero
	assert.NotContains(t, markdown, "| p95 Margin |")
}

func TestFormatMarkdownResult_MixModeRevenueTable(t *testing.T) {
	result := createMockSimulationResult()
	result.Revenue.Mode = "mix"
	result.Revenue.ByPlan = []pricing.PlanRevenue{
		{PlanName: "free", Price: 0, Share: floatPtr(0.7), Users: 700, Revenue: 0},
		{PlanName: "pro", Price: 29, Share: floatPtr(0.3), Users: 300, Revenue: 8700},
	}
	result.Revenue.Total = 8700

	markdown := FormatMarkdownResult(result)

	assert.Contains(t, markdown, "### Revenue Mix")
	assert.Contains(t, markdown, "| free | 70.0% | 700 | $0.00 | $0.00 |")
	assert.Contains(t, markdown, "| pro | 30.0% | 300 | $29.00 | $8700.00 |")
}

func TestFormatMarkdownResult_HybridRevenueTable(t *testing.T) {
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

	markdown := FormatMarkdownResult(result)

	assert.Contains(t, markdown, "| Booked Margin | 30.0% |")
	assert.Contains(t, markdown, "| Operating Margin | 68.8% |")
	assert.Contains(t, markdown, "| Booked Cost per User | $7.00 |")
	assert.Contains(t, markdown, "| Operating Cost per User | $12.46 |")
	assert.Contains(t, markdown, "### Hybrid Revenue")
	assert.Contains(t, markdown, "| Total Revenue | $6500.00 |")
	assert.Contains(t, markdown, "| Recurring Revenue | $2000.00 |")
	assert.Contains(t, markdown, "| One-Time Revenue | $5000.00 |")
	assert.Contains(t, markdown, "| Minimum Uplift | $250.00 |")
	assert.Contains(t, markdown, "| base_platform_fee | $500.00 |")
}

func TestFormatMarkdownResult_PaymentFees(t *testing.T) {
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

	markdown := FormatMarkdownResult(result)

	assert.Contains(t, markdown, "### Payment Fees")
	assert.Contains(t, markdown, "| Total Payment Fees | $32.00 |")
	assert.Contains(t, markdown, "| Monthly Fees | $29.00 |")
	assert.Contains(t, markdown, "| Annual Amortized Fees | $3.00 |")
	assert.Contains(t, markdown, "| Operating Fees | $20.00 |")
	assert.Contains(t, markdown, "| One-Time Fees | $12.00 |")
	assert.Contains(t, markdown, "| Revenue Percentage Fee | $29.00 |")
	assert.Contains(t, markdown, "| Fixed Transaction Fees | $3.00 |")
	assert.Contains(t, markdown, "| Paid User Fees | $32.00 |")
}

func TestFormatMarkdownResult_Calibration(t *testing.T) {
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

	markdown := FormatMarkdownResult(result)

	assert.Contains(t, markdown, "### Calibration")
	assert.Contains(t, markdown, "| Revenue | $1000.00 | $1100.00 | $100.00 |")
	assert.Contains(t, markdown, "| Payment Fees | $30.00 | $32.00 | $2.00 |")
	assert.Contains(t, markdown, "| Plan Mix pro | - | - | 0.05 |")
}
