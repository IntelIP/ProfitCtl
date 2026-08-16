package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolValidate = "profitctl_validate_scenario"
	toolSimulate = "profitctl_simulate_scenario"
	toolCompare  = "profitctl_compare_scenarios"
	toolJudge    = "profitctl_judge_standards"

	mcpImplementationVersion = "0.2.0"
)

type singleScenarioInput struct {
	ScenarioPath string `json:"scenario_path"`
}

type compareScenariosInput struct {
	ScenarioPaths []string `json:"scenario_paths"`
}

type standardsInput struct {
	TargetPath string `json:"target_path"`
}

// Run serves the current MCP protocol over stdin and stdout until the client disconnects.
func (s *Server) Run(ctx context.Context) error {
	return s.Serve(ctx, &mcp.StdioTransport{})
}

// Serve runs the MCP protocol on a supplied transport. It supports in-memory
// transport in tests while production uses Run and stdio.
func (s *Server) Serve(ctx context.Context, transport mcp.Transport) error {
	return s.MCPServer().Run(ctx, transport)
}

// MCPServer returns the four-tool MCP server for the already-fixed workspace.
func (s *Server) MCPServer() *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{Name: "profitctl-local", Version: mcpImplementationVersion},
		&mcp.ServerOptions{
			Instructions: "Use only existing root-relative files in the configured local workspace. These four tools are read-only; do not use them for provider lookup, repository detection, scenario creation, or writes.",
			Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
			Capabilities: &mcp.ServerCapabilities{},
		},
	)

	s.registerValidateTool(server)
	s.registerSimulateTool(server)
	s.registerCompareTool(server)
	s.registerJudgeTool(server)
	return server
}

func (s *Server) registerValidateTool(server *mcp.Server) {
	server.AddTool(&mcp.Tool{
		Name:         toolValidate,
		Title:        "Validate ProfitCtl scenario",
		Description:  "Use this when the user needs to confirm one existing local scenario is structurally valid before simulation.",
		InputSchema:  singlePathSchema("scenario_path", "Existing root-relative .yml or .yaml scenario path."),
		OutputSchema: resultSchema(),
		Annotations:  readOnlyAnnotations(),
	}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		started := time.Now()
		var input singleScenarioInput
		if err := decodeArguments(request.Params.Arguments, &input); err != nil {
			return toolResult(failEnvelope(toolValidate, started, OutcomeInvalidInput, "invalid_arguments", "The tool arguments do not match the required schema.", "scenario_path")), nil
		}
		return toolResult(s.validateScenario(ctx, input.ScenarioPath, started)), nil
	})
}

func (s *Server) registerSimulateTool(server *mcp.Server) {
	server.AddTool(&mcp.Tool{
		Name:         toolSimulate,
		Title:        "Simulate ProfitCtl scenario",
		Description:  "Use this when the user needs margin, stress, and covenant evidence for one existing validated local scenario.",
		InputSchema:  singlePathSchema("scenario_path", "Existing root-relative .yml or .yaml scenario path."),
		OutputSchema: resultSchema(),
		Annotations:  readOnlyAnnotations(),
	}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		started := time.Now()
		var input singleScenarioInput
		if err := decodeArguments(request.Params.Arguments, &input); err != nil {
			return toolResult(failEnvelope(toolSimulate, started, OutcomeInvalidInput, "invalid_arguments", "The tool arguments do not match the required schema.", "scenario_path")), nil
		}
		return toolResult(s.simulateScenario(ctx, input.ScenarioPath, started)), nil
	})
}

func (s *Server) registerCompareTool(server *mcp.Server) {
	server.AddTool(&mcp.Tool{
		Name:        toolCompare,
		Title:       "Compare ProfitCtl scenarios",
		Description: "Use this when the user needs to compare two to four known local scenarios without blending their assumptions.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"scenario_paths": map[string]any{
					"type":        "array",
					"description": "Two to four distinct existing root-relative .yml or .yaml scenario paths.",
					"minItems":    2,
					"maxItems":    4,
					"uniqueItems": true,
					"items": map[string]any{
						"type":      "string",
						"minLength": 1,
					},
				},
			},
			"required": []string{"scenario_paths"},
		},
		OutputSchema: resultSchema(),
		Annotations:  readOnlyAnnotations(),
	}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		started := time.Now()
		var input compareScenariosInput
		if err := decodeArguments(request.Params.Arguments, &input); err != nil {
			return toolResult(failEnvelope(toolCompare, started, OutcomeInvalidInput, "invalid_arguments", "The tool arguments do not match the required schema.", "scenario_paths")), nil
		}
		return toolResult(s.compareScenarios(ctx, input.ScenarioPaths, started)), nil
	})
}

func (s *Server) registerJudgeTool(server *mcp.Server) {
	server.AddTool(&mcp.Tool{
		Name:         toolJudge,
		Title:        "Judge ProfitCtl standards",
		Description:  "Use this when the user needs evidence-quality gaps checked for one existing local scenario or recommendation artifact.",
		InputSchema:  singlePathSchema("target_path", "Existing root-relative .yml, .yaml, .md, or .txt target path."),
		OutputSchema: resultSchema(),
		Annotations:  readOnlyAnnotations(),
	}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		started := time.Now()
		var input standardsInput
		if err := decodeArguments(request.Params.Arguments, &input); err != nil {
			return toolResult(failEnvelope(toolJudge, started, OutcomeInvalidInput, "invalid_arguments", "The tool arguments do not match the required schema.", "target_path")), nil
		}
		return toolResult(s.judgeStandards(ctx, input.TargetPath, started)), nil
	})
}

func (s *Server) validateScenario(ctx context.Context, rawPath string, started time.Time) ResultEnvelope {
	scenario, scenarioErr := s.loadScenario(rawPath)
	if scenarioErr != nil {
		return scenarioFailure(toolValidate, started, scenarioErr)
	}
	input := inputFor(scenario)
	run := s.run(ctx, s.binary, []string{"validate", "--file", scenario.relative})
	if result, handled := processFailure(toolValidate, started, input, run); handled {
		return result
	}
	if run.exitCode == 0 {
		envelope := newEnvelope(toolValidate, started, input)
		envelope.Outcome = OutcomePassed
		envelope.Result = map[string]any{"valid": true}
		return envelope
	}
	if run.exitCode == 2 {
		envelope := newEnvelope(toolValidate, started, input)
		envelope.Outcome = OutcomeInvalidScenario
		envelope.Error = &ErrorInfo{Code: "profitctl_validation_failed", Message: "ProfitCtl rejected the scenario during validation.", Field: "scenario_path"}
		return envelope
	}
	return runtimeFailure(toolValidate, started, input, "profitctl_validation_failed")
}

func (s *Server) simulateScenario(ctx context.Context, rawPath string, started time.Time) ResultEnvelope {
	scenario, scenarioErr := s.loadScenario(rawPath)
	if scenarioErr != nil {
		return scenarioFailure(toolSimulate, started, scenarioErr)
	}
	input := inputFor(scenario)
	run := s.run(ctx, s.binary, []string{"simulate", "--file", scenario.relative, "--json"})
	if result, handled := processFailure(toolSimulate, started, input, run); handled {
		return result
	}
	if run.exitCode == 2 {
		envelope := newEnvelope(toolSimulate, started, input)
		envelope.Outcome = OutcomeInvalidScenario
		envelope.Error = &ErrorInfo{Code: "profitctl_validation_failed", Message: "ProfitCtl rejected the scenario before simulation.", Field: "scenario_path"}
		return envelope
	}
	if run.exitCode != 0 && run.exitCode != 1 {
		return runtimeFailure(toolSimulate, started, input, "profitctl_simulation_failed")
	}
	payload, ok := jsonObject(run.stdout)
	if !ok {
		return runtimeFailure(toolSimulate, started, input, "profitctl_invalid_json")
	}
	envelope := newEnvelope(toolSimulate, started, input)
	envelope.Result = payload
	if run.exitCode == 1 {
		envelope.Outcome = OutcomeCovenantFailed
	} else {
		envelope.Outcome = OutcomePassed
	}
	return envelope
}

func (s *Server) compareScenarios(ctx context.Context, paths []string, started time.Time) ResultEnvelope {
	if len(paths) < 2 || len(paths) > 4 {
		return failEnvelope(toolCompare, started, OutcomeInvalidInput, "scenario_count_out_of_range", "Provide between two and four distinct scenario paths.", "scenario_paths")
	}

	scenarios := make([]loadedScenario, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		scenario, scenarioErr := s.loadScenario(path)
		if scenarioErr != nil {
			return scenarioFailure(toolCompare, started, scenarioErr)
		}
		if _, duplicate := seen[scenario.relative]; duplicate {
			return failEnvelope(toolCompare, started, OutcomeInvalidInput, "duplicate_scenario_path", "Each comparison scenario must resolve to a distinct file.", "scenario_paths")
		}
		seen[scenario.relative] = struct{}{}
		scenarios = append(scenarios, scenario)
	}

	input := inputFor(scenarios...)
	args := make([]string, 0, len(scenarios)+2)
	args = append(args, "compare")
	for _, scenario := range scenarios {
		args = append(args, scenario.relative)
	}
	args = append(args, "--json")
	run := s.run(ctx, s.binary, args)
	if result, handled := processFailure(toolCompare, started, input, run); handled {
		return result
	}
	if run.exitCode == 2 {
		envelope := newEnvelope(toolCompare, started, input)
		envelope.Outcome = OutcomeInvalidScenario
		envelope.Error = &ErrorInfo{Code: "profitctl_validation_failed", Message: "ProfitCtl rejected one or more scenarios before comparison.", Field: "scenario_paths"}
		return envelope
	}
	if run.exitCode != 0 && run.exitCode != 1 {
		return runtimeFailure(toolCompare, started, input, "profitctl_comparison_failed")
	}
	payload, ok := jsonObject(run.stdout)
	if !ok {
		return runtimeFailure(toolCompare, started, input, "profitctl_invalid_json")
	}
	envelope := newEnvelope(toolCompare, started, input)
	envelope.Result = payload
	if run.exitCode == 1 {
		envelope.Outcome = OutcomeCovenantFailed
	} else {
		envelope.Outcome = OutcomePassed
	}
	return envelope
}

func (s *Server) judgeStandards(ctx context.Context, rawPath string, started time.Time) ResultEnvelope {
	target, targetErr := s.loadStandardsTarget(rawPath)
	if targetErr != nil {
		return scenarioFailure(toolJudge, started, targetErr)
	}
	input := inputFor(target)
	run := s.run(ctx, s.standardsBinary, []string{target.relative})
	if result, handled := processFailure(toolJudge, started, input, run); handled {
		return result
	}
	if run.exitCode != 0 && run.exitCode != 1 {
		return runtimeFailure(toolJudge, started, input, "profitctl_standards_failed")
	}
	payload, ok := jsonObject(run.stdout)
	if !ok {
		return runtimeFailure(toolJudge, started, input, "profitctl_invalid_json")
	}
	envelope := newEnvelope(toolJudge, started, input)
	envelope.Result = payload
	if run.exitCode == 1 {
		envelope.Outcome = OutcomeStandardsFailed
	} else {
		envelope.Outcome = OutcomePassed
	}
	return envelope
}

func (s *Server) run(ctx context.Context, binary string, args []string) processResult {
	deadline, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.runner.Run(deadline, binary, args, s.workspaceRoot, s.maxOutput)
}

func inputFor(scenarios ...loadedScenario) InputEvidence {
	input := emptyInput()
	for _, scenario := range scenarios {
		input.Paths = append(input.Paths, scenario.relative)
		input.SHA256 = append(input.SHA256, scenario.digest)
	}
	return input
}

func scenarioFailure(tool string, started time.Time, scenarioErr *scenarioError) ResultEnvelope {
	return failEnvelope(tool, started, scenarioErr.outcome, scenarioErr.code, scenarioErr.message, scenarioErr.field)
}

func processFailure(tool string, started time.Time, input InputEvidence, run processResult) (ResultEnvelope, bool) {
	if run.outputLimitExceeded {
		envelope := newEnvelope(tool, started, input)
		envelope.Outcome = OutcomeOutputLimitExceeded
		envelope.Error = &ErrorInfo{Code: "output_limit_exceeded", Message: "ProfitCtl output exceeded the 1 MiB local limit."}
		return envelope, true
	}
	if run.deadlineExceeded {
		envelope := newEnvelope(tool, started, input)
		envelope.Outcome = OutcomeDeadlineExceeded
		envelope.Error = &ErrorInfo{Code: "deadline_exceeded", Message: "ProfitCtl did not finish within the 30 second local limit."}
		return envelope, true
	}
	return ResultEnvelope{}, false
}

func runtimeFailure(tool string, started time.Time, input InputEvidence, code string) ResultEnvelope {
	envelope := newEnvelope(tool, started, input)
	envelope.Outcome = OutcomeRuntimeFailed
	envelope.Error = &ErrorInfo{Code: code, Message: "ProfitCtl could not produce a usable local result."}
	return envelope
}

func jsonObject(data []byte) (json.RawMessage, bool) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || !json.Valid(trimmed) {
		return nil, false
	}
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, false
	}
	return json.RawMessage(bytes.Clone(trimmed)), true
}

func toolResult(envelope ResultEnvelope) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%s: %s", envelope.Tool, envelope.Outcome)},
		},
		StructuredContent: envelope,
	}
}

func decodeArguments(raw json.RawMessage, target any) error {
	if len(raw) == 0 {
		return fmt.Errorf("arguments are required")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("arguments must contain one JSON object")
	}
	return nil
}

func readOnlyAnnotations() *mcp.ToolAnnotations {
	falseValue := false
	return &mcp.ToolAnnotations{
		ReadOnlyHint:    true,
		DestructiveHint: &falseValue,
		OpenWorldHint:   &falseValue,
		IdempotentHint:  true,
	}
}

func singlePathSchema(field, description string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			field: map[string]any{
				"type":        "string",
				"description": description,
				"minLength":   1,
			},
		},
		"required": []string{field},
	}
}

func resultSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"schema_version": map[string]any{"type": "string", "const": ResultSchemaVersion},
			"tool":           map[string]any{"type": "string"},
			"outcome": map[string]any{
				"type": "string",
				"enum": []string{
					string(OutcomePassed),
					string(OutcomeCovenantFailed),
					string(OutcomeStandardsFailed),
					string(OutcomeInvalidInput),
					string(OutcomeInvalidScenario),
					string(OutcomeRuntimeFailed),
					string(OutcomeDeadlineExceeded),
					string(OutcomeOutputLimitExceeded),
				},
			},
			"input": map[string]any{"type": "object"},
			"duration_ms": map[string]any{
				"type":    "integer",
				"minimum": 0,
			},
			"result": map[string]any{},
			"error":  map[string]any{},
		},
		"required": []string{"schema_version", "tool", "outcome", "input", "duration_ms", "result", "error"},
	}
}
