package config

import (
	"errors"
	"fmt"

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
	v := validator.New()

	if err := v.Struct(cfg); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := validateDistributionParams(cfg); err != nil {
		return err
	}

	if err := validatePricingConsistency(cfg); err != nil {
		return err
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
	if cfg.Pricing == nil || len(cfg.Pricing.Plans) < 2 {
		return nil
	}

	plans := cfg.Pricing.Plans
	for i := 1; i < len(plans); i++ {
		if plans[i-1].Limits == nil || plans[i].Limits == nil {
			return fmt.Errorf("plan limits are required: %w", ErrPricingInconsistency)
		}
		if plans[i].Limits.Users <= plans[i-1].Limits.Users {
			return fmt.Errorf("plan %s (%d users) must have higher limit than %s (%d users): %w",
				plans[i].Name, plans[i].Limits.Users,
				plans[i-1].Name, plans[i-1].Limits.Users,
				ErrPricingInconsistency,
			)
		}
	}

	return nil
}
