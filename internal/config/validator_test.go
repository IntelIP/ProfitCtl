package config

import (
	"testing"

	"github.com/IntelIP/ProfitCtl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	validConfig := &Config{
		Project: &ProjectInfo{Name: "test"},
		FixedCosts: []types.FixedCost{
			{
				Name:   "Server",
				Amount: 1000,
				Period: types.PeriodMonthly,
				Layer:  types.LayerInfrastructure,
			},
		},
		VariableCosts: []types.VariableCost{
			{
				Name:         "API Calls",
				CostPerUnit:  0.0001,
				UnitsPerUser: 10000,
				Distribution: types.DistNormal,
				Mean:         float64Ptr(10000),
				StdDev:       float64Ptr(2000),
				Layer:        types.LayerApplication,
			},
		},
		Pricing: &PricingConfig{
			Plans: []PricingPlan{
				{
					Name:   "Basic",
					Price:  10,
					Limits: &PlanLimits{Users: 1000},
				},
				{
					Name:   "Pro",
					Price:  29,
					Limits: &PlanLimits{Users: 5000},
				},
			},
		},
		Covenants: []Covenant{
			{
				Type:     "threshold",
				Field:    "margin",
				Operator: "gte",
				Value:    20,
				Message:  "Gross margin must be >= 20%",
			},
		},
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err, "Valid config should pass")
}

func TestValidateConfig_NormalDistributionMissingMean(t *testing.T) {
	invalidConfig := &Config{
		VariableCosts: []types.VariableCost{
			{
				Name:         "API Calls",
				CostPerUnit:  0.0001,
				UnitsPerUser: 10000,
				Distribution: types.DistNormal,
				StdDev:       float64Ptr(2000),
				Layer:        types.LayerApplication,
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Config without mean should fail")
	assert.Contains(t, err.Error(), "mean")
}

func TestValidateConfig_UniformDistributionInvalidRange(t *testing.T) {
	invalidConfig := &Config{
		VariableCosts: []types.VariableCost{
			{
				Name:         "Storage",
				CostPerUnit:  0.023,
				UnitsPerUser: 5,
				Distribution: types.DistUniform,
				Min:          float64Ptr(10),
				Max:          float64Ptr(2),
				Layer:        types.LayerInfrastructure,
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Config with min >= max should fail")
	assert.Contains(t, err.Error(), "must be set and > min")
}

func TestValidateConfig_PricingPlansInvalid(t *testing.T) {
	invalidConfig := &Config{
		Pricing: &PricingConfig{
			Plans: []PricingPlan{
				{
					Name:   "Basic",
					Price:  10,
					Limits: &PlanLimits{Users: 1000},
				},
				{
					Name:   "Pro",
					Price:  29,
					Limits: &PlanLimits{Users: 500},
				},
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Config with invalid pricing tiers should fail")
	assert.Contains(t, err.Error(), "higher limit")
}

func TestValidateConfig_AllowsUnlimitedFinalPlan(t *testing.T) {
	validConfig := &Config{
		Pricing: &PricingConfig{
			Plans: []PricingPlan{
				{
					Name:   "Basic",
					Price:  10,
					Limits: &PlanLimits{Users: 1000},
				},
				{
					Name:   "Unlimited",
					Price:  29,
					Limits: nil,
				},
			},
		},
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err, "Config with a final unlimited plan should pass")
}

func TestValidateConfig_RejectsUnlimitedNonFinalPlan(t *testing.T) {
	invalidConfig := &Config{
		Pricing: &PricingConfig{
			Plans: []PricingPlan{
				{
					Name:   "Unlimited",
					Price:  29,
					Limits: nil,
				},
				{
					Name:   "Enterprise",
					Price:  99,
					Limits: &PlanLimits{Users: 5000},
				},
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Config with a non-final unlimited plan should fail")
	assert.Contains(t, err.Error(), "last pricing plan")
}

func TestValidateConfig_MixModeValid(t *testing.T) {
	validConfig := &Config{
		Pricing: &PricingConfig{
			Mode: "mix",
			Plans: []PricingPlan{
				{Name: "Free", Price: 0, Share: float64Ptr(0.7)},
				{Name: "Pro", Price: 29, Share: float64Ptr(0.3)},
			},
		},
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err, "Mix mode config should pass when shares sum to 1")
}

func TestValidateConfig_MixModeRejectsLimits(t *testing.T) {
	invalidConfig := &Config{
		Pricing: &PricingConfig{
			Mode: "mix",
			Plans: []PricingPlan{
				{Name: "Free", Price: 0, Share: float64Ptr(0.7)},
				{Name: "Pro", Price: 29, Share: float64Ptr(0.3), Limits: &PlanLimits{Users: 100}},
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Mix mode should reject limits")
	assert.Contains(t, err.Error(), "does not support limits")
}

func TestValidateConfig_MixModeRejectsShareTotalMismatch(t *testing.T) {
	invalidConfig := &Config{
		Pricing: &PricingConfig{
			Mode: "mix",
			Plans: []PricingPlan{
				{Name: "Free", Price: 0, Share: float64Ptr(0.6)},
				{Name: "Pro", Price: 29, Share: float64Ptr(0.3)},
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Mix mode should require shares to sum to 1")
	assert.Contains(t, err.Error(), "must sum to 1.0")
}

func TestValidateConfig_HybridModeValid(t *testing.T) {
	validConfig := &Config{
		Pricing: &PricingConfig{
			Mode: "hybrid",
			Contract: &HybridContract{
				BasePlatformFee:  500,
				PerSeatFee:       25,
				WorkspaceMinimum: 2000,
				IncludedSeats:    10,
				Pilot: &PilotConfig{
					SetupFee:       5000,
					MonthlyFee:     1500,
					DurationMonths: 3,
				},
			},
		},
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err, "Hybrid mode config should pass")
}

func TestValidateConfig_HybridModeRejectsPlans(t *testing.T) {
	invalidConfig := &Config{
		Pricing: &PricingConfig{
			Mode: "hybrid",
			Plans: []PricingPlan{
				{Name: "Basic", Price: 10, Limits: &PlanLimits{Users: 100}},
			},
			Contract: &HybridContract{BasePlatformFee: 500},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Hybrid mode should reject plans")
	assert.Contains(t, err.Error(), "does not support plans")
}

func TestValidateConfig_HybridModeRequiresContract(t *testing.T) {
	invalidConfig := &Config{
		Pricing: &PricingConfig{Mode: "hybrid"},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Hybrid mode should require contract")
	assert.Contains(t, err.Error(), "requires contract")
}

func TestValidateConfig_HybridModeRejectsInvalidPilot(t *testing.T) {
	invalidConfig := &Config{
		Pricing: &PricingConfig{
			Mode: "hybrid",
			Contract: &HybridContract{
				BasePlatformFee: 500,
				Pilot: &PilotConfig{
					DurationMonths: 0,
				},
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Hybrid mode should validate pilot duration")
	assert.Contains(t, err.Error(), "duration_months")
}

func TestValidateConfig_PaymentFeesValid(t *testing.T) {
	validConfig := &Config{
		PaymentFees: &PaymentFeesConfig{
			PaidUserFee: &PaymentFeeProfile{
				MonthlyPercent: 2.9,
				PerTransaction: 0.30,
			},
			BillingMix: &BillingMixConfig{
				MonthlyShare: 0.75,
				AnnualShare:  0.25,
			},
			Assumptions: &PaymentFeeAssumptions{
				AnnualPrepaidMonths: 12,
			},
		},
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err, "Payment fees config should pass")
}

func TestValidateConfig_PaymentFeesRejectsNegativeValues(t *testing.T) {
	invalidConfig := &Config{
		PaymentFees: &PaymentFeesConfig{
			PaidUserFee: &PaymentFeeProfile{
				MonthlyPercent: -1,
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Payment fees should reject negative values")
	assert.Contains(t, err.Error(), "monthly_percent")
}

func TestValidateConfig_PaymentFeesRejectsBillingMixMismatch(t *testing.T) {
	invalidConfig := &Config{
		PaymentFees: &PaymentFeesConfig{
			PaidUserFee: &PaymentFeeProfile{
				MonthlyPercent: 2.9,
			},
			BillingMix: &BillingMixConfig{
				MonthlyShare: 0.8,
				AnnualShare:  0.1,
			},
			Assumptions: &PaymentFeeAssumptions{
				AnnualPrepaidMonths: 12,
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Payment fees should require billing mix to sum to 1")
	assert.Contains(t, err.Error(), "must sum to 1.0")
}

func TestValidateConfig_CalibrationValid(t *testing.T) {
	validConfig := &Config{
		Calibration: &CalibrationConfig{
			Period:       "2026-03",
			Source:       "stripe_export",
			GrossRevenue: 125000,
			PaymentFees:  4380,
			FreeUsers:    12000,
			PaidUsers: &CalibrationPaidUsers{
				Monthly: 1800,
				Annual:  400,
			},
			PlanMix: map[string]float64{
				"starter": 0.65,
				"pro":     0.35,
			},
			Usage: map[string]float64{
				"api_calls": 3200000,
			},
		},
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err, "Calibration config should pass")
}

func TestValidateConfig_CalibrationRejectsPlanMixMismatch(t *testing.T) {
	invalidConfig := &Config{
		Calibration: &CalibrationConfig{
			PlanMix: map[string]float64{
				"starter": 0.5,
				"pro":     0.2,
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err, "Calibration should require plan mix to sum to 1")
	assert.Contains(t, err.Error(), "plan_mix shares must sum to 1.0")
}

func TestValidateConfig_WorkspaceHybridRequiresWorkspaceAssumption(t *testing.T) {
	invalidConfig := &Config{
		Pricing: &PricingConfig{
			Mode: PricingModeWorkspaceHybrid,
			Plans: []PricingPlan{
				{Name: "Minimum", Price: 0, Share: float64Ptr(0.8), WorkspaceMinimum: float64Ptr(149)},
				{Name: "Pro", Price: 39, Share: float64Ptr(0.2), WorkspaceMinimum: float64Ptr(299)},
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "average_users_per_workspace")
}

func TestValidateConfig_WorkspaceHybridAllowsMinimums(t *testing.T) {
	validConfig := &Config{
		Pricing: &PricingConfig{
			Mode: PricingModeWorkspaceHybrid,
			Workspace: &WorkspaceConfig{
				AverageUsersPerWorkspace: 2,
			},
			Plans: []PricingPlan{
				{Name: "Minimum", Price: 0, Share: float64Ptr(0.8), WorkspaceMinimum: float64Ptr(149)},
				{Name: "Pro", Price: 39, Share: float64Ptr(0.2), WorkspaceMinimum: float64Ptr(299)},
			},
		},
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err)
}

func TestValidateConfig_PaidUserScopedVariableCostsRequirePricing(t *testing.T) {
	invalidConfig := &Config{
		VariableCosts: []types.VariableCost{
			{
				Name:         "Billing",
				CostPerUnit:  1,
				UnitsPerUser: 1,
				Distribution: types.DistNormal,
				Mean:         float64Ptr(1),
				StdDev:       float64Ptr(0.1),
				Layer:        types.LayerService,
				UserScope:    types.UserScopePaidUsers,
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_scope=paid_users")
}

func TestValidateConfig_DefaultsEconomicsLayerToDelivery(t *testing.T) {
	validConfig := &Config{
		FixedCosts: []types.FixedCost{
			{
				Name:   "Runtime Infra",
				Amount: 100,
				Period: types.PeriodMonthly,
				Layer:  types.LayerInfrastructure,
			},
		},
		VariableCosts: []types.VariableCost{
			{
				Name:         "Runtime Tokens",
				CostPerUnit:  0.01,
				UnitsPerUser: 100,
				Distribution: types.DistNormal,
				Mean:         float64Ptr(100),
				StdDev:       float64Ptr(10),
				Layer:        types.LayerService,
			},
		},
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err)
	assert.Equal(t, types.EconomicsLayerDelivery, validConfig.FixedCosts[0].EconomicsLayer)
	assert.Equal(t, types.EconomicsLayerDelivery, validConfig.VariableCosts[0].EconomicsLayer)
}

func TestValidateConfig_InvalidEconomicsLayer(t *testing.T) {
	invalidConfig := &Config{
		FixedCosts: []types.FixedCost{
			{
				Name:           "Bad Layer",
				Amount:         100,
				Period:         types.PeriodMonthly,
				Layer:          types.LayerInfrastructure,
				EconomicsLayer: types.EconomicsLayer("unsupported"),
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "EconomicsLayer")
	assert.Contains(t, err.Error(), "oneof")
}

func TestValidateConfig_InvalidEconomicsAllocation(t *testing.T) {
	invalidConfig := &Config{
		FixedCosts: []types.FixedCost{
			{
				Name:           "AI Dev Tooling",
				Amount:         500,
				Period:         types.PeriodMonthly,
				Layer:          types.LayerService,
				EconomicsLayer: types.EconomicsLayerProductization,
				Allocation: &types.EconomicsAllocation{
					Mode: types.EconomicsAllocationMode("unsupported"),
				},
			},
		},
	}

	err := ValidateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Allocation.Mode")
	assert.Contains(t, err.Error(), "oneof")
}

func float64Ptr(f float64) *float64 {
	return &f
}
