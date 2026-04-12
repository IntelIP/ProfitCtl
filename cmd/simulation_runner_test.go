package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
