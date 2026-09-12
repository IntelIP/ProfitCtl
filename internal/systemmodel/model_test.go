package systemmodel

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func fixture(t *testing.T, name string) Model {
	t.Helper()
	f, err := os.Open("../../examples/system/" + name + ".json")
	require.NoError(t, err)
	defer f.Close()
	var m Model
	require.NoError(t, Decode(f, &m))
	return m
}
func TestSystemCascade(t *testing.T) {
	base := fixture(t, "condere-synthetic-baseline")
	cheap := fixture(t, "condere-synthetic-cheaper")
	adverse := fixture(t, "condere-synthetic-adverse")
	r, err := Evaluate(base)
	require.NoError(t, err)
	require.Equal(t, 191.5, *r.Total)
	require.InDelta(t, 80.85, *r.Margin, 1e-9)
	c, err := Compare(base, cheap)
	require.NoError(t, err)
	require.Equal(t, -50.0, *c.Delta)
	c, err = Compare(base, adverse)
	require.NoError(t, err)
	require.Equal(t, 85.0, *c.Delta)
	require.Contains(t, c.ChangedInputs, "edge:assessments-to-calls")
	for _, d := range c.Proposed.Drivers {
		if d.ID == "job-seconds" {
			require.Equal(t, 50000.0, *d.Quantity)
			require.Contains(t, d.Inputs, "edge:assessments-to-calls")
		}
	}
	for _, s := range c.Services {
		if s.ID == "jobs" {
			require.Equal(t, 60.0, *s.Delta)
		}
	}
}
func TestSystemUnknowns(t *testing.T) {
	m := fixture(t, "condere-production")
	r, err := Evaluate(m)
	require.NoError(t, err)
	require.Equal(t, "incomplete", r.Status)
	require.Nil(t, r.Total)
	require.Nil(t, r.Margin)
	require.Greater(t, len(r.Gaps), 10)
	b := fixture(t, "condere-synthetic-baseline")
	b.Drivers[0].Value = m.Drivers[0].Value
	b.Drivers[1].Inputs[0].Factor.Value = number(0)
	r, err = Evaluate(b)
	require.NoError(t, err)
	require.Nil(t, r.Drivers[1].Quantity, "unknown times zero must remain unknown")
	require.Nil(t, r.Total)
	require.Equal(t, 20.5, r.KnownCost)
}
func TestSystemValidation(t *testing.T) {
	cases := map[string]func(*Model){
		"cycle": func(m *Model) {
			m.Drivers[0].Value = nil
			m.Drivers[0].Inputs = []Link{{ID: "cycle", From: "calls", InputUnit: "model-call", OutputUnit: "assessment", Factor: *m.Drivers[7].Value}}
		},
		"dangling":                    func(m *Model) { m.Drivers[1].Inputs[0].From = "absent" },
		"units":                       func(m *Model) { m.Drivers[1].Inputs[0].InputUnit = "bytes" },
		"duplicate driver":            func(m *Model) { m.Drivers[1].ID = m.Drivers[0].ID },
		"duplicate edge":              func(m *Model) { m.Drivers[2].Inputs[0].ID = m.Drivers[1].Inputs[0].ID },
		"duplicate service":           func(m *Model) { m.Services[1].ID = m.Services[0].ID },
		"duplicate line":              func(m *Model) { m.Services[0].Costs = append(m.Services[0].Costs, m.Services[0].Costs[0]) },
		"bad service dependency":      func(m *Model) { m.Services[0].DependsOn = []string{"absent"} },
		"missing price evidence":      func(m *Model) { m.Services[0].Costs[0].Price.Evidence = nil },
		"unsupported schema":          func(m *Model) { m.SchemaVersion = "v999" },
		"negative":                    func(m *Model) { m.Drivers[0].Value.Value = number(-1) },
		"infinite":                    func(m *Model) { m.Drivers[0].Value.Value = number(math.Inf(1)) },
		"unknown without acquisition": func(m *Model) { m.Drivers[0].Value = &Assumption{} },
		"both root and inputs":        func(m *Model) { m.Drivers[1].Value = m.Drivers[0].Value },
		"free and charged":            func(m *Model) { m.Services[0].FreeReason = "free" },
		"unpriced service":            func(m *Model) { m.Services[0].Costs = nil },
		"too many drivers":            func(m *Model) { m.Drivers = make([]Driver, 513) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m := fixture(t, "condere-synthetic-baseline")
			mutate(&m)
			_, err := Evaluate(m)
			require.Error(t, err)
		})
	}
	m := fixture(t, "condere-synthetic-baseline")
	m.Drivers[0].Value.Value = number(math.MaxFloat64)
	_, err := Evaluate(m)
	require.ErrorContains(t, err, "overflow")
}
func TestStrictSystemJSON(t *testing.T) {
	for _, s := range []string{`{"name":"a","name":"b"}`, `{"extra":1}`, `null {}`, `{"schema_version":"profitctl.system/v1"} false`, strings.Repeat("[", 34) + strings.Repeat("]", 34), strings.Repeat(" ", MaxInputBytes+1)} {
		var m Model
		require.Error(t, Decode(strings.NewReader(s), &m))
	}
	m := fixture(t, "condere-synthetic-baseline")
	raw, err := json.Marshal(m)
	require.NoError(t, err)
	var copy Model
	require.NoError(t, Decode(bytes.NewReader(raw), &copy))
	require.NoError(t, copy.Validate())
}
