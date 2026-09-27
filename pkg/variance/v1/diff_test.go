package v1

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	cost "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
)

func TestCostDiffVarianceGolden(t *testing.T) {
	var cases []struct {
		Name     string  `json:"name"`
		FQ       float64 `json:"forecast_quantity"`
		AQ       float64 `json:"actual_quantity"`
		FP       float64 `json:"forecast_price"`
		AP       float64 `json:"actual_price"`
		Dominant string  `json:"dominant"`
		Variance float64 `json:"variance"`
	}
	b, err := os.ReadFile("../../../test/fixtures/cost-variance/v1/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			f := observation("forecast", tc.FQ, tc.FP, tc.FQ*tc.FP)
			a := observation("actual", tc.AQ, tc.AP, tc.AQ*tc.AP)
			report, err := Diff([]cost.CostObservation{f}, []cost.CostObservation{a})
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(report.VarianceAbsolute.Amount-tc.Variance) > Rounding || math.Abs(report.Residual.Amount) > Rounding {
				t.Fatalf("nonreconciling report: %+v", report)
			}
			switch tc.Dominant {
			case "quantity":
				if math.Abs(report.DriverContributions[0].Amount.Amount) <= math.Abs(report.DriverContributions[1].Amount.Amount) && report.DriverContributions[0].Driver == "usage_volume" {
					t.Fatal("wrong dominant")
				}
			case "unit_price":
				if report.DriverContributions[0].Driver != "price_change" {
					t.Fatal("wrong dominant")
				}
			}
			explained, err := Explain(*report)
			if err != nil {
				t.Fatal(err)
			}
			if explained.Confidence != "unknown" || len(explained.DriverContributions[0].MissingEvidence) == 0 || len(explained.DriverContributions[0].RemedialActions) == 0 {
				t.Fatal("unsupported causal claim")
			}
			if _, err := json.Marshal(explained); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestCostDiffVarianceUpstashAndPeriods(t *testing.T) {
	b, err := os.ReadFile("../../../test/fixtures/cost_contract/v1/upstash_idle_polling.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Observation cost.CostObservation `json:"observation"`
	}
	if err := json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	f := fixture.Observation
	a := f
	a.ID = "actual"
	a.Quantity.Value *= 1.5
	a.TotalCost.Amount *= 1.5
	r, err := Diff([]cost.CostObservation{f}, []cost.CostObservation{a})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.Residual.Amount) > Rounding || r.DriverContributions[0].Attribution != "unknown" || r.DriverContributions[0].ConfidenceScore > .35 {
		t.Fatalf("unsupported claim: %+v", r)
	}
	// Independent windows are matched by identity, not by array position.
	f2, a2 := f, a
	f2.ID = "next_forecast"
	a2.ID = "next_actual"
	f2.Window.Start = "2026-08-01T00:00:00Z"
	f2.Window.End = "2026-09-01T00:00:00Z"
	a2.Window = f2.Window
	f2.Evidence.Quantity.Source.CapturedAt = "2026-09-01"
	f2.Evidence.UnitPrice.Source.CapturedAt = "2026-09-01"
	f2.Evidence.TotalCost.Source.CapturedAt = "2026-09-01"
	a2.Evidence = f2.Evidence
	r, err = Diff([]cost.CostObservation{f2, f}, []cost.CostObservation{a, a2})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Periods) != 2 || math.Abs(r.VarianceAbsolute.Amount-1.386) > Rounding {
		t.Fatalf("bad multi-period result: %+v", r)
	}
	if _, err = Explain(*r); err != nil {
		t.Fatal(err)
	}
}

func TestCostDiffVarianceResidualAndValidation(t *testing.T) {
	f := observation("f", 10, 2, 21)
	a := observation("a", 20, 3, 65)
	r, err := Diff([]cost.CostObservation{f}, []cost.CostObservation{a})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.Residual.Amount-4) > Rounding {
		t.Fatalf("residual %v", r.Residual)
	}
	r.Periods[0].Residual.Amount = 0
	if _, err = Explain(*r); err == nil {
		t.Fatal("tampered report accepted")
	}
	r.Periods[0].Residual.Amount = 4
	r.DriverContributions[0].Amount.Amount++
	if _, err = Explain(*r); err == nil {
		t.Fatal("tampered driver accepted")
	}
	a.Window.End = "2026-09-01T00:00:00Z"
	if _, err = Diff([]cost.CostObservation{f}, []cost.CostObservation{a}); err == nil {
		t.Fatal("unmatched window accepted")
	}
	if _, err = Diff([]cost.CostObservation{f}, nil); err == nil {
		t.Fatal("unavailable actual accepted as zero")
	}
}
