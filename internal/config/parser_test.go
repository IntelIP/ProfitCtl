package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/IntelIP/ProfitCtl/pkg/types"
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
	assert.Equal(t, "tiered", cfg.Pricing.Mode)
	assert.Equal(t, "basic", cfg.Pricing.Plans[0].Name)
	assert.Equal(t, "pro", cfg.Pricing.Plans[1].Name)
	assert.Equal(t, "enterprise", cfg.Pricing.Plans[2].Name)
}

func TestParseConfig_AllowsUnlimitedFinalPlan(t *testing.T) {
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
    - name: unlimited
      price: 29

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Pricing)
	assert.Len(t, cfg.Pricing.Plans, 2)
	assert.Equal(t, "tiered", cfg.Pricing.Mode)
	assert.NotNil(t, cfg.Pricing.Plans[0].Limits)
	assert.Nil(t, cfg.Pricing.Plans[1].Limits)
}

func TestParseConfig_MixMode(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

pricing:
  mode: mix
  plans:
    - name: free
      price: 0
      share: 0.7
    - name: pro
      price: 29
      share: 0.3

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Pricing)
	assert.Equal(t, "mix", cfg.Pricing.Mode)
	assert.Len(t, cfg.Pricing.Plans, 2)
	assert.NotNil(t, cfg.Pricing.Plans[0].Share)
	assert.Nil(t, cfg.Pricing.Plans[0].Limits)
	assert.Equal(t, 0.7, *cfg.Pricing.Plans[0].Share)
}

func TestParseConfig_MixModePreservesReferenceLimits(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

pricing:
  mode: mix
  plans:
    - name: free
      price: 0
      share: 0.7
      limits:
        users: 10
    - name: pro
      price: 29
      share: 0.3
      limits:
        users: 100

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Pricing)
	assert.Equal(t, PricingModeMix, cfg.Pricing.Mode)
	assert.NotNil(t, cfg.Pricing.Plans[0].Limits)
	assert.Equal(t, 10, cfg.Pricing.Plans[0].Limits.Users)
	assert.NotNil(t, cfg.Pricing.Plans[1].Limits)
	assert.Equal(t, 100, cfg.Pricing.Plans[1].Limits.Users)
}

func TestParseConfig_HybridMode(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

pricing:
  mode: hybrid
  contract:
    base_platform_fee: 500
    per_seat_fee: 25
    workspace_minimum: 2000
    included_seats: 10
    pilot:
      setup_fee: 5000
      monthly_fee: 1500
      duration_months: 3

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Pricing)
	assert.Equal(t, "hybrid", cfg.Pricing.Mode)
	assert.NotNil(t, cfg.Pricing.Contract)
	assert.Equal(t, 500.0, cfg.Pricing.Contract.BasePlatformFee)
	assert.Equal(t, 25.0, cfg.Pricing.Contract.PerSeatFee)
	assert.Equal(t, 2000.0, cfg.Pricing.Contract.WorkspaceMinimum)
	assert.Equal(t, 10, cfg.Pricing.Contract.IncludedSeats)
	assert.NotNil(t, cfg.Pricing.Contract.Pilot)
	assert.Equal(t, 5000.0, cfg.Pricing.Contract.Pilot.SetupFee)
	assert.Equal(t, 1500.0, cfg.Pricing.Contract.Pilot.MonthlyFee)
	assert.Equal(t, 3, cfg.Pricing.Contract.Pilot.DurationMonths)
}

func TestParseConfig_WorkspaceHybridMode(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

pricing:
  mode: workspace_hybrid
  workspace:
    average_users_per_workspace: 2
  plans:
    - name: minimum
      price: 0
      share: 0.8
      workspace_minimum: 149
    - name: pro
      price: 39
      share: 0.2
      workspace_minimum: 299

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Pricing)
	assert.Equal(t, PricingModeWorkspaceHybrid, cfg.Pricing.Mode)
	assert.NotNil(t, cfg.Pricing.Workspace)
	assert.Equal(t, 2, cfg.Pricing.Workspace.AverageUsersPerWorkspace)
	assert.NotNil(t, cfg.Pricing.Plans[0].WorkspaceMinimum)
	assert.Equal(t, 149.0, *cfg.Pricing.Plans[0].WorkspaceMinimum)
}

func TestParseConfig_PaymentFees(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

payment_fees:
  processor: stripe
  currency: usd
  free_user_fee:
    monthly_percent: 0
    per_transaction: 0
  paid_user_fee:
    monthly_percent: 2.9
    annual_percent: 2.0
    per_transaction: 0.30
  billing_mix:
    monthly_share: 0.75
    annual_share: 0.25
  plan_burden:
    - plan: pro
      applies_to: annual
      fee_multiplier: 0.5
  assumptions:
    annual_discount_percent: 10
    annual_prepaid_months: 12

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.PaymentFees)
	assert.Equal(t, "stripe", cfg.PaymentFees.Processor)
	assert.Equal(t, "usd", cfg.PaymentFees.Currency)
	assert.NotNil(t, cfg.PaymentFees.PaidUserFee)
	assert.Equal(t, 2.9, cfg.PaymentFees.PaidUserFee.MonthlyPercent)
	assert.NotNil(t, cfg.PaymentFees.BillingMix)
	assert.Equal(t, 0.75, cfg.PaymentFees.BillingMix.MonthlyShare)
	assert.Len(t, cfg.PaymentFees.PlanBurden, 1)
	assert.Equal(t, 0.5, cfg.PaymentFees.PlanBurden[0].FeeMultiplier)
}

func TestParseConfig_Calibration(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

calibration:
  period: 2026-03
  source: stripe_export
  gross_revenue: 125000
  payment_fees: 4380
  free_users: 12000
  paid_users:
    monthly: 1800
    annual: 400
  plan_mix:
    starter: 0.65
    pro: 0.35
  usage:
    api_calls: 3200000

fixed_costs: []
variable_costs: []
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Calibration)
	assert.Equal(t, "2026-03", cfg.Calibration.Period)
	assert.Equal(t, 125000.0, cfg.Calibration.GrossRevenue)
	assert.NotNil(t, cfg.Calibration.PaidUsers)
	assert.Equal(t, 1800, cfg.Calibration.PaidUsers.Monthly)
	assert.Equal(t, 0.65, cfg.Calibration.PlanMix["starter"])
}

func TestParseConfig_CalibrationFile(t *testing.T) {
	tmpDir := t.TempDir()
	calibrationFile := filepath.Join(tmpDir, "calibration.yml")
	err := os.WriteFile(calibrationFile, []byte(`
period: 2026-03
source: stripe_export
gross_revenue: 125000
payment_fees: 4380
free_users: 12000
paid_users:
  monthly: 1800
  annual: 400
plan_mix:
  starter: 0.65
  pro: 0.35
`), 0644)
	assert.NoError(t, err)

	configFile := filepath.Join(tmpDir, "profit.yml")
	err = os.WriteFile(configFile, []byte(`
simulation:
  base_users: 100
  billable_users: 25
  growth_factor: 1.5
  iterations: 10000

calibration_file: calibration.yml

fixed_costs: []
variable_costs: []
`), 0644)
	assert.NoError(t, err)

	cfg, err := ParseConfig(configFile)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Calibration)
	assert.Equal(t, "2026-03", cfg.Calibration.Period)
	assert.NotNil(t, cfg.Simulation.BillableUsers)
	assert.Equal(t, 25, *cfg.Simulation.BillableUsers)
}

func TestParseConfig_CalibrationAndCalibrationFileConflict(t *testing.T) {
	tmpDir := t.TempDir()
	calibrationFile := filepath.Join(tmpDir, "calibration.yml")
	err := os.WriteFile(calibrationFile, []byte("period: 2026-03\n"), 0644)
	assert.NoError(t, err)

	configFile := filepath.Join(tmpDir, "profit.yml")
	err = os.WriteFile(configFile, []byte(`
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

calibration:
  period: 2026-03

calibration_file: calibration.yml

fixed_costs: []
variable_costs: []
`), 0644)
	assert.NoError(t, err)

	_, err = ParseConfig(configFile)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
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

func TestParseConfig_DefaultsVariableCostUserScopeToAllUsers(t *testing.T) {
	configYAML := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000

pricing:
  plans:
    - name: starter
      price: 49

fixed_costs: []
variable_costs:
  - name: Billing
    cost_per_unit: 1
    units_per_user: 1
    distribution: normal
    mean: 1
    stddev: 0.1
    layer: service
`

	tmpFile := createTempConfigFile(t, configYAML)
	cfg, err := ParseConfig(tmpFile)

	assert.NoError(t, err)
	assert.Len(t, cfg.VariableCosts, 1)
	assert.Equal(t, types.UserScopeAllUsers, cfg.VariableCosts[0].UserScope)
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
