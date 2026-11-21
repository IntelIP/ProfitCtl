package mcpserver

import (
	"context"
	"errors"
	"io/fs"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/decision"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolCreateDecision = "profitctl_create_decision"
	toolReadDecision   = "profitctl_read_decision"
	toolListDecisions  = "profitctl_list_decisions"
	toolReviseDecision = "profitctl_revise_decision"
)

type createDecisionInput struct {
	ID   string        `json:"id" jsonschema:"Stable lowercase decision identifier shared by the CLI and both clients."`
	Case decision.Case `json:"case" jsonschema:"Sourced economic alternatives, quality constraints, recommendation or threshold, and material missing evidence."`
}

type readDecisionInput struct {
	ID       string `json:"id"`
	Revision int    `json:"revision,omitempty" jsonschema:"Positive revision number; omit to read the latest saved decision."`
}

type listDecisionsInput struct {
	Query string `json:"query,omitempty" jsonschema:"Optional project, question, or decision ID substring."`
	After string `json:"after,omitempty" jsonschema:"Exclusive cursor returned on the previous page."`
	Limit int    `json:"limit,omitempty" jsonschema:"Page size from 1 to 50; omit for 20."`
}

type reviseDecisionInput struct {
	ID               string        `json:"id"`
	ExpectedRevision int           `json:"expected_revision" jsonschema:"Latest revision you inspected. If it changed, read again before merging edits."`
	Case             decision.Case `json:"case" jsonschema:"Full revised case, retaining relevant evidence, assumptions, and choices."`
}

func (s *Server) registerDecisionTools(server *mcp.Server) {
	writeAnnotations := readOnlyAnnotations()
	writeAnnotations.ReadOnlyHint = false // Append-only, conditional, and safe to retry with identical inputs.
	mcp.AddTool(server, &mcp.Tool{
		Name: toolCreateDecision, Title: "Create an economic decision",
		Description: "Save the first revision of a sourced recommendation or an unresolved economic question in the configured workspace. Use sanitized evidence references.",
		Annotations: writeAnnotations, OutputSchema: resultSchema(),
	}, func(_ context.Context, _ *mcp.CallToolRequest, in createDecisionInput) (*mcp.CallToolResult, ResultEnvelope, error) {
		start := time.Now()
		rev, err := decision.Save(s.workspaceRoot, in.ID, 0, in.Case)
		return decisionResult(toolCreateDecision, start, rev, err)
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: toolReadDecision, Title: "Read an economic decision",
		Description: "Retrieve the latest or a named historical revision with its assumptions, evidence, forecast, and observed outcome.",
		Annotations: readOnlyAnnotations(), OutputSchema: resultSchema(),
	}, func(_ context.Context, _ *mcp.CallToolRequest, in readDecisionInput) (*mcp.CallToolResult, ResultEnvelope, error) {
		start := time.Now()
		rev, err := decision.Read(s.workspaceRoot, in.ID, in.Revision)
		return decisionResult(toolReadDecision, start, rev, err)
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: toolListDecisions, Title: "Find an economic decision",
		Description: "Locate a case in this workspace by project, question, or ID before reading its evidence. Follow next_cursor even when a page contains no matching cases.",
		Annotations: readOnlyAnnotations(), OutputSchema: resultSchema(),
	}, func(_ context.Context, _ *mcp.CallToolRequest, in listDecisionsInput) (*mcp.CallToolResult, ResultEnvelope, error) {
		start := time.Now()
		page, err := decision.List(s.workspaceRoot, in.Query, in.After, in.Limit)
		return decisionResult(toolListDecisions, start, page, err)
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: toolReviseDecision, Title: "Revise an economic decision",
		Description: "Save a new revision for changed assumptions, a chosen option, or a measured result. Read and supply the current revision so concurrent work is preserved.",
		Annotations: writeAnnotations, OutputSchema: resultSchema(),
	}, func(_ context.Context, _ *mcp.CallToolRequest, in reviseDecisionInput) (*mcp.CallToolResult, ResultEnvelope, error) {
		start := time.Now()
		if in.ExpectedRevision < 1 {
			env := failEnvelope(toolReviseDecision, start, OutcomeInvalidInput, "invalid_revision", "Supply a positive base revision from read_decision.", "expected_revision")
			return &mcp.CallToolResult{IsError: true}, env, nil
		}
		rev, err := decision.Save(s.workspaceRoot, in.ID, in.ExpectedRevision, in.Case)
		return decisionResult(toolReviseDecision, start, rev, err)
	})
}

func decisionResult(tool string, start time.Time, value any, err error) (*mcp.CallToolResult, ResultEnvelope, error) {
	if err == nil {
		env := newEnvelope(tool, start, emptyInput())
		env.Outcome = OutcomePassed
		env.Result = value
		return nil, env, nil
	}
	var env ResultEnvelope
	switch {
	case errors.Is(err, decision.ErrConflict):
		env = failEnvelope(tool, start, OutcomeInvalidInput, "revision_conflict", "This revision has a different saved case. Read the latest revision and reconcile the changes.", "id")
	case errors.Is(err, fs.ErrNotExist):
		env = failEnvelope(tool, start, OutcomeInvalidInput, "revision_not_found", "This decision or revision could not be found.", "id")
	case errors.Is(err, decision.ErrInvalid):
		env = failEnvelope(tool, start, OutcomeInvalidInput, "invalid_decision", err.Error(), "")
	default:
		env = runtimeFailure(tool, start, emptyInput(), "decision_store_failed")
	}
	return &mcp.CallToolResult{IsError: true}, env, nil
}
