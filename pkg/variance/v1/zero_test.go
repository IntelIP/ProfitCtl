package v1

import (
	"encoding/json"
	"math"
	"testing"

	cost "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
)

func TestZeroForecastPeriodsAndAggregate(t *testing.T) {
	f := observation("f", 0, 2, 0)
	a := observation("a", 2, 2, 4)
	nextF, nextA := f, a
	nextF.ID, nextA.ID = "next-f", "next-a"
	nextF.Window.Start, nextF.Window.End = stamp, "2026-09-01T00:00:00Z"
	nextA.Window = nextF.Window
	for _, tc := range []struct {
		name             string
		forecast, actual []cost.CostObservation
		aggregateNull    bool
	}{
		{"zero aggregate", []cost.CostObservation{f}, []cost.CostObservation{a}, true},
		{"mixed periods", []cost.CostObservation{f, observation("other-f", 3, 2, 6)}, []cost.CostObservation{a, observation("other-a", 4, 2, 8)}, false},
		{"zero periods", []cost.CostObservation{f, nextF}, []cost.CostObservation{a, nextA}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.forecast) > 1 && tc.name == "mixed periods" {
				tc.forecast[1].Window = nextF.Window
				tc.actual[1].Window = nextF.Window
			}
			r, err := Diff(tc.forecast, tc.actual)
			if err != nil {
				t.Fatal(err)
			}
			if (r.VariancePercent == nil) != tc.aggregateNull || r.Periods[0].VariancePercent != nil {
				t.Fatalf("wrong undefined percentages: %+v", r)
			}
			if !tc.aggregateNull && (r.VariancePercent == nil || math.Abs(*r.VariancePercent-100) > Rounding || r.Periods[1].VariancePercent == nil) {
				t.Fatalf("lost numeric percentage: %+v", r)
			}
			sum := r.Residual.Amount
			for _, c := range r.DriverContributions {
				sum += c.Amount.Amount
			}
			near(t, sum, r.VarianceAbsolute.Amount)
			explained, err := Explain(*r)
			if err != nil {
				t.Fatal(err)
			}
			if (explained.VariancePercent == nil) != tc.aggregateNull || explained.Periods[0].VariancePercent != nil {
				t.Fatal("explanation lost undefined percentage")
			}
			b, err := json.Marshal(explained)
			if err != nil {
				t.Fatal(err)
			}
			var raw map[string]any
			if err = json.Unmarshal(b, &raw); err != nil {
				t.Fatal(err)
			}
			if tc.aggregateNull && raw["variance_percent"] != nil {
				t.Fatalf("expected null: %s", b)
			}
			if raw["periods"].([]any)[0].(map[string]any)["variance_percent"] != nil {
				t.Fatalf("expected period null: %s", b)
			}
			tampered := *r
			if tc.aggregateNull {
				v := 0.0
				tampered.VariancePercent = &v
			} else {
				tampered.VariancePercent = nil
			}
			if _, err = Explain(tampered); err == nil {
				t.Fatal("invalid percentage accepted")
			}
		})
	}
	in := input(cost.DriverFixed, 0, 2, 2, 2, 0, 4)
	result, err := Compare(in)
	if err != nil {
		t.Fatal(err)
	}
	if result.PercentageVariance != nil || result.AbsoluteVariance.Amount != 4 {
		t.Fatalf("canonical comparison: %+v", result)
	}
	b, err := MarshalJSONResult(result)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err = json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["percentage_variance"] != nil {
		t.Fatalf("canonical percentage not null: %s", b)
	}
	sum := result.Residual.Amount
	for _, c := range result.Contributions {
		sum += c.Amount.Amount
	}
	near(t, sum, result.AbsoluteVariance.Amount)
}
