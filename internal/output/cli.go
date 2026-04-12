package output

import (
	"fmt"
	"os"
)

// FormatCLIResult formats simulation results as human-readable CLI output.
func FormatCLIResult(result SimulationResult, verbose bool) {
	fmt.Println("=== profitctl Simulation Results ===")
	fmt.Println()

	fmt.Printf("Scenario: %d users\n", result.Users)
	if result.BillableUsers > 0 && result.BillableUsers != result.Users {
		fmt.Printf("Billable seats: %d\n", result.BillableUsers)
	}
	fmt.Println("───────────────────────────────")

	if result.Revenue.Mode != "" {
		fmt.Printf("Pricing mode: %s\n", result.Revenue.Mode)
	}

	if result.Margin.GrossMargin != 0 {
		if hasOperatingView(result) {
			fmt.Printf("Booked margin: %.2f%%\n", result.Margin.GrossMargin)
			fmt.Printf("Operating margin: %.2f%%\n", result.OperatingMargin.GrossMargin)
		} else {
			fmt.Printf("Mean margin: %.2f%%\n", result.Margin.GrossMargin)
		}
	}

	if result.StressResult.MonteCarlo.P95 > 0 && result.Revenue.Total > 0 {
		if !hasOperatingView(result) {
			fmt.Printf("p95 margin: %.2f%%\n", bookedStressMargin(result, result.StressResult.P95CostPerUser))
		}
	}

	if result.StressResult.P95CostPerUser > 0 {
		fmt.Printf("Worst-case cost per user: $%.2f\n", result.StressResult.P95CostPerUser)
	}

	fmt.Println()

	fmt.Printf("Fixed COGS: $%.2f/month\n", result.FixedCosts.Monthly)
	fmt.Printf("  Infrastructure: $%.2f\n", result.FixedCosts.ByLayer.Infrastructure)
	fmt.Printf("  Application: $%.2f\n", result.FixedCosts.ByLayer.Application)
	fmt.Printf("  Service: $%.2f\n", result.FixedCosts.ByLayer.Service)
	fmt.Println()

	if result.VariableCosts.Total > 0 && result.Users > 0 {
		fmt.Printf("Variable COGS: $%.4f/user\n", result.VariableCosts.PerUser)
		if hasOperatingView(result) {
			fmt.Printf("Booked cost per user: $%.2f\n", result.Margin.CostPerUser)
			fmt.Printf("Operating cost per user: $%.2f\n", result.OperatingMargin.CostPerUser)
		} else {
			fmt.Printf("Cost per user: $%.2f\n", result.Margin.CostPerUser)
		}
	} else {
		fmt.Printf("Variable COGS: $0.00/user\n")
		if hasOperatingView(result) {
			fmt.Printf("Booked cost per user: $%.2f\n", result.Margin.CostPerUser)
			fmt.Printf("Operating cost per user: $%.2f\n", result.OperatingMargin.CostPerUser)
		} else {
			fmt.Printf("Cost per user: $%.2f\n", result.Margin.CostPerUser)
		}
	}
	fmt.Println()

	if result.Revenue.Mode == "hybrid" && result.Revenue.Total > 0 {
		fmt.Printf("Revenue: $%.2f/month\n", result.Revenue.Total)
		fmt.Printf("  Recurring: $%.2f\n", result.Revenue.RecurringTotal)
		if result.Revenue.OneTimeTotal > 0 {
			fmt.Printf("  One-time: $%.2f\n", result.Revenue.OneTimeTotal)
		}
		if result.Revenue.MinimumUplift > 0 {
			fmt.Printf("  Minimum uplift: $%.2f\n", result.Revenue.MinimumUplift)
		}
		for _, component := range result.Revenue.Components {
			fmt.Printf("  %s: $%.2f\n", component.Name, component.Amount)
		}
		fmt.Println()
	}

	if result.PaymentFees.Total > 0 {
		fmt.Printf("Payment Fees: $%.2f\n", result.PaymentFees.Total)
		fmt.Printf("  Monthly fees: $%.2f\n", result.PaymentFees.MonthlyAmount)
		if result.PaymentFees.AnnualAmortizedAmount > 0 {
			fmt.Printf("  Annual amortized fees: $%.2f\n", result.PaymentFees.AnnualAmortizedAmount)
		}
		if result.PaymentFees.OperatingAmount > 0 && result.PaymentFees.OperatingAmount != result.PaymentFees.Total {
			fmt.Printf("  Operating fees: $%.2f\n", result.PaymentFees.OperatingAmount)
		}
		if result.PaymentFees.OneTimeAmount > 0 {
			fmt.Printf("  One-time fees: $%.2f\n", result.PaymentFees.OneTimeAmount)
		}
		fmt.Printf("  Revenue percentage fee: $%.2f\n", result.PaymentFees.PercentageAmount)
		if result.PaymentFees.FixedAmount > 0 {
			fmt.Printf("  Fixed transaction fees: $%.2f\n", result.PaymentFees.FixedAmount)
		}
		if result.PaymentFees.FreeUserAmount > 0 {
			fmt.Printf("  Free user fees: $%.2f\n", result.PaymentFees.FreeUserAmount)
		}
		if result.PaymentFees.PaidUserAmount > 0 {
			fmt.Printf("  Paid user fees: $%.2f\n", result.PaymentFees.PaidUserAmount)
		}
		fmt.Println()
	}

	if result.Calibration.Period != "" || result.Calibration.Source != "" {
		fmt.Println("Calibration:")
		if result.Calibration.Period != "" {
			fmt.Printf("  Period: %s\n", result.Calibration.Period)
		}
		if result.Calibration.Source != "" {
			fmt.Printf("  Source: %s\n", result.Calibration.Source)
		}
		fmt.Printf("  Revenue delta: $%.2f\n", result.Calibration.RevenueDelta)
		fmt.Printf("  Payment fee delta: $%.2f\n", result.Calibration.PaymentFeesDelta)
		fmt.Printf("  Free user delta: %.2f\n", result.Calibration.FreeUsersDelta)
		fmt.Printf("  Paid monthly delta: %.2f\n", result.Calibration.PaidMonthlyDelta)
		fmt.Printf("  Paid annual delta: %.2f\n", result.Calibration.PaidAnnualDelta)
		fmt.Println()
	}

	covenantStatus := "✅ PASSED"
	if !result.Covenants.Passed {
		covenantStatus = "❌ FAILED"
		fmt.Printf("Covenant Status: %s\n", covenantStatus)
		if len(result.Covenants.Violations) > 0 {
			fmt.Println("\nViolations:")
			for _, violation := range result.Covenants.Violations {
				fmt.Printf("  • %s\n", violation.Message)
				fmt.Printf("    Expected: %s %.2f, Actual: %.2f\n",
					formatOperator(violation.Operator), violation.Value, violation.Actual)
			}
		}
	} else {
		fmt.Printf("Covenant Status: %s\n", covenantStatus)
	}
	fmt.Println()

	if verbose && len(result.ScaleResult.Scenarios) > 0 {
		if result.Revenue.Mode == "mix" && len(result.Revenue.ByPlan) > 0 {
			fmt.Println("Revenue Mix:")
			for _, plan := range result.Revenue.ByPlan {
				if plan.Share != nil {
					fmt.Printf("  %s: %.1f%% -> %d users -> $%.2f\n",
						plan.PlanName, *plan.Share*100, plan.Users, plan.Revenue)
				}
			}
			fmt.Println()
		}

		fmt.Println("Scale Simulation:")
		for _, scenario := range result.ScaleResult.Scenarios {
			fmt.Printf("  Step %d: %d users - Total Cost: $%.2f\n",
				scenario.Step, scenario.Users, scenario.TotalCosts.GrandTotal)
		}
		fmt.Println()
	}

	if verbose {
		fmt.Printf("Stress Test Details (%d iterations):\n", result.StressResult.MonteCarlo.Iterations)
		fmt.Printf("  Mean: $%.2f\n", result.StressResult.MeanCostPerUser)
		fmt.Printf("  P95: $%.2f\n", result.StressResult.P95CostPerUser)
		fmt.Printf("  P99: $%.2f\n", result.StressResult.P99CostPerUser)
		if hasOperatingView(result) {
			fmt.Printf("  P95 booked margin: %.2f%%\n", bookedStressMargin(result, result.StressResult.P95CostPerUser))
			fmt.Printf("  P95 operating margin: %.2f%%\n", operatingStressMargin(result, result.StressResult.P95CostPerUser))
			fmt.Printf("  P99 booked margin: %.2f%%\n", bookedStressMargin(result, result.StressResult.P99CostPerUser))
			fmt.Printf("  P99 operating margin: %.2f%%\n", operatingStressMargin(result, result.StressResult.P99CostPerUser))
		}
		fmt.Printf("  Worst Case Total Cost: $%.2f\n", result.StressResult.WorstCaseTotalCost)
		fmt.Println()
	}
}

// FormatCLIQuietResult returns a single-line summary suitable for scripts.
func FormatCLIQuietResult(result SimulationResult) string {
	status := "PASSED"
	if !result.Covenants.Passed {
		status = "FAILED"
	}
	return fmt.Sprintf("%s margin=%.2f%% cost_per_user=%.2f violations=%d",
		status,
		result.Margin.GrossMargin,
		result.Margin.CostPerUser,
		len(result.Covenants.Violations),
	)
}

// PrintCLIResult writes CLI output to stdout.
func PrintCLIResult(result SimulationResult, verbose bool) {
	FormatCLIResult(result, verbose)
}

// PrintCLIQuietResult writes quiet output to stdout.
func PrintCLIQuietResult(result SimulationResult) {
	fmt.Println(FormatCLIQuietResult(result))
}

// PrintCLIResultToFile writes CLI output to a file.
func PrintCLIResultToFile(result SimulationResult, verbose bool, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	oldStdout := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = oldStdout }()

	FormatCLIResult(result, verbose)
	return nil
}

func formatOperator(op string) string {
	switch op {
	case "gt":
		return ">"
	case "lt":
		return "<"
	case "gte":
		return ">="
	case "lte":
		return "<="
	case "eq":
		return "=="
	default:
		return op
	}
}
