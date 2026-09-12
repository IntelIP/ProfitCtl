package mcpserver

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHTTPHandlerServesCurrentStatelessProtocol(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "scenarios", "valid.yml"), "simulation:\n  iterations: 1\n")

	runner := &fakeRunner{respond: func(_ string, args []string) processResult {
		if args[0] == "simulate" {
			return processResult{exitCode: 0, stdout: []byte(`{"covenants":{"passed":true}}`)}
		}
		return processResult{exitCode: 0}
	}}
	server := newTestServer(t, root, runner)
	httpServer := httptest.NewServer(server.HTTPHandler())
	defer httpServer.Close()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "profitctl-http-test", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: httpServer.URL}, nil)
	if err != nil {
		t.Fatalf("connect HTTP MCP client: %v", err)
	}
	defer session.Close()

	if got := session.InitializeResult().ProtocolVersion; got != "2026-07-28" {
		t.Fatalf("protocol version = %q, want 2026-07-28", got)
	}
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list HTTP MCP tools: %v", err)
	}
	if len(tools.Tools) != 4 {
		t.Fatalf("tool count = %d, want 4", len(tools.Tools))
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolSimulate,
		Arguments: map[string]any{"scenario_path": "scenarios/valid.yml"},
	})
	if err != nil {
		t.Fatalf("call HTTP MCP tool: %v", err)
	}
	if result.IsError {
		t.Fatalf("HTTP MCP tool returned error: %#v", result.Content)
	}
	encoded, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal HTTP result: %v", err)
	}
	var envelope ResultEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatalf("decode HTTP result: %v", err)
	}
	if envelope.Outcome != OutcomePassed {
		t.Fatalf("HTTP outcome = %q, want %q", envelope.Outcome, OutcomePassed)
	}
}
