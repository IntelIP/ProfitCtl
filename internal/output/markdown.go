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
	if result.Revenue.Mode != "" {
		markdown += fmt.Sprintf("| Pricing Mode | %s |\n", result.Revenue.Mode)
	}

	if result.Margin.GrossMargin != 0 {
		if hasOperatingView(result) {
			markdown += fmt.Sprintf("| Booked Margin | %.1f%% |\n", result.Margin.GrossMargin)
			markdown += fmt.Sprintf("| Operating Margin | %.1f%% |\n", result.OperatingMargin.GrossMargin)
		} else {
			markdown += fmt.Sprintf("| Margin | %.1f%% |\n", result.Margin.GrossMargin)
		}
	}

	// Stress test margins (if available)
	if result.StressResult.P95CostPerUser > 0 && result.Revenue.Total > 0 {
		if !hasOperatingView(result) {
			markdown += fmt.Sprintf("| p95 Margin | %.1f%% |\n", bookedStressMargin(result, result.StressResult.P95CostPerUser))
		}
	}

	if result.Margin.CostPerUser > 0 {
		if hasOperatingView(result) {
			markdown += fmt.Sprintf("| Booked Cost per User | $%.2f |\n", result.Margin.CostPerUser)
			markdown += fmt.Sprintf("| Operating Cost per User | $%.2f |\n", result.OperatingMargin.CostPerUser)
		} else {
			markdown += fmt.Sprintf("| Cost per User | $%.2f |\n", result.Margin.CostPerUser)
		}
	}

	// Covenant status
	covenantStatus := "✅ PASSED"
	if !result.Covenants.Passed {
		covenantStatus = "❌ FAILED"
	}
	markdown += fmt.Sprintf("| Covenant Status | %s |\n", covenantStatus)
	markdown += "\n"

	if result.Revenue.Mode == "mix" && len(result.Revenue.ByPlan) > 0 {
		markdown += "### Revenue Mix\n\n"
		markdown += "| Plan | Share | Users | Price | Revenue |\n"
		markdown += "|------|-------|-------|-------|---------|\n"
		for _, plan := range result.Revenue.ByPlan {
			shareDisplay := "-"
			if plan.Share != nil {
				shareDisplay = fmt.Sprintf("%.1f%%", *plan.Share*100)
			}
			markdown += fmt.Sprintf("| %s | %s | %d | $%.2f | $%.2f |\n",
				plan.PlanName, shareDisplay, plan.Users, plan.Price, plan.Revenue)
		}
		markdown += "\n"
	}

	if result.Revenue.Mode == "hybrid" && result.Revenue.Total > 0 {
		markdown += "### Hybrid Revenue\n\n"
		markdown += "| Component | Amount |\n"
		markdown += "|-----------|--------|\n"
		markdown += fmt.Sprintf("| Total Revenue | $%.2f |\n", result.Revenue.Total)
		markdown += fmt.Sprintf("| Recurring Revenue | $%.2f |\n", result.Revenue.RecurringTotal)
		if result.Revenue.OneTimeTotal > 0 {
			markdown += fmt.Sprintf("| One-Time Revenue | $%.2f |\n", result.Revenue.OneTimeTotal)
		}
		if result.Revenue.MinimumUplift > 0 {
			markdown += fmt.Sprintf("| Minimum Uplift | $%.2f |\n", result.Revenue.MinimumUplift)
		}
		for _, component := range result.Revenue.Components {
			markdown += fmt.Sprintf("| %s | $%.2f |\n", component.Name, component.Amount)
		}
		markdown += "\n"
	}

	if result.PaymentFees.Total > 0 {
		markdown += "### Payment Fees\n\n"
		markdown += "| Component | Amount |\n"
		markdown += "|-----------|--------|\n"
		markdown += fmt.Sprintf("| Total Payment Fees | $%.2f |\n", result.PaymentFees.Total)
		markdown += fmt.Sprintf("| Monthly Fees | $%.2f |\n", result.PaymentFees.MonthlyAmount)
		if result.PaymentFees.AnnualAmortizedAmount > 0 {
			markdown += fmt.Sprintf("| Annual Amortized Fees | $%.2f |\n", result.PaymentFees.AnnualAmortizedAmount)
		}
		if result.PaymentFees.OperatingAmount > 0 && result.PaymentFees.OperatingAmount != result.PaymentFees.Total {
			markdown += fmt.Sprintf("| Operating Fees | $%.2f |\n", result.PaymentFees.OperatingAmount)
		}
		if result.PaymentFees.OneTimeAmount > 0 {
			markdown += fmt.Sprintf("| One-Time Fees | $%.2f |\n", result.PaymentFees.OneTimeAmount)
		}
		markdown += fmt.Sprintf("| Revenue Percentage Fee | $%.2f |\n", result.PaymentFees.PercentageAmount)
		if result.PaymentFees.FixedAmount > 0 {
			markdown += fmt.Sprintf("| Fixed Transaction Fees | $%.2f |\n", result.PaymentFees.FixedAmount)
		}
		if result.PaymentFees.FreeUserAmount > 0 {
			markdown += fmt.Sprintf("| Free User Fees | $%.2f |\n", result.PaymentFees.FreeUserAmount)
		}
		if result.PaymentFees.PaidUserAmount > 0 {
			markdown += fmt.Sprintf("| Paid User Fees | $%.2f |\n", result.PaymentFees.PaidUserAmount)
		}
		markdown += "\n"
	}

	if result.Calibration.Period != "" || result.Calibration.Source != "" {
		markdown += "### Calibration\n\n"
		markdown += "| Metric | Actual | Modeled | Delta |\n"
		markdown += "|--------|--------|---------|-------|\n"
		markdown += fmt.Sprintf("| Revenue | $%.2f | $%.2f | $%.2f |\n",
			result.Calibration.RevenueActual, result.Calibration.RevenueModeled, result.Calibration.RevenueDelta)
		markdown += fmt.Sprintf("| Payment Fees | $%.2f | $%.2f | $%.2f |\n",
			result.Calibration.PaymentFeesActual, result.Calibration.PaymentFeesModeled, result.Calibration.PaymentFeesDelta)
		markdown += fmt.Sprintf("| Free Users | %d | %.2f | %.2f |\n",
			result.Calibration.FreeUsersActual, result.Calibration.FreeUsersModeled, result.Calibration.FreeUsersDelta)
		markdown += fmt.Sprintf("| Paid Monthly Users | %d | %.2f | %.2f |\n",
			result.Calibration.PaidMonthlyActual, result.Calibration.PaidMonthlyModeled, result.Calibration.PaidMonthlyDelta)
		markdown += fmt.Sprintf("| Paid Annual Users | %d | %.2f | %.2f |\n",
			result.Calibration.PaidAnnualActual, result.Calibration.PaidAnnualModeled, result.Calibration.PaidAnnualDelta)
		for plan, delta := range result.Calibration.PlanMixDeltas {
			markdown += fmt.Sprintf("| Plan Mix %s | - | - | %.2f |\n", plan, delta)
		}
		markdown += "\n"
	}

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
		if hasOperatingView(result) {
			markdown += fmt.Sprintf("| P95 Booked Margin | %.2f%% |\n", bookedStressMargin(result, result.StressResult.P95CostPerUser))
			markdown += fmt.Sprintf("| P95 Operating Margin | %.2f%% |\n", operatingStressMargin(result, result.StressResult.P95CostPerUser))
			markdown += fmt.Sprintf("| P99 Booked Margin | %.2f%% |\n", bookedStressMargin(result, result.StressResult.P99CostPerUser))
			markdown += fmt.Sprintf("| P99 Operating Margin | %.2f%% |\n", operatingStressMargin(result, result.StressResult.P99CostPerUser))
		}
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
