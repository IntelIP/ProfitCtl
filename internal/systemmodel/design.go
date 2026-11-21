package systemmodel

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const DesignSchemaVersion = "profitctl.design/v1"
const DesignReportVersion = "profitctl.design-report/v1"

// Design embeds two monthly graph templates. In each graph the volume driver
// declares its unit. Each scenario binds its volume.
type Design struct {
	SchemaVersion string           `json:"schema_version"`
	Name          string           `json:"name"`
	Question      string           `json:"question"`
	Workload      string           `json:"workload"`
	CostScope     string           `json:"cost_scope"`
	VolumeDriver  string           `json:"volume_driver"`
	VolumeUnit    string           `json:"volume_unit"`
	Expected      Assumption       `json:"expected_monthly"`
	High          *Assumption      `json:"high_monthly,omitempty"`
	ValidThrough  Assumption       `json:"valid_through_monthly"`
	Peak          PeakInput        `json:"peak"`
	Baseline      Model            `json:"baseline"`
	Proposed      Model            `json:"proposed"`
	Sensitivity   FactorChange     `json:"sensitivity"`
	Conditions    DesignConditions `json:"conditions"`
}
type PeakInput struct {
	Arrivals  Assumption     `json:"arrivals_per_hour"`
	Duration  Assumption     `json:"duration_hours"`
	Occupancy []CapacityPool `json:"occupancy"`
}
type CapacityPool struct {
	Option      string     `json:"option"`
	BusySeconds string     `json:"busy_seconds_driver"`
	Slots       Assumption `json:"available_slots"`
}
type FactorChange struct {
	Description string     `json:"description"`
	EdgeID      string     `json:"edge_id"`
	Replacement Assumption `json:"replacement_factor"`
}
type DesignConditions struct {
	Latency         string `json:"latency"`
	Retries         string `json:"retries"`
	Confidentiality string `json:"confidentiality"`
	Integrity       string `json:"integrity"`
	Availability    string `json:"availability"`
	LeastPrivilege  string `json:"least_privilege"`
	Source          string `json:"source"`
}
type MonthlyPoint struct {
	Volume   *float64 `json:"volume"`
	Baseline *float64 `json:"baseline_usd"`
	Proposed *float64 `json:"proposed_usd"`
	Delta    *float64 `json:"delta_usd"`
}
type PoolOccupancy struct {
	Option      string   `json:"option"`
	BusySeconds string   `json:"busy_seconds_driver"`
	Required    *float64 `json:"required_average_slots"`
	Slots       *float64 `json:"available_slots"`
	Status      string   `json:"status"`
}
type PeakReport struct {
	Arrivals  *float64        `json:"arrivals_per_hour"`
	Duration  *float64        `json:"duration_hours"`
	Volume    *float64        `json:"volume"`
	Baseline  *float64        `json:"baseline_incremental_usd"`
	Proposed  *float64        `json:"proposed_incremental_usd"`
	Delta     *float64        `json:"delta_usd"`
	Occupancy []PoolOccupancy `json:"occupancy"`
	Capacity  string          `json:"capacity_status"`
	Scope     string          `json:"scope"`
}
type Crossover struct {
	Status         string   `json:"status"`
	Volume         *float64 `json:"volume_per_month"`
	FirstCheaper   *float64 `json:"first_whole_proposed_cheaper"`
	ValidThrough   *float64 `json:"valid_through_monthly"`
	Interpretation string   `json:"interpretation"`
	PeakCapacity   string   `json:"peak_capacity"`
}
type FactorResult struct {
	Description string        `json:"description"`
	EdgeID      string        `json:"edge_id"`
	FactorUnit  string        `json:"factor_unit"`
	Original    *float64      `json:"original_factor"`
	Replacement *float64      `json:"replacement_factor"`
	Expected    MonthlyPoint  `json:"expected_monthly"`
	High        *MonthlyPoint `json:"high_monthly"`
	BreakEven   Crossover     `json:"break_even"`
	Peak        PeakReport    `json:"peak"`
}
type CostCurve struct {
	Fixed *float64 `json:"fixed_monthly_usd"`
	Unit  *float64 `json:"marginal_usd_per_volume_unit"`
}
type DesignReport struct {
	SchemaVersion string                `json:"schema_version"`
	Name          string                `json:"name"`
	Status        string                `json:"status"`
	Summary       string                `json:"summary"`
	Question      string                `json:"question"`
	Workload      string                `json:"workload"`
	CostScope     string                `json:"cost_scope"`
	VolumeUnit    string                `json:"volume_unit"`
	BaselineName  string                `json:"baseline_name"`
	ProposedName  string                `json:"proposed_name"`
	BaselineCurve CostCurve             `json:"baseline_curve"`
	ProposedCurve CostCurve             `json:"proposed_curve"`
	Expected      MonthlyPoint          `json:"expected_monthly"`
	High          *MonthlyPoint         `json:"high_monthly"`
	Peak          PeakReport            `json:"peak"`
	BreakEven     Crossover             `json:"break_even"`
	Sensitivity   FactorResult          `json:"sensitivity"`
	Conditions    DesignConditions      `json:"conditions"`
	Assumptions   map[string]Assumption `json:"assumptions"`
	Missing       []Gap                 `json:"missing_inputs"`
}

func (d Design) sourceInputs() map[string]Assumption {
	items := map[string]Assumption{
		"expected_monthly": d.Expected, "valid_through_monthly": d.ValidThrough,
		"peak.arrivals_per_hour": d.Peak.Arrivals, "peak.duration_hours": d.Peak.Duration,
		"sensitivity.replacement_factor": d.Sensitivity.Replacement,
	}
	if d.High != nil {
		items["high_monthly"] = *d.High
	}
	for _, pool := range d.Peak.Occupancy {
		items["peak.available_slots."+pool.Option+"."+pool.BusySeconds] = pool.Slots
	}
	for option, model := range map[string]Model{"baseline": d.Baseline, "proposed": d.Proposed} {
		items[option+".revenue"] = model.Revenue
		for _, driver := range model.Drivers {
			if driver.Value != nil {
				items[option+".driver."+driver.ID] = *driver.Value
			}
			for _, edge := range driver.Inputs {
				items[option+".edge."+edge.ID] = edge.Factor
			}
		}
		for _, service := range model.Services {
			for _, line := range service.Costs {
				items[option+".price."+service.ID+"/"+line.ID] = line.Price
			}
		}
	}
	return items
}

func (d Design) validate() error {
	if d.SchemaVersion != DesignSchemaVersion || !idPattern.MatchString(d.Name) {
		return fmt.Errorf("expected %s and a design name using an identifier", DesignSchemaVersion)
	}
	for _, item := range []struct{ name, text string }{
		{"question", d.Question}, {"workload", d.Workload}, {"cost_scope", d.CostScope}, {"volume_unit", d.VolumeUnit},
		{"sensitivity.description", d.Sensitivity.Description}, {"conditions.latency", d.Conditions.Latency},
		{"conditions.retries", d.Conditions.Retries}, {"conditions.confidentiality", d.Conditions.Confidentiality},
		{"conditions.integrity", d.Conditions.Integrity}, {"conditions.availability", d.Conditions.Availability},
		{"conditions.least_privilege", d.Conditions.LeastPrivilege}, {"conditions.source", d.Conditions.Source},
	} {
		if strings.TrimSpace(item.text) == "" {
			return fmt.Errorf("%s is required", item.name)
		}
	}
	for _, model := range []Model{d.Baseline, d.Proposed} {
		found := false
		for _, driver := range model.Drivers {
			if driver.ID == d.VolumeDriver {
				if driver.Value != nil || len(driver.Inputs) > 0 || driver.Unit != d.VolumeUnit {
					return fmt.Errorf("%s: %s must be an unbound independent root of unit %s", model.Name, d.VolumeDriver, d.VolumeUnit)
				}
				found = true
			}
		}
		if !found {
			return fmt.Errorf("%s: volume driver %s is absent", model.Name, d.VolumeDriver)
		}
	}
	seen := map[string]bool{}
	options := map[string]bool{}
	for _, p := range d.Peak.Occupancy {
		if p.Option != "baseline" && p.Option != "proposed" {
			return fmt.Errorf("invalid occupancy option %q", p.Option)
		}
		key := p.Option + "/" + p.BusySeconds
		if seen[key] {
			return fmt.Errorf("duplicate occupancy pool %s", key)
		}
		seen[key], options[p.Option] = true, true
	}
	if !options["baseline"] || !options["proposed"] {
		return fmt.Errorf("declare peak occupancy for each option")
	}
	if d.Peak.Duration.Value != nil && *d.Peak.Duration.Value <= 0 {
		return fmt.Errorf("duration_hours must be positive")
	}
	if d.High != nil && d.High.Value != nil && d.Expected.Value != nil && *d.High.Value < *d.Expected.Value {
		return fmt.Errorf("high_monthly must be at least expected_monthly")
	}
	for _, volume := range []*Assumption{&d.Expected, d.High} {
		if volume != nil && volume.Value != nil && d.ValidThrough.Value != nil && *volume.Value > *d.ValidThrough.Value {
			return fmt.Errorf("monthly volume exceeds valid_through_monthly")
		}
	}
	return nil
}

func bindVolume(m Model, root string, volume Assumption) Model {
	m.Drivers = append([]Driver(nil), m.Drivers...)
	for i := range m.Drivers {
		if m.Drivers[i].ID == root {
			m.Drivers[i].Value = &volume
		}
	}
	return m
}

func compareMonth(a, b Model, root string, v Assumption) (MonthlyPoint, error) {
	c, err := Compare(bindVolume(a, root, v), bindVolume(b, root, v))
	return MonthlyPoint{Volume: v.Value, Baseline: c.Baseline.Total, Proposed: c.Proposed.Total, Delta: c.Delta}, err
}

func compareMonths(a, b Model, d Design) (MonthlyPoint, *MonthlyPoint, error) {
	expected, err := compareMonth(a, b, d.VolumeDriver, d.Expected)
	if err != nil || d.High == nil {
		return expected, nil, err
	}
	high, err := compareMonth(a, b, d.VolumeDriver, *d.High)
	return expected, &high, err
}

func changeFactor(m Model, change FactorChange) (Model, Link, error) {
	m.Drivers = append([]Driver(nil), m.Drivers...)
	for i, driver := range m.Drivers {
		for j, edge := range driver.Inputs {
			if edge.ID == change.EdgeID {
				m.Drivers[i].Inputs = append([]Link(nil), driver.Inputs...)
				m.Drivers[i].Inputs[j].Factor = change.Replacement
				return m, edge, nil
			}
		}
	}
	return Model{}, Link{}, fmt.Errorf("proposed sensitivity edge %q is absent", change.EdgeID)
}

// EvaluateDesign compares sourced graph templates without reading files or calling providers.
func EvaluateDesign(d Design) (DesignReport, error) {
	if err := d.validate(); err != nil {
		return DesignReport{}, err
	}
	r := DesignReport{SchemaVersion: DesignReportVersion, Name: d.Name, Status: "complete", Question: d.Question, Workload: d.Workload, CostScope: d.CostScope, VolumeUnit: d.VolumeUnit,
		BaselineName: d.Baseline.Name, ProposedName: d.Proposed.Name, Conditions: d.Conditions, Assumptions: d.sourceInputs(), Missing: []Gap{}}
	// Sort the source paths so incomplete reports and invalid-input diagnostics are reproducible.
	keys := make([]string, 0, len(r.Assumptions))
	for k := range r.Assumptions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := r.Assumptions[k]
		if err := v.validate(); err != nil {
			return DesignReport{}, fmt.Errorf("%s: %w", k, err)
		}
		if v.Value == nil {
			r.Missing = append(r.Missing, Gap{k, *v.Missing})
		}
	}
	var err error
	r.Expected, r.High, err = compareMonths(d.Baseline, d.Proposed, d)
	if err != nil {
		return DesignReport{}, err
	}
	base, err := analyzeCurve(d.Baseline, d.VolumeDriver)
	if err != nil {
		return DesignReport{}, err
	}
	proposed, err := analyzeCurve(d.Proposed, d.VolumeDriver)
	if err != nil {
		return DesignReport{}, err
	}
	r.BaselineCurve, r.ProposedCurve = base.curve, proposed.curve
	r.Peak, err = analyzePeak(d.Peak, base, proposed)
	if err != nil {
		return DesignReport{}, err
	}
	if r.Peak.Volume != nil && d.ValidThrough.Value != nil && *r.Peak.Volume > *d.ValidThrough.Value {
		return DesignReport{}, fmt.Errorf("arrivals in one peak window exceed valid_through_monthly")
	}
	r.BreakEven = crossover(base.curve, proposed.curve, d.ValidThrough.Value)
	r.BreakEven.PeakCapacity = r.Peak.Capacity
	changed, edge, err := changeFactor(d.Proposed, d.Sensitivity)
	if err != nil {
		return DesignReport{}, err
	}
	s := FactorResult{Description: d.Sensitivity.Description, EdgeID: edge.ID, FactorUnit: edge.OutputUnit + "/" + edge.InputUnit, Original: edge.Factor.Value, Replacement: d.Sensitivity.Replacement.Value}
	s.Expected, s.High, err = compareMonths(d.Baseline, changed, d)
	if err != nil {
		return DesignReport{}, err
	}
	alternative, err := analyzeCurve(changed, d.VolumeDriver)
	if err != nil {
		return DesignReport{}, err
	}
	s.Peak, err = analyzePeak(d.Peak, base, alternative)
	if err != nil {
		return DesignReport{}, err
	}
	s.BreakEven = crossover(base.curve, alternative.curve, d.ValidThrough.Value)
	s.BreakEven.PeakCapacity = s.Peak.Capacity
	r.Sensitivity = s
	if len(r.Missing) > 0 {
		r.Status = "incomplete"
	}
	r.Summary = designSummary(r)
	return r, nil
}

func designSummary(r DesignReport) string {
	money := func(v *float64, precision int) string {
		if v == nil {
			return "unknown"
		}
		return fmt.Sprintf("$%.*f", precision, *v)
	}
	quantity := func(v *float64) string {
		if v == nil {
			return "unknown"
		}
		return fmt.Sprintf("%g", *v)
	}
	point := func(p MonthlyPoint) string {
		return fmt.Sprintf("At a monthly volume of %s, costs are %s and %s.", quantity(p.Volume), money(p.Baseline, 2), money(p.Proposed, 2))
	}
	parts := []string{fmt.Sprintf("Modeled %s versus %s (%s).", r.BaselineName, r.ProposedName, r.VolumeUnit), point(r.Expected)}
	if r.High != nil {
		parts = append(parts, point(*r.High))
	}
	parts = append(parts, fmt.Sprintf("The %s-hour peak brings %s arrivals and adds %s versus %s, excluding monthly floors.", quantity(r.Peak.Duration), quantity(r.Peak.Volume), money(r.Peak.Baseline, 3), money(r.Peak.Proposed, 3)))
	cross := func(c Crossover) string {
		result := c.Interpretation
		if c.Status == "crosses" {
			result = fmt.Sprintf("Break-even is %.2f per month", *c.Volume)
		}
		if c.FirstCheaper != nil {
			result += fmt.Sprintf(". The first whole monthly volume favoring the proposed design is %.0f", *c.FirstCheaper)
		}
		return result + "."
	}
	parts = append(parts, cross(r.BreakEven), r.Sensitivity.Description+".", cross(r.Sensitivity.BreakEven))
	if high := r.Sensitivity.High; high != nil && r.High.Delta != nil && high.Delta != nil && ((*high.Delta < 0 && *r.High.Delta > 0) || (*high.Delta > 0 && *r.High.Delta < 0)) {
		parts = append(parts, fmt.Sprintf("The high-volume choice reverses at %s versus %s.", money(high.Baseline, 2), money(high.Proposed, 2)))
	}
	for _, scenario := range []struct {
		label string
		peak  PeakReport
	}{{"base case", r.Peak}, {"sensitivity case", r.Sensitivity.Peak}} {
		for _, pool := range scenario.peak.Occupancy {
			if pool.Status == "over_capacity" {
				parts = append(parts, fmt.Sprintf("In the %s, the %s pool %s needs %g average slots with %g available.", scenario.label, pool.Option, pool.BusySeconds, *pool.Required, *pool.Slots))
				break
			}
		}
	}
	return strings.Join(append(parts, r.CostScope), " ")
}

func multiply(a, b *float64) (*float64, error) {
	if a == nil || b == nil {
		return nil, nil
	}
	v := *a * *b
	if !finite(v) {
		return nil, fmt.Errorf("design quantity overflow")
	}
	return number(v), nil
}

func crossover(a, b CostCurve, limit *float64) Crossover {
	r := Crossover{Status: "missing", ValidThrough: limit, Interpretation: "Crossover requires both marginal rates, monthly floors and a valid range"}
	if a.Fixed == nil || b.Fixed == nil || a.Unit == nil || b.Unit == nil || limit == nil {
		return r
	}
	fixed, marginal := *b.Fixed-*a.Fixed, *b.Unit-*a.Unit
	// Comparing unlike scales with a fixed epsilon would erase small, real costs.
	if marginal == 0 {
		r.Status = "no_cross"
		switch {
		case fixed == 0:
			r.Status = "equal"
			r.Interpretation = "Both options cost the same throughout the declared range"
		case fixed < 0:
			r.Interpretation = "Proposed costs less throughout the declared range"
			r.FirstCheaper = number(0)
		default:
			r.Interpretation = "Baseline costs less throughout the declared range"
		}
		return r
	}
	q := -fixed / marginal
	if fixed < 0 {
		r.FirstCheaper = number(0)
	}
	if q < 0 {
		r.Status = "no_cross"
		if fixed < 0 {
			r.Interpretation = "Proposed costs less throughout the declared range"
		} else {
			r.Interpretation = "Baseline costs less throughout the declared range"
		}
		return r
	}
	if q > *limit {
		r.Status = "outside_range"
		r.Interpretation = "Crossover lies beyond the declared monthly volume range"
		return r
	}
	// A strict win starts after an exact integer tie. Allow a few representable
	// float steps for arithmetic in the source graph when locating that boundary.
	nearest := math.Round(q)
	if math.Abs(q-nearest) <= 4*math.Abs(q-math.Nextafter(q, 0)) {
		q = nearest
	}
	q = math.Abs(q) // Canonicalize signed zero in JSON.
	r.Status = "crosses"
	r.Volume = number(q)
	if marginal < 0 {
		r.Interpretation = "Proposed costs less above break-even within the declared range"
		first := math.Floor(q) + 1
		if first > q && first <= *limit {
			r.FirstCheaper = number(first)
		}
	} else {
		r.Interpretation = "Proposed costs less below break-even within the declared range"
	}
	return r
}
