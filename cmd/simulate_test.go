package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func createTempConfigFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_config.yml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	assert.NoError(t, err)
	return tmpFile
}

func createValidConfig() string {
	return `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 1000

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
    value: -1000
    message: Gross margin must be >= 20%
`
}

func TestRunSimulate_ValidConfig(t *testing.T) {
	tmpFile := createTempConfigFile(t, createValidConfig())

	cmd := &cobra.Command{}
	cmd.Flags().StringP("file", "f", tmpFile, "Config file")

	err := runSimulate(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunSimulate_MissingConfigFile(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("file", "f", "/nonexistent/file.yml", "Config file")

	err := runSimulate(cmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
	var exitErr *ExitError
	assert.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
}

func TestRunSimulate_InvalidYAML(t *testing.T) {
	invalidYAML := `
simulation:
  base_users: 100
  invalid: [unclosed bracket
`
	tmpFile := createTempConfigFile(t, invalidYAML)

	cmd := &cobra.Command{}
	cmd.Flags().StringP("file", "f", tmpFile, "Config file")

	err := runSimulate(cmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestRunSimulate_MissingSimulationConfig(t *testing.T) {
	configWithoutSimulation := `
fixed_costs: []
variable_costs: []
`
	tmpFile := createTempConfigFile(t, configWithoutSimulation)

	cmd := &cobra.Command{}
	cmd.Flags().StringP("file", "f", tmpFile, "Config file")

	err := runSimulate(cmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulation configuration is required")
}

func TestRunSimulate_NoPricingConfig(t *testing.T) {
	configNoPricing := `
simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 1000

fixed_costs:
  - name: Server
    amount: 500
    period: monthly
    layer: infrastructure

variable_costs: []
`
	tmpFile := createTempConfigFile(t, configNoPricing)

	cmd := &cobra.Command{}
	cmd.Flags().StringP("file", "f", tmpFile, "Config file")

	assert.NotPanics(t, func() {
		err := runSimulate(cmd, []string{})
		assert.NoError(t, err)
	})
}

func TestRunSimulate_ConflictingOutputFlags(t *testing.T) {
	oldJSON, oldMD := jsonOutput, markdownOutput
	defer func() {
		jsonOutput = oldJSON
		markdownOutput = oldMD
	}()

	jsonOutput = true
	markdownOutput = true

	tmpFile := createTempConfigFile(t, createValidConfig())
	cmd := &cobra.Command{}
	cmd.Flags().StringP("file", "f", tmpFile, "Config file")

	err := runSimulate(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be used together")
}

func TestSimulateCmd_Flags(t *testing.T) {
	assert.NotNil(t, simulateCmd)

	flags := simulateCmd.Flags()
	assert.NotNil(t, flags.Lookup("json"))
	assert.NotNil(t, flags.Lookup("markdown"))
	assert.NotNil(t, flags.Lookup("quiet"))
	assert.NotNil(t, flags.Lookup("verbose"))
}
