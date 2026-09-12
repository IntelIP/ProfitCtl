package cmd

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestSystemCLI(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"evaluate", "../examples/system/condere-synthetic-baseline.json"}, 0},
		{[]string{"evaluate", "../examples/system/condere-production.json"}, 1},
		{[]string{"compare", "../examples/system/condere-synthetic-baseline.json", "../examples/system/condere-synthetic-adverse.json"}, 0},
		{[]string{"receipts", "../examples/system/condere-sanitized-receipts.json"}, 1},
	} {
		cmd := systemCommand()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs(tc.args)
		err := cmd.Execute()
		require.Equal(t, tc.code, ExitCode(err))
		require.True(t, json.Valid(buf.Bytes()))
	}
	cmd := systemCommand()
	cmd.SetArgs([]string{"evaluate", "missing.json"})
	require.Equal(t, 2, ExitCode(cmd.Execute()))
	bad := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(bad, []byte(`{"schema_version":"bad","unexpected":1}`), 0600))
	cmd = systemCommand()
	cmd.SetArgs([]string{"evaluate", bad})
	require.Equal(t, 2, ExitCode(cmd.Execute()))
}
