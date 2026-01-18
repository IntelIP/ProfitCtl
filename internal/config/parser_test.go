package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/profitctl/profitctl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func createTempConfigFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_config.yml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	assert.NoError(t, err)
	return tmpFile
}

func TestParseConfig_ValidConfig(t *testing.T) {
	configYAML := `
project:
  name: Test Project

fixed_costs:
  - name: Server
    amount: 500
    period: monthly
    layer: infrastructure

variable_costs:
  - name: API Calls
    cost_per_unit: 0.0001
    units_per_user: 10000
    distribution: normal
    mean: 10000
    stddev: 2000
    layer: application

pricing:
  plans:
    - name: basic
      price: 10
      limits:
        users: 1000

covenants:
  - type: threshold
    field: margin
    operator: gte
    value: 20
    message: Gross margin must be >= 20%

simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "Test Project", cfg.Project.Name)
	assert.Len(t, cfg.FixedCosts, 1)
	assert.Equal(t, "Server", cfg.FixedCosts[0].Name)
	assert.Equal(t, 500.0, cfg.FixedCosts[0].Amount)
	assert.Len(t, cfg.VariableCosts, 1)
	assert.NotNil(t, cfg.Pricing)
	assert.Len(t, cfg.Pricing.Plans, 1)
	assert.Len(t, cfg.Covenants, 1)
	assert.NotNil(t, cfg.Simulation)
	assert.Equal(t, 100, cfg.Simulation.BaseUsers)
}

func TestParseConfig_MinimalConfig(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Simulation)
	assert.Empty(t, cfg.FixedCosts)
	assert.Empty(t, cfg.VariableCosts)
}

func TestParseConfig_AllCostTypes(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

fixed_costs:
  - name: Monthly Server
    amount: 500
    period: monthly
    layer: infrastructure
  - name: Yearly License
    amount: 12000
    period: yearly
    layer: application
  - name: Daily Backup
    amount: 10
    period: daily
    layer: service

variable_costs:
  - name: Normal Distribution
    cost_per_unit: 0.01
    units_per_user: 100
    distribution: normal
    mean: 100
    stddev: 10
    layer: infrastructure
  - name: Uniform Distribution
    cost_per_unit: 0.005
    units_per_user: 50
    distribution: uniform
    min: 40
    max: 60
    layer: application
  - name: Exponential Distribution
    cost_per_unit: 0.001
    units_per_user: 10
    distribution: exponential
    rate: 0.1
    layer: service
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Len(t, cfg.FixedCosts, 3)
	assert.Len(t, cfg.VariableCosts, 3)

	// Verify periods
	assert.Equal(t, types.PeriodMonthly, cfg.FixedCosts[0].Period)
	assert.Equal(t, types.PeriodYearly, cfg.FixedCosts[1].Period)
	assert.Equal(t, types.PeriodDaily, cfg.FixedCosts[2].Period)

	// Verify distributions
	assert.Equal(t, types.DistNormal, cfg.VariableCosts[0].Distribution)
	assert.Equal(t, types.DistUniform, cfg.VariableCosts[1].Distribution)
	assert.Equal(t, types.DistExponential, cfg.VariableCosts[2].Distribution)
}

func TestParseConfig_AllPricingTiers(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

pricing:
  plans:
    - name: basic
      price: 10
      limits:
        users: 1000
    - name: pro
      price: 29
      limits:
        users: 5000
    - name: enterprise
      price: 99
      limits:
        users: 10000

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Pricing)
	assert.Len(t, cfg.Pricing.Plans, 3)
	assert.Equal(t, "basic", cfg.Pricing.Plans[0].Name)
	assert.Equal(t, "pro", cfg.Pricing.Plans[1].Name)
	assert.Equal(t, "enterprise", cfg.Pricing.Plans[2].Name)
}

func TestParseConfig_MultipleCovenants(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

covenants:
  - type: threshold
    field: margin
    operator: gte
    value: 20
    message: Gross margin must be >= 20%
  - type: threshold
    field: cost_per_user
    operator: lte
    value: 5
    message: Cost per user must be <= $5
  - type: threshold
    field: p95_margin
    operator: gte
    value: 15
    message: p95 margin must be >= 15%

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Len(t, cfg.Covenants, 3)
	assert.Equal(t, "margin", cfg.Covenants[0].Field)
	assert.Equal(t, "cost_per_user", cfg.Covenants[1].Field)
	assert.Equal(t, "p95_margin", cfg.Covenants[2].Field)
}

func TestParseConfig_FileNotFound(t *testing.T) {
	_, err := ParseConfig("/nonexistent/file.yml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestParseConfig_InvalidYAML(t *testing.T) {
	invalidYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  invalid: [unclosed bracket
`

	tmpFile := createTempConfigFile(t, invalidYAML)
	_, err := ParseConfig(tmpFile)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestParseConfig_EmptyConfig(t *testing.T) {
	emptyYAML := ""
	tmpFile := createTempConfigFile(t, emptyYAML)
	cfg, err := ParseConfig(tmpFile)

	// Empty YAML should not error, but may have empty slices
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestParseConfig_OptionalFieldsOmitted(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	// Optional fields should be nil or empty
	assert.Nil(t, cfg.Project)
	assert.Nil(t, cfg.Pricing)
	assert.Empty(t, cfg.Covenants)
}

func TestParseConfig_PricingPlanWithoutLimits(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

pricing:
  plans:
    - name: unlimited
      price: 50

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Pricing)
	assert.Len(t, cfg.Pricing.Plans, 1)
	assert.Nil(t, cfg.Pricing.Plans[0].Limits) // Should allow nil limits
}

func TestParseConfig_AllLayers(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

fixed_costs:
  - name: Infrastructure Cost
    amount: 100
    period: monthly
    layer: infrastructure
  - name: Application Cost
    amount: 200
    period: monthly
    layer: application
  - name: Service Cost
    amount: 300
    period: monthly
    layer: service
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Len(t, cfg.FixedCosts, 3)
	assert.Equal(t, types.LayerInfrastructure, cfg.FixedCosts[0].Layer)
	assert.Equal(t, types.LayerApplication, cfg.FixedCosts[1].Layer)
	assert.Equal(t, types.LayerService, cfg.FixedCosts[2].Layer)
}