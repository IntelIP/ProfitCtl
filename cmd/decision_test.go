package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/decision"
	"github.com/stretchr/testify/require"
)

func TestDecisionCLI(t *testing.T) {
	root := t.TempDir()
	input := `{"project":"example","question":"Keep cache?","workload":"accepted plans","source_revision":"fixture-v1","window":"May","alternatives":[{"name":"keep","cost":"unknown","quality":"freshness unknown"},{"name":"remove","cost":"unknown","quality":"current data"}],"evidence":[{"kind":"unknown","source":"usage export","finding":"not yet available"}],"missing":["Count repeat lookups"]}`
	run := func(stdin string, args ...string) (decision.Revision, int) {
		t.Helper()
		cmd := decisionCommand()
		cmd.SetIn(strings.NewReader(stdin))
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs(append([]string{"--workspace", root}, args...))
		err := cmd.Execute()
		var record decision.Revision
		if err == nil {
			require.NoError(t, json.Unmarshal(out.Bytes(), &record))
		}
		return record, ExitCode(err)
	}
	r1, code := run(input, "create", "sample", "--input", "-")
	require.Zero(t, code)
	require.Equal(t, 1, r1.Revision)
	r2, code := run(strings.ReplaceAll(input, "Count repeat lookups", "Collect two weeks"), "revise", "sample", "--expected", "1", "--input", "-")
	require.Zero(t, code)
	require.Equal(t, 2, r2.Revision)
	latest, code := run("", "read", "sample")
	require.Zero(t, code)
	require.Equal(t, r2, latest)
	old, code := run("", "read", "sample", "--revision", "1")
	require.Zero(t, code)
	require.Equal(t, r1, old)
	_, code = run(input, "revise", "sample", "--expected", "1", "--input", "-")
	require.Equal(t, 4, code)
	_, code = run(input, "revise", "sample", "--input", "-")
	require.Equal(t, 2, code)
	_, code = run(input+" {}", "create", "another", "--input", "-")
	require.Equal(t, 2, code)
}
