package mcpserver

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type recordedCall struct {
	binary string
	args   []string
	dir    string
}

type fakeRunner struct {
	mu      sync.Mutex
	calls   []recordedCall
	respond func(binary string, args []string) processResult
}

func (runner *fakeRunner) Run(_ context.Context, binary string, args []string, dir string, _ int) processResult {
	runner.mu.Lock()
	runner.calls = append(runner.calls, recordedCall{
		binary: binary,
		args:   append([]string(nil), args...),
		dir:    dir,
	})
	runner.mu.Unlock()
	if runner.respond == nil {
		return processResult{exitCode: 0}
	}
	return runner.respond(binary, args)
}

func (runner *fakeRunner) Calls() []recordedCall {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	return append([]recordedCall(nil), runner.calls...)
}

func TestMCPToolsExposeReadOnlyContractsAndStructuredResult(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "scenarios", "valid.yml"), "simulation:\n  iterations: 1\n")

	runner := &fakeRunner{respond: func(_ string, args []string) processResult {
		switch args[0] {
		case "simulate":
			return processResult{exitCode: 0, stdout: []byte(`{"covenants":{"passed":true}}`)}
		case "compare":
			return processResult{exitCode: 0, stdout: []byte(`{"items":[]}`)}
		default:
			return processResult{exitCode: 0}
		}
	}}
	server := newTestServer(t, root, runner)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- server.Serve(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "profitctl-test", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect MCP client: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools.Tools) != 4 {
		t.Fatalf("tool count = %d, want 4", len(tools.Tools))
	}
	wantTools := map[string]bool{
		toolValidate: false,
		toolSimulate: false,
		toolCompare:  false,
		toolJudge:    false,
	}
	for _, tool := range tools.Tools {
		if _, ok := wantTools[tool.Name]; !ok {
			t.Fatalf("unexpected tool %q", tool.Name)
		}
		wantTools[tool.Name] = true
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint || tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint || tool.Annotations.OpenWorldHint == nil || *tool.Annotations.OpenWorldHint {
			t.Fatalf("tool %q annotations are not read-only and closed-world: %#v", tool.Name, tool.Annotations)
		}
	}
	for name, present := range wantTools {
		if !present {
			t.Fatalf("missing tool %q", name)
		}
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolSimulate,
		Arguments: map[string]any{"scenario_path": "scenarios/valid.yml"},
	})
	if err != nil {
		t.Fatalf("call simulate tool: %v", err)
	}
	if result.IsError {
		t.Fatalf("simulate tool returned MCP error: %#v", result.Content)
	}
	var envelope ResultEnvelope
	encoded, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured result: %v", err)
	}
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatalf("decode structured result: %v", err)
	}
	if envelope.SchemaVersion != ResultSchemaVersion || envelope.Outcome != OutcomePassed {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
	if len(envelope.Input.Paths) != 1 || envelope.Input.Paths[0] != "scenarios/valid.yml" || len(envelope.Input.SHA256) != 1 {
		t.Fatalf("unexpected input evidence: %#v", envelope.Input)
	}

	calls := runner.Calls()
	if len(calls) != 1 || calls[0].binary != server.binary || calls[0].dir != server.workspaceRoot {
		t.Fatalf("unexpected child call: %#v", calls)
	}
	if got := calls[0].args; len(got) != 4 || got[0] != "simulate" || got[1] != "--file" || got[2] != "scenarios/valid.yml" || got[3] != "--json" {
		t.Fatalf("unexpected child args: %#v", got)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("MCP server did not stop after cancellation")
	}
}

func TestPathBoundariesFailClosedBeforeChildExecution(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.yml")
	writeTestFile(t, outside, "simulation:\n  iterations: 1\n")
	writeTestFile(t, filepath.Join(root, "scenarios", "escape-calibration.yml"), "simulation:\n  iterations: 1\ncalibration_file: ../../outside.yml\n")
	writeTestFile(t, filepath.Join(root, "scenarios", "too-many.yml"), "simulation:\n  iterations: 10001\n")
	if err := os.Symlink(outside, filepath.Join(root, "scenarios", "symlink-escape.yml")); err != nil {
		t.Fatalf("create symlink fixture: %v", err)
	}

	runner := &fakeRunner{}
	server := newTestServer(t, root, runner)
	cases := []struct {
		name string
		path string
		code string
	}{
		{name: "parent", path: "../outside.yml", code: "path_escape_rejected"},
		{name: "absolute", path: outside, code: "path_must_be_root_relative"},
		{name: "symlink", path: "scenarios/symlink-escape.yml", code: "path_escape_rejected"},
		{name: "calibration", path: "scenarios/escape-calibration.yml", code: "calibration_path_outside_workspace"},
		{name: "iteration cap", path: "scenarios/too-many.yml", code: "simulation_iterations_exceed_limit"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			envelope := server.simulateScenario(context.Background(), testCase.path, time.Now())
			if envelope.Outcome != OutcomeInvalidInput || envelope.Error == nil || envelope.Error.Code != testCase.code {
				t.Fatalf("unexpected envelope: %#v", envelope)
			}
		})
	}
	if calls := runner.Calls(); len(calls) != 0 {
		t.Fatalf("invalid paths started child processes: %#v", calls)
	}
}

func TestCompareAndJudgeUseBoundedLocalExecutables(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "scenarios", "a.yml"), "simulation:\n  iterations: 1\n")
	writeTestFile(t, filepath.Join(root, "scenarios", "b.yml"), "simulation:\n  iterations: 1\n")
	writeTestFile(t, filepath.Join(root, "decisions", "incomplete.md"), "Recommendation only")

	runner := &fakeRunner{respond: func(_ string, args []string) processResult {
		if args[0] == "compare" {
			return processResult{exitCode: 1, stdout: []byte(`{"items":[]}`)}
		}
		return processResult{exitCode: 1, stdout: []byte(`{"passed":false,"files":[]}`)}
	}}
	server := newTestServer(t, root, runner)

	comparison := server.compareScenarios(context.Background(), []string{"scenarios/a.yml", "scenarios/b.yml"}, time.Now())
	if comparison.Outcome != OutcomeCovenantFailed {
		t.Fatalf("compare outcome = %q, want %q", comparison.Outcome, OutcomeCovenantFailed)
	}
	judgment := server.judgeStandards(context.Background(), "decisions/incomplete.md", time.Now())
	if judgment.Outcome != OutcomeStandardsFailed {
		t.Fatalf("judge outcome = %q, want %q", judgment.Outcome, OutcomeStandardsFailed)
	}

	calls := runner.Calls()
	if len(calls) != 2 {
		t.Fatalf("child calls = %d, want 2", len(calls))
	}
	if calls[0].binary != server.binary {
		t.Fatalf("unexpected compare executable: %#v", calls[0])
	}
	if got := calls[0].args; len(got) != 4 || got[0] != "compare" || got[1] != "scenarios/a.yml" || got[2] != "scenarios/b.yml" || got[3] != "--json" {
		t.Fatalf("unexpected compare call: %#v", calls[0])
	}
	if calls[1].binary != server.standardsBinary || len(calls[1].args) != 1 || calls[1].args[0] != "decisions/incomplete.md" {
		t.Fatalf("unexpected standards call: %#v", calls[1])
	}
}

func newTestServer(t *testing.T, root string, runner processRunner) *Server {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve test executable: %v", err)
	}
	server, err := newWithRunner(Config{
		WorkspaceRoot:   root,
		Binary:          binary,
		StandardsBinary: binary,
	}, runner)
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	return server
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
