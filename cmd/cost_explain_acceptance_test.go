package cmd

import (
	"context"
	"math"
	"testing"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
	variancev1 "github.com/IntelIP/ProfitCtl/pkg/variance/v1"
	"github.com/stretchr/testify/require"
)

// Operator-owned regressions from rejected agent candidates. Workers may read,
// but cannot edit this file or the controller's acceptance commands.
func TestCostExplainAcceptanceActualReason(t *testing.T) {
	f, _ := varianceFixture(t, "forecast", 10, 2, 20, costv1.DriverFixed)
	for _, tc := range []struct {
		name, evidence string
		valid          bool
	}{
		{"canonical actual reason", "actual cost: invoice not delivered", true},
		{"contradictory substring", "actual cost: not invoice not delivered", false},
		{"reason only in forecast clause", "actual cost: ledger pending; forecast invoice not delivered", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Diff(context.Background(), &f, nil)
			require.NoError(t, err)
			result.UnavailableReason = "invoice not delivered"
			result.MissingEvidence = []string{tc.evidence}
			report, err := Explain(context.Background(), result)
			if !tc.valid {
				require.ErrorContains(t, err, "missing_evidence")
				return
			}
			require.NoError(t, err)
			require.Equal(t, result.UnavailableReason, report.UnavailableReason)
			require.Nil(t, report.Actual)
		})
	}
}

func TestCostExplainAcceptanceAccumulationOverflow(t *testing.T) {
	f, fd := varianceFixture(t, "forecast", 10, 2, 20, costv1.DriverFixed)
	a, ad := varianceFixture(t, "actual", 20, 2, 40, costv1.DriverFixed)
	result, err := variancev1.Compare(variancev1.Input{Forecast: f, Actual: &a, Drivers: []variancev1.DriverPair{{Forecast: fd, Actual: ad}}})
	require.NoError(t, err)
	require.NotEmpty(t, result.Contributions)
	contribution := result.Contributions[0]
	result.Contributions = result.Contributions[:0]
	for _, amount := range []float64{math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64} {
		contribution.Amount.Amount = amount
		result.Contributions = append(result.Contributions, contribution)
	}
	_, err = Explain(context.Background(), &result)
	require.Error(t, err, "Finite inputs must not allow an infinite accumulated contribution total")
}
