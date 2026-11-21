package systemmodel

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func designFixture(t *testing.T) Design {
	t.Helper()
	f, err := os.Open("../../examples/design/condere-queue.json")
	require.NoError(t, err)
	defer f.Close()
	var d Design
	require.NoError(t, Decode(f, &d))
	return d
}

func designGap() Assumption {
	return Assumption{Missing: &Acquisition{Question: "What is the observed value?", Source: "Scoped load-test or billing export", Owner: "Workload owner"}}
}

func TestDesignCondereDecision(t *testing.T) {
	d := designFixture(t)
	r, err := EvaluateDesign(d)
	require.NoError(t, err)
	require.Equal(t, DesignReportVersion, r.SchemaVersion)
	require.Equal(t, "complete", r.Status)
	require.Empty(t, r.Missing)
	require.InDelta(t, 8.16, *r.Expected.Baseline, 1e-10)
	require.InDelta(t, 20.48, *r.Expected.Proposed, 1e-10)
	require.InDelta(t, -11.8, *r.High.Delta, 1e-10)
	require.InDelta(t, 2.448, *r.Peak.Baseline, 1e-10)
	require.InDelta(t, 1.644, *r.Peak.Proposed, 1e-10)
	require.InDelta(t, 1119.402985, *r.BreakEven.Volume, 1e-6)
	require.Equal(t, 1120.0, *r.BreakEven.FirstCheaper)
	require.Equal(t, 0.0, *r.BaselineCurve.Fixed)
	require.InDelta(t, .0408, *r.BaselineCurve.Unit, 1e-10)
	require.Equal(t, 15.0, *r.ProposedCurve.Fixed)
	require.InDelta(t, .0274, *r.ProposedCurve.Unit, 1e-10)
	require.Equal(t, "within_assumed_capacity", r.BreakEven.PeakCapacity)
	for i, n := range []float64{17, 4, 7} {
		require.InDelta(t, n, *r.Peak.Occupancy[i].Required, 1e-10)
	}
	require.InDelta(t, 89.0, *r.Sensitivity.High.Proposed, 1e-10)
	require.InDelta(t, 3947.368421, *r.Sensitivity.BreakEven.Volume, 1e-6)
	require.Equal(t, 3948.0, *r.Sensitivity.BreakEven.FirstCheaper)
	require.Equal(t, "over_capacity", r.Sensitivity.BreakEven.PeakCapacity)
	require.Equal(t, 11.0, *r.Sensitivity.Peak.Occupancy[2].Required)
	require.Contains(t, r.Summary, "$89.00")
	require.Contains(t, r.Summary, "monthly floors")
	require.Contains(t, r.Conditions.Confidentiality, "opaque assessment reference")
	require.Equal(t, d.Peak.Arrivals, r.Assumptions["peak.arrivals_per_hour"])
	// Evaluating the sensitivity must leave both graph templates and the next report intact.
	first, err := json.Marshal(r)
	require.NoError(t, err)
	second, err := EvaluateDesign(d)
	require.NoError(t, err)
	repeat, err := json.Marshal(second)
	require.NoError(t, err)
	require.Equal(t, first, repeat)
}

func TestDesignPeakWindowAndUnrelatedRoots(t *testing.T) {
	d := designFixture(t)
	d.Name = "independent-volume"
	d.Expected.Value = number(30)
	d.Peak.Duration.Value = number(.5)
	// This driver adds monthly warm-up work and attempt-dependent busy seconds
	// before applying the same price. Only one term belongs to the peak.
	prep := &d.Proposed.Drivers[1]
	prep.Inputs = append(prep.Inputs, Link{ID: "plan-warmup", From: "queue-plan", InputUnit: "month", OutputUnit: "second", Factor: unitQuantity()})
	r, err := EvaluateDesign(d)
	require.NoError(t, err)
	require.InDelta(t, 15.000040, *r.ProposedCurve.Fixed, 1e-10)
	require.InDelta(t, .0274, *r.ProposedCurve.Unit, 1e-10)
	require.Equal(t, 30.0, *r.Peak.Volume)
	require.InDelta(t, .822, *r.Peak.Proposed, 1e-10)
	require.Equal(t, 4.0, *r.Peak.Occupancy[1].Required)
	gap := designGap()
	d.Proposed.Drivers[3].Value = &gap
	d.Peak.Occupancy[1].Slots = gap
	r, err = EvaluateDesign(d)
	require.NoError(t, err)
	require.Equal(t, "incomplete", r.Status)
	require.Nil(t, r.ProposedCurve.Fixed)
	require.Nil(t, r.Expected.Proposed)
	require.Nil(t, r.BreakEven.Volume)
	require.Equal(t, "missing", r.BreakEven.Status)
	require.Equal(t, "unknown", r.BreakEven.PeakCapacity)
	require.Equal(t, "over_capacity", r.Sensitivity.BreakEven.PeakCapacity)
	require.InDelta(t, .822, *r.Peak.Proposed, 1e-10)
	require.Len(t, r.Missing, 2)
	require.Equal(t, "peak.available_slots.proposed.prepare-seconds", r.Missing[0].Input)
	require.Equal(t, "proposed.driver.queue-plan", r.Missing[1].Input)
}

func TestDesignUnknownDriversAndPriceScopes(t *testing.T) {
	d := designFixture(t)
	d.Expected = designGap()
	d.Peak.Arrivals = designGap()
	d.Sensitivity.Replacement = designGap()
	r, err := EvaluateDesign(d)
	require.NoError(t, err)
	require.Nil(t, r.Expected.Baseline)
	require.Nil(t, r.Peak.Proposed)
	require.Nil(t, r.Peak.Occupancy[0].Required)
	require.Nil(t, r.Sensitivity.BreakEven.Volume)
	require.NotNil(t, r.BreakEven.Volume)
	require.NotNil(t, r.High.Proposed)
	d = designFixture(t)
	d.Proposed.Services[1].Costs[0].Price = designGap()
	r, err = EvaluateDesign(d)
	require.NoError(t, err)
	require.Nil(t, r.ProposedCurve.Fixed)
	require.NotNil(t, r.Peak.Proposed)
	d = designFixture(t)
	d.Proposed.Services[1].Costs[1].Price = designGap()
	d.High = nil
	r, err = EvaluateDesign(d)
	require.NoError(t, err)
	require.Nil(t, r.High)
	require.NotNil(t, r.ProposedCurve.Fixed)
	require.Nil(t, r.Peak.Proposed)
	require.NotNil(t, r.Peak.Baseline)
	require.Nil(t, r.Peak.Delta)
	require.NotNil(t, r.Peak.Occupancy[1].Required)
}

func TestDesignCrossoverCases(t *testing.T) {
	for _, tc := range []struct {
		name         string
		a, b         CostCurve
		limit        float64
		status       string
		equal, first *float64
	}{
		{"equal", CostCurve{number(1), number(2)}, CostCurve{number(1), number(2)}, 20, "equal", nil, nil},
		{"parallel", CostCurve{number(10), number(2)}, CostCurve{number(1), number(2)}, 20, "no_cross", nil, number(0)},
		{"negative intersection", CostCurve{number(10), number(3)}, CostCurve{number(1), number(2)}, 20, "no_cross", nil, number(0)},
		{"exact integer tie", CostCurve{number(0), number(3)}, CostCurve{number(2), number(2)}, 20, "crosses", number(2), number(3)},
		{"reverse direction", CostCurve{number(2), number(2)}, CostCurve{number(0), number(3)}, 20, "crosses", number(2), number(0)},
		{"tie at zero", CostCurve{number(2), number(2)}, CostCurve{number(2), number(3)}, 20, "crosses", number(0), nil},
		{"last point ties", CostCurve{number(0), number(3)}, CostCurve{number(2), number(2)}, 2, "crosses", number(2), nil},
		{"outside", CostCurve{number(0), number(3)}, CostCurve{number(2), number(2)}, 1, "outside_range", nil, nil},
		{"missing", CostCurve{number(0), nil}, CostCurve{number(2), number(2)}, 20, "missing", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := crossover(tc.a, tc.b, &tc.limit)
			require.Equal(t, tc.status, r.Status)
			require.Equal(t, tc.equal, r.Volume)
			require.Equal(t, tc.first, r.FirstCheaper)
		})
	}
}

func TestDesignRejectsInvalidContracts(t *testing.T) {
	for name, change := range map[string]func(*Design){
		"root value override":        func(d *Design) { d.Baseline.Drivers[0].Value = &d.Expected },
		"mismatched unit":            func(d *Design) { d.Proposed.Drivers[0].Unit = "byte" },
		"dependent volume":           func(d *Design) { d.VolumeDriver = "single-seconds" },
		"unknown sensitivity edge":   func(d *Design) { d.Sensitivity.EdgeID = "other" },
		"missing scope":              func(d *Design) { d.CostScope = " " },
		"missing control":            func(d *Design) { d.Conditions.Integrity = "" },
		"ungrounded price":           func(d *Design) { d.Proposed.Services[1].Costs[0].Price.Evidence = nil },
		"invalid occupancy unit":     func(d *Design) { d.Peak.Occupancy[0].BusySeconds = "attempts" },
		"no proposed capacity":       func(d *Design) { d.Peak.Occupancy = d.Peak.Occupancy[:1] },
		"unguarded unknown":          func(d *Design) { d.Peak.Occupancy[0].Slots = Assumption{} },
		"duration zero":              func(d *Design) { d.Peak.Duration.Value = number(0) },
		"outside monthly range":      func(d *Design) { d.Expected.Value = number(5001); d.High = nil },
		"peak outside monthly range": func(d *Design) { d.Peak.Arrivals.Value = number(5001) },
		"arithmetic overflow":        func(d *Design) { d.Peak.Arrivals.Value = number(math.MaxFloat64); d.Peak.Duration.Value = number(2) },
	} {
		t.Run(name, func(t *testing.T) {
			d := designFixture(t)
			change(&d)
			_, err := EvaluateDesign(d)
			require.Error(t, err)
		})
	}
}
