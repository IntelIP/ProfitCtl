package config

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/IntelIP/ProfitCtl/pkg/types"
	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidMean          = errors.New("mean must be > 0 for normal distribution")
	ErrInvalidStdDev        = errors.New("stddev must be > 0 for normal distribution")
	ErrInvalidMin           = errors.New("min must be set and < max for uniform distribution")
	ErrInvalidMax           = errors.New("max must be set and > min for uniform distribution")
	ErrInvalidRate          = errors.New("rate must be > 0 for exponential distribution")
	ErrPricingInconsistency = errors.New("pricing inconsistency")
)

func ValidateConfig(cfg *Config) error {
	normalizeConfig(cfg)

	v := validator.New()

	if err := v.Struct(cfg); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := validateNestedStructs(cfg, v); err != nil {
		return err
	}

	if err := validateDistributionParams(cfg); err != nil {
		return err
	}

	if err := validatePricingConsistency(cfg); err != nil {
		return err
	}

	if err := validatePaymentFees(cfg); err != nil {
		return err
	}

	if err := validateCalibration(cfg); err != nil {
		return err
	}

	return nil
}

func validateNestedStructs(cfg *Config, v *validator.Validate) error {
	if cfg.Project != nil {
		if err := v.Struct(cfg.Project); err != nil {
			return fmt.Errorf("project validation failed: %w", err)
		}
	}

	for _, fixedCost := range cfg.FixedCosts {
		if err := v.Struct(fixedCost); err != nil {
			return fmt.Errorf("fixed cost %q validation failed: %w", fixedCost.Name, err)
		}
		if fixedCost.Allocation != nil {
			if err := v.Struct(fixedCost.Allocation); err != nil {
				return fmt.Errorf("fixed cost %q allocation validation failed: %w", fixedCost.Name, err)
			}
		}
	}

	for _, variableCost := range cfg.VariableCosts {
		if err := v.Struct(variableCost); err != nil {
			return fmt.Errorf("variable cost %q validation failed: %w", variableCost.Name, err)
		}
		if variableCost.Allocation != nil {
			if err := v.Struct(variableCost.Allocation); err != nil {
				return fmt.Errorf("variable cost %q allocation validation failed: %w", variableCost.Name, err)
			}
		}
	}

	if cfg.Pricing != nil {
		if err := v.Struct(cfg.Pricing); err != nil {
			return fmt.Errorf("pricing validation failed: %w", err)
		}
		if cfg.Pricing.Workspace != nil {
			if err := v.Struct(cfg.Pricing.Workspace); err != nil {
				return fmt.Errorf("workspace pricing validation failed: %w", err)
			}
		}
		for _, plan := range cfg.Pricing.Plans {
			if err := v.Struct(plan); err != nil {
				return fmt.Errorf("pricing plan %q validation failed: %w", plan.Name, err)
			}
		}
	}

	for _, covenant := range cfg.Covenants {
		if err := v.Struct(covenant); err != nil {
			return fmt.Errorf("covenant %q validation failed: %w", covenant.Message, err)
		}
	}

	if cfg.Simulation != nil {
		if err := v.Struct(cfg.Simulation); err != nil {
			return fmt.Errorf("simulation validation failed: %w", err)
		}
	}

	return nil
}

func validateDistributionParams(cfg *Config) error {
	for _, vc := range cfg.VariableCosts {
		switch vc.Distribution {
		case types.DistNormal:
			if vc.Mean == nil || *vc.Mean <= 0 {
				return fmt.Errorf("variable cost %q: %w", vc.Name, ErrInvalidMean)
			}
			if vc.StdDev == nil || *vc.StdDev <= 0 {
				return fmt.Errorf("variable cost %q: %w", vc.Name, ErrInvalidStdDev)
			}
		case types.DistUniform:
			if vc.Min == nil || vc.Max == nil {
				return fmt.Errorf("variable cost %q: %w", vc.Name, ErrInvalidMin)
			}
			if *vc.Min >= *vc.Max {
				return fmt.Errorf("variable cost %q: %w", vc.Name, ErrInvalidMax)
			}
		case types.DistExponential:
			if vc.Rate == nil || *vc.Rate <= 0 {
				return fmt.Errorf("variable cost %q: %w", vc.Name, ErrInvalidRate)
			}
		}
	}

	return nil
}

func validatePricingConsistency(cfg *Config) error {
	for _, variableCost := range cfg.VariableCosts {
		if variableCost.UserScope == types.UserScopePaidUsers {
			if cfg.Pricing == nil {
				return fmt.Errorf("variable cost %q uses user_scope=paid_users but pricing is not configured: %w", variableCost.Name, ErrPricingInconsistency)
			}
		}
	}
	if cfg.Pricing == nil {
		return nil
	}

	switch normalizedPricingMode(cfg.Pricing.Mode, cfg.Pricing) {
	case PricingModeTiered:
		return validateTieredPricing(cfg.Pricing)
	case PricingModeMix:
		return validateMixPricing(cfg.Pricing)
	case PricingModeHybrid:
		return validateHybridPricing(cfg.Pricing)
	case PricingModeWorkspaceHybrid:
		return validateWorkspaceHybridPricing(cfg.Pricing)
	default:
		return fmt.Errorf("unsupported pricing mode %q: %w", cfg.Pricing.Mode, ErrPricingInconsistency)
	}
}

func validateTieredPricing(pricing *PricingConfig) error {
	if len(pricing.Plans) == 0 {
		return fmt.Errorf("tiered pricing requires plans: %w", ErrPricingInconsistency)
	}
	if pricing.Contract != nil {
		return fmt.Errorf("tiered pricing does not support contract: %w", ErrPricingInconsistency)
	}

	plans := pricing.Plans
	for i, plan := range plans {
		if err := validatePricingPlanCohort(plan); err != nil {
			return err
		}
		if plan.Share != nil {
			return fmt.Errorf("tiered pricing does not support share on plan %s: %w", plan.Name, ErrPricingInconsistency)
		}
		if plan.WorkspaceMinimum != nil {
			return fmt.Errorf("tiered pricing does not support workspace_minimum on plan %s: %w", plan.Name, ErrPricingInconsistency)
		}
		if plan.Limits == nil {
			if i != len(plans)-1 {
				return fmt.Errorf("only the last pricing plan may omit limits: %w", ErrPricingInconsistency)
			}
			continue
		}

		if i == 0 {
			continue
		}

		previousPlan := plans[i-1]
		if previousPlan.Limits == nil {
			return fmt.Errorf("unlimited pricing plan must be last: %w", ErrPricingInconsistency)
		}
		if plan.Limits.Users <= previousPlan.Limits.Users {
			return fmt.Errorf("plan %s (%d users) must have higher limit than %s (%d users): %w",
				plan.Name, plan.Limits.Users,
				previousPlan.Name, previousPlan.Limits.Users,
				ErrPricingInconsistency,
			)
		}
	}

	return nil
}

func validateMixPricing(pricing *PricingConfig) error {
	const shareTolerance = 1e-9
	if len(pricing.Plans) == 0 {
		return fmt.Errorf("pricing mode mix requires plans: %w", ErrPricingInconsistency)
	}
	if pricing.Contract != nil {
		return fmt.Errorf("pricing mode mix does not support contract: %w", ErrPricingInconsistency)
	}

	plans := pricing.Plans
	totalShare := 0.0
	for _, plan := range plans {
		if err := validatePricingPlanCohort(plan); err != nil {
			return err
		}
		if plan.Limits != nil {
			return fmt.Errorf("pricing mode mix does not support limits on plan %s: %w", plan.Name, ErrPricingInconsistency)
		}
		if plan.Share == nil {
			return fmt.Errorf("pricing mode mix requires share on each plan: %w", ErrPricingInconsistency)
		}
		if plan.WorkspaceMinimum != nil {
			return fmt.Errorf("pricing mode mix does not support workspace_minimum on plan %s: %w", plan.Name, ErrPricingInconsistency)
		}
		if *plan.Share <= 0 || *plan.Share > 1 {
			return fmt.Errorf("pricing share for plan %s must be > 0 and <= 1: %w", plan.Name, ErrPricingInconsistency)
		}

		totalShare += *plan.Share
	}

	if math.Abs(totalShare-1.0) > shareTolerance {
		return fmt.Errorf("pricing shares must sum to 1.0: %w", ErrPricingInconsistency)
	}

	return nil
}

func validateWorkspaceHybridPricing(pricing *PricingConfig) error {
	if pricing.Contract != nil {
		return fmt.Errorf("pricing mode workspace_hybrid does not support contract: %w", ErrPricingInconsistency)
	}
	if pricing.Workspace == nil || pricing.Workspace.AverageUsersPerWorkspace < 1 {
		return fmt.Errorf("pricing mode workspace_hybrid requires workspace.average_users_per_workspace >= 1: %w", ErrPricingInconsistency)
	}
	if len(pricing.Plans) == 0 {
		return fmt.Errorf("pricing mode workspace_hybrid requires plans: %w", ErrPricingInconsistency)
	}

	const shareTolerance = 1e-9
	totalShare := 0.0
	for _, plan := range pricing.Plans {
		if err := validatePricingPlanCohort(plan); err != nil {
			return err
		}
		if plan.Limits != nil {
			return fmt.Errorf("pricing mode workspace_hybrid does not support limits on plan %s: %w", plan.Name, ErrPricingInconsistency)
		}
		if plan.Share == nil {
			return fmt.Errorf("pricing mode workspace_hybrid requires share on each plan: %w", ErrPricingInconsistency)
		}
		if *plan.Share <= 0 || *plan.Share > 1 {
			return fmt.Errorf("pricing share for plan %s must be > 0 and <= 1: %w", plan.Name, ErrPricingInconsistency)
		}
		if plan.WorkspaceMinimum != nil && *plan.WorkspaceMinimum < 0 {
			return fmt.Errorf("pricing workspace_minimum for plan %s must be >= 0: %w", plan.Name, ErrPricingInconsistency)
		}

		totalShare += *plan.Share
	}

	if math.Abs(totalShare-1.0) > shareTolerance {
		return fmt.Errorf("pricing shares must sum to 1.0: %w", ErrPricingInconsistency)
	}

	return nil
}

func validateHybridPricing(pricing *PricingConfig) error {
	if pricing.Contract == nil {
		return fmt.Errorf("pricing mode hybrid requires contract: %w", ErrPricingInconsistency)
	}
	if len(pricing.Plans) > 0 {
		return fmt.Errorf("pricing mode hybrid does not support plans: %w", ErrPricingInconsistency)
	}

	contract := pricing.Contract
	if contract.BasePlatformFee < 0 {
		return fmt.Errorf("hybrid pricing base_platform_fee must be >= 0: %w", ErrPricingInconsistency)
	}
	if contract.PerSeatFee < 0 {
		return fmt.Errorf("hybrid pricing per_seat_fee must be >= 0: %w", ErrPricingInconsistency)
	}
	if contract.WorkspaceMinimum < 0 {
		return fmt.Errorf("hybrid pricing workspace_minimum must be >= 0: %w", ErrPricingInconsistency)
	}
	if contract.IncludedSeats < 0 {
		return fmt.Errorf("hybrid pricing included_seats must be >= 0: %w", ErrPricingInconsistency)
	}
	if contract.Pilot == nil {
		return nil
	}
	if contract.Pilot.SetupFee < 0 {
		return fmt.Errorf("hybrid pricing pilot setup_fee must be >= 0: %w", ErrPricingInconsistency)
	}
	if contract.Pilot.MonthlyFee < 0 {
		return fmt.Errorf("hybrid pricing pilot monthly_fee must be >= 0: %w", ErrPricingInconsistency)
	}
	if contract.Pilot.DurationMonths < 1 {
		return fmt.Errorf("hybrid pricing pilot duration_months must be >= 1: %w", ErrPricingInconsistency)
	}

	return nil
}

func validatePaymentFees(cfg *Config) error {
	if cfg.PaymentFees == nil {
		return nil
	}

	fees := cfg.PaymentFees
	if fees.FreeUserFee == nil && fees.PaidUserFee == nil && fees.PercentOfRevenue == 0 && fees.FixedPerPayment == 0 && fees.MonthlyTransactions == 0 {
		return fmt.Errorf("payment_fees requires either legacy fee inputs or cohort fee profiles: %w", ErrPricingInconsistency)
	}
	if fees.PercentOfRevenue < 0 {
		return fmt.Errorf("payment_fees percent_of_revenue must be >= 0: %w", ErrPricingInconsistency)
	}
	if fees.FixedPerPayment < 0 {
		return fmt.Errorf("payment_fees fixed_per_payment must be >= 0: %w", ErrPricingInconsistency)
	}
	if fees.MonthlyTransactions < 0 {
		return fmt.Errorf("payment_fees monthly_transactions must be >= 0: %w", ErrPricingInconsistency)
	}
	if fees.BillingMix != nil {
		if fees.BillingMix.MonthlyShare < 0 || fees.BillingMix.AnnualShare < 0 {
			return fmt.Errorf("payment_fees billing_mix shares must be >= 0: %w", ErrPricingInconsistency)
		}
		if math.Abs((fees.BillingMix.MonthlyShare+fees.BillingMix.AnnualShare)-1.0) > 1e-9 {
			return fmt.Errorf("payment_fees billing_mix shares must sum to 1.0: %w", ErrPricingInconsistency)
		}
	}
	if fees.Assumptions != nil {
		if fees.Assumptions.AnnualDiscountPercent < 0 {
			return fmt.Errorf("payment_fees annual_discount_percent must be >= 0: %w", ErrPricingInconsistency)
		}
		if fees.Assumptions.AnnualPrepaidMonths < 1 {
			return fmt.Errorf("payment_fees annual_prepaid_months must be >= 1: %w", ErrPricingInconsistency)
		}
	}
	if err := validatePaymentFeeProfile("free_user_fee", fees.FreeUserFee); err != nil {
		return err
	}
	if err := validatePaymentFeeProfile("paid_user_fee", fees.PaidUserFee); err != nil {
		return err
	}
	for _, burden := range fees.PlanBurden {
		if strings.TrimSpace(burden.Plan) == "" {
			return fmt.Errorf("payment_fees plan_burden requires plan: %w", ErrPricingInconsistency)
		}
		switch burden.AppliesTo {
		case "", "all", "monthly", "annual":
		default:
			return fmt.Errorf("payment_fees plan_burden applies_to must be one of all, monthly, annual: %w", ErrPricingInconsistency)
		}
		if burden.FeeMultiplier < 0 {
			return fmt.Errorf("payment_fees plan_burden fee_multiplier must be >= 0: %w", ErrPricingInconsistency)
		}
	}

	return nil
}

func validateCalibration(cfg *Config) error {
	if cfg.Calibration == nil {
		return nil
	}

	calibration := cfg.Calibration
	if calibration.GrossRevenue < 0 {
		return fmt.Errorf("calibration gross_revenue must be >= 0: %w", ErrPricingInconsistency)
	}
	if calibration.PaymentFees < 0 {
		return fmt.Errorf("calibration payment_fees must be >= 0: %w", ErrPricingInconsistency)
	}
	if calibration.FreeUsers < 0 {
		return fmt.Errorf("calibration free_users must be >= 0: %w", ErrPricingInconsistency)
	}
	if calibration.PaidUsers != nil {
		if calibration.PaidUsers.Monthly < 0 || calibration.PaidUsers.Annual < 0 {
			return fmt.Errorf("calibration paid_users must be >= 0: %w", ErrPricingInconsistency)
		}
	}
	if len(calibration.PlanMix) > 0 {
		total := 0.0
		for plan, share := range calibration.PlanMix {
			if share < 0 || share > 1 {
				return fmt.Errorf("calibration plan_mix share for %s must be between 0 and 1: %w", plan, ErrPricingInconsistency)
			}
			total += share
		}
		if math.Abs(total-1.0) > 1e-9 {
			return fmt.Errorf("calibration plan_mix shares must sum to 1.0: %w", ErrPricingInconsistency)
		}
	}
	for key, value := range calibration.Usage {
		if value < 0 {
			return fmt.Errorf("calibration usage value for %s must be >= 0: %w", key, ErrPricingInconsistency)
		}
	}

	return nil
}

func validatePaymentFeeProfile(name string, profile *PaymentFeeProfile) error {
	if profile == nil {
		return nil
	}
	if profile.MonthlyPercent < 0 {
		return fmt.Errorf("payment_fees %s monthly_percent must be >= 0: %w", name, ErrPricingInconsistency)
	}
	if profile.AnnualPercent < 0 {
		return fmt.Errorf("payment_fees %s annual_percent must be >= 0: %w", name, ErrPricingInconsistency)
	}
	if profile.PerTransaction < 0 {
		return fmt.Errorf("payment_fees %s per_transaction must be >= 0: %w", name, ErrPricingInconsistency)
	}

	return nil
}

func validatePricingPlanCohort(plan PricingPlan) error {
	switch plan.Cohort {
	case "", "free", "paid":
		return nil
	default:
		return fmt.Errorf("plan %s cohort must be free or paid: %w", plan.Name, ErrPricingInconsistency)
	}
}
