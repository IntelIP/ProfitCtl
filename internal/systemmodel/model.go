// Package systemmodel evaluates sourced monthly workload graphs without network calls.
package systemmodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strings"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
)

const SchemaVersion = "profitctl.system/v1"
const MaxInputBytes = 2 * 1024 * 1024

type Acquisition struct {
	Question string `json:"question"`
	Source   string `json:"source"`
	Owner    string `json:"owner"`
}
type Assumption struct {
	Value    *float64         `json:"value"`
	Evidence *costv1.Evidence `json:"evidence,omitempty"`
	Missing  *Acquisition     `json:"missing,omitempty"`
}
type Link struct {
	ID         string     `json:"id"`
	From       string     `json:"from"`
	InputUnit  string     `json:"input_unit"`
	OutputUnit string     `json:"output_unit"`
	Factor     Assumption `json:"factor"`
}
type Driver struct {
	ID     string      `json:"id"`
	Unit   string      `json:"unit"`
	Value  *Assumption `json:"value,omitempty"`
	Inputs []Link      `json:"inputs,omitempty"`
}
type CostLine struct {
	ID     string     `json:"id"`
	Driver string     `json:"driver"`
	Unit   string     `json:"unit"`
	Price  Assumption `json:"usd_per_unit"`
}
type Service struct {
	ID         string     `json:"id"`
	Provider   string     `json:"provider"`
	Role       string     `json:"role"`
	Evidence   []string   `json:"evidence"`
	DependsOn  []string   `json:"depends_on,omitempty"`
	Costs      []CostLine `json:"costs"`
	FreeReason string     `json:"free_reason,omitempty"`
}
type Model struct {
	SchemaVersion string     `json:"schema_version"`
	Name          string     `json:"name"`
	Period        string     `json:"period"`
	Currency      string     `json:"currency"`
	Revenue       Assumption `json:"revenue_usd"`
	Drivers       []Driver   `json:"drivers"`
	Services      []Service  `json:"services"`
}
type Gap struct {
	Input       string      `json:"input"`
	Acquisition Acquisition `json:"acquisition"`
}
type DriverResult struct {
	ID        string   `json:"id"`
	Unit      string   `json:"unit"`
	Quantity  *float64 `json:"quantity"`
	DependsOn []string `json:"depends_on"`
	Inputs    []string `json:"root_inputs"`
}
type LineResult struct {
	ID     string   `json:"id"`
	Driver string   `json:"driver"`
	Cost   *float64 `json:"cost_usd"`
}
type ServiceResult struct {
	Provider   string       `json:"provider"`
	Role       string       `json:"role"`
	Evidence   []string     `json:"evidence"`
	DependsOn  []string     `json:"depends_on"`
	FreeReason string       `json:"free_reason,omitempty"`
	ID         string       `json:"id"`
	KnownCost  float64      `json:"known_cost_usd"`
	Total      *float64     `json:"total_cost_usd"`
	Lines      []LineResult `json:"lines"`
}
type Report struct {
	SchemaVersion string          `json:"schema_version"`
	Name          string          `json:"name"`
	Status        string          `json:"status"`
	Drivers       []DriverResult  `json:"drivers"`
	Services      []ServiceResult `json:"services"`
	Gaps          []Gap           `json:"missing_inputs"`
	KnownCost     float64         `json:"known_cost_usd"`
	Total         *float64        `json:"total_cost_usd"`
	Revenue       *float64        `json:"revenue_usd"`
	Margin        *float64        `json:"operating_margin_percent"`
	Limitations   []string        `json:"limitations"`
}

// Decode rejects unknown and duplicate fields, trailing data and oversized inputs.
func Decode(r io.Reader, target any) error {
	data, err := io.ReadAll(io.LimitReader(r, MaxInputBytes+1))
	if err != nil {
		return err
	}
	if len(data) > MaxInputBytes {
		return fmt.Errorf("input exceeds %d bytes", MaxInputBytes)
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("input must contain one JSON value")
	}
	return nil
}
func rejectDuplicateKeys(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	var walk func(int) error
	walk = func(depth int) error {
		if depth > 32 {
			return fmt.Errorf("JSON nesting exceeds 32 levels")
		}
		t, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		keys := map[string]bool{}
		for d.More() {
			if delim == '{' {
				key, err := d.Token()
				if err != nil {
					return err
				}
				s, ok := key.(string)
				if !ok || keys[s] {
					return fmt.Errorf("duplicate or invalid JSON key")
				}
				keys[s] = true
			}
			if err := walk(depth + 1); err != nil {
				return err
			}
		}
		_, err = d.Token()
		return err
	}
	return walk(0)
}

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,79}$`)

func finite(v float64) bool     { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 }
func number(v float64) *float64 { return &v }
func (a Assumption) validate() error {
	if a.Value == nil {
		if a.Missing == nil || strings.TrimSpace(a.Missing.Question) == "" || strings.TrimSpace(a.Missing.Source) == "" || strings.TrimSpace(a.Missing.Owner) == "" {
			return fmt.Errorf("unknown value requires question, source and owner")
		}
		if a.Evidence != nil {
			return fmt.Errorf("unknown value cannot claim numeric evidence")
		}
		return nil
	}
	if !finite(*a.Value) {
		return fmt.Errorf("value must be finite and nonnegative")
	}
	if a.Missing != nil || a.Evidence == nil {
		return fmt.Errorf("known value requires evidence and no missing-input declaration")
	}
	return a.Evidence.Validate()
}
func (m Model) Validate() error {
	if m.SchemaVersion != SchemaVersion || strings.TrimSpace(m.Name) == "" || m.Period != "monthly" || m.Currency != "USD" {
		return fmt.Errorf("expected %s, name, monthly period and USD currency", SchemaVersion)
	}
	if len(m.Drivers) == 0 || len(m.Drivers) > 512 || len(m.Services) == 0 || len(m.Services) > 256 {
		return fmt.Errorf("expected 1..512 drivers and 1..256 services")
	}
	if err := m.Revenue.validate(); err != nil {
		return fmt.Errorf("revenue: %w", err)
	}
	drivers := map[string]Driver{}
	edges := map[string]bool{}
	for _, d := range m.Drivers {
		if !idPattern.MatchString(d.ID) || drivers[d.ID].ID != "" || strings.TrimSpace(d.Unit) == "" {
			return fmt.Errorf("invalid/duplicate driver ID or missing unit: %s", d.ID)
		}
		drivers[d.ID] = d
		if (d.Value == nil) == (len(d.Inputs) == 0) {
			return fmt.Errorf("driver %s requires either a value or inputs", d.ID)
		}
		if d.Value != nil {
			if err := d.Value.validate(); err != nil {
				return fmt.Errorf("driver %s: %w", d.ID, err)
			}
		}
		for _, e := range d.Inputs {
			if !idPattern.MatchString(e.ID) || edges[e.ID] {
				return fmt.Errorf("invalid/duplicate edge %s", e.ID)
			}
			edges[e.ID] = true
			if err := e.Factor.validate(); err != nil {
				return fmt.Errorf("edge %s: %w", e.ID, err)
			}
		}
	}
	if len(edges) > 1024 {
		return fmt.Errorf("too many driver edges")
	}
	for _, d := range m.Drivers {
		for _, e := range d.Inputs {
			p, ok := drivers[e.From]
			if !ok || e.InputUnit != p.Unit || e.OutputUnit != d.Unit {
				return fmt.Errorf("edge %s has missing parent or mismatched units", e.ID)
			}
		}
	}
	services := map[string]bool{}
	for _, s := range m.Services {
		if !idPattern.MatchString(s.ID) || services[s.ID] || strings.TrimSpace(s.Provider) == "" || strings.TrimSpace(s.Role) == "" || len(s.Evidence) == 0 {
			return fmt.Errorf("service %s needs unique ID, provider, role and evidence", s.ID)
		}
		services[s.ID] = true
		for _, ref := range s.Evidence {
			if strings.TrimSpace(ref) == "" {
				return fmt.Errorf("service %s has empty evidence", s.ID)
			}
		}
		if (len(s.Costs) == 0) == (strings.TrimSpace(s.FreeReason) == "") {
			return fmt.Errorf("service %s requires cost lines or an explicit free reason", s.ID)
		}
		seen := map[string]bool{}
		for _, c := range s.Costs {
			d, ok := drivers[c.Driver]
			if !idPattern.MatchString(c.ID) || seen[c.ID] || !ok || c.Unit != d.Unit {
				return fmt.Errorf("invalid cost line %s/%s", s.ID, c.ID)
			}
			seen[c.ID] = true
			if err := c.Price.validate(); err != nil {
				return fmt.Errorf("price %s/%s: %w", s.ID, c.ID, err)
			}
		}
	}
	for _, s := range m.Services {
		seen := map[string]bool{}
		for _, id := range s.DependsOn {
			if !services[id] || id == s.ID || seen[id] {
				return fmt.Errorf("invalid dependency %s -> %s", s.ID, id)
			}
			seen[id] = true
		}
	}
	// Service calls may be bidirectional; calculation edges must be acyclic.
	state := map[string]int{}
	var visit func(string) error
	visit = func(id string) error {
		if state[id] == 1 {
			return fmt.Errorf("driver cycle at %s", id)
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, e := range drivers[id].Inputs {
			if err := visit(e.From); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for id := range drivers {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}
func Evaluate(m Model) (Report, error) {
	if err := m.Validate(); err != nil {
		return Report{}, err
	}
	r := Report{SchemaVersion: SchemaVersion, Name: m.Name, Status: "complete", Gaps: []Gap{}, Drivers: []DriverResult{}, Services: []ServiceResult{}, Revenue: m.Revenue.Value, Limitations: []string{"Consequences follow declared links, not automatically inferred causality.", "Monthly planning arithmetic is not live billing, capacity, latency or quality proof."}}
	gap := func(id string, a Assumption) {
		if a.Value == nil {
			r.Gaps = append(r.Gaps, Gap{id, *a.Missing})
		}
	}
	gap("revenue", m.Revenue)
	defs := map[string]Driver{}
	results := map[string]DriverResult{}
	for _, d := range m.Drivers {
		defs[d.ID] = d
		if d.Value != nil {
			gap("driver:"+d.ID, *d.Value)
		}
		for _, e := range d.Inputs {
			gap("edge:"+e.ID, e.Factor)
		}
	}
	var eval func(string) (DriverResult, error)
	eval = func(id string) (DriverResult, error) {
		if x, ok := results[id]; ok {
			return x, nil
		}
		d := defs[id]
		x := DriverResult{ID: id, Unit: d.Unit, DependsOn: []string{}, Inputs: []string{}}
		if d.Value != nil {
			x.Quantity = d.Value.Value
			x.Inputs = []string{"driver:" + id}
		} else {
			total := 0.0
			known := true
			inputs := map[string]bool{}
			for _, e := range d.Inputs {
				p, err := eval(e.From)
				if err != nil {
					return x, err
				}
				x.DependsOn = append(x.DependsOn, e.From)
				inputs["edge:"+e.ID] = true
				for _, root := range p.Inputs {
					inputs[root] = true
				}
				if p.Quantity == nil || e.Factor.Value == nil {
					known = false
					continue
				}
				total += *p.Quantity * *e.Factor.Value
				if !finite(total) {
					return x, fmt.Errorf("driver %s arithmetic overflow", id)
				}
			}
			if known {
				x.Quantity = number(total)
			}
			for root := range inputs {
				x.Inputs = append(x.Inputs, root)
			}
			sort.Strings(x.Inputs)
		}
		results[id] = x
		return x, nil
	}
	for _, d := range m.Drivers {
		x, err := eval(d.ID)
		if err != nil {
			return Report{}, err
		}
		r.Drivers = append(r.Drivers, x)
	}
	allKnown := true
	for _, s := range m.Services {
		x := ServiceResult{ID: s.ID, Provider: s.Provider, Role: s.Role, Evidence: s.Evidence, DependsOn: s.DependsOn, FreeReason: s.FreeReason, Lines: []LineResult{}}
		known := true
		for _, c := range s.Costs {
			gap("price:"+s.ID+"/"+c.ID, c.Price)
			line := LineResult{ID: c.ID, Driver: c.Driver}
			q := results[c.Driver].Quantity
			if q == nil || c.Price.Value == nil {
				known = false
			} else {
				v := *q * *c.Price.Value
				if !finite(v) {
					return Report{}, fmt.Errorf("cost %s/%s arithmetic overflow", s.ID, c.ID)
				}
				line.Cost = number(v)
				x.KnownCost += v
			}
			x.Lines = append(x.Lines, line)
		}
		if !finite(x.KnownCost) {
			return Report{}, fmt.Errorf("service total overflow")
		}
		if known {
			x.Total = number(x.KnownCost)
		} else {
			allKnown = false
		}
		r.KnownCost += x.KnownCost
		r.Services = append(r.Services, x)
	}
	if !finite(r.KnownCost) {
		return Report{}, fmt.Errorf("system total overflow")
	}
	if allKnown {
		r.Total = number(r.KnownCost)
	}
	if len(r.Gaps) > 0 {
		r.Status = "incomplete"
	}
	if r.Total != nil && r.Revenue != nil && *r.Revenue > 0 {
		v := 100 * (1 - *r.Total / *r.Revenue)
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return Report{}, fmt.Errorf("margin overflow")
		}
		r.Margin = &v
	}
	sort.Slice(r.Gaps, func(i, j int) bool { return r.Gaps[i].Input < r.Gaps[j].Input })
	return r, nil
}
