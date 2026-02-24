package config

import (
	"fmt"
	"os"

	"github.com/IntelIP/ProfitCtl/pkg/types"
	"gopkg.in/yaml.v3"
)

// Config represents the complete profit.yml configuration.
type Config struct {
	Project       *ProjectInfo         `yaml:"project"`
	FixedCosts    []types.FixedCost    `yaml:"fixed_costs"`
	VariableCosts []types.VariableCost `yaml:"variable_costs"`
	Pricing       *PricingConfig       `yaml:"pricing"`
	Covenants     []Covenant           `yaml:"covenants"`
	Simulation    *SimulationConfig    `yaml:"simulation"`
}

type ProjectInfo struct {
	Name string `yaml:"name" validate:"required"`
}

type PricingConfig struct {
	Plans []PricingPlan `yaml:"plans" validate:"required"`
}

type PricingPlan struct {
	Name   string      `yaml:"name" validate:"required"`
	Price  float64     `yaml:"price" validate:"required,min=0"`
	Limits *PlanLimits `yaml:"limits" validate:"required"`
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
	BaseUsers    int     `yaml:"base_users" validate:"required,min=1"`
	GrowthFactor float64 `yaml:"growth_factor" validate:"required,min=1.0"`
	Iterations   int     `yaml:"iterations" validate:"required,min=1"`
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

	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}
