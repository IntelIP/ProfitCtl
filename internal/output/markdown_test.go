package output

import (
	"os"
	"strings"
	"testing"

	"github.com/profitctl/profitctl/internal/covenant"
	"github.com/profitctl/profitctl/internal/pricing"
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