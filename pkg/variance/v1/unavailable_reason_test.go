package v1_test

import (
	"context"
	"strings"
	"testing"

	"github.com/IntelIP/ProfitCtl/cmd"
	cost "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
	variance "github.com/IntelIP/ProfitCtl/pkg/variance/v1"
)

func TestCompareUnavailableReasonMatchesExplain(t *testing.T) {
	e := cost.Evidence{
		Kind: cost.EvidenceUserSupplied, Measurement: cost.MeasurementDeclared,
		Source:     cost.SourceReference{Type: cost.SourceUserSupplied, ArtifactIdentity: "fixture://forecast", CapturedAt: "2026-08-01T00:00:00Z"},
		Confidence: cost.ConfidenceLow, ConfidenceRationale: "locally supplied forecast",
	}
	forecast := cost.CostObservation{
		SchemaVersion: cost.SchemaVersion, ID: "forecast",
		DriverIDs:  []string{"driver"},
		Window:     cost.TimeWindow{Start: "2026-07-01T00:00:00Z", End: "2026-08-01T00:00:00Z"},
		Quantity:   cost.Quantity{Value: 1, Unit: "command"},
		UnitPrice:  cost.UnitPrice{Amount: cost.Money{Amount: 20, Currency: "USD"}, Per: cost.Quantity{Value: 1, Unit: "command"}},
		TotalCost:  cost.Money{Amount: 20, Currency: "USD"},
		Dimensions: cost.Dimensions{Workload: "worker"},
		Evidence:   cost.ClaimEvidence{Quantity: e, UnitPrice: e, TotalCost: e},
	}
	for _, reason := range []string{"", " \t\n "} {
		t.Run("invalid_"+strings.ReplaceAll(reason, "\n", "newline"), func(t *testing.T) {
			_, err := variance.Compare(variance.Input{Forecast: forecast, UnavailableReason: reason})
			if err == nil || !strings.Contains(err.Error(), "unavailable_reason") {
				t.Fatalf("expected actionable unavailable_reason error for %q, got %v", reason, err)
			}
		})
	}
	reason := "  invoice not delivered  "
	result, err := variance.Compare(variance.Input{Forecast: forecast, UnavailableReason: reason})
	if err != nil {
		t.Fatal(err)
	}
	if result.UnavailableReason != reason || len(result.MissingEvidence) != 1 || !strings.Contains(result.MissingEvidence[0], reason) {
		t.Fatalf("supplied reason not preserved and visible: %+v", result)
	}
	explanation, err := cmd.Explain(context.Background(), &result)
	if err != nil {
		t.Fatalf("Compare result must be explainable: %v", err)
	}
	if explanation.UnavailableReason != reason || len(explanation.MissingEvidence) != 1 || !strings.Contains(explanation.MissingEvidence[0], reason) {
		t.Fatalf("explanation lost supplied reason: %+v", explanation)
	}
}
