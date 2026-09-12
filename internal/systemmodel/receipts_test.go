package systemmodel

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestReceiptCosts(t *testing.T) {
	raw, err := os.ReadFile("../../examples/system/condere-sanitized-receipts.json")
	require.NoError(t, err)
	var x ReceiptExport
	require.NoError(t, json.Unmarshal(raw, &x))
	r, err := SummarizeReceipts(x, raw)
	require.NoError(t, err)
	require.Nil(t, r.Total)
	require.Equal(t, .02, r.KnownCost)
	require.Len(t, r.SourceSHA256, 64)
	// Explicit zero with reported status is known; absent/uncertain status is not.
	x.Receipts[1].Metadata = map[string]json.RawMessage{"cost_status": json.RawMessage(`"reported"`)}
	r, err = SummarizeReceipts(x, raw)
	require.NoError(t, err)
	require.Equal(t, .02, *r.Total)
	x.Receipts[1].Metadata = nil
	r, err = SummarizeReceipts(x, raw)
	require.NoError(t, err)
	require.Nil(t, r.Total)
	x.Receipts[0].Metadata["private_payload"] = json.RawMessage(`"DO_NOT_EMIT"`)
	r, err = SummarizeReceipts(x, raw)
	require.NoError(t, err)
	out, err := json.Marshal(r)
	require.NoError(t, err)
	require.NotContains(t, string(out), "DO_NOT_EMIT")
	require.NotContains(t, string(out), x.AssessmentID)
	x.Receipts[0].InputTokens = nil
	r, err = SummarizeReceipts(x, raw)
	require.NoError(t, err)
	for _, g := range r.Groups {
		if g.Provider == "openrouter" {
			require.Nil(t, g.InputTokens)
		}
	}
	x.Receipts[0].Cost = number(-1)
	_, err = SummarizeReceipts(x, raw)
	require.Error(t, err)
}
