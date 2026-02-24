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
	fmt.Println("───────────────────────────────")

	if result.Margin.GrossMargin != 0 {
		fmt.Printf("Mean margin: %.2f%%\n", result.Margin.GrossMargin)
	}

	if result.StressResult.MonteCarlo.P95 > 0 && result.Revenue.Total > 0 {
		p95TotalCost := result.StressResult.P95CostPerUser * float64(result.Users)
		p95Margin := ((result.Revenue.Total - p95TotalCost) / result.Revenue.Total) * 100
		fmt.Printf("p95 margin: %.2f%%\n", p95Margin)
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
		fmt.Printf("Cost per user: $%.2f\n", result.Margin.CostPerUser)
	} else {
		fmt.Printf("Variable COGS: $0.00/user\n")
		fmt.Printf("Cost per user: $%.2f\n", result.Margin.CostPerUser)
	}
	fmt.Println()

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
