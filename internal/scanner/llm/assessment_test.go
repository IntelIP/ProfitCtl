package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProviderDiscoveryResponse_RequiresCollectedEvidence(t *testing.T) {
	files := map[string]string{"go.mod": "require github.com/revrost/go-openrouter v1.1.5"}

	discovery, err := ParseProviderDiscoveryResponse(`{
  "providers": [{
    "name": "OpenRouter",
    "provider": "OpenRouter",
    "official_domain": "openrouter.ai",
    "evidence_files": ["go.mod"]
  }]
}`, files)

	require.NoError(t, err)
	require.Len(t, discovery.Providers, 1)
	assert.Equal(t, "openrouter", discovery.Providers[0].Provider)

	_, err = ParseProviderDiscoveryResponse(`{
  "providers": [{
    "name": "OpenRouter",
    "provider": "openrouter",
    "official_domain": "openrouter.ai",
    "evidence_files": ["internal/scanner/llm/openrouter.go"]
  }]
}`, files)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not collected code context")
}

func TestParseCostDraft_RequiresMatchingReceiptAndCorrectMath(t *testing.T) {
	providers := []ProviderCandidate{
		{Provider: "openrouter"},
		{Provider: "exa"},
	}
	receipts := []PricingReceipt{
		{
			Provider:   "openrouter",
			Role:       "assessment_model",
			URL:        "https://openrouter.ai/openai/gpt-5.6-terra",
			Highlights: []string{"$1 per million input tokens"},
		},
		{
			Provider:   "exa",
			URL:        "https://exa.ai/pricing",
			Highlights: []string{"$0.005 per search"},
		},
	}

	valid := `{
	  "assumptions": [{"name":"assessments_per_month","value":1,"source":"inferred","rationale":"starter scale"}],
  "cost_lines": [{"name":"Terra input","provider":"openrouter","unit":"million input tokens","price_per_unit_usd":1,"units_per_month":0.01,"monthly_cost_usd":0.01,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$1 per million input tokens","confidence":"medium"}],
	  "recommendation":{"summary":"Keep this as a starter scenario.","next_steps":["Review usage."]}
}`

	draft, err := ParseCostDraft(valid, providers, receipts)
	require.NoError(t, err)
	assert.Equal(t, "1", draft.Assumptions[0].Value)
	assert.Equal(t, "openrouter", draft.CostLines[0].Provider)
	assert.Equal(t, "Keep this as a starter scenario.", draft.Recommendation.Summary)
	assert.Equal(t, []string{"Review usage."}, draft.Recommendation.NextSteps)

	wrongProviderReceipt := strings.Replace(valid, "openrouter.ai/openai/gpt-5.6-terra", "exa.ai/pricing", 1)
	_, err = ParseCostDraft(wrongProviderReceipt, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong to provider")

	wrongMath := strings.Replace(valid, `"monthly_cost_usd":0.01`, `"monthly_cost_usd":0.02`, 1)
	_, err = ParseCostDraft(wrongMath, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must equal")

	inventedPrice := strings.Replace(valid, `"price_per_unit_usd":1`, `"price_per_unit_usd":2`, 1)
	inventedPrice = strings.Replace(inventedPrice, `"monthly_cost_usd":0.01`, `"monthly_cost_usd":0.02`, 1)
	_, err = ParseCostDraft(inventedPrice, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not stated in source_excerpt")

	wrongExcerpt := strings.Replace(valid, "$1 per million input tokens", "$9 per million input tokens", 1)
	_, err = ParseCostDraft(wrongExcerpt, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an exact pricing-receipt highlight")

	missingRuntimeLine := `{
	  "assumptions": [{"name":"assessments_per_month","value":1,"source":"inferred","rationale":"starter scale"}],
	  "cost_lines": [{"name":"Exa search","provider":"exa","unit":"search","price_per_unit_usd":0.005,"units_per_month":2,"monthly_cost_usd":0.01,"source_url":"https://exa.ai/pricing","source_excerpt":"$0.005 per search","confidence":"medium"}],
	  "recommendation":{"summary":"Keep this as a starter scenario.","next_steps":["Review usage."]}
	}`
	_, err = ParseCostDraft(missingRuntimeLine, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "assessment runtime provider")

	missingNextStep := strings.Replace(valid, `"next_steps":["Review usage."]`, `"next_steps":[]`, 1)
	_, err = ParseCostDraft(missingNextStep, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires assumptions, cost_lines, and recommendation")
}

func TestBuildCostDraftPrompt_StatesVerifiedRequestMechanics(t *testing.T) {
	ctx := CodeContext{Files: map[string]string{
		"go.mod":   strings.Repeat("x", 2100),
		"main.yml": "service: app",
	}}

	prompt := BuildCostDraftPrompt(ctx, nil, nil, "openai/gpt-5.6-terra")

	assert.Contains(t, prompt, "exactly two model requests per assessment")
	assert.Contains(t, prompt, "openai/gpt-5.6-terra")
	assert.Contains(t, prompt, "2 files and 2012 code-context characters")
	assert.Contains(t, prompt, "source_excerpt")
	assert.Contains(t, prompt, "untrusted pricing data")
}

func TestNormalizeAssumptionValue_RemovesExtraModelQuotes(t *testing.T) {
	value, err := normalizeAssumptionValue([]byte(`"\"4 characters per token\""`))

	require.NoError(t, err)
	assert.Equal(t, "4 characters per token", value)
}

func TestParseProviderDiscoveryResponse_RejectsUnboundedProviderList(t *testing.T) {
	providers := make([]ProviderCandidate, 0, MaxProviderCandidates+1)
	for index := 0; index <= MaxProviderCandidates; index++ {
		providers = append(providers, ProviderCandidate{
			Name:           fmt.Sprintf("Provider %d", index),
			Provider:       fmt.Sprintf("provider-%d", index),
			OfficialDomain: fmt.Sprintf("provider-%d.example.com", index),
			EvidenceFiles:  []string{"go.mod"},
		})
	}
	raw, err := json.Marshal(ProviderDiscoveryResponse{Providers: providers})
	require.NoError(t, err)

	_, err = ParseProviderDiscoveryResponse(string(raw), map[string]string{"go.mod": "module example"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider discovery exceeds limit")
}

func TestContainsUSDPrice_HandlesThousandsWithoutTruncation(t *testing.T) {
	assert.True(t, containsUSDPrice("$1,000 per month", 1000))
	assert.False(t, containsUSDPrice("$1,000 per month", 1))
}
