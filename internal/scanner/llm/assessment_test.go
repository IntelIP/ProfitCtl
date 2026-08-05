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
    "evidence": [{"file":"go.mod","excerpt":"github.com/revrost/go-openrouter"}]
  }]
}`, files)

	require.NoError(t, err)
	require.Len(t, discovery.Providers, 1)
	assert.Equal(t, "openrouter", discovery.Providers[0].Provider)
	assert.Equal(t, []string{"go.mod"}, discovery.Providers[0].EvidenceFiles)

	_, err = ParseProviderDiscoveryResponse(`{
  "providers": [{
    "name": "OpenRouter",
    "provider": "openrouter",
    "official_domain": "openrouter.ai",
    "evidence": [{"file":"internal/scanner/llm/openrouter.go","excerpt":"openrouter"}]
  }]
}`, files)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file is not collected code context")

	_, err = ParseProviderDiscoveryResponse(`{
  "providers": [{
    "name": "OpenRouter",
    "provider": "openrouter",
    "official_domain": "openrouter.example",
    "evidence": [{"file":"go.mod","excerpt":"github.com/revrost/go-openrouter"}]
  }]
}`, files)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `official_domain must be "openrouter.ai"`)

	_, err = ParseProviderDiscoveryResponse(`{
  "providers": [{
    "name": "Unknown",
    "provider": "unknown",
    "official_domain": "unknown.example",
    "evidence": [{"file":"go.mod","excerpt":"github.com/revrost/go-openrouter"}]
  }]
}`, files)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in the trusted provider registry")

	_, err = ParseProviderDiscoveryResponse(`{
  "providers": [{
    "name": "Stripe",
    "provider": "stripe",
    "official_domain": "stripe.com",
    "evidence": [{"file":"go.mod","excerpt":"github.com/revrost/go-openrouter"}]
  }]
}`, files)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `excerpt does not identify provider "stripe"`)
}

func TestProviderEvidenceMatches_RequiresMarkerBoundaries(t *testing.T) {
	assert.True(t, ProviderEvidenceMatches("exa", "EXA_API_KEY"))
	assert.False(t, ProviderEvidenceMatches("exa", "example configuration"))
}

func TestParseCostDraft_RequiresMatchingReceiptAndCorrectMath(t *testing.T) {
	providers := []ProviderCandidate{}
	receipts := []PricingReceipt{
		{
			Provider:   "openrouter",
			Role:       "assessment_model",
			URL:        "https://openrouter.ai/openai/gpt-5.6-terra",
			Highlights: []string{"$1 per million input tokens", "$6 per million output tokens"},
		},
		{
			Provider:   "exa",
			Role:       "codebacked_provider",
			URL:        "https://exa.ai/pricing",
			Highlights: []string{"$0.005 per search"},
		},
	}

	valid := `{
	  "assumptions": [{"name":"assessments_per_month","value":1,"source":"inferred","rationale":"starter scale"}],
	  "cost_lines": [
	    {"name":"Terra input","provider":"openrouter","role":"assessment_model","unit":"million input tokens","price_per_unit_usd":1,"units_per_month":0.01,"monthly_cost_usd":0.01,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$1 per million input tokens","confidence":"medium"},
	    {"name":"Terra output","provider":"openrouter","role":"assessment_model","unit":"million output tokens","price_per_unit_usd":6,"units_per_month":0.002,"monthly_cost_usd":0.012,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$6 per million output tokens","confidence":"medium"}
	  ],
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
	assert.Contains(t, err.Error(), "not the single price stated in source_excerpt")

	wrongExcerpt := strings.Replace(valid, "$1 per million input tokens", "$9 per million input tokens", 1)
	_, err = ParseCostDraft(wrongExcerpt, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an exact pricing-receipt highlight")

	missingRuntimeLine := `{
	  "assumptions": [{"name":"assessments_per_month","value":1,"source":"inferred","rationale":"starter scale"}],
	  "cost_lines": [{"name":"Terra input","provider":"openrouter","role":"assessment_model","unit":"million input tokens","price_per_unit_usd":1,"units_per_month":0.01,"monthly_cost_usd":0.01,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$1 per million input tokens","confidence":"medium"}],
	  "recommendation":{"summary":"Keep this as a starter scenario.","next_steps":["Review usage."]}
	}`
	_, err = ParseCostDraft(missingRuntimeLine, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `role "assessment_model" usage "output"`)

	missingNextStep := strings.Replace(valid, `"next_steps":["Review usage."]`, `"next_steps":[]`, 1)
	_, err = ParseCostDraft(missingNextStep, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires assumptions, cost_lines, and recommendation")
}

func TestParseCostDraft_RequiresEveryDiscoveredProvider(t *testing.T) {
	providers := []ProviderCandidate{
		{Provider: "stripe"},
	}
	receipts := []PricingReceipt{
		{
			Provider:   "openrouter",
			Role:       "assessment_model",
			URL:        "https://openrouter.ai/openai/gpt-5.6-terra",
			Highlights: []string{"$1 per million input tokens", "$6 per million output tokens"},
		},
		{
			Provider:   "stripe",
			Role:       "codebacked_provider",
			URL:        "https://stripe.com/pricing",
			Highlights: []string{"$0.30 per transaction"},
		},
	}
	missingStripe := `{
	  "assumptions": [{"name":"assessments_per_month","value":1,"source":"inferred","rationale":"starter scale"}],
	  "cost_lines": [
	    {"name":"Terra input","provider":"openrouter","role":"assessment_model","unit":"million input tokens","price_per_unit_usd":1,"units_per_month":0.01,"monthly_cost_usd":0.01,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$1 per million input tokens","confidence":"medium"},
	    {"name":"Terra output","provider":"openrouter","role":"assessment_model","unit":"million output tokens","price_per_unit_usd":6,"units_per_month":0.002,"monthly_cost_usd":0.012,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$6 per million output tokens","confidence":"medium"}
	  ],
	  "recommendation":{"summary":"Keep this as a starter scenario.","next_steps":["Review usage."]}
	}`

	_, err := ParseCostDraft(missingStripe, providers, receipts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `provider "stripe" role "codebacked_provider"`)
}

func TestParseCostDraft_BindsOnePriceToItsExactUnit(t *testing.T) {
	providers := []ProviderCandidate{}
	receipts := []PricingReceipt{{
		Provider:   "openrouter",
		Role:       "assessment_model",
		URL:        "https://openrouter.ai/openai/gpt-5.6-terra",
		Highlights: []string{"$1 / $6 per 1M tokens", "$1 per 1M input tokens", "$6 per 1M output tokens"},
	}}
	base := `{
	  "assumptions": [{"name":"assessments_per_month","value":1,"source":"inferred","rationale":"starter scale"}],
	  "cost_lines": [
	    {"name":"Terra input","provider":"openrouter","role":"assessment_model","unit":"million input tokens","price_per_unit_usd":1,"units_per_month":0.01,"monthly_cost_usd":0.01,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$1 per 1M input tokens","confidence":"medium"},
	    {"name":"Terra output","provider":"openrouter","role":"assessment_model","unit":"million output tokens","price_per_unit_usd":6,"units_per_month":0.002,"monthly_cost_usd":0.012,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$6 per 1M output tokens","confidence":"medium"}
	  ],
	  "recommendation":{"summary":"Keep this as a starter scenario.","next_steps":["Review usage."]}
	}`

	_, err := ParseCostDraft(base, providers, receipts)
	require.NoError(t, err)

	ambiguous := strings.Replace(base, "$1 per 1M input tokens", "$1 / $6 per 1M tokens", 1)
	_, err = ParseCostDraft(ambiguous, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one USD price")

	wrongUnit := strings.Replace(base, `"unit":"million input tokens"`, `"unit":"million output tokens"`, 1)
	_, err = ParseCostDraft(wrongUnit, providers, receipts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unit is not fully stated")
}

func TestParseCostDraft_SeparatesCodebackedAndRuntimeRoles(t *testing.T) {
	providers := []ProviderCandidate{{Provider: "openrouter"}}
	receipts := []PricingReceipt{{
		Provider:   "openrouter",
		Role:       "assessment_model_and_codebacked_provider",
		URL:        "https://openrouter.ai/openai/gpt-5.6-terra",
		Highlights: []string{"$1 per million input tokens", "$6 per million output tokens"},
	}}
	runtimeOnly := `{
	  "assumptions": [{"name":"assessments_per_month","value":1,"source":"inferred","rationale":"starter scale"}],
	  "cost_lines": [
	    {"name":"Terra assessment input","provider":"openrouter","role":"assessment_model","unit":"million input tokens","price_per_unit_usd":1,"units_per_month":0.01,"monthly_cost_usd":0.01,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$1 per million input tokens","confidence":"medium"},
	    {"name":"Terra assessment output","provider":"openrouter","role":"assessment_model","unit":"million output tokens","price_per_unit_usd":6,"units_per_month":0.002,"monthly_cost_usd":0.012,"source_url":"https://openrouter.ai/openai/gpt-5.6-terra","source_excerpt":"$6 per million output tokens","confidence":"medium"}
	  ],
	  "recommendation":{"summary":"Keep this as a starter scenario.","next_steps":["Review usage."]}
	}`

	_, err := ParseCostDraft(runtimeOnly, providers, receipts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `role "codebacked_provider"`)
}

func TestParseCostDraft_RejectsCompoundPaymentFeeExcerpt(t *testing.T) {
	providers := []ProviderCandidate{{Provider: "stripe"}}
	receipts := []PricingReceipt{{
		Provider:   "stripe",
		Role:       "codebacked_provider",
		URL:        "https://stripe.com/pricing",
		Highlights: []string{"2.9% + $0.30 per transaction"},
	}}
	compound := `{
	  "assumptions": [{"name":"transactions","value":100,"source":"inferred","rationale":"starter scale"}],
	  "cost_lines": [{"name":"Stripe transaction","provider":"stripe","role":"codebacked_provider","unit":"transaction","price_per_unit_usd":0.30,"units_per_month":100,"monthly_cost_usd":30,"source_url":"https://stripe.com/pricing","source_excerpt":"2.9% + $0.30 per transaction","confidence":"medium"}],
	  "recommendation":{"summary":"Keep this as a starter scenario.","next_steps":["Review usage."]}
	}`

	_, err := ParseCostDraft(compound, providers, receipts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "additional percentage or cent price component")
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
	assert.Contains(t, prompt, "exactly one USD price")
	assert.Contains(t, prompt, "## main.yml")
	assert.Contains(t, prompt, "service: app")
	assert.Contains(t, prompt, "separate assessment_model input-token and output-token lines")
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
			Evidence:       []ProviderEvidence{{File: "go.mod", Excerpt: "module example"}},
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

func TestNormalizePricingHighlights_MatchesPromptBoundary(t *testing.T) {
	highlights := []string{"  $1 per request; $2 per response  ", strings.Repeat("x", MaxPricingHighlightCharacters+20)}

	normalized := NormalizePricingHighlights(highlights)

	require.Len(t, normalized, 3)
	assert.Equal(t, "$1 per request", normalized[0])
	assert.Equal(t, "$2 per response", normalized[1])
	assert.Len(t, normalized[2], MaxPricingHighlightCharacters)
}

func TestParseStructuredJSON_DoesNotMergeFailedCandidates(t *testing.T) {
	var target struct {
		Required []string `json:"required"`
	}
	raw := "{\"required\":[\"leaked\"],\"broken\": }\n```json\n{}\n```"

	err := parseStructuredJSON(raw, &target)

	require.NoError(t, err)
	assert.Empty(t, target.Required)
}
