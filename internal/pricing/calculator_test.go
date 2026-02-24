package pricing

import (
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCalculateRevenue_EmptyPricing(t *testing.T) {
	result := CalculateRevenue(nil, 1000)
	
	assert.Equal(t, 0.0, result.Total, "Revenue should be zero for nil pricing")
	assert.Empty(t, result.ByPlan, "No plans should be present")
}

func TestCalculateRevenue_NoPlans(t *testing.T) {
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{},
	}
	
	result := CalculateRevenue(pricing, 1000)
	
	assert.Equal(t, 0.0, result.Total, "Revenue should be zero for empty plans")
	assert.Empty(t, result.ByPlan, "No plans should be present")
}

func TestCalculateRevenue_ZeroUsers(t *testing.T) {
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{
			{
				Name:  "Basic",
				Price: 10,
				Limits: &config.PlanLimits{
					Users: 1000,
				},
			},
		},
	}
	
	result := CalculateRevenue(pricing, 0)
	
	assert.Equal(t, 0.0, result.Total, "Revenue should be zero for zero users")
	assert.Empty(t, result.ByPlan, "No plans should be used")
}

func TestCalculateRevenue_SinglePlan(t *testing.T) {
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{
			{
				Name:  "Basic",
				Price: 10,
				Limits: &config.PlanLimits{
					Users: 1000,
				},
			},
		},
	}
	
	result := CalculateRevenue(pricing, 500)
	
	assert.Equal(t, 5000.0, result.Total, "Revenue should be 500 * $10 = $5000")
	assert.Len(t, result.ByPlan, 1, "Should have one plan")
	assert.Equal(t, "Basic", result.ByPlan[0].PlanName)
	assert.Equal(t, 10.0, result.ByPlan[0].Price)
	assert.Equal(t, 500, result.ByPlan[0].Users)
	assert.Equal(t, 5000.0, result.ByPlan[0].Revenue)
}

func TestCalculateRevenue_TieredPricing(t *testing.T) {
	// Example from plan: basic=1000, pro=5000, enterprise=10000
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{
			{
				Name:  "basic",
				Price: 10,
				Limits: &config.PlanLimits{
					Users: 1000,
				},
			},
			{
				Name:  "pro",
				Price: 29,
				Limits: &config.PlanLimits{
					Users: 5000,
				},
			},
			{
				Name:  "enterprise",
				Price: 99,
				Limits: &config.PlanLimits{
					Users: 10000,
				},
			},
		},
	}
	
	// Test case: 2500 users
	// First 1000 users at $10 = $10,000
	// Next 1500 users at $29 = $43,500
	// Total = $53,500
	result := CalculateRevenue(pricing, 2500)
	
	expectedTotal := 53500.0
	assert.Equal(t, expectedTotal, result.Total, "Total revenue for 2500 users")
	assert.Len(t, result.ByPlan, 2, "Should use two plans")
	
	// First plan: basic
	assert.Equal(t, "basic", result.ByPlan[0].PlanName)
	assert.Equal(t, 10.0, result.ByPlan[0].Price)
	assert.Equal(t, 1000, result.ByPlan[0].Users)
	assert.Equal(t, 10000.0, result.ByPlan[0].Revenue)
	
	// Second plan: pro
	assert.Equal(t, "pro", result.ByPlan[1].PlanName)
	assert.Equal(t, 29.0, result.ByPlan[1].Price)
	assert.Equal(t, 1500, result.ByPlan[1].Users)
	assert.Equal(t, 43500.0, result.ByPlan[1].Revenue)
}

func TestCalculateRevenue_TieredPricing_AllTiers(t *testing.T) {
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{
			{
				Name:  "basic",
				Price: 10,
				Limits: &config.PlanLimits{
					Users: 1000,
				},
			},
			{
				Name:  "pro",
				Price: 29,
				Limits: &config.PlanLimits{
					Users: 5000,
				},
			},
			{
				Name:  "enterprise",
				Price: 99,
				Limits: &config.PlanLimits{
					Users: 10000,
				},
			},
		},
	}
	
	// Test case: 7500 users
	// First 1000 users at $10 = $10,000
	// Next 4000 users at $29 = $116,000
	// Next 2500 users at $99 = $247,500
	// Total = $373,500
	result := CalculateRevenue(pricing, 7500)
	
	expectedTotal := 373500.0
	assert.Equal(t, expectedTotal, result.Total, "Total revenue for 7500 users")
	assert.Len(t, result.ByPlan, 3, "Should use all three plans")
	
	// First plan: basic (1000 users)
	assert.Equal(t, "basic", result.ByPlan[0].PlanName)
	assert.Equal(t, 1000, result.ByPlan[0].Users)
	assert.Equal(t, 10000.0, result.ByPlan[0].Revenue)
	
	// Second plan: pro (4000 users: 5000 - 1000)
	assert.Equal(t, "pro", result.ByPlan[1].PlanName)
	assert.Equal(t, 4000, result.ByPlan[1].Users)
	assert.Equal(t, 116000.0, result.ByPlan[1].Revenue)
	
	// Third plan: enterprise (2500 users: 7500 - 5000)
	assert.Equal(t, "enterprise", result.ByPlan[2].PlanName)
	assert.Equal(t, 2500, result.ByPlan[2].Users)
	assert.Equal(t, 247500.0, result.ByPlan[2].Revenue)
}

func TestCalculateRevenue_ExactTierBoundaries(t *testing.T) {
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{
			{
				Name:  "basic",
				Price: 10,
				Limits: &config.PlanLimits{
					Users: 1000,
				},
			},
			{
				Name:  "pro",
				Price: 29,
				Limits: &config.PlanLimits{
					Users: 5000,
				},
			},
		},
	}
	
	// Test exactly at tier boundary: 1000 users
	result := CalculateRevenue(pricing, 1000)
	
	assert.Equal(t, 10000.0, result.Total, "Revenue at tier boundary")
	assert.Len(t, result.ByPlan, 1, "Should only use first plan")
	assert.Equal(t, 1000, result.ByPlan[0].Users)
}

func TestCalculateRevenue_ExceedsAllTiers(t *testing.T) {
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{
			{
				Name:  "basic",
				Price: 10,
				Limits: &config.PlanLimits{
					Users: 1000,
				},
			},
			{
				Name:  "pro",
				Price: 29,
				Limits: &config.PlanLimits{
					Users: 5000,
				},
			},
		},
	}
	
	// Test exceeding all tiers: 10000 users
	// First 1000 at $10 = $10,000
	// Next 4000 (users 1001-5000) at $29 = $116,000
	// Remaining 5000 (users 5001-10000) at $29 = $145,000 (last tier applies to remaining)
	result := CalculateRevenue(pricing, 10000)
	
	expectedTotal := 271000.0 // 10000 + 116000 + 145000
	assert.Equal(t, expectedTotal, result.Total, "Revenue when exceeding all tiers")
	assert.Len(t, result.ByPlan, 2, "Should use both plans")
	assert.Equal(t, 1000, result.ByPlan[0].Users)
	assert.Equal(t, 9000, result.ByPlan[1].Users) // 4000 (tier) + 5000 (remaining) = 9000 total at pro rate
}

func TestCalculateRevenue_PlanWithoutLimits(t *testing.T) {
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{
			{
				Name:  "basic",
				Price: 10,
				Limits: &config.PlanLimits{
					Users: 1000,
				},
			},
			{
				Name:  "unlimited",
				Price: 50,
				Limits: nil, // No limits - applies to all remaining users
			},
		},
	}
	
	// Test: 2500 users
	// First 1000 at $10 = $10,000
	// Next 1500 at $50 = $75,000
	result := CalculateRevenue(pricing, 2500)
	
	expectedTotal := 85000.0
	assert.Equal(t, expectedTotal, result.Total, "Revenue with unlimited plan")
	assert.Len(t, result.ByPlan, 2, "Should use both plans")
	assert.Equal(t, 1000, result.ByPlan[0].Users)
	assert.Equal(t, 1500, result.ByPlan[1].Users)
	assert.Equal(t, 75000.0, result.ByPlan[1].Revenue)
}

func TestCalculateRevenue_InvalidTierOrder(t *testing.T) {
	// Test case where tier limits are not in ascending order
	// This should still work but may produce unexpected results
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{
			{
				Name:  "pro",
				Price: 29,
				Limits: &config.PlanLimits{
					Users: 5000,
				},
			},
			{
				Name:  "basic",
				Price: 10,
				Limits: &config.PlanLimits{
					Users: 1000, // This is less than previous, should be skipped
				},
			},
		},
	}
	
	// With 2500 users, only pro should apply (basic tier is invalid)
	result := CalculateRevenue(pricing, 2500)
	
	assert.Len(t, result.ByPlan, 1, "Should only use first valid plan")
	assert.Equal(t, "pro", result.ByPlan[0].PlanName)
	assert.Equal(t, 72500.0, result.Total) // 2500 * $29
}

func TestCalculateRevenue_PrecisionRounding(t *testing.T) {
	pricing := &config.PricingConfig{
		Plans: []config.PricingPlan{
			{
				Name:  "basic",
				Price: 10.333, // Price with many decimals
				Limits: &config.PlanLimits{
					Users: 100,
				},
			},
		},
	}
	
	result := CalculateRevenue(pricing, 100)
	
	// Should round to 2 decimal places
	expectedRevenue := 1033.30 // 100 * 10.333 = 1033.30 (rounded)
	assert.Equal(t, expectedRevenue, result.Total, "Revenue should be rounded to 2 decimals")
	assert.Equal(t, expectedRevenue, result.ByPlan[0].Revenue, "Plan revenue should be rounded")
}