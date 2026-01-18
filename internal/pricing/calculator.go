package pricing

import (
	"math"

	"github.com/profitctl/profitctl/internal/config"
)

// RevenueResult represents the calculated revenue from pricing plans
type RevenueResult struct {
	Total   float64
	ByPlan  []PlanRevenue
}

// PlanRevenue represents revenue from a single pricing plan
type PlanRevenue struct {
	PlanName string
	Price    float64
	Users    int
	Revenue  float64
}

// CalculateRevenue calculates total revenue from pricing plans for a given number of users
// Uses tiered pricing where each plan applies to a range of users
func CalculateRevenue(pricing *config.PricingConfig, users int) RevenueResult {
	if pricing == nil || len(pricing.Plans) == 0 {
		return RevenueResult{
			Total:  0,
			ByPlan: []PlanRevenue{},
		}
	}

	if users <= 0 {
		return RevenueResult{
			Total:  0,
			ByPlan: []PlanRevenue{},
		}
	}

	result := RevenueResult{
		ByPlan: make([]PlanRevenue, 0, len(pricing.Plans)),
	}

	remainingUsers := users
	previousLimit := 0

	// Iterate through plans in order
	for i, plan := range pricing.Plans {
		if plan.Limits == nil {
			// If no limits specified, this plan applies to all remaining users
			planUsers := remainingUsers
			revenue := plan.Price * float64(planUsers)
			
			result.ByPlan = append(result.ByPlan, PlanRevenue{
				PlanName: plan.Name,
				Price:    plan.Price,
				Users:    planUsers,
				Revenue:  math.Round(revenue*100) / 100,
			})
			
			result.Total += revenue
			break
		}

		planLimit := plan.Limits.Users
		
		// Calculate how many users fall into this tier
		// Users in this tier = min(remainingUsers, planLimit - previousLimit)
		tierCapacity := planLimit - previousLimit
		
		if tierCapacity <= 0 {
			// Skip invalid tiers (limit <= previous limit)
			continue
		}

		isLastPlan := i == len(pricing.Plans)-1
		planUsers := remainingUsers
		
		// If this is the last plan and we have remaining users, apply to all remaining
		// Otherwise, cap at tier capacity
		if !isLastPlan && planUsers > tierCapacity {
			planUsers = tierCapacity
		}

		revenue := plan.Price * float64(planUsers)

		result.ByPlan = append(result.ByPlan, PlanRevenue{
			PlanName: plan.Name,
			Price:    plan.Price,
			Users:    planUsers,
			Revenue:  math.Round(revenue*100) / 100,
		})

		result.Total += revenue
		remainingUsers -= planUsers
		previousLimit = planLimit

		if remainingUsers <= 0 {
			break
		}
	}

	// Round total to 2 decimal places
	result.Total = math.Round(result.Total*100) / 100

	return result
}