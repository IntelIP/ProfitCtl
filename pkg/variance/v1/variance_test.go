package v1

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	cost "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
)

const stamp = "2026-08-01T00:00:00Z"

var window = cost.TimeWindow{Start: "2026-07-01T00:00:00Z", End: stamp}

func evidence(source cost.SourceType, kind cost.EvidenceKind, measurement cost.MeasurementKind, confidence cost.Confidence) cost.Evidence {
	return cost.Evidence{Kind: kind, Measurement: measurement, Source: cost.SourceReference{Type: source, ArtifactIdentity: "fixture://variance", CapturedAt: stamp}, Confidence: confidence, ConfidenceRationale: "deterministic test evidence"}
}
func observation(id string, q, p, total float64) cost.CostObservation {
	e := evidence(cost.SourceUserSupplied, cost.EvidenceUserSupplied, cost.MeasurementDeclared, cost.ConfidenceHigh)
	return cost.CostObservation{SchemaVersion: cost.SchemaVersion, ID: id, DriverIDs: []string{"driver"}, Window: window, Quantity: cost.Quantity{Value: q, Unit: "command"}, UnitPrice: cost.UnitPrice{Amount: cost.Money{Amount: p, Currency: "USD"}, Per: cost.Quantity{Value: 1, Unit: "command"}}, TotalCost: cost.Money{Amount: total, Currency: "USD"}, Dimensions: cost.Dimensions{Workload: "worker"}, Evidence: cost.ClaimEvidence{Quantity: e, UnitPrice: e, TotalCost: e}}
}
func driver(kind cost.DriverKind, q, p float64) cost.CostDriver {
	o := observation("base", q, p, q*p)
	d := cost.CostDriver{SchemaVersion: cost.SchemaVersion, ID: "driver", Name: "test driver", Kind: kind, Quantity: o.Quantity, UnitPrice: o.UnitPrice, Window: window, Dimensions: o.Dimensions, Evidence: cost.DriverEvidence{Quantity: o.Evidence.Quantity, UnitPrice: o.Evidence.UnitPrice}}
	if kind == cost.DriverVariable || kind == cost.DriverCadence {
		d.Per = &cost.Quantity{Value: 1, Unit: "second"}
	}
	return d
}
func input(kind cost.DriverKind, fq, aq, fp, ap, ft, at float64) Input {
	a := observation("actual", aq, ap, at)
	return Input{Forecast: observation("forecast", fq, fp, ft), Actual: &a, Drivers: []DriverPair{{Forecast: driver(kind, fq, fp), Actual: driver(kind, aq, ap)}}}
}
func near(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > Rounding {
		t.Fatalf("got %g, want %g", got, want)
	}
}
func TestVarianceGoldenScenarios(t *testing.T) {
	var cases []struct {
		Name     string          `json:"name"`
		Kind     cost.DriverKind `json:"kind"`
		FQ       float64         `json:"forecast_quantity"`
		AQ       float64         `json:"actual_quantity"`
		FP       float64         `json:"forecast_price"`
		AP       float64         `json:"actual_price"`
		Dominant string          `json:"dominant"`
		Variance float64         `json:"variance"`
	}
	data, err := os.ReadFile("../../../test/fixtures/cost-variance/v1/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			in := input(tc.Kind, tc.FQ, tc.AQ, tc.FP, tc.AP, tc.FQ*tc.FP, tc.AQ*tc.AP)
			r, err := Compare(in)
			if err != nil {
				t.Fatal(err)
			}
			near(t, r.AbsoluteVariance.Amount, tc.Variance)
			near(t, r.Residual.Amount, 0)
			c := r.Contributions[0]
			near(t, c.Amount.Amount, tc.Variance)
			switch tc.Dominant {
			case "quantity":
				if math.Abs(c.Quantity) <= math.Abs(c.UnitPrice) {
					t.Fatal("quantity not dominant")
				}
			case "unit_price":
				if math.Abs(c.UnitPrice) <= math.Abs(c.Quantity) {
					t.Fatal("price not dominant")
				}
			}
			if tc.Kind == cost.DriverCadence || tc.Kind == cost.DriverConcurrency {
				if c.Attribution != "unknown" || len(c.MissingEvidence) == 0 {
					t.Fatal("missing receipt must lower confidence")
				}
			}
		})
	}
}
func TestVarianceMixedResidualAndJSON(t *testing.T) {
	in := input(cost.DriverVariable, 10, 20, 2, 3, 21, 65) // model: 20 -> 60; ledger: 21 -> 65
	r, err := Compare(in)
	if err != nil {
		t.Fatal(err)
	}
	near(t, r.AbsoluteVariance.Amount, 44)
	near(t, r.Contributions[0].Quantity, 25)
	near(t, r.Contributions[0].UnitPrice, 15)
	near(t, r.Residual.Amount, 4)
	b, err := MarshalJSONResult(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Result
	if err = json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Window != window || decoded.Forecast.Currency != "USD" || decoded.Contributions[0].Sources[0].ArtifactIdentity == "" || decoded.Contributions[0].ConfidenceScore == 0 || decoded.Residual.Amount != 4 {
		t.Fatalf("incomplete JSON: %s", b)
	}
}
func TestVarianceMultipleDriversAndDistribution(t *testing.T) {
	in := input(cost.DriverVariable, 10, 12, 2, 2, 40, 59)
	second := DriverPair{Forecast: driver(cost.DriverFixed, 10, 2), Actual: driver(cost.DriverFixed, 10, 3)}
	second.Forecast.ID = "second"
	second.Actual.ID = "second"
	in.Forecast.DriverIDs = append(in.Forecast.DriverIDs, "second")
	in.Actual.DriverIDs = append(in.Actual.DriverIDs, "second")
	in.Drivers = append(in.Drivers, second)
	r, err := Compare(in)
	if err != nil {
		t.Fatal(err)
	}
	near(t, r.AbsoluteVariance.Amount, 19)
	near(t, r.Residual.Amount, 5)
	near(t, r.Contributions[0].Amount.Amount, 10)
	near(t, r.Contributions[1].Amount.Amount, 4)
	if len(r.MissingEvidence) == 0 {
		t.Fatal("residual must name missing evidence")
	}
	// A changed stochastic expectation contributes separately, without a causal claim.
	in = input(cost.DriverVariable, 10, 10, 2, 2, 40, 60)
	m1, m2 := 2.0, 3.0
	s := 1.0
	in.Drivers[0].Forecast.Distribution = &cost.Distribution{Type: cost.DistributionNormal, Mean: &m1, StdDev: &s}
	in.Drivers[0].Actual.Distribution = &cost.Distribution{Type: cost.DistributionNormal, Mean: &m2, StdDev: &s}
	r, err = Compare(in)
	if err != nil {
		t.Fatal(err)
	}
	near(t, r.Contributions[0].Distribution, 20)
	near(t, r.Residual.Amount, 0)
}

func TestVarianceUnavailableAndZeroDenominator(t *testing.T) {
	in := input(cost.DriverFixed, 0, 1, 2, 2, 0, 2)
	r, err := Compare(in)
	if err != nil {
		t.Fatal(err)
	}
	if r.PercentageVariance != nil {
		t.Fatal("zero denominator must have null percentage")
	}
	in.Actual = nil
	in.Drivers = nil
	in.UnavailableReason = "invoice not delivered"
	r, err = Compare(in)
	if err != nil {
		t.Fatal(err)
	}
	if r.Actual != nil || r.AbsoluteVariance != nil || r.Residual != nil || r.Status != "unavailable" {
		t.Fatal("unavailable cost cannot become zero")
	}
	in.UnavailableReason = ""
	if _, err = Compare(in); err == nil {
		t.Fatal("missing reason accepted")
	}
}
func TestVarianceConversionsAndReceipts(t *testing.T) {
	in := input(cost.DriverCadence, 1, 60, 2, 2, 2, 120)
	in.Drivers[0].Forecast.Quantity.Unit = "hour"
	in.Drivers[0].Forecast.UnitPrice.Per.Unit = "hour"
	in.Drivers[0].Actual.Quantity.Unit = "minute"
	in.Drivers[0].Actual.UnitPrice.Per.Unit = "minute"
	// One hour at $2/hour, sixty minutes at $2/minute.
	in.Receipts = []Receipt{{DriverID: "driver", Kind: "deployment", ForecastObservationID: "forecast", ActualObservationID: "actual", Source: cost.SourceReference{Type: cost.SourceRuntimeLedger, ArtifactIdentity: "deploy://exact", CapturedAt: stamp}}}
	r, err := Compare(in)
	if err != nil {
		t.Fatal(err)
	}
	near(t, r.Residual.Amount, 0)
	if r.Contributions[0].Attribution != "modeled" || r.Contributions[0].Confidence != cost.ConfidenceHigh {
		t.Fatal("exact receipt should permit high modeled confidence")
	}
	in.Receipts[0].ActualObservationID = "wrong"
	if _, err = Compare(in); err == nil {
		t.Fatal("inexact receipt accepted")
	}
}
func TestVarianceCurrencyAndValidation(t *testing.T) {
	in := input(cost.DriverVariable, 10, 20, 2, 2, 20, 40)
	in.Actual.TotalCost.Currency = "EUR"
	in.Actual.UnitPrice.Amount.Currency = "EUR"
	in.Drivers[0].Actual.UnitPrice.Amount.Currency = "EUR"
	if _, err := Compare(in); err == nil {
		t.Fatal("missing exchange rate accepted")
	}
	in.ExchangeRates = []ExchangeRate{{From: "EUR", To: "USD", Rate: 2, Source: cost.SourceReference{Type: cost.SourceUserSupplied, ArtifactIdentity: "fx://snapshot", CapturedAt: stamp}}}
	r, err := Compare(in)
	if err != nil {
		t.Fatal(err)
	}
	near(t, r.AbsoluteVariance.Amount, 60)
	near(t, r.Residual.Amount, 0)
	in.Actual.Window.End = "2026-09-01T00:00:00Z"
	if _, err = Compare(in); err == nil {
		t.Fatal("mismatched window accepted")
	}
}
