package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/IntelIP/ProfitCtl/pkg/types"
	"gopkg.in/yaml.v3"
)

// Config represents the complete profit.yml configuration.
type Config struct {
	Project         *ProjectInfo         `yaml:"project"`
	FixedCosts      []types.FixedCost    `yaml:"fixed_costs"`
	VariableCosts   []types.VariableCost `yaml:"variable_costs"`
	Pricing         *PricingConfig       `yaml:"pricing"`
	PaymentFees     *PaymentFeesConfig   `yaml:"payment_fees"`
	Calibration     *CalibrationConfig   `yaml:"calibration"`
	CalibrationFile string               `yaml:"calibration_file"`
	Covenants       []Covenant           `yaml:"covenants"`
	Simulation      *SimulationConfig    `yaml:"simulation"`
}

type ProjectInfo struct {
	Name string `yaml:"name" validate:"required"`
}

type PricingConfig struct {
	Mode     string          `yaml:"mode"`
	Plans    []PricingPlan   `yaml:"plans"`
	Contract *HybridContract `yaml:"contract"`
}

type PricingPlan struct {
	Name   string      `yaml:"name" validate:"required"`
	Price  float64     `yaml:"price" validate:"required,min=0"`
	Share  *float64    `yaml:"share"`
	Cohort string      `yaml:"cohort"`
	Limits *PlanLimits `yaml:"limits"`
}

type HybridContract struct {
	BasePlatformFee  float64      `yaml:"base_platform_fee"`
	PerSeatFee       float64      `yaml:"per_seat_fee"`
	WorkspaceMinimum float64      `yaml:"workspace_minimum"`
	IncludedSeats    int          `yaml:"included_seats"`
	Pilot            *PilotConfig `yaml:"pilot"`
}

type PilotConfig struct {
	SetupFee       float64 `yaml:"setup_fee"`
	MonthlyFee     float64 `yaml:"monthly_fee"`
	DurationMonths int     `yaml:"duration_months"`
}

type PaymentFeesConfig struct {
	Processor           string                 `yaml:"processor"`
	Currency            string                 `yaml:"currency"`
	FreeUserFee         *PaymentFeeProfile     `yaml:"free_user_fee"`
	PaidUserFee         *PaymentFeeProfile     `yaml:"paid_user_fee"`
	BillingMix          *BillingMixConfig      `yaml:"billing_mix"`
	PlanBurden          []PlanBurdenConfig     `yaml:"plan_burden"`
	Assumptions         *PaymentFeeAssumptions `yaml:"assumptions"`
	PercentOfRevenue    float64                `yaml:"percent_of_revenue"`
	FixedPerPayment     float64                `yaml:"fixed_per_payment"`
	MonthlyTransactions int                    `yaml:"monthly_transactions"`
}

type PaymentFeeProfile struct {
	MonthlyPercent float64 `yaml:"monthly_percent"`
	AnnualPercent  float64 `yaml:"annual_percent"`
	PerTransaction float64 `yaml:"per_transaction"`
}

type BillingMixConfig struct {
	MonthlyShare float64 `yaml:"monthly_share"`
	AnnualShare  float64 `yaml:"annual_share"`
}

type PlanBurdenConfig struct {
	Plan          string  `yaml:"plan"`
	AppliesTo     string  `yaml:"applies_to"`
	FeeMultiplier float64 `yaml:"fee_multiplier"`
}

type PaymentFeeAssumptions struct {
	AnnualDiscountPercent float64 `yaml:"annual_discount_percent"`
	AnnualPrepaidMonths   int     `yaml:"annual_prepaid_months"`
}

type CalibrationConfig struct {
	Period       string                `yaml:"period"`
	Source       string                `yaml:"source"`
	GrossRevenue float64               `yaml:"gross_revenue"`
	PaymentFees  float64               `yaml:"payment_fees"`
	FreeUsers    int                   `yaml:"free_users"`
	PaidUsers    *CalibrationPaidUsers `yaml:"paid_users"`
	PlanMix      map[string]float64    `yaml:"plan_mix"`
	Usage        map[string]float64    `yaml:"usage"`
}

type CalibrationPaidUsers struct {
	Monthly int `yaml:"monthly"`
	Annual  int `yaml:"annual"`
}

type PlanLimits struct {
	Users int `yaml:"users" validate:"required,min=1"`
}

type Covenant struct {
	Type     string  `yaml:"type" validate:"required,oneof=threshold"`
	Field    string  `yaml:"field,omitempty" validate:"required_if=Type threshold"`
	Operator string  `yaml:"operator,omitempty" validate:"required_if=Type threshold,oneof=gt lt gte lte eq"`
	Value    float64 `yaml:"value,omitempty" validate:"required_if=Type threshold"`
	Message  string  `yaml:"message" validate:"required"`
}

type SimulationConfig struct {
	BaseUsers     int     `yaml:"base_users" validate:"required,min=1"`
	BillableUsers *int    `yaml:"billable_users" validate:"omitempty,min=0"`
	GrowthFactor  float64 `yaml:"growth_factor" validate:"required,min=1.0"`
	Iterations    int     `yaml:"iterations" validate:"required,min=1"`
}

// ParseConfig reads and validates a profit.yml configuration file.
func ParseConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	normalizeConfig(&cfg)

	if cfg.Calibration != nil && strings.TrimSpace(cfg.CalibrationFile) != "" {
		return nil, fmt.Errorf("failed to validate config: calibration and calibration_file are mutually exclusive")
	}

	if strings.TrimSpace(cfg.CalibrationFile) != "" {
		calibrationPath := cfg.CalibrationFile
		if !filepath.IsAbs(calibrationPath) {
			calibrationPath = filepath.Join(filepath.Dir(filename), calibrationPath)
		}

		calibration, err := LoadCalibrationFile(calibrationPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load calibration file: %w", err)
		}
		cfg.Calibration = calibration
	}

	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}

func normalizeConfig(cfg *Config) {
	if cfg.Pricing == nil {
		return
	}

	cfg.Pricing.Mode = normalizedPricingMode(cfg.Pricing.Mode, cfg.Pricing)
	for i := range cfg.Pricing.Plans {
		cfg.Pricing.Plans[i].Cohort = strings.ToLower(strings.TrimSpace(cfg.Pricing.Plans[i].Cohort))
	}
	if cfg.PaymentFees == nil {
		return
	}

	if strings.TrimSpace(cfg.PaymentFees.Processor) == "" {
		cfg.PaymentFees.Processor = "stripe"
	}
	if strings.TrimSpace(cfg.PaymentFees.Currency) == "" {
		cfg.PaymentFees.Currency = "usd"
	}
	if cfg.PaymentFees.BillingMix == nil {
		cfg.PaymentFees.BillingMix = &BillingMixConfig{MonthlyShare: 1.0, AnnualShare: 0}
	}
	if cfg.PaymentFees.Assumptions == nil {
		cfg.PaymentFees.Assumptions = &PaymentFeeAssumptions{AnnualPrepaidMonths: 12}
	}
	if cfg.PaymentFees.Assumptions.AnnualPrepaidMonths == 0 {
		cfg.PaymentFees.Assumptions.AnnualPrepaidMonths = 12
	}
	for i := range cfg.PaymentFees.PlanBurden {
		cfg.PaymentFees.PlanBurden[i].AppliesTo = strings.ToLower(strings.TrimSpace(cfg.PaymentFees.PlanBurden[i].AppliesTo))
		if cfg.PaymentFees.PlanBurden[i].FeeMultiplier == 0 {
			cfg.PaymentFees.PlanBurden[i].FeeMultiplier = 1.0
		}
	}
}

func normalizedPricingMode(mode string, pricing *PricingConfig) string {
	if strings.TrimSpace(mode) == "" {
		if pricing != nil && pricing.Contract != nil {
			return "hybrid"
		}
		return "tiered"
	}

	return strings.ToLower(mode)
}
