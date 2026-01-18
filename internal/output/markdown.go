package output

import (
	"fmt"
	"os"
)

// FormatMarkdownResult formats simulation results as markdown
func FormatMarkdownResult(result SimulationResult) string {
	var markdown string

	// Header
	markdown += "## profitctl Results\n\n"

	// Summary table
	markdown += "| Metric | Value |\n"
	markdown += "|--------|-------|\n"
	markdown += fmt.Sprintf("| Users | %d |\n", result.Users)

	if result.Margin.GrossMargin != 0 {
		markdown += fmt.Sprintf("| Margin | %.1f%% |\n", result.Margin.GrossMargin)
	}

	// Stress test margins (if available)
	if result.StressResult.P95CostPerUser > 0 && result.Revenue.Total > 0 {
		p95TotalCost := result.StressResult.P95CostPerUser * float64(result.Users)
		p95Margin := ((result.Revenue.Total - p95TotalCost) / result.Revenue.Total) * 100
		markdown += fmt.Sprintf("| p95 Margin | %.1f%% |\n", p95Margin)
	}

	if result.Margin.CostPerUser > 0 {
		markdown += fmt.Sprintf("| Cost per User | $%.2f |\n", result.Margin.CostPerUser)
	}

	// Covenant status
	covenantStatus := "✅ PASSED"
	if !result.Covenants.Passed {
		covenantStatus = "❌ FAILED"
	}
	markdown += fmt.Sprintf("| Covenant Status | %s |\n", covenantStatus)
	markdown += "\n"

	// Cost breakdown by layer
	markdown += "### Cost Breakdown\n\n"
	markdown += "| Layer | Fixed | Variable |\n"
	markdown += "|-------|-------|----------|\n"
	markdown += fmt.Sprintf("| Infrastructure | $%.2f | $%.2f |\n",
		result.FixedCosts.ByLayer.Infrastructure,
		result.VariableCosts.ByLayer.Infrastructure)
	markdown += fmt.Sprintf("| Application | $%.2f | $%.2f |\n",
		result.FixedCosts.ByLayer.Application,
		result.VariableCosts.ByLayer.Application)
	markdown += fmt.Sprintf("| Service | $%.2f | $%.2f |\n",
		result.FixedCosts.ByLayer.Service,
		result.VariableCosts.ByLayer.Service)
	markdown += "\n"

	// Covenant violations (if any)
	if !result.Covenants.Passed && len(result.Covenants.Violations) > 0 {
		markdown += "### Covenant Violations\n\n"
		for _, violation := range result.Covenants.Violations {
			markdown += fmt.Sprintf("- **%s**: %s (Expected: %s %.2f, Actual: %.2f)\n",
				violation.Field, violation.Message,
				formatOperator(violation.Operator), violation.Value, violation.Actual)
		}
		markdown += "\n"
	}

	// Stress test summary (if available)
	if result.StressResult.P95CostPerUser > 0 {
		markdown += "### Stress Test Results\n\n"
		markdown += "| Metric | Value |\n"
		markdown += "|--------|-------|\n"
		markdown += fmt.Sprintf("| Mean Cost per User | $%.2f |\n", result.StressResult.MeanCostPerUser)
		markdown += fmt.Sprintf("| P95 Cost per User | $%.2f |\n", result.StressResult.P95CostPerUser)
		markdown += fmt.Sprintf("| P99 Cost per User | $%.2f |\n", result.StressResult.P99CostPerUser)
		markdown += fmt.Sprintf("| Worst Case Total Cost | $%.2f |\n", result.StressResult.WorstCaseTotalCost)
		markdown += "\n"
	}

	return markdown
}

// PrintMarkdownResult writes markdown output to stdout
func PrintMarkdownResult(result SimulationResult) {
	markdown := FormatMarkdownResult(result)
	fmt.Print(markdown)
}

// WriteMarkdownResult writes markdown output to a file
func WriteMarkdownResult(result SimulationResult, filename string) error {
	markdown := FormatMarkdownResult(result)
	return os.WriteFile(filename, []byte(markdown), 0644)
}