package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/systemmodel"
	"github.com/stretchr/testify/require"
)

func TestSystemDesignFileAndStdin(t *testing.T) {
	const path = "../examples/design/condere-queue.json"
	input, err := os.ReadFile(path)
	require.NoError(t, err)
	run := func(arg string, stdin []byte, code int) []byte {
		cmd := systemCommand()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetIn(bytes.NewReader(stdin))
		cmd.SetArgs([]string{"design", arg})
		require.Equal(t, code, ExitCode(cmd.Execute()))
		return out.Bytes()
	}
	require.Equal(t, run(path, nil, 0), run("-", input, 0))
	run("-", []byte(`{"path":"/private/example","prompt":"scan"}`), 2)
	run("-", []byte(`{"name":"first","name":"second"}`), 2)
	var d systemmodel.Design
	require.NoError(t, systemmodel.Decode(bytes.NewReader(input), &d))
	d.Peak.Arrivals = systemmodel.Assumption{Missing: &systemmodel.Acquisition{Question: "How many arrivals?", Source: "Hourly admission counts", Owner: "Service owner"}}
	b, err := json.Marshal(d)
	require.NoError(t, err)
	var r systemmodel.DesignReport
	require.NoError(t, json.Unmarshal(run("-", b, 1), &r))
	require.Equal(t, "incomplete", r.Status)
	require.Nil(t, r.Peak.Volume)
	require.NotNil(t, r.Expected.Delta)
}
