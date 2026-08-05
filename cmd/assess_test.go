package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/research/exa"
	"github.com/IntelIP/ProfitCtl/internal/scanner/llm"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sequenceProvider struct {
	responses []string
	calls     int
	prompts   []string
}

func (s *sequenceProvider) Chat(_ context.Context, messages []llm.Message) (string, error) {
	if len(messages) != 1 {
		return "", errors.New("expected one user message")
	}
	if s.calls == len(s.responses) {
		return "", errors.New("unexpected model call")
	}
	s.prompts = append(s.prompts, messages[0].Content)
	response := s.responses[s.calls]
	s.calls++
	return response, nil
}

func (s *sequenceProvider) IsAvailable(context.Context) bool { return true }

type fakeExaSearcher struct {
	response exa.SearchResponse
	requests []exa.SearchRequest
}

func (s *fakeExaSearcher) Search(_ context.Context, request exa.SearchRequest) (exa.SearchResponse, error) {
	s.requests = append(s.requests, request)
	return s.response, nil
}

func TestRunAssess_WritesSourceBackedCostDraft(t *testing.T) {
	tempDir := t.TempDir()
	reportPath := filepath.Join(tempDir, "assessment.json")
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module test\nrequire github.com/revrost/go-openrouter v1.1.5\n"), 0644))

	oldPath, oldModel, oldOut := assessPath, assessModel, assessOut
	oldNow := assessNow
	oldProvider, oldSearcher := newProvider, newExaSearcher
	t.Cleanup(func() {
		assessPath, assessModel, assessOut = oldPath, oldModel, oldOut
		assessNow = oldNow
		newProvider, newExaSearcher = oldProvider, oldSearcher
	})
	t.Setenv("OPENROUTER_API_KEY", "test-openrouter-key")
	t.Setenv("EXA_API_KEY", "test-exa-key")
	assessPath = tempDir
	assessModel = defaultAssessModel
	assessOut = reportPath
	assessNow = func() time.Time {
		return time.Date(2026, time.August, 4, 20, 0, 0, 0, time.UTC)
	}

	provider := &sequenceProvider{responses: []string{
		`{"providers":[{"name":"OpenRouter","provider":"openrouter","official_domain":"openrouter.ai","evidence":[{"file":"go.mod","excerpt":"github.com/revrost/go-openrouter"}]}]}`,
		`{"assumptions":[{"name":"assessments_per_month","value":"1","source":"inferred","rationale":"starter scale"}],"cost_lines":[{"name":"Application OpenRouter input","provider":"openrouter","role":"codebacked_provider","unit":"million input tokens","price_per_unit_usd":1,"units_per_month":0.1,"monthly_cost_usd":0.1,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$1 per 1M input tokens","confidence":"medium"},{"name":"Assessment Terra input","provider":"openrouter","role":"assessment_model","unit":"million input tokens","price_per_unit_usd":1,"units_per_month":0.01,"monthly_cost_usd":0.01,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$1 per 1M input tokens","confidence":"medium"},{"name":"Assessment Terra output","provider":"openrouter","role":"assessment_model","unit":"million output tokens","price_per_unit_usd":6,"units_per_month":0.002,"monthly_cost_usd":0.012,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$6 per 1M output tokens","confidence":"medium"},{"name":"Exa pricing searches","provider":"exa","role":"assessment_research","unit":"search","price_per_unit_usd":0.005,"units_per_month":2,"monthly_cost_usd":0.01,"source_url":"https://exa.ai/pricing","source_excerpt":"$0.005 per search","confidence":"medium"}],"recommendation":{"summary":"Use this as a starter scenario.","next_steps":["Review it."]}}`,
	}}
	searcher := &fakeExaSearcher{response: exa.SearchResponse{
		RequestID: "exa-request",
		Results: []exa.SearchResult{
			{
				Title:      "GPT-5.6 Terra",
				URL:        "https://openrouter.ai/openai/gpt-5.6-terra",
				Highlights: []string{"$1 per 1M input tokens", "$6 per 1M output tokens"},
			},
			{
				Title:      "Exa pricing",
				URL:        "https://exa.ai/pricing",
				Highlights: []string{"$0.005 per search"},
			},
		},
	}}
	newProvider = func(apiKey, model string) (llm.LLMProvider, error) {
		assert.Equal(t, "test-openrouter-key", apiKey)
		assert.Equal(t, defaultAssessModel, model)
		return provider, nil
	}
	newExaSearcher = func(apiKey string) (exa.Searcher, error) {
		assert.Equal(t, "test-exa-key", apiKey)
		return searcher, nil
	}

	command := &cobra.Command{}
	command.SetContext(context.Background())
	require.NoError(t, runAssess(command, nil))

	contents, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	var report costAssessmentReport
	require.NoError(t, json.Unmarshal(contents, &report))
	assert.Equal(t, assessmentSchemaVersion, report.SchemaVersion)
	assert.Equal(t, defaultAssessModel, report.Model)
	assert.Equal(t, 2, report.ModelRequestsPerAssessment)
	assert.Equal(t, 2, report.ExaRequestsPerAssessment)
	assert.Equal(t, 1, report.AnalyzedFiles)
	require.Len(t, report.Providers, 1)
	assert.Equal(t, "openrouter", report.Providers[0].Provider)
	require.Len(t, report.PricingReceipts, 2)
	assert.Equal(t, "exa-request", report.PricingReceipts[0].RequestID)
	assert.Equal(t, "assessment_model_and_codebacked_provider", report.PricingReceipts[0].Role)
	assert.Equal(t, "2026-08-04T20:00:00Z", report.PricingReceipts[0].CapturedAt)
	assert.Equal(t, "assessment_research", report.PricingReceipts[1].Role)
	require.Len(t, report.Draft.CostLines, 4)
	assert.InDelta(t, 0.132, report.EstimatedMonthlyCostUSD, 0.000001)
	assert.Len(t, searcher.requests, 2)
	assert.Equal(t, "OpenRouter openai/gpt-5.6-terra pricing", searcher.requests[0].Query)
	assert.Equal(t, []string{"openrouter.ai"}, searcher.requests[0].IncludeDomains)
	assert.Equal(t, "Exa pricing", searcher.requests[1].Query)
	assert.Equal(t, []string{"exa.ai"}, searcher.requests[1].IncludeDomains)
	assert.Len(t, provider.prompts, 2)
	assert.Contains(t, provider.prompts[1], "exactly two model requests per assessment")
	assert.Contains(t, provider.prompts[1], "exactly 2 Exa pricing-search requests")
	assert.Contains(t, provider.prompts[1], "github.com/revrost/go-openrouter")
}

func TestPricingSearchTarget_BindsOpenRouterToAssessmentModel(t *testing.T) {
	query, requirement, err := pricingSearchTarget(llm.ProviderCandidate{Name: "OpenRouter", Provider: "openrouter"}, "openai/gpt-5.6-terra")
	require.NoError(t, err)
	assert.Equal(t, "OpenRouter openai/gpt-5.6-terra pricing", query)
	assert.Equal(t, "/openai/gpt-5.6-terra", requirement.ExactPath)

	result, ok := firstOfficialResult([]exa.SearchResult{
		{URL: "https://openrouter.ai/pricing"},
		{URL: "https://openrouter.ai/openai/gpt-5.6-terra-pro"},
		{URL: "https://openrouter.ai/openai/gpt-5.6-terra"},
	}, "openrouter.ai", requirement)
	assert.True(t, ok)
	assert.Equal(t, "https://openrouter.ai/openai/gpt-5.6-terra", result.URL)

	query, requirement, err = pricingSearchTarget(llm.ProviderCandidate{Name: "Exa", Provider: "exa"}, defaultAssessModel)
	require.NoError(t, err)
	assert.Equal(t, "Exa pricing", query)
	assert.Equal(t, []string{"/pricing"}, requirement.Fragments)
}

func TestFirstOfficialResult_RejectsPrefixModelSlug(t *testing.T) {
	_, ok := firstOfficialResult([]exa.SearchResult{
		{URL: "https://openrouter.ai/openai/gpt-5.6-terra-pro"},
	}, "openrouter.ai", pricingPathRequirement{ExactPath: "/openai/gpt-5.6-terra"})

	assert.False(t, ok)
}

func TestFirstOfficialResult_RejectsLookalikeDomains(t *testing.T) {
	_, ok := firstOfficialResult([]exa.SearchResult{
		{URL: "https://openrouter.ai.evil.example/openai/gpt-5.6-terra"},
		{URL: "https://evil-openrouter.ai/openai/gpt-5.6-terra"},
	}, "openrouter.ai", pricingPathRequirement{ExactPath: "/openai/gpt-5.6-terra"})

	assert.False(t, ok)
}

func TestFirstOfficialResult_RejectsNonPricingProviderPage(t *testing.T) {
	result, ok := firstOfficialResult([]exa.SearchResult{
		{URL: "https://stripe.com/docs/payments"},
		{URL: "https://stripe.com/blog/pricing-models"},
		{URL: "https://stripe.com/pricing"},
	}, "stripe.com", pricingPathRequirement{Fragments: []string{"/pricing"}})

	require.True(t, ok)
	assert.Equal(t, "https://stripe.com/pricing", result.URL)
}

func TestEnsureAssessmentRuntimeReceipts_AddsModelAndResearchReceipts(t *testing.T) {
	searcher := &fakeExaSearcher{response: exa.SearchResponse{
		RequestID: "terra-price",
		Results: []exa.SearchResult{
			{
				Title:      "GPT-5.6 Terra",
				URL:        "https://openrouter.ai/openai/gpt-5.6-terra",
				Highlights: []string{"$1 / $6 per 1M tokens"},
			},
			{
				Title:      "Exa pricing",
				URL:        "https://exa.ai/pricing",
				Highlights: []string{"$0.005 per search"},
			},
		},
	}}

	receipts, err := ensureAssessmentRuntimeReceipts(context.Background(), searcher, []llm.PricingReceipt{{
		Provider: "buildkite",
		URL:      "https://buildkite.com/pricing/",
	}}, defaultAssessModel)

	require.NoError(t, err)
	require.Len(t, receipts, 3)
	assert.Equal(t, "openrouter", receipts[1].Provider)
	assert.Equal(t, "assessment_model", receipts[1].Role)
	assert.Equal(t, "exa", receipts[2].Provider)
	assert.Equal(t, "assessment_research", receipts[2].Role)
	assert.Equal(t, "OpenRouter openai/gpt-5.6-terra pricing", searcher.requests[0].Query)
	assert.Equal(t, "Exa pricing", searcher.requests[1].Query)
}

func TestRetrievePricingReceipt_RejectsUntrustedDomain(t *testing.T) {
	searcher := &fakeExaSearcher{}

	_, err := retrievePricingReceipt(context.Background(), searcher, llm.ProviderCandidate{
		Name:           "OpenRouter",
		Provider:       "openrouter",
		OfficialDomain: "openrouter.example",
	}, defaultAssessModel, "assessment_model")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `trusted official domain "openrouter.ai"`)
	assert.Empty(t, searcher.requests)
}

func TestRunAssess_RequiresBothKeys(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("EXA_API_KEY", "")

	command := &cobra.Command{}
	command.SetContext(context.Background())
	err := runAssess(command, nil)

	require.Error(t, err)
	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
	assert.Contains(t, err.Error(), "OPENROUTER_API_KEY")
}

func TestRunAssess_RequiresExaKey(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "test-openrouter-key")
	t.Setenv("EXA_API_KEY", "")

	command := &cobra.Command{}
	command.SetContext(context.Background())
	err := runAssess(command, nil)

	require.Error(t, err)
	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
	assert.Contains(t, err.Error(), "EXA_API_KEY")
}
