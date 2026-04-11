package pricing

import (
	"math"
	"sort"
	"strings"

	"github.com/IntelIP/ProfitCtl/internal/config"
)

// RevenueResult represents the calculated revenue from pricing plans
type RevenueResult struct {
	Mode           string
	Total          float64
	RecurringTotal float64
	OneTimeTotal   float64
	MinimumUplift  float64
	Components     []RevenueComponent
	ByPlan         []PlanRevenue
}

// PlanRevenue represents revenue from a single pricing plan
type PlanRevenue struct {
	PlanName string
	Price    float64
	Share    *float64
	Cohort   string
	Users    int
	Revenue  float64
}

type RevenueComponent struct {
	Name   string
	Amount float64
}

// CalculateRevenue calculates total revenue from pricing plans for a given number of users
// Uses tiered pricing where each plan applies to a range of users
func CalculateRevenue(pricing *config.PricingConfig, users int) RevenueResult {
	if pricing == nil {
		return RevenueResult{
			Mode:       "tiered",
			Total:      0,
			Components: []RevenueComponent{},
			ByPlan:     []PlanRevenue{},
		}
	}

	mode := resolvedPricingMode(pricing)
	if len(pricing.Plans) == 0 && mode != "hybrid" {
		return RevenueResult{
			Mode:       mode,
			Total:      0,
			Components: []RevenueComponent{},
			ByPlan:     []PlanRevenue{},
		}
	}

	if users <= 0 && mode != "hybrid" {
		return RevenueResult{
			Mode:       mode,
			Total:      0,
			Components: []RevenueComponent{},
			ByPlan:     []PlanRevenue{},
		}
	}

	result := RevenueResult{
		Mode:       mode,
		Components: []RevenueComponent{},
		ByPlan:     make([]PlanRevenue, 0, len(pricing.Plans)),
	}

	switch result.Mode {
	case "mix":
		return calculateMixRevenue(pricing, users, result)
	case "hybrid":
		return calculateHybridRevenue(pricing, users, result)
	default:
		return calculateTieredRevenue(pricing, users, result)
	}
}

func calculateTieredRevenue(pricing *config.PricingConfig, users int, result RevenueResult) RevenueResult {
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
				Share:    nil,
				Cohort:   plan.Cohort,
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
			Share:    nil,
			Cohort:   plan.Cohort,
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
	result.RecurringTotal = result.Total

	return result
}

func calculateMixRevenue(pricing *config.PricingConfig, users int, result RevenueResult) RevenueResult {
	type planAllocation struct {
		index     int
		users     int
		remainder float64
	}

	allocations := make([]planAllocation, len(pricing.Plans))
	allocatedUsers := 0

	for i, plan := range pricing.Plans {
		share := 0.0
		if plan.Share != nil {
			share = *plan.Share
		}

		exactUsers := float64(users) * share
		baseUsers := int(math.Floor(exactUsers))
		allocations[i] = planAllocation{
			index:     i,
			users:     baseUsers,
			remainder: exactUsers - float64(baseUsers),
		}
		allocatedUsers += baseUsers
	}

	remainingUsers := users - allocatedUsers
	sort.SliceStable(allocations, func(i, j int) bool {
		return allocations[i].remainder > allocations[j].remainder
	})
	for i := 0; i < remainingUsers && i < len(allocations); i++ {
		allocations[i].users++
	}

	sort.SliceStable(allocations, func(i, j int) bool {
		return allocations[i].index < allocations[j].index
	})

	for _, allocation := range allocations {
		plan := pricing.Plans[allocation.index]
		revenue := plan.Price * float64(allocation.users)

		result.ByPlan = append(result.ByPlan, PlanRevenue{
			PlanName: plan.Name,
			Price:    plan.Price,
			Share:    plan.Share,
			Cohort:   plan.Cohort,
			Users:    allocation.users,
			Revenue:  math.Round(revenue*100) / 100,
		})
		result.Total += revenue
	}

	result.Total = math.Round(result.Total*100) / 100
	result.RecurringTotal = result.Total
	return result
}

func calculateHybridRevenue(pricing *config.PricingConfig, users int, result RevenueResult) RevenueResult {
	if pricing.Contract == nil {
		return result
	}

	contract := pricing.Contract
	billableSeats := users - contract.IncludedSeats
	if billableSeats < 0 {
		billableSeats = 0
	}

	baseFee := roundCurrency(contract.BasePlatformFee)
	seatRevenue := roundCurrency(float64(billableSeats) * contract.PerSeatFee)
	steadyStateRecurring := roundCurrency(baseFee + seatRevenue)
	minimumUplift := 0.0
	if contract.WorkspaceMinimum > steadyStateRecurring {
		minimumUplift = roundCurrency(contract.WorkspaceMinimum - steadyStateRecurring)
	}
	steadyStateRecurring = roundCurrency(steadyStateRecurring + minimumUplift)

	result.RecurringTotal = steadyStateRecurring
	result.MinimumUplift = minimumUplift

	if baseFee > 0 {
		result.Components = append(result.Components, RevenueComponent{Name: "base_platform_fee", Amount: baseFee})
	}
	if seatRevenue > 0 {
		result.Components = append(result.Components, RevenueComponent{Name: "seat_revenue", Amount: seatRevenue})
	}
	if minimumUplift > 0 {
		result.Components = append(result.Components, RevenueComponent{Name: "minimum_uplift", Amount: minimumUplift})
	}

	if contract.Pilot != nil {
		pilotMonthlyFee := roundCurrency(contract.Pilot.MonthlyFee)
		setupFee := roundCurrency(contract.Pilot.SetupFee)

		if pilotMonthlyFee > 0 {
			result.Components = append(result.Components, RevenueComponent{Name: "pilot_monthly_fee", Amount: pilotMonthlyFee})
		}
		if setupFee > 0 {
			result.Components = append(result.Components, RevenueComponent{Name: "pilot_setup_fee", Amount: setupFee})
		}

		result.OneTimeTotal = setupFee
		result.Total = roundCurrency(pilotMonthlyFee + setupFee)
		return result
	}

	result.Total = steadyStateRecurring
	return result
}

func resolvedPricingMode(pricing *config.PricingConfig) string {
	if pricing == nil {
		return "tiered"
	}

	return normalizedMode(pricing.Mode, pricing.Contract != nil)
}

func normalizedMode(mode string, hasContract bool) string {
	trimmed := strings.ToLower(strings.TrimSpace(mode))
	if trimmed == "" {
		if hasContract {
			return "hybrid"
		}
		return "tiered"
	}

	return trimmed
}

func roundCurrency(value float64) float64 {
	return math.Round(value*100) / 100
}
