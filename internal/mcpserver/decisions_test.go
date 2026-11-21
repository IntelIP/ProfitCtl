package mcpserver

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDecisionModePublishesScopedToolsAndPreservesHistory(t *testing.T) {
	root := t.TempDir()
	server := newTestServer(t, root, &fakeRunner{})
	server.enableDecisions = true
	ctx := context.Background()
	serverSide, clientSide := mcp.NewInMemoryTransports()
	go server.Serve(ctx, serverSide)
	client := mcp.NewClient(&mcp.Implementation{Name: "decision-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	list, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tools) != 9 {
		t.Fatalf("decision server advertises %d tools, want 9", len(list.Tools))
	}
	for _, tool := range list.Tools {
		write := tool.Name == toolCreateDecision || tool.Name == toolReviseDecision
		if tool.Annotations == nil || tool.Annotations.ReadOnlyHint == write || tool.Annotations.OpenWorldHint == nil || *tool.Annotations.OpenWorldHint {
			t.Fatalf("unexpected authority for %q: %#v", tool.Name, tool.Annotations)
		}
	}

	call := func(name string, args map[string]any) (string, int) {
		t.Helper()
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(result.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		var env struct {
			Error  *ErrorInfo `json:"error"`
			Result struct {
				Revision int `json:"revision"`
			} `json:"result"`
		}
		if err := json.Unmarshal(encoded, &env); err != nil {
			t.Fatal(err)
		}
		if result.IsError != (env.Error != nil) {
			t.Fatalf("%s error status disagrees with payload: %s", name, encoded)
		}
		if env.Error != nil {
			return env.Error.Code, 0
		}
		return "", env.Result.Revision
	}
	record := map[string]any{
		"project": "Condere", "question": "Which depth?", "source_revision": "test", "workload": "reviewed assessment", "window": "pilot",
		"alternatives": []map[string]string{{"name": "standard", "cost": "unmeasured", "quality": "review required"}, {"name": "deep", "cost": "unmeasured", "quality": "review required"}},
		"evidence":     []map[string]string{{"kind": "unknown", "source": "pilot review", "finding": "sample pending"}},
		"missing":      []string{"Observe acceptance and spend on matching work."},
	}
	args := map[string]any{"id": "depth", "case": record}
	if code, rev := call(toolCreateDecision, args); code != "" || rev != 1 {
		t.Fatalf("create: code=%q revision=%d", code, rev)
	}
	found, err := session.CallTool(ctx, &mcp.CallToolParams{Name: toolListDecisions, Arguments: map[string]any{"query": "Condere"}})
	if err != nil || found.IsError {
		t.Fatalf("find decision: %v, %#v", err, found)
	}
	var matches struct {
		Result struct {
			Cases []struct {
				ID string `json:"id"`
			} `json:"cases"`
		} `json:"result"`
	}
	data, err := json.Marshal(found.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &matches); err != nil || len(matches.Result.Cases) != 1 || matches.Result.Cases[0].ID != "depth" {
		t.Fatalf("unexpected decision search result: %v, %s", err, data)
	}
	args["expected_revision"] = 1
	record["choice"] = "standard"
	if code, rev := call(toolReviseDecision, args); code != "" || rev != 2 {
		t.Fatalf("revise: code=%q revision=%d", code, rev)
	}
	record["choice"] = "deep"
	if code, _ := call(toolReviseDecision, args); code != "revision_conflict" {
		t.Fatalf("stale update code=%q", code)
	}
	if code, rev := call(toolReadDecision, map[string]any{"id": "depth", "revision": 1}); code != "" || rev != 1 {
		t.Fatalf("historical read: code=%q revision=%d", code, rev)
	}

	// Even if the server was configured for local writes, HTTP retains the
	// original read-only authority.
	httpServer := httptest.NewServer(server.HTTPHandler())
	defer httpServer.Close()
	httpSession, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: httpServer.URL}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer httpSession.Close()
	readOnly, err := httpSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(readOnly.Tools) != 4 {
		t.Fatalf("HTTP advertises %d tools, want 4", len(readOnly.Tools))
	}
}
