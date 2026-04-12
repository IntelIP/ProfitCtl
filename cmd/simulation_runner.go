package cmd

import (
	"fmt"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/internal/cost"
	"github.com/IntelIP/ProfitCtl/internal/covenant"
	"github.com/IntelIP/ProfitCtl/internal/output"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/IntelIP/ProfitCtl/internal/simulation"
)

func buildSimulationResult(cfgFile string) (output.SimulationResult, error) {
	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		return output.SimulationResult{}, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Simulation == nil {
		return output.SimulationResult{}, fmt.Errorf("simulation configuration is required")
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
	var operatingMarginResult pricing.MarginResult
	billableUsers := resolveBillableUsers(cfg.Simulation, baseScenario.Users)

	if cfg.Pricing != nil {
		revenueResult = pricing.CalculateRevenueWithSeats(cfg.Pricing, baseScenario.Users, billableUsers)
	}
	if cfg.PaymentFees != nil && revenueResult.Total > 0 {
		paymentFeeResult = pricing.CalculatePaymentFeesWithSeats(cfg.PaymentFees, cfg.Pricing, revenueResult, baseScenario.Users, billableUsers)
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

	operatingRevenue := revenueResult.RecurringTotal
	if operatingRevenue == 0 {
		operatingRevenue = revenueResult.Total
	}
	operatingPaymentFees := paymentFeeResult.OperatingAmount
	if operatingPaymentFees == 0 {
		operatingPaymentFees = paymentFeeResult.Total
	}
	operatingTotalCost := totalCostResult.GrandTotal + operatingPaymentFees
	if operatingRevenue > 0 {
		operatingMarginResult = pricing.CalculateMargins(
			operatingRevenue,
			operatingTotalCost,
			totalCostResult.GrandByLayer,
			baseScenario.Users,
		)
	}

	var p95Margin, p99Margin float64
	var p95OperatingMargin, p99OperatingMargin float64
	if revenueResult.Total > 0 {
		p95TotalCost := stressResult.P95CostPerUser*float64(baseScenario.Users) + paymentFeeResult.Total
		p95Margin = ((revenueResult.Total - p95TotalCost) / revenueResult.Total) * 100

		p99TotalCost := stressResult.P99CostPerUser*float64(baseScenario.Users) + paymentFeeResult.Total
		p99Margin = ((revenueResult.Total - p99TotalCost) / revenueResult.Total) * 100
	}
	if operatingRevenue > 0 {
		p95OperatingTotalCost := stressResult.P95CostPerUser*float64(baseScenario.Users) + operatingPaymentFees
		p95OperatingMargin = ((operatingRevenue - p95OperatingTotalCost) / operatingRevenue) * 100

		p99OperatingTotalCost := stressResult.P99CostPerUser*float64(baseScenario.Users) + operatingPaymentFees
		p99OperatingMargin = ((operatingRevenue - p99OperatingTotalCost) / operatingRevenue) * 100
	}

	covenantResults := covenant.SimulationResults{
		Margin:               marginResult.GrossMargin,
		OperatingMargin:      operatingMarginResult.GrossMargin,
		CostPerUser:          marginResult.CostPerUser,
		OperatingCostPerUser: operatingMarginResult.CostPerUser,
		P95Margin:            p95Margin,
		P95OperatingMargin:   p95OperatingMargin,
		P95CostPerUser:       stressResult.P95CostPerUser,
		P99Margin:            p99Margin,
		P99OperatingMargin:   p99OperatingMargin,
		P99CostPerUser:       stressResult.P99CostPerUser,
	}
	covenantValidation := covenant.ValidateCovenants(cfg.Covenants, covenantResults)

	return output.SimulationResult{
		Users:           baseScenario.Users,
		BillableUsers:   derefUsers(billableUsers, baseScenario.Users),
		Months:          12,
		GrowthFactor:    scaleConfig.GrowthFactor,
		FixedCosts:      baseScenario.FixedCosts,
		VariableCosts:   baseScenario.VariableCost,
		TotalCosts:      totalCostResult,
		Revenue:         revenueResult,
		PaymentFees:     paymentFeeResult,
		Calibration:     calibrationResult,
		Margin:          marginResult,
		OperatingMargin: operatingMarginResult,
		ScaleResult:     scaleResult,
		StressResult:    stressResult,
		Covenants:       covenantValidation,
	}, nil
}

func resolveBillableUsers(cfg *config.SimulationConfig, scenarioUsers int) *int {
	if cfg == nil || cfg.BillableUsers == nil {
		return nil
	}

	billableUsers := *cfg.BillableUsers
	if billableUsers < 0 {
		billableUsers = 0
	}
	if billableUsers > scenarioUsers {
		billableUsers = scenarioUsers
	}
	return &billableUsers
}

func derefUsers(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}
