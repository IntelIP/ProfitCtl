package cmd

import (
	"fmt"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/internal/cost"
	"github.com/IntelIP/ProfitCtl/internal/covenant"
	"github.com/IntelIP/ProfitCtl/internal/output"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/IntelIP/ProfitCtl/internal/simulation"
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

	if jsonOutput && markdownOutput {
		return wrapExit(2, fmt.Errorf("--json and --markdown cannot be used together"))
	}
	if quietOutput && (jsonOutput || markdownOutput) {
		return wrapExit(2, fmt.Errorf("--quiet cannot be combined with --json or --markdown"))
	}

	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		return wrapExit(2, fmt.Errorf("failed to parse config: %w", err))
	}

	if cfg.Simulation == nil {
		return wrapExit(2, fmt.Errorf("simulation configuration is required"))
	}

	costEngine := cost.NewCostEngine(cfg)

	scaleConfig := simulation.NewScaleConfig(cfg.Simulation)
	scaleResult := simulation.RunScaleSimulation(cfg.FixedCosts, cfg.VariableCosts, scaleConfig, 12)
	const summaryMonths = 1

	baseUsers := cfg.Simulation.BaseUsers
	if len(scaleResult.Scenarios) > 0 {
		baseUsers = scaleResult.Scenarios[0].Users
	}
	stressResult := simulation.RunStressTest(cfg.FixedCosts, cfg.VariableCosts, baseUsers, summaryMonths, cfg.Simulation.Iterations)

	baseScenario := scaleResult.Scenarios[0]
	var revenueResult pricing.RevenueResult
	var paymentFeeResult pricing.PaymentFeeResult
	var calibrationResult pricing.CalibrationResult
	var marginResult pricing.MarginResult

	if cfg.Pricing != nil {
		revenueResult = pricing.CalculateRevenue(cfg.Pricing, baseScenario.Users)
	}
	if cfg.PaymentFees != nil && revenueResult.Total > 0 {
		paymentFeeResult = pricing.CalculatePaymentFees(cfg.PaymentFees, cfg.Pricing, revenueResult, baseScenario.Users)
	}
	calibrationResult = pricing.CalculateCalibration(cfg.Calibration, revenueResult, paymentFeeResult)

	totalCostResult := costEngine.CalculateTotalCosts(baseScenario.Users, summaryMonths)
	adjustedTotalCost := totalCostResult.GrandTotal + paymentFeeResult.Total
	if revenueResult.Total > 0 {
		marginResult = pricing.CalculateMargins(
			revenueResult.Total,
			adjustedTotalCost,
			totalCostResult.GrandByLayer,
			baseScenario.Users,
		)
	}

	var p95Margin, p99Margin float64
	if revenueResult.Total > 0 {
		p95TotalCost := stressResult.P95CostPerUser*float64(baseScenario.Users) + paymentFeeResult.Total
		p95Margin = ((revenueResult.Total - p95TotalCost) / revenueResult.Total) * 100

		p99TotalCost := stressResult.P99CostPerUser*float64(baseScenario.Users) + paymentFeeResult.Total
		p99Margin = ((revenueResult.Total - p99TotalCost) / revenueResult.Total) * 100
	}

	covenantResults := covenant.SimulationResults{
		Margin:         marginResult.GrossMargin,
		CostPerUser:    marginResult.CostPerUser,
		P95Margin:      p95Margin,
		P95CostPerUser: stressResult.P95CostPerUser,
		P99Margin:      p99Margin,
		P99CostPerUser: stressResult.P99CostPerUser,
	}
	covenantValidation := covenant.ValidateCovenants(cfg.Covenants, covenantResults)

	outputResult := output.SimulationResult{
		Users:         baseScenario.Users,
		Months:        12,
		GrowthFactor:  scaleConfig.GrowthFactor,
		FixedCosts:    baseScenario.FixedCosts,
		VariableCosts: baseScenario.VariableCost,
		TotalCosts:    totalCostResult,
		Revenue:       revenueResult,
		PaymentFees:   paymentFeeResult,
		Calibration:   calibrationResult,
		Margin:        marginResult,
		ScaleResult:   scaleResult,
		StressResult:  stressResult,
		Covenants:     covenantValidation,
	}

	switch {
	case jsonOutput:
		if err := output.PrintJSONResult(outputResult); err != nil {
			return wrapExit(3, fmt.Errorf("failed to output JSON: %w", err))
		}
	case markdownOutput:
		output.PrintMarkdownResult(outputResult)
	case quietOutput:
		output.PrintCLIQuietResult(outputResult)
	default:
		output.PrintCLIResult(outputResult, verboseOutput)
	}

	if !covenantValidation.Passed {
		return wrapExitSilent(1, fmt.Errorf("covenant validation failed"))
	}

	return nil
}
