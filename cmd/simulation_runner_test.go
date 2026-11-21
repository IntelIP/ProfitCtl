package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimulationReplaysStressDraws(t *testing.T) {
	// This fixture includes normal, uniform and exponential distributions.
	data, err := os.ReadFile("../test/fixtures/valid_config.yml")
	require.NoError(t, err)
	data = bytes.Replace(data, []byte("iterations: 10000"), []byte("iterations: 100"), 1)
	file := filepath.Join(t.TempDir(), "profit.yml")
	run := func(input []byte, seed int64) []byte {
		t.Helper()
		require.NoError(t, os.WriteFile(file, input, 0600))
		result, err := buildSimulationResult(file)
		require.NoError(t, err)
		require.Equal(t, seed, result.StressResult.MonteCarlo.Seed)
		require.Positive(t, result.StressResult.MonteCarlo.StdDev)
		encoded, err := output.FormatJSONResult(result)
		require.NoError(t, err)
		return encoded
	}
	initial := run(data, 1)
	require.JSONEq(t, string(initial), string(run(data, 1)))
	alternative := bytes.Replace(data, []byte("simulation:"), []byte("simulation:\n  seed: 27"), 1)
	changed := run(alternative, 27)
	// Check actual sampled metrics too; a changed seed field alone proves nothing.
	var initialMetrics, changedMetrics output.JSONResult
	require.NoError(t, json.Unmarshal(initial, &initialMetrics))
	require.NoError(t, json.Unmarshal(changed, &changedMetrics))
	require.NotEqual(t, initialMetrics.StressTest.WorstCaseTotalCost, changedMetrics.StressTest.WorstCaseTotalCost)
}

func TestBuildSimulationResult_UsesBillableUsersForHybridRevenue(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "profit.yml")
	err := os.WriteFile(configFile, []byte(`
simulation:
  base_users: 100
  billable_users: 12
  growth_factor: 1.1
  iterations: 100

fixed_costs:
  - name: base infra
    amount: 100
    period: monthly
    layer: infrastructure

variable_costs: []

pricing:
  mode: hybrid
  contract:
    base_platform_fee: 500
    per_seat_fee: 25

covenants:
  - type: threshold
    field: operating_margin
    operator: gte
    value: 80
    message: Operating margin must be >= 80%
`), 0644)
	require.NoError(t, err)

	result, err := buildSimulationResult(configFile)
	require.NoError(t, err)

	assert.Equal(t, 100, result.Users)
	assert.Equal(t, 800.0, result.Revenue.Total)
	assert.Equal(t, 800.0, result.Revenue.RecurringTotal)
	assert.True(t, result.Covenants.Passed)
}
