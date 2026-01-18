package config

import (
	"testing"

	"github.com/profitctl/profitctl/pkg/types"
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

func float64Ptr(f float64) *float64 {
	return &f
}
