package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func varianceObservation(t *testing.T, id string, quantity float64) costv1.CostObservation {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "test", "fixtures", "cost_contract", "v1", "upstash_idle_polling.json"))
	require.NoError(t, err)
	var fixture struct {
		Observation costv1.CostObservation `json:"observation"`
	}
	require.NoError(t, json.Unmarshal(data, &fixture))
	observation := fixture.Observation
	observation.ID = id
	observation.Quantity.Value = quantity
	observation.TotalCost.Amount = quantity * observation.UnitPrice.Amount.Amount
	require.NoError(t, observation.Validate())
	return observation
}

func varianceWriteJSON(t *testing.T, file string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(file, data, 0600))
}

func varianceExecute(t *testing.T, command *cobra.Command, args ...string) ([]byte, error) {
	t.Helper()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&bytes.Buffer{})
	command.SetArgs(args)
	err := command.Execute()
	return output.Bytes(), err
}

func TestCostDiffAndCostExplainJSON(t *testing.T) {
	directory := t.TempDir()
	forecastPath := filepath.Join(directory, "forecast.json")
	actualPath := filepath.Join(directory, "actual.json")
	diffPath := filepath.Join(directory, "diff.json")
	explainPath := filepath.Join(directory, "explain.json")
	varianceWriteJSON(t, forecastPath, varianceObservation(t, "forecast", 1000000))
	varianceWriteJSON(t, actualPath, varianceObservation(t, "actual", 1500000))

	output, err := varianceExecute(t, newCostDiffCommand(), "--forecast", forecastPath, "--actual", actualPath, "--output", diffPath)
	require.NoError(t, err)
	disk, err := os.ReadFile(diffPath)
	require.NoError(t, err)
	require.Equal(t, output, disk)
	var diff map[string]any
	require.NoError(t, json.Unmarshal(output, &diff))
	require.NotEmpty(t, diff["schema_version"])

	output, err = varianceExecute(t, newCostExplainCommand(), "--input", diffPath, "--output", explainPath)
	require.NoError(t, err)
	disk, err = os.ReadFile(explainPath)
	require.NoError(t, err)
	require.Equal(t, output, disk)
	var explanation map[string]any
	require.NoError(t, json.Unmarshal(output, &explanation))
	require.NotEmpty(t, explanation["schema_version"])
}

func TestCostDiffInputErrors(t *testing.T) {
	directory := t.TempDir()
	forecastPath := filepath.Join(directory, "forecast.json")
	actualPath := filepath.Join(directory, "actual.json")
	valid := varianceObservation(t, "forecast", 1000000)
	varianceWriteJSON(t, forecastPath, valid)
	varianceWriteJSON(t, actualPath, varianceObservation(t, "actual", 1500000))

	cases := []struct {
		name   string
		mutate func()
		want   string
	}{
		{"missing flag", func() {}, "--forecast and --actual"},
		{"malformed JSON", func() { require.NoError(t, os.WriteFile(actualPath, []byte("{"), 0600)) }, "actual:"},
		{"missing evidence", func() { bad := valid; bad.Evidence.Quantity = costv1.Evidence{}; varianceWriteJSON(t, actualPath, bad) }, "evidence.quantity"},
		{"zero forecast", func() {
			zero := valid
			zero.Quantity.Value = 0
			zero.TotalCost.Amount = 0
			varianceWriteJSON(t, forecastPath, zero)
		}, "zero"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			varianceWriteJSON(t, forecastPath, valid)
			varianceWriteJSON(t, actualPath, varianceObservation(t, "actual", 1500000))
			test.mutate()
			args := []string{"--forecast", forecastPath, "--actual", actualPath}
			if test.name == "missing flag" {
				args = nil
			}
			_, err := varianceExecute(t, newCostDiffCommand(), args...)
			require.Equal(t, 2, ExitCode(err))
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestCostExplainInvalidAndOutputAlias(t *testing.T) {
	directory := t.TempDir()
	input := filepath.Join(directory, "diff.json")
	require.NoError(t, os.WriteFile(input, []byte("{}{}"), 0600))
	_, err := varianceExecute(t, newCostExplainCommand(), "--input", input)
	require.Equal(t, 2, ExitCode(err))
	require.ErrorContains(t, err, "decode diff report")
	_, err = varianceExecute(t, newCostExplainCommand(), "--input", input, "--output", input)
	require.Equal(t, 2, ExitCode(err))
	require.ErrorContains(t, err, "--output must not reference input")
	_, err = varianceExecute(t, newCostExplainCommand())
	require.Equal(t, 2, ExitCode(err))
	require.ErrorContains(t, err, "--input is required")
	_, err = varianceExecute(t, newCostDiffCommand(), "--forecast", input, "--actual", input, "--output", input)
	require.Equal(t, 2, ExitCode(err))
	require.ErrorContains(t, err, "--output must not reference input")
}

func TestCostVarianceCommandFlags(t *testing.T) {
	for _, name := range []string{"forecast", "actual", "output"} {
		require.NotNil(t, costDiffCmd.Flags().Lookup(name))
	}
	for _, name := range []string{"input", "output"} {
		require.NotNil(t, costExplainCmd.Flags().Lookup(name))
	}
	for _, name := range []string{"cost_diff", "cost_explain"} {
		command, _, err := rootCmd.Find([]string{name})
		require.NoError(t, err)
		require.Equal(t, name, command.Name())
	}
}
