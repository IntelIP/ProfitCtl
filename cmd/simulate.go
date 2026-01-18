package cmd

import (
	"fmt"
	"os"

	"github.com/profitctl/profitctl/internal/config"
	"github.com/profitctl/profitctl/internal/cost"
	"github.com/profitctl/profitctl/internal/covenant"
	"github.com/profitctl/profitctl/internal/output"
	"github.com/profitctl/profitctl/internal/pricing"
	"github.com/profitctl/profitctl/internal/simulation"
	"github.com/spf13/cobra"
)

var (
	jsonOutput     bool
	markdownOutput bool
	quietOutput    bool
	verboseOutput  bool
)

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Run cost and profit simulations",
	Long: `Run scale simulations, p95/p99 stress testing, and calculate margins.
Requires a valid profit.yml configuration file.`,
	RunE: runSimulate,
}

func init() {
	simulateCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	simulateCmd.Flags().BoolVar(&markdownOutput, "markdown", false, "Output results as markdown")
	simulateCmd.Flags().BoolVarP(&quietOutput, "quiet", "q", false, "Minimal output")
	simulateCmd.Flags().BoolVarP(&verboseOutput, "verbose", "v", false, "Verbose output")
}

func runSimulate(cmd *cobra.Command, args []string) error {
	cfgFile, _ := cmd.Flags().GetString("file")

	// Parse configuration
	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate required configuration
	if cfg.Simulation == nil {
		return fmt.Errorf("simulation configuration is required")
	}

	// Create cost engine
	costEngine := cost.NewCostEngine(cfg)

	// Run scale simulation
	scaleConfig := simulation.NewScaleConfig(cfg.Simulation)
	scaleResult := simulation.RunScaleSimulation(cfg.FixedCosts, cfg.VariableCosts, scaleConfig, 12) // 12 months

	// Run stress test for base scenario (first scale point)
	baseUsers := cfg.Simulation.BaseUsers
	if len(scaleResult.Scenarios) > 0 {
		baseUsers = scaleResult.Scenarios[0].Users
	}
	stressResult := simulation.RunStressTest(cfg.FixedCosts, cfg.VariableCosts, baseUsers, 12, cfg.Simulation.Iterations)

	// Calculate revenue and margins for base scenario
	baseScenario := scaleResult.Scenarios[0]
	var revenueResult pricing.RevenueResult
	var marginResult pricing.MarginResult
	var totalCostResult cost.TotalCostResult

	if cfg.Pricing != nil {
		revenueResult = pricing.CalculateRevenue(cfg.Pricing, baseScenario.Users)
	}

	// Calculate margins using cost engine
	totalCostResult = costEngine.CalculateTotalCosts(baseScenario.Users, 12)
	if revenueResult.Total > 0 {
		marginResult = pricing.CalculateMargins(
			revenueResult.Total,
			totalCostResult.GrandTotal,
			totalCostResult.GrandByLayer,
			baseScenario.Users,
		)
	}

	// Calculate stress test margins (if revenue is available)
	var p95Margin, p99Margin float64
	if revenueResult.Total > 0 {
		p95TotalCost := stressResult.P95CostPerUser * float64(baseScenario.Users)
		p95Margin = ((revenueResult.Total - p95TotalCost) / revenueResult.Total) * 100

		p99TotalCost := stressResult.P99CostPerUser * float64(baseScenario.Users)
		p99Margin = ((revenueResult.Total - p99TotalCost) / revenueResult.Total) * 100
	}

	// Validate covenants
	covenantResults := covenant.SimulationResults{
		Margin:         marginResult.GrossMargin,
		CostPerUser:    marginResult.CostPerUser,
		P95Margin:      p95Margin,
		P95CostPerUser: stressResult.P95CostPerUser,
		P99Margin:      p99Margin,
		P99CostPerUser: stressResult.P99CostPerUser,
	}
	covenantValidation := covenant.ValidateCovenants(cfg.Covenants, covenantResults)

	// Build output result
	outputResult := output.SimulationResult{
		Users:         baseScenario.Users,
		Months:        12,
		GrowthFactor:   scaleConfig.GrowthFactor,
		FixedCosts:     baseScenario.FixedCosts,
		VariableCosts:  baseScenario.VariableCost,
		TotalCosts:     totalCostResult,
		Revenue:        revenueResult,
		Margin:         marginResult,
		ScaleResult:    scaleResult,
		StressResult:   stressResult,
		Covenants:      covenantValidation,
	}

	// Output results based on format flags
	if jsonOutput {
		if err := output.PrintJSONResult(outputResult); err != nil {
			return fmt.Errorf("failed to output JSON: %w", err)
		}
	} else if markdownOutput {
		output.PrintMarkdownResult(outputResult)
	} else {
		output.PrintCLIResult(outputResult, verboseOutput)
	}

	// Exit code based on covenant validation (Task 18 - CI Integration)
	if !covenantValidation.Passed {
		os.Exit(1) // Covenant breach - exit code 1
	}

	return nil
}

func init() {
	// This ensures simulateCmd is initialized when the package is loaded
}