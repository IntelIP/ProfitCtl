package mcpserver

import (
	"context"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/systemmodel"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const toolCompareOutcomes = "profitctl_compare_outcomes"

func (s *Server) registerOutcomesTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: toolCompareOutcomes, Title: "Compare cost per accepted outcome",
		Description: "Calculate scoped USD per accepted outcome and the candidate acceptance rate needed to match the first option. Use comparable cohorts, include failed work and keep unmeasured counts or charges null. No source access or model calls.",
		Annotations: readOnlyAnnotations(), OutputSchema: resultSchema(),
	}, func(_ context.Context, _ *mcp.CallToolRequest, in systemmodel.OutcomeComparison) (*mcp.CallToolResult, ResultEnvelope, error) {
		started := time.Now()
		report, err := systemmodel.CompareOutcomes(in)
		if err != nil {
			env := failEnvelope(toolCompareOutcomes, started, OutcomeInvalidInput, "invalid_cohorts", err.Error(), "options")
			return &mcp.CallToolResult{IsError: true}, env, nil
		}
		env := newEnvelope(toolCompareOutcomes, started, emptyInput())
		env.Outcome = OutcomePassed
		env.Result = report
		return nil, env, nil
	})
}
