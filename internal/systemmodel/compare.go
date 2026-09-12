package systemmodel

import (
	"fmt"
	"reflect"
	"sort"
)

type ServiceDelta struct {
	ID     string   `json:"id"`
	Before *float64 `json:"before_usd"`
	After  *float64 `json:"after_usd"`
	Delta  *float64 `json:"delta_usd"`
}
type Comparison struct {
	SchemaVersion string         `json:"schema_version"`
	Baseline      Report         `json:"baseline"`
	Proposed      Report         `json:"proposed"`
	ChangedInputs []string       `json:"changed_inputs"`
	Services      []ServiceDelta `json:"services"`
	Delta         *float64       `json:"total_delta_usd"`
}

func modelInputs(m Model) map[string]any {
	out := map[string]any{"revenue": m.Revenue}
	for _, d := range m.Drivers {
		out["driver:"+d.ID] = d.Value
		out["unit:"+d.ID] = d.Unit
		for _, e := range d.Inputs {
			out["edge:"+e.ID] = e
		}
	}
	for _, s := range m.Services {
		out["service:"+s.ID] = s
	}
	return out
}
func Compare(a, b Model) (Comparison, error) {
	if a.Period != b.Period || a.Currency != b.Currency {
		return Comparison{}, fmt.Errorf("comparison requires matching period and currency")
	}
	before, err := Evaluate(a)
	if err != nil {
		return Comparison{}, fmt.Errorf("baseline: %w", err)
	}
	after, err := Evaluate(b)
	if err != nil {
		return Comparison{}, fmt.Errorf("proposed: %w", err)
	}
	out := Comparison{SchemaVersion: SchemaVersion, Baseline: before, Proposed: after, ChangedInputs: []string{}, Services: []ServiceDelta{}}
	av, bv := modelInputs(a), modelInputs(b)
	keys := map[string]bool{}
	for k := range av {
		keys[k] = true
	}
	for k := range bv {
		keys[k] = true
	}
	for k := range keys {
		if !reflect.DeepEqual(av[k], bv[k]) {
			out.ChangedInputs = append(out.ChangedInputs, k)
		}
	}
	sort.Strings(out.ChangedInputs)
	am, bm := map[string]*float64{}, map[string]*float64{}
	ids := map[string]bool{}
	for _, s := range before.Services {
		am[s.ID] = s.Total
		ids[s.ID] = true
	}
	for _, s := range after.Services {
		bm[s.ID] = s.Total
		ids[s.ID] = true
	}
	order := []string{}
	for id := range ids {
		order = append(order, id)
	}
	sort.Strings(order)
	for _, id := range order {
		x, ok := am[id]
		if !ok {
			x = number(0)
		}
		y, ok := bm[id]
		if !ok {
			y = number(0)
		}
		d := ServiceDelta{ID: id, Before: x, After: y}
		if x != nil && y != nil {
			d.Delta = number(*y - *x)
		}
		out.Services = append(out.Services, d)
	}
	if before.Total != nil && after.Total != nil {
		out.Delta = number(*after.Total - *before.Total)
	}
	return out, nil
}
