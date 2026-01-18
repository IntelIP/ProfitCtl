package config

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/profitctl/profitctl/pkg/types"
)

var (
	ErrInvalidMean   = errors.New("mean must be > 0 for normal distribution")
	ErrInvalidStdDev = errors.New("stddev must be > 0 for normal distribution")
	ErrInvalidMin    = errors.New("min must be set and < max for uniform distribution")
	ErrInvalidMax    = errors.New("max must be set and > min for uniform distribution")
	ErrInvalidRate   = errors.New("rate must be > 0 for exponential distribution")
)

func ValidateConfig(config *Config) error {
	v := validator.New()

	if err := v.Struct(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := validateDistributionParams(config); err != nil {
		return err
	}

	if err := validatePricingConsistency(config); err != nil {
		return err
	}

	return nil
}

func validateDistributionParams(config *Config) error {
	for _, vc := range config.VariableCosts {
		switch vc.Distribution {
		case types.DistNormal:
			if vc.Mean == nil || *vc.Mean < 0 {
				return fmt.Errorf("variable cost '%s': %w", vc.Name, ErrInvalidMean)
			}
			if vc.StdDev == nil || *vc.StdDev < 0 {
				return fmt.Errorf("variable cost '%s': %w", vc.Name, ErrInvalidStdDev)
			}
		case types.DistUniform:
			if vc.Min == nil || vc.Max == nil {
				return fmt.Errorf("variable cost '%s': %w", vc.Name, ErrInvalidMin)
			}
			if *vc.Min >= *vc.Max {
				return fmt.Errorf("variable cost '%s': %w", vc.Name, ErrInvalidMax)
			}
		case types.DistExponential:
			if vc.Rate == nil || *vc.Rate < 0 {
				return fmt.Errorf("variable cost '%s': %w", vc.Name, ErrInvalidRate)
			}
		}
	}

	return nil
}

func validatePricingConsistency(config *Config) error {
	if config.Pricing == nil {
		return nil
	}

	plans := config.Pricing.Plans
	for i := 1; i < len(plans); i++ {
		if plans[i].Limits.Users <= plans[i-1].Limits.Users {
			return fmt.Errorf("plan %s (%d users) must have higher limit than %s (%d users): %w",
				plans[i].Name, plans[i].Limits.Users,
				plans[i-1].Name, plans[i-1].Limits.Users,
				errors.New("pricing inconsistency"))
		}
	}

	return nil
}
