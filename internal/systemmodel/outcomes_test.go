package systemmodel

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutcomeComparisonUsesAllSpendAndMeasuredDenominators(t *testing.T) {
	f, err := os.Open("../../examples/system/accepted-outcomes.json")
	require.NoError(t, err)
	defer f.Close()
	var input OutcomeComparison
	require.NoError(t, Decode(f, &input))
	report, err := CompareOutcomes(input)
	require.NoError(t, err)
	require.Equal(t, "incomplete", report.Status)
	require.Equal(t, 7.0, *report.Options[0].PerAccepted)
	require.Nil(t, report.Options[1].PerAccepted)
	require.InDelta(t, 0.8, *report.Options[1].BreakEvenRate, 1e-12)

	baseline := &input.Options[0]
	baseline.Costs = baseline.Costs[:1]
	report, err = CompareOutcomes(input)
	require.NoError(t, err)
	require.Equal(t, 5.0, *report.Options[0].PerAccepted)
	require.Greater(t, *report.Options[1].BreakEvenRate, 1.0)
	baseline.Costs = append(baseline.Costs, OutcomeCharge{ID: "unallocated", Basis: "unknown", Source: "Missing billed usage"})
	report, err = CompareOutcomes(input)
	require.NoError(t, err)
	require.Equal(t, 30.0, *report.Options[0].KnownSubtotal)
	require.Nil(t, report.Options[0].PerAccepted)
	require.Nil(t, report.Options[1].BreakEvenRate)

	*baseline.Accepted = 0
	baseline.Costs = baseline.Costs[:1]
	report, err = CompareOutcomes(input)
	require.NoError(t, err)
	require.Nil(t, report.Options[0].PerAccepted)
	require.Nil(t, report.Options[1].BreakEvenRate)
	require.Contains(t, report.Options[0].Limitations, "No accepted outcomes; unit cost is undefined.")
	*baseline.Accepted = *baseline.Attempts + 1
	_, err = CompareOutcomes(input)
	require.ErrorContains(t, err, "exceeds attempts")
	*baseline.Accepted = *baseline.Attempts
	baseline.Costs = append(baseline.Costs, baseline.Costs[0])
	_, err = CompareOutcomes(input)
	require.ErrorContains(t, err, "unique charge")
}
