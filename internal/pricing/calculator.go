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
	PaidUsers      int
	Components     []RevenueComponent
	ByPlan         []PlanRevenue
}

// PlanRevenue represents revenue from a single pricing plan
type PlanRevenue struct {
	PlanName   string
	Price      float64
	Share      *float64
	Cohort     string
	Users      int
	Workspaces int
	Revenue    float64
}

type RevenueComponent struct {
	Name   string
	Amount float64
}

// CalculateRevenue calculates total revenue from pricing plans for a given number of users.
func CalculateRevenue(pricing *config.PricingConfig, users int) RevenueResult {
	return CalculateRevenueWithSeats(pricing, users, nil)
}

// CalculateRevenueWithSeats calculates revenue with an optional billable-seat override.
func CalculateRevenueWithSeats(pricing *config.PricingConfig, users int, billableUsers *int) RevenueResult {
	if pricing == nil {
		return RevenueResult{
			Mode:       config.PricingModeTiered,
			Components: []RevenueComponent{},
			ByPlan:     []PlanRevenue{},
		}
	}

	mode := resolvedPricingMode(pricing)
	if len(pricing.Plans) == 0 && mode != config.PricingModeHybrid {
		return RevenueResult{
			Mode:       mode,
			Components: []RevenueComponent{},
			ByPlan:     []PlanRevenue{},
		}
	}

	if users <= 0 && mode != config.PricingModeHybrid {
		return RevenueResult{
			Mode:       mode,
			Components: []RevenueComponent{},
			ByPlan:     []PlanRevenue{},
		}
	}

	result := RevenueResult{
		Mode:       mode,
		Components: []RevenueComponent{},
		ByPlan:     make([]PlanRevenue, 0, len(pricing.Plans)),
	}

	switch mode {
	case config.PricingModeMix:
		return calculateMixRevenue(pricing, users, result)
	case config.PricingModeWorkspaceHybrid:
		return calculateWorkspaceHybridRevenue(pricing, users, result)
	case config.PricingModeHybrid:
		hybridSeats := users
		if billableUsers != nil {
			hybridSeats = *billableUsers
		}
		return calculateHybridRevenue(pricing, users, hybridSeats, result)
	default:
		return calculateTieredRevenue(pricing, users, result)
	}
}

func calculateTieredRevenue(pricing *config.PricingConfig, users int, result RevenueResult) RevenueResult {
	remainingUsers := users
	previousLimit := 0

	for i, plan := range pricing.Plans {
		if plan.Limits == nil {
			planUsers := remainingUsers
			revenue := plan.Price * float64(planUsers)

			result.ByPlan = append(result.ByPlan, PlanRevenue{
				PlanName: plan.Name,
				Price:    plan.Price,
				Cohort:   plan.Cohort,
				Users:    planUsers,
				Revenue:  roundCurrency(revenue),
			})

			result.Total += revenue
			if isMonetizedPlan(plan) {
				result.PaidUsers += planUsers
			}
			break
		}

		planLimit := plan.Limits.Users
		tierCapacity := planLimit - previousLimit
		if tierCapacity <= 0 {
			continue
		}

		isLastPlan := i == len(pricing.Plans)-1
		planUsers := remainingUsers
		if !isLastPlan && planUsers > tierCapacity {
			planUsers = tierCapacity
		}

		revenue := plan.Price * float64(planUsers)
		result.ByPlan = append(result.ByPlan, PlanRevenue{
			PlanName: plan.Name,
			Price:    plan.Price,
			Cohort:   plan.Cohort,
			Users:    planUsers,
			Revenue:  roundCurrency(revenue),
		})

		result.Total += revenue
		if isMonetizedPlan(plan) {
			result.PaidUsers += planUsers
		}
		remainingUsers -= planUsers
		previousLimit = planLimit
		if remainingUsers <= 0 {
			break
		}
	}

	result.Total = roundCurrency(result.Total)
	result.RecurringTotal = result.Total
	return result
}

type planAllocation struct {
	index      int
	users      int
	fractional float64
}

func allocateUsersByShare(pricing *config.PricingConfig, users int) []planAllocation {
	allocations := make([]planAllocation, len(pricing.Plans))
	allocatedUsers := 0

	for i, plan := range pricing.Plans {
		share := 0.0
		if plan.Share != nil {
			share = *plan.Share
		}

		rawUsers := share * float64(users)
		baseUsers := int(math.Floor(rawUsers))
		allocations[i] = planAllocation{
			index:      i,
			users:      baseUsers,
			fractional: rawUsers - float64(baseUsers),
		}
		allocatedUsers += baseUsers
	}

	remainingUsers := users - allocatedUsers
	sort.SliceStable(allocations, func(i, j int) bool {
		return allocations[i].fractional > allocations[j].fractional
	})
	for i := 0; i < remainingUsers; i++ {
		allocations[i%len(allocations)].users++
	}
	sort.SliceStable(allocations, func(i, j int) bool {
		return allocations[i].index < allocations[j].index
	})

	return allocations
}

func calculateMixRevenue(pricing *config.PricingConfig, users int, result RevenueResult) RevenueResult {
	allocations := allocateUsersByShare(pricing, users)
	for _, allocation := range allocations {
		plan := pricing.Plans[allocation.index]
		revenue := plan.Price * float64(allocation.users)
		result.ByPlan = append(result.ByPlan, PlanRevenue{
			PlanName: plan.Name,
			Price:    plan.Price,
			Share:    plan.Share,
			Cohort:   plan.Cohort,
			Users:    allocation.users,
			Revenue:  roundCurrency(revenue),
		})
		result.Total += revenue
		if isMonetizedPlan(plan) {
			result.PaidUsers += allocation.users
		}
	}

	result.Total = roundCurrency(result.Total)
	result.RecurringTotal = result.Total
	return result
}

func calculateWorkspaceHybridRevenue(pricing *config.PricingConfig, users int, result RevenueResult) RevenueResult {
	allocations := allocateUsersByShare(pricing, users)
	avgUsersPerWorkspace := 1
	if pricing.Workspace != nil && pricing.Workspace.AverageUsersPerWorkspace > 0 {
		avgUsersPerWorkspace = pricing.Workspace.AverageUsersPerWorkspace
	}

	for _, allocation := range allocations {
		plan := pricing.Plans[allocation.index]
		workspaces := 0
		if allocation.users > 0 {
			workspaces = int(math.Ceil(float64(allocation.users) / float64(avgUsersPerWorkspace)))
		}

		seatRevenue := plan.Price * float64(allocation.users)
		minimumRevenue := 0.0
		if plan.WorkspaceMinimum != nil {
			minimumRevenue = *plan.WorkspaceMinimum * float64(workspaces)
		}
		revenue := math.Max(seatRevenue, minimumRevenue)

		result.ByPlan = append(result.ByPlan, PlanRevenue{
			PlanName:   plan.Name,
			Price:      plan.Price,
			Share:      plan.Share,
			Cohort:     plan.Cohort,
			Users:      allocation.users,
			Workspaces: workspaces,
			Revenue:    roundCurrency(revenue),
		})
		result.Total += revenue
		if isMonetizedPlan(plan) {
			result.PaidUsers += allocation.users
		}
	}

	result.Total = roundCurrency(result.Total)
	result.RecurringTotal = result.Total
	return result
}

func calculateHybridRevenue(pricing *config.PricingConfig, users int, hybridSeats int, result RevenueResult) RevenueResult {
	if pricing.Contract == nil {
		return result
	}

	contract := pricing.Contract
	billableSeats := hybridSeats - contract.IncludedSeats
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
	if users > 0 && steadyStateRecurring > 0 {
		result.PaidUsers = users
	}

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
		return config.PricingModeTiered
	}
	return normalizedMode(pricing.Mode, pricing.Contract != nil)
}

func normalizedMode(mode string, hasContract bool) string {
	trimmed := strings.ToLower(strings.TrimSpace(mode))
	if trimmed == "" {
		if hasContract {
			return config.PricingModeHybrid
		}
		return config.PricingModeTiered
	}
	return trimmed
}

func isMonetizedPlan(plan config.PricingPlan) bool {
	if plan.Price > 0 {
		return true
	}
	return plan.WorkspaceMinimum != nil && *plan.WorkspaceMinimum > 0
}

func roundCurrency(value float64) float64 {
	return math.Round(value*100) / 100
}
