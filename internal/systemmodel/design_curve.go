package systemmodel

import (
	"fmt"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
)

// The coefficient of the variable term is evaluated at one unit. This is a
// normalization of the graph equation, independent of every planning forecast.
func unitQuantity() Assumption {
	return Assumption{Value: number(1), Evidence: &costv1.Evidence{
		Kind: costv1.EvidencePredicted, Measurement: costv1.MeasurementDerived,
		Source:     costv1.SourceReference{Type: costv1.SourceProfitCtlDerived, ArtifactIdentity: DesignSchemaVersion + "#unit-coefficient", CapturedAt: "2026-11-21"},
		Confidence: costv1.ConfidenceHigh, ConfidenceRationale: "One unit defines the marginal coefficient of the linear workload graph.",
	}}
}

type designCurve struct {
	curve           CostCurve
	marginalDrivers []DriverResult
}

// Keep only paths from the selected roots. A derived driver can sum fixed and
// volume-dependent terms. Removing edges keeps unknown fixed costs out of
// incremental cost, and unknown marginal rates out of the already-paid floor.
func graphComponent(m Model, root string, variable bool) (Report, error) {
	drivers := make(map[string]Driver, len(m.Drivers))
	for _, d := range m.Drivers {
		drivers[d.ID] = d
	}
	selected := map[string]bool{}
	var selectDriver func(string) bool
	selectDriver = func(id string) bool {
		if keep, ok := selected[id]; ok {
			return keep
		}
		d := drivers[id]
		if len(d.Inputs) == 0 {
			selected[id] = (id == root) == variable
			return selected[id]
		}
		selected[id] = false
		for _, edge := range d.Inputs {
			if selectDriver(edge.From) {
				selected[id] = true
			}
		}
		return selected[id]
	}
	out := m
	out.Drivers = nil
	for _, d := range m.Drivers {
		if !selectDriver(d.ID) {
			continue
		}
		if d.ID == root {
			unit := unitQuantity()
			d.Value = &unit
		} else if len(d.Inputs) > 0 {
			edges := make([]Link, 0, len(d.Inputs))
			for _, e := range d.Inputs {
				if selected[e.From] {
					edges = append(edges, e)
				}
			}
			d.Inputs = edges
		}
		out.Drivers = append(out.Drivers, d)
	}
	if len(out.Drivers) == 0 {
		return Report{Total: number(0)}, nil
	}
	out.Services = make([]Service, 0, len(m.Services))
	for _, s := range m.Services {
		lines := make([]CostLine, 0, len(s.Costs))
		for _, line := range s.Costs {
			if selected[line.Driver] {
				lines = append(lines, line)
			}
		}
		s.Costs = lines
		if len(lines) == 0 {
			s.FreeReason = "No cost lines belong to this graph component."
		}
		out.Services = append(out.Services, s)
	}
	return Evaluate(out)
}

func analyzeCurve(m Model, root string) (designCurve, error) {
	fixed, err := graphComponent(m, root, false)
	if err != nil {
		return designCurve{}, err
	}
	marginal, err := graphComponent(m, root, true)
	if err != nil {
		return designCurve{}, err
	}
	return designCurve{CostCurve{Fixed: fixed.Total, Unit: marginal.Total}, marginal.Drivers}, nil
}

func analyzePeak(in PeakInput, base, proposed designCurve) (PeakReport, error) {
	r := PeakReport{Arrivals: in.Arrivals.Value, Duration: in.Duration.Value, Occupancy: []PoolOccupancy{}, Capacity: "within_assumed_capacity",
		Scope: "Incremental cost of all work caused by arrivals during this window, including work drained afterward. Already-paid monthly floors are excluded. Average concurrent slots assume steady arrivals and one busy slot per second of service. Arrival clustering, cold starts, retries and queue delay require measurement."}
	var err error
	r.Volume, err = multiply(r.Arrivals, r.Duration)
	if err != nil {
		return PeakReport{}, err
	}
	r.Baseline, err = multiply(r.Volume, base.curve.Unit)
	if err != nil {
		return PeakReport{}, err
	}
	r.Proposed, err = multiply(r.Volume, proposed.curve.Unit)
	if err != nil {
		return PeakReport{}, err
	}
	if r.Baseline != nil && r.Proposed != nil {
		r.Delta = number(*r.Proposed - *r.Baseline)
	}
	for _, p := range in.Occupancy {
		drivers := base.marginalDrivers
		if p.Option == "proposed" {
			drivers = proposed.marginalDrivers
		}
		var driver *DriverResult
		for i := range drivers {
			if drivers[i].ID == p.BusySeconds {
				driver = &drivers[i]
				break
			}
		}
		if driver == nil || driver.Unit != "second" {
			return PeakReport{}, fmt.Errorf("%s occupancy %q must be a volume-linked busy-time driver in seconds", p.Option, p.BusySeconds)
		}
		pool := PoolOccupancy{Option: p.Option, BusySeconds: p.BusySeconds, Slots: p.Slots.Value, Status: "unknown"}
		if driver.Quantity != nil {
			hoursPerUnit := *driver.Quantity / 3600
			pool.Required, err = multiply(r.Arrivals, &hoursPerUnit)
			if err != nil {
				return PeakReport{}, err
			}
		}
		if pool.Required != nil && pool.Slots != nil {
			switch {
			case *pool.Required > *pool.Slots:
				pool.Status = "over_capacity"
			case *pool.Required == *pool.Slots:
				pool.Status = "at_capacity"
			default:
				pool.Status = "within_assumed_capacity"
			}
		}
		// A known overload is material even if another pool has unknown capacity.
		switch {
		case pool.Status == "over_capacity":
			r.Capacity = pool.Status
		case pool.Status == "unknown" && r.Capacity != "over_capacity":
			r.Capacity = pool.Status
		case pool.Status == "at_capacity" && r.Capacity == "within_assumed_capacity":
			r.Capacity = pool.Status
		}
		r.Occupancy = append(r.Occupancy, pool)
	}
	return r, nil
}
