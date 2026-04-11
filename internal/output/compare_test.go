package output

import (
	"encoding/json"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/covenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildComparisonResult(t *testing.T) {
	baseline := createMockSimulationResult()
	baseline.Revenue.Mode = "tiered"
	baseline.Revenue.Total = 10000
	baseline.Margin.GrossMargin = 30
	baseline.Margin.CostPerUser = 7

	mix := createMockSimulationResult()
	mix.Revenue.Mode = "mix"
	mix.Revenue.Total = 12500
	mix.Margin.GrossMargin = 42
	mix.Margin.CostPerUser = 5
	mix.Covenants.Passed = true

	hybrid := createMockSimulationResult()
	hybrid.Revenue.Mode = "hybrid"
	hybrid.Revenue.Total = 9000
	hybrid.Margin.GrossMargin = 18
	hybrid.Margin.CostPerUser = 9
	hybrid.Covenants.Passed = false
	hybrid.Covenants.Violations = []covenant.Violation{{Message: "margin breach"}}

	result := BuildComparisonResult([]ComparisonItem{
		{Name: "tiered", File: "tiered.yml", Result: baseline},
		{Name: "mix", File: "mix.yml", Result: mix},
		{Name: "hybrid", File: "hybrid.yml", Result: hybrid},
	})

	assert.Equal(t, "tiered", result.Baseline)
	require.Len(t, result.Scenarios, 3)
	assert.Equal(t, 2500.0, result.Scenarios[1].RevenueDelta)
	assert.Equal(t, 12.0, result.Scenarios[1].MarginDelta)
	assert.Equal(t, -2.0, result.Scenarios[1].CostPerUserDelta)
	assert.Equal(t, "mix", result.Leaders.HighestRevenue)
	assert.Equal(t, "mix", result.Leaders.HighestMargin)
	assert.Equal(t, "mix", result.Leaders.LowestCostPerUser)
	assert.Equal(t, "mix", result.Leaders.BestCovenantHealth)
}

func TestFormatCLIComparisonResult(t *testing.T) {
	result := BuildComparisonResult([]ComparisonItem{
		{Name: "tiered", File: "tiered.yml", Result: createMockSimulationResult()},
		{Name: "mix", File: "mix.yml", Result: createMockSimulationResult()},
	})

	output := FormatCLIComparisonResult(result)
	assert.Contains(t, output, "=== profitctl Scenario Comparison ===")
	assert.Contains(t, output, "Baseline: tiered")
	assert.Contains(t, output, "Delta vs baseline:")
	assert.Contains(t, output, "Leaders:")
}

func TestFormatMarkdownComparisonResult(t *testing.T) {
	result := BuildComparisonResult([]ComparisonItem{
		{Name: "tiered", File: "tiered.yml", Result: createMockSimulationResult()},
		{Name: "mix", File: "mix.yml", Result: createMockSimulationResult()},
	})

	output := FormatMarkdownComparisonResult(result)
	assert.Contains(t, output, "## profitctl Scenario Comparison")
	assert.Contains(t, output, "| Scenario | Mode | Revenue | Payment Fees | Total Cost | Margin | Cost/User | Covenants |")
	assert.Contains(t, output, "### Delta vs Baseline")
	assert.Contains(t, output, "### Leaders")
}

func TestFormatJSONComparisonResult(t *testing.T) {
	result := BuildComparisonResult([]ComparisonItem{
		{Name: "tiered", File: "tiered.yml", Result: createMockSimulationResult()},
		{Name: "mix", File: "mix.yml", Result: createMockSimulationResult()},
	})

	data, err := FormatJSONComparisonResult(result)
	require.NoError(t, err)

	var decoded ComparisonResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "tiered", decoded.Baseline)
	require.Len(t, decoded.Scenarios, 2)
}
