package mcpserver

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/systemmodel"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestCompareOutcomesToolPreservesUnknownAcceptance(t *testing.T) {
	server := newTestServer(t, t.TempDir(), &fakeRunner{})
	server.enableDecisions = true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	st, ct := mcp.NewInMemoryTransports()
	go server.Serve(ctx, st)
	client := mcp.NewClient(&mcp.Implementation{Name: "outcomes-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer session.Close()
	input, err := os.ReadFile("../../examples/system/accepted-outcomes.json")
	require.NoError(t, err)
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: toolCompareOutcomes, Arguments: json.RawMessage(input)})
	require.NoError(t, err)
	require.False(t, res.IsError, "%v", res.Content)
	raw, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)
	var envelope struct{ Result systemmodel.OutcomeReport }
	require.NoError(t, json.Unmarshal(raw, &envelope))
	require.Equal(t, "incomplete", envelope.Result.Status)
	require.Nil(t, envelope.Result.Options[1].PerAccepted)
	require.InDelta(t, 0.8, *envelope.Result.Options[1].BreakEvenRate, 1e-12)
}
