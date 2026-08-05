package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/research/exa"
	"github.com/IntelIP/ProfitCtl/internal/scanner/llm"
	"github.com/spf13/cobra"
)

const (
	defaultAssessModel      = llm.DefaultOpenRouterModel
	assessmentSchemaVersion = "profitctl.assessment/v1"
)

var (
	assessPath  = defaultDetectPath
	assessOut   string
	assessModel = defaultAssessModel
	assessNow   = time.Now

	newExaSearcher = func(apiKey string) (exa.Searcher, error) {
		return exa.NewClient(apiKey)
	}
)

type costAssessmentReport struct {
	SchemaVersion              string                  `json:"schema_version"`
	Path                       string                  `json:"path"`
	Model                      string                  `json:"model"`
	ModelRequestsPerAssessment int                     `json:"model_requests_per_assessment"`
	ExaRequestsPerAssessment   int                     `json:"exa_requests_per_assessment"`
	AnalyzedFiles              int                     `json:"analyzed_files"`
	Providers                  []llm.ProviderCandidate `json:"providers"`
	PricingReceipts            []llm.PricingReceipt    `json:"pricing_receipts"`
	Draft                      llm.CostDraft           `json:"draft"`
	EstimatedMonthlyCostUSD    float64                 `json:"estimated_monthly_cost_usd"`
}

var assessCmd = &cobra.Command{
	Use:   "assess",
	Short: "Build a source-backed starter cost model from a codebase",
	Long:  "Detect code-backed providers, retrieve official-domain pricing through Exa, and use GPT-5.6 Terra to produce a labeled starter cost model.",
	RunE:  runAssess,
}

func init() {
	assessCmd.Flags().StringVar(&assessPath, "path", defaultDetectPath, "Directory to assess recursively")
	assessCmd.Flags().StringVar(&assessOut, "out", "", "Write JSON report to file instead of stdout")
	assessCmd.Flags().StringVar(&assessModel, "model", defaultAssessModel, "OpenRouter model id used for the assessment")
}

func runAssess(cmd *cobra.Command, args []string) error {
	openRouterKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if openRouterKey == "" {
		return wrapExit(2, errors.New("OPENROUTER_API_KEY is required for assess"))
	}
	exaKey := strings.TrimSpace(os.Getenv("EXA_API_KEY"))
	if exaKey == "" {
		return wrapExit(2, errors.New("EXA_API_KEY is required for assess"))
	}

	collector := newCollector()
	files, err := collector.Collect(assessPath)
	if err != nil {
		return wrapExit(2, fmt.Errorf("collect assessment files: %w", err))
	}
	if len(files) == 0 {
		return wrapExit(2, fmt.Errorf("no supported project configuration files found under %q", assessPath))
	}

	provider, err := newProvider(openRouterKey, assessModel)
	if err != nil {
		return wrapExit(3, fmt.Errorf("initialize OpenRouter provider: %w", err))
	}
	codeContext := llm.CodeContext{Files: files}
	discoveryPrompt := llm.BuildProviderDiscoveryPrompt(codeContext)
	discoveryRaw, err := provider.Chat(cmd.Context(), []llm.Message{{Role: "user", Content: discoveryPrompt}})
	if err != nil {
		return wrapExit(3, fmt.Errorf("discover code-backed providers: %w", err))
	}
	discovery, err := llm.ParseProviderDiscoveryResponse(discoveryRaw, files)
	if err != nil {
		return wrapExit(3, fmt.Errorf("parse provider discovery: %w", err))
	}
	if len(discovery.Providers) == 0 {
		return wrapExit(2, errors.New("no code-backed billable providers were detected"))
	}

	searcher, err := newExaSearcher(exaKey)
	if err != nil {
		return wrapExit(3, fmt.Errorf("initialize Exa client: %w", err))
	}
	receipts, err := retrievePricingReceipts(cmd.Context(), searcher, discovery.Providers, assessModel)
	if err != nil {
		return wrapExit(3, err)
	}
	receipts, err = ensureAssessmentRuntimeReceipts(cmd.Context(), searcher, receipts, assessModel)
	if err != nil {
		return wrapExit(3, err)
	}

	draftPrompt := llm.BuildCostDraftPrompt(codeContext, discovery.Providers, receipts, assessModel)
	draftRaw, err := provider.Chat(cmd.Context(), []llm.Message{{
		Role:    "user",
		Content: draftPrompt,
	}})
	if err != nil {
		return wrapExit(3, fmt.Errorf("build cost-model draft: %w", err))
	}
	draft, err := llm.ParseCostDraft(draftRaw, discovery.Providers, receipts)
	if err != nil {
		return wrapExit(3, fmt.Errorf("parse cost-model draft: %w", err))
	}
	estimatedMonthlyCost := 0.0
	for _, line := range draft.CostLines {
		estimatedMonthlyCost += line.MonthlyCostUSD
	}

	report := costAssessmentReport{
		SchemaVersion:              assessmentSchemaVersion,
		Path:                       assessPath,
		Model:                      assessModel,
		ModelRequestsPerAssessment: 2,
		ExaRequestsPerAssessment:   len(receipts),
		AnalyzedFiles:              len(files),
		Providers:                  discovery.Providers,
		PricingReceipts:            receipts,
		Draft:                      draft,
		EstimatedMonthlyCostUSD:    estimatedMonthlyCost,
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return wrapExit(3, fmt.Errorf("marshal assessment report: %w", err))
	}
	if assessOut != "" {
		if err := os.WriteFile(assessOut, append(encoded, '\n'), 0600); err != nil {
			return wrapExit(3, fmt.Errorf("write assessment report to %s: %w", assessOut, err))
		}
		return nil
	}
	fmt.Println(string(encoded))
	return nil
}

func retrievePricingReceipts(ctx context.Context, searcher exa.Searcher, providers []llm.ProviderCandidate, assessmentModel string) ([]llm.PricingReceipt, error) {
	receipts := make([]llm.PricingReceipt, 0, len(providers))
	for _, provider := range providers {
		receipt, err := retrievePricingReceipt(ctx, searcher, provider, assessmentModel, "codebacked_provider")
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}

func ensureAssessmentRuntimeReceipts(ctx context.Context, searcher exa.Searcher, receipts []llm.PricingReceipt, assessmentModel string) ([]llm.PricingReceipt, error) {
	runtimes := []struct {
		candidate llm.ProviderCandidate
		role      string
	}{
		{
			candidate: llm.ProviderCandidate{
				Name:           "OpenRouter",
				Provider:       defaultDetectProvider,
				OfficialDomain: "openrouter.ai",
			},
			role: "assessment_model",
		},
		{
			candidate: llm.ProviderCandidate{
				Name:           "Exa",
				Provider:       "exa",
				OfficialDomain: "exa.ai",
			},
			role: "assessment_research",
		},
	}

	for _, runtime := range runtimes {
		found := false
		for index := range receipts {
			if receipts[index].Provider == runtime.candidate.Provider {
				receipts[index].Role = runtime.role + "_and_codebacked_provider"
				found = true
				break
			}
		}
		if found {
			continue
		}
		receipt, err := retrievePricingReceipt(ctx, searcher, runtime.candidate, assessmentModel, runtime.role)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}

func retrievePricingReceipt(ctx context.Context, searcher exa.Searcher, provider llm.ProviderCandidate, assessmentModel, role string) (llm.PricingReceipt, error) {
	trustedDomain, supported := llm.TrustedOfficialDomain(provider.Provider)
	if !supported {
		return llm.PricingReceipt{}, fmt.Errorf("provider %q is not in the trusted provider registry", provider.Provider)
	}
	if !strings.EqualFold(strings.TrimSpace(provider.OfficialDomain), trustedDomain) {
		return llm.PricingReceipt{}, fmt.Errorf("provider %q must use trusted official domain %q", provider.Provider, trustedDomain)
	}
	provider.OfficialDomain = trustedDomain
	query, pathRequirement, err := pricingSearchTarget(provider, assessmentModel)
	if err != nil {
		return llm.PricingReceipt{}, err
	}
	response, err := searcher.Search(ctx, exa.SearchRequest{
		Query:          query,
		Type:           "auto",
		NumResults:     3,
		IncludeDomains: []string{provider.OfficialDomain},
		Contents:       exa.Contents{Highlights: true},
	})
	if err != nil {
		return llm.PricingReceipt{}, fmt.Errorf("retrieve %s pricing through Exa: %w", provider.Name, err)
	}
	result, ok := firstOfficialResult(response.Results, provider.OfficialDomain, pathRequirement)
	if !ok {
		return llm.PricingReceipt{}, fmt.Errorf("Exa returned no official %s pricing result", provider.Name)
	}
	return llm.PricingReceipt{
		Provider:   provider.Provider,
		Role:       role,
		Domain:     provider.OfficialDomain,
		RequestID:  response.RequestID,
		CapturedAt: assessNow().UTC().Format(time.RFC3339),
		Title:      result.Title,
		URL:        result.URL,
		Highlights: llm.NormalizePricingHighlights(result.Highlights),
	}, nil
}

type pricingPathRequirement struct {
	ExactPath string
	Fragments []string
}

func pricingSearchTarget(provider llm.ProviderCandidate, assessmentModel string) (string, pricingPathRequirement, error) {
	if provider.Provider == defaultDetectProvider && strings.TrimSpace(assessmentModel) != "" {
		modelPath := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(assessmentModel)), "/")
		return provider.Name + " " + assessmentModel + " pricing", pricingPathRequirement{
			ExactPath: "/" + modelPath,
		}, nil
	}
	fragments, supported := llm.TrustedPricingPathFragments(provider.Provider)
	if !supported || len(fragments) == 0 {
		return "", pricingPathRequirement{}, fmt.Errorf("provider %q has no trusted pricing-page paths", provider.Provider)
	}
	return provider.Name + " pricing", pricingPathRequirement{Fragments: fragments}, nil
}

func firstOfficialResult(results []exa.SearchResult, domain string, requirement pricingPathRequirement) (exa.SearchResult, bool) {
	for _, result := range results {
		parsed, err := url.Parse(result.URL)
		if err != nil {
			continue
		}
		host := strings.ToLower(parsed.Hostname())
		if host == domain || strings.HasSuffix(host, "."+domain) {
			if !pricingPathMatches(parsed.Path, requirement) {
				continue
			}
			return result, true
		}
	}
	return exa.SearchResult{}, false
}

func pricingPathMatches(actual string, requirement pricingPathRequirement) bool {
	actual = "/" + strings.Trim(strings.ToLower(strings.TrimSpace(actual)), "/")
	if requirement.ExactPath != "" {
		required := "/" + strings.Trim(strings.ToLower(strings.TrimSpace(requirement.ExactPath)), "/")
		return actual == required
	}
	for _, fragment := range requirement.Fragments {
		fragment = "/" + strings.Trim(strings.ToLower(strings.TrimSpace(fragment)), "/")
		if actual == fragment || strings.HasSuffix(actual, fragment) || strings.Contains(actual, fragment+"/") {
			return true
		}
	}
	return false
}
