package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/internal/scanner"
	"github.com/IntelIP/ProfitCtl/internal/scanner/llm"
	"github.com/IntelIP/ProfitCtl/pkg/types"
)

type FileReport struct {
	File     string   `json:"file"`
	Passed   bool     `json:"passed"`
	Issues   []string `json:"issues,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type JudgeReport struct {
	Passed bool         `json:"passed"`
	Files  []FileReport `json:"files"`
}

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		paths = []string{
			"skills/profitctl-cost-aware/references/templates",
			"test/fixtures/assessment_valid.json",
			"test/fixtures/agent_recommendation_valid.md",
		}
	}

	files, err := collectStandardsFiles(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect standards files: %v\n", err)
		os.Exit(2)
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no standards files found")
		os.Exit(2)
	}

	report := JudgeReport{Passed: true}
	for _, file := range files {
		fileReport := judgeFile(file)
		if !fileReport.Passed {
			report.Passed = false
		}
		report.Files = append(report.Files, fileReport)
	}

	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal report: %v\n", err)
		os.Exit(2)
	}
	fmt.Println(string(encoded))
	if !report.Passed {
		os.Exit(1)
	}
}

func collectStandardsFiles(paths []string) ([]string, error) {
	var files []string
	for _, raw := range paths {
		path := filepath.Clean(raw)
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if isStandardsFile(path) {
				files = append(files, path)
			}
			continue
		}
		err = filepath.WalkDir(path, func(candidate string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if isYAML(candidate) {
				files = append(files, candidate)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func isStandardsFile(path string) bool {
	return isYAML(path) || isRecommendationArtifact(path) || isAssessmentArtifact(path)
}

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yml" || ext == ".yaml"
}

func isRecommendationArtifact(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".txt"
}

func isAssessmentArtifact(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".json")
}

func judgeFile(file string) FileReport {
	if isAssessmentArtifact(file) {
		return judgeAssessmentFile(file)
	}
	if isRecommendationArtifact(file) {
		return judgeRecommendationFile(file)
	}
	return judgeScenarioFile(file)
}

func judgeScenarioFile(file string) FileReport {
	report := FileReport{File: file, Passed: true}
	cfg, err := config.ParseConfig(file)
	if err != nil {
		report.Passed = false
		report.Issues = append(report.Issues, fmt.Sprintf("scenario does not parse or validate: %v", err))
		return report
	}

	sourceCounts := map[types.CostSourceType]int{}
	for _, cost := range cfg.FixedCosts {
		judgeSource(&report, "fixed cost", cost.Name, cost.EconomicsLayer, cost.Allocation, cost.Source)
		if cost.Source != nil {
			sourceCounts[cost.Source.Type]++
		}
	}
	for _, cost := range cfg.VariableCosts {
		judgeSource(&report, "variable cost", cost.Name, cost.EconomicsLayer, cost.Allocation, cost.Source)
		if cost.Source != nil {
			sourceCounts[cost.Source.Type]++
		}
	}

	if len(cfg.Covenants) == 0 {
		report.Issues = append(report.Issues, "scenario has no covenants")
	}
	if !hasCovenant(cfg, "margin") && !hasCovenant(cfg, "p95_margin") {
		report.Issues = append(report.Issues, "scenario lacks margin or p95_margin covenant")
	}
	if !hasCovenant(cfg, "cost_per_user") && !hasCovenant(cfg, "p95_cost_per_user") {
		report.Issues = append(report.Issues, "scenario lacks cost_per_user or p95_cost_per_user covenant")
	}

	totalSources := 0
	for _, count := range sourceCounts {
		totalSources += count
	}
	if totalSources > 0 && sourceCounts[types.CostSourceTemplate] == totalSources {
		report.Warnings = append(report.Warnings, "all cost inputs are template-derived; confidence should stay low/medium until calibrated")
	}
	if sourceCounts[types.CostSourceTelemetry] == 0 && sourceCounts[types.CostSourceInvoice] == 0 {
		report.Warnings = append(report.Warnings, "no telemetry or invoice-backed cost inputs present")
	}

	if len(report.Issues) > 0 {
		report.Passed = false
	}
	return report
}

func judgeSource(report *FileReport, kind, name string, economicsLayer types.EconomicsLayer, allocation *types.EconomicsAllocation, source *types.CostSource) {
	label := fmt.Sprintf("%s %q", kind, name)
	if source == nil {
		report.Issues = append(report.Issues, label+" is missing source provenance")
		return
	}
	if source.Type == types.CostSourceTemplate && source.Confidence == types.CostSourceConfidenceHigh {
		report.Issues = append(report.Issues, label+" uses high confidence for template-derived input")
	}
	if source.Type == types.CostSourceProviderCatalog && source.URL == "" && source.Note == "" {
		report.Issues = append(report.Issues, label+" provider_catalog source needs url or note")
	}
	if (source.Type == types.CostSourceTelemetry || source.Type == types.CostSourceInvoice) && source.CapturedAt == "" && source.Note == "" {
		report.Issues = append(report.Issues, label+" telemetry/invoice source needs captured_at or note")
	}
	if types.NormalizeEconomicsLayer(economicsLayer) != types.EconomicsLayerDelivery && allocation == nil {
		report.Issues = append(report.Issues, label+" non-delivery cost needs allocation")
	}
}

func hasCovenant(cfg *config.Config, field string) bool {
	for _, covenant := range cfg.Covenants {
		if covenant.Field == field {
			return true
		}
	}
	return false
}

type assessmentArtifact struct {
	SchemaVersion              string                  `json:"schema_version"`
	Path                       string                  `json:"path"`
	Model                      string                  `json:"model"`
	ModelRequestsPerAssessment int                     `json:"model_requests_per_assessment"`
	ExaRequestsPerAssessment   int                     `json:"exa_requests_per_assessment"`
	AnalyzedFiles              int                     `json:"analyzed_files"`
	Providers                  []llm.ProviderCandidate `json:"providers"`
	PricingReceipts            []llm.PricingReceipt    `json:"pricing_receipts"`
	Draft                      json.RawMessage         `json:"draft"`
	EstimatedMonthlyCostUSD    float64                 `json:"estimated_monthly_cost_usd"`
}

func judgeAssessmentFile(file string) FileReport {
	report := FileReport{File: file, Passed: true}
	data, err := os.ReadFile(file)
	if err != nil {
		report.Passed = false
		report.Issues = append(report.Issues, fmt.Sprintf("assessment artifact cannot be read: %v", err))
		return report
	}

	var artifact assessmentArtifact
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&artifact); err != nil {
		report.Passed = false
		report.Issues = append(report.Issues, fmt.Sprintf("assessment artifact does not match profitctl.assessment/v1: %v", err))
		return report
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		report.Passed = false
		if err == nil {
			report.Issues = append(report.Issues, "assessment artifact contains trailing JSON data")
		} else {
			report.Issues = append(report.Issues, fmt.Sprintf("assessment artifact contains trailing data: %v", err))
		}
		return report
	}

	if artifact.SchemaVersion != "profitctl.assessment/v1" {
		report.Issues = append(report.Issues, "assessment schema_version must be profitctl.assessment/v1")
	}
	if strings.TrimSpace(artifact.Path) == "" || strings.TrimSpace(artifact.Model) == "" {
		report.Issues = append(report.Issues, "assessment requires path and model")
	}
	if artifact.ModelRequestsPerAssessment != 2 {
		report.Issues = append(report.Issues, "assessment must report exactly two model requests")
	}
	if artifact.ExaRequestsPerAssessment != len(artifact.PricingReceipts) {
		report.Issues = append(report.Issues, "assessment Exa request count must equal pricing receipt count")
	}
	if artifact.ExaRequestsPerAssessment < 1 || artifact.ExaRequestsPerAssessment > llm.MaxProviderCandidates+2 {
		report.Issues = append(report.Issues, "assessment Exa request count is outside the bounded limit")
	}
	if artifact.AnalyzedFiles < 1 || len(artifact.Providers) == 0 {
		report.Issues = append(report.Issues, "assessment requires analyzed files and at least one code-backed provider")
	}
	if len(artifact.Providers) > llm.MaxProviderCandidates {
		report.Issues = append(report.Issues, fmt.Sprintf("assessment provider count exceeds limit of %d", llm.MaxProviderCandidates))
	}
	assessmentRoot, assessmentRootErr := resolveAssessmentRoot(file, artifact.Path)
	if assessmentRootErr != nil {
		report.Issues = append(report.Issues, fmt.Sprintf("assessment path cannot be resolved: %v", assessmentRootErr))
	}

	providerDomains := make(map[string]string, len(artifact.Providers))
	declaredProviders := make(map[string]struct{}, len(artifact.Providers))
	for index, provider := range artifact.Providers {
		providerKey := strings.ToLower(strings.TrimSpace(provider.Provider))
		if strings.TrimSpace(provider.Name) == "" || providerKey == "" || strings.TrimSpace(provider.OfficialDomain) == "" || len(provider.Evidence) == 0 {
			report.Issues = append(report.Issues, fmt.Sprintf("providers[%d] requires name, provider, official_domain, and exact evidence", index))
			continue
		}
		trustedDomain, supported := llm.TrustedOfficialDomain(providerKey)
		if !supported {
			report.Issues = append(report.Issues, fmt.Sprintf("providers[%d].provider is not in the trusted provider registry", index))
			continue
		}
		if _, duplicate := declaredProviders[providerKey]; duplicate {
			report.Issues = append(report.Issues, fmt.Sprintf("providers[%d].provider duplicates %q", index, providerKey))
		}
		if strings.ToLower(strings.TrimSpace(provider.OfficialDomain)) != trustedDomain {
			report.Issues = append(report.Issues, fmt.Sprintf("providers[%d].official_domain does not match the trusted provider registry", index))
		}
		for evidenceIndex, evidence := range provider.Evidence {
			if strings.TrimSpace(evidence.File) == "" || strings.TrimSpace(evidence.Excerpt) == "" {
				report.Issues = append(report.Issues, fmt.Sprintf("providers[%d].evidence[%d] requires file and excerpt", index, evidenceIndex))
				continue
			}
			if !llm.ProviderEvidenceMatches(providerKey, evidence.Excerpt) {
				report.Issues = append(report.Issues, fmt.Sprintf("providers[%d].evidence[%d].excerpt does not identify provider", index, evidenceIndex))
			}
			if assessmentRootErr == nil {
				content, err := readAssessmentEvidence(assessmentRoot, evidence.File)
				if err != nil {
					report.Issues = append(report.Issues, fmt.Sprintf("providers[%d].evidence[%d].file cannot be read from assessed path: %v", index, evidenceIndex, err))
				} else if !strings.Contains(providerDetectionContent(evidence.File, content), evidence.Excerpt) {
					report.Issues = append(report.Issues, fmt.Sprintf("providers[%d].evidence[%d].excerpt is not exact assessed code", index, evidenceIndex))
				}
			}
		}
		providerDomains[providerKey] = trustedDomain
		declaredProviders[providerKey] = struct{}{}
	}
	if assessmentRootErr == nil {
		files, err := scanner.NewCollector().Collect(assessmentRoot)
		if err != nil {
			report.Issues = append(report.Issues, fmt.Sprintf("assessed files cannot be recollected: %v", err))
		} else {
			if artifact.AnalyzedFiles != len(files) {
				report.Issues = append(report.Issues, "analyzed_files must equal the recollected supported file count")
			}
			discoveryJSON, err := json.Marshal(llm.ProviderDiscoveryResponse{Providers: artifact.Providers})
			if err != nil {
				report.Issues = append(report.Issues, fmt.Sprintf("assessment providers cannot be encoded: %v", err))
			} else if _, err := llm.ParseProviderDiscoveryResponse(string(discoveryJSON), files); err != nil {
				report.Issues = append(report.Issues, fmt.Sprintf("assessment providers do not match provider-discovery rules: %v", err))
			}
			for _, detected := range detectCodeBackedProviders(files) {
				if !declaresAnyProvider(declaredProviders, detected.Aliases) {
					report.Issues = append(report.Issues, fmt.Sprintf(
						"assessed files contain code-backed provider %q in %s but the assessment omits it",
						detected.Provider,
						detected.File,
					))
				}
			}
		}
	}

	receiptProviders := make(map[string]struct{}, len(artifact.PricingReceipts))
	hasAssessmentModelReceipt := false
	hasAssessmentResearchReceipt := false
	for index, receipt := range artifact.PricingReceipts {
		providerKey := strings.ToLower(strings.TrimSpace(receipt.Provider))
		roles, rolesValid := receiptRoleTokens(receipt.Role)
		if !rolesValid {
			report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d].role is invalid", index))
		}
		if hasReceiptRole(roles, llm.CostRoleAssessmentModel) && providerKey != "openrouter" {
			report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d].assessment_model role requires provider \"openrouter\"", index))
		}
		if hasReceiptRole(roles, llm.CostRoleAssessmentResearch) && providerKey != "exa" {
			report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d].assessment_research role requires provider \"exa\"", index))
		}
		receiptProviders[providerKey] = struct{}{}
		isAssessmentModelReceipt := hasReceiptRole(roles, llm.CostRoleAssessmentModel) && providerKey == "openrouter"
		if isAssessmentModelReceipt {
			hasAssessmentModelReceipt = true
		}
		if hasReceiptRole(roles, llm.CostRoleAssessmentResearch) && providerKey == "exa" {
			hasAssessmentResearchReceipt = true
		}
		if strings.TrimSpace(receipt.RequestID) == "" || strings.TrimSpace(receipt.CapturedAt) == "" || strings.TrimSpace(receipt.Title) == "" || len(receipt.Highlights) == 0 {
			report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d] requires request_id, captured_at, title, and highlights", index))
		}
		if _, err := time.Parse(time.RFC3339, receipt.CapturedAt); err != nil {
			report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d].captured_at must be RFC3339", index))
		}
		if !urlMatchesDomain(receipt.URL, receipt.Domain) {
			report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d].url does not match domain", index))
		}
		trustedDomain, supported := llm.TrustedOfficialDomain(providerKey)
		if !supported || strings.ToLower(strings.TrimSpace(receipt.Domain)) != trustedDomain {
			report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d].domain does not match the trusted provider registry", index))
		}
		if expectedDomain, exists := providerDomains[providerKey]; exists && expectedDomain != receipt.Domain {
			report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d].domain does not match detected provider", index))
		}
		if isAssessmentModelReceipt {
			if !urlPathMatchesExact(receipt.URL, artifact.Model) {
				report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d].url does not match selected assessment model", index))
			}
		} else {
			fragments, supported := llm.TrustedPricingPathFragments(providerKey)
			if !supported || !urlPathMatchesFragments(receipt.URL, fragments) {
				report.Issues = append(report.Issues, fmt.Sprintf("pricing_receipts[%d].url does not match a trusted pricing path", index))
			}
		}
	}
	for _, provider := range artifact.Providers {
		if _, exists := receiptProviders[provider.Provider]; !exists {
			report.Issues = append(report.Issues, fmt.Sprintf("provider %q has no pricing receipt", provider.Provider))
		}
	}
	if !hasAssessmentModelReceipt || !hasAssessmentResearchReceipt {
		report.Issues = append(report.Issues, "assessment requires OpenRouter model and Exa research runtime receipts")
	}

	var draftObject map[string]json.RawMessage
	if err := json.Unmarshal(artifact.Draft, &draftObject); err != nil || draftObject == nil {
		report.Issues = append(report.Issues, "assessment draft must be a JSON object")
	} else if draft, err := llm.ParseCostDraft(string(artifact.Draft), artifact.Providers, artifact.PricingReceipts); err != nil {
		report.Issues = append(report.Issues, fmt.Sprintf("assessment draft is invalid: %v", err))
	} else {
		lineProviders := make(map[string]struct{}, len(draft.CostLines))
		total := 0.0
		exaRuntimeUnits := 0.0
		modelInputUnits := 0.0
		modelOutputUnits := 0.0
		codeBackedUnits := make(map[string]float64, len(artifact.Providers))
		for _, line := range draft.CostLines {
			lineProviders[line.Provider] = struct{}{}
			total += line.MonthlyCostUSD
			if line.Role == llm.CostRoleCodebackedProvider {
				codeBackedUnits[line.Provider] += line.UnitsPerMonth
			}
			if line.Provider == "exa" && line.Role == llm.CostRoleAssessmentResearch {
				exaRuntimeUnits += line.UnitsPerMonth * pricedUnitScale(line.Unit)
			}
			if line.Provider == "openrouter" && line.Role == llm.CostRoleAssessmentModel {
				unit := strings.ToLower(line.Unit)
				if strings.Contains(unit, "input") {
					modelInputUnits += line.UnitsPerMonth
				}
				if strings.Contains(unit, "output") {
					modelOutputUnits += line.UnitsPerMonth
				}
			}
		}
		for _, provider := range artifact.Providers {
			if _, exists := lineProviders[provider.Provider]; !exists {
				report.Issues = append(report.Issues, fmt.Sprintf("provider %q has no cost line", provider.Provider))
			}
			if codeBackedUnits[provider.Provider] <= 0 {
				report.Issues = append(report.Issues, fmt.Sprintf("provider %q code-backed units_per_month must be positive", provider.Provider))
			}
		}
		if math.Abs(total-artifact.EstimatedMonthlyCostUSD) > 0.000001*math.Max(1, math.Abs(total)) {
			report.Issues = append(report.Issues, "estimated_monthly_cost_usd must equal the sum of cost lines")
		}
		if exaRuntimeUnits+0.000001 < float64(artifact.ExaRequestsPerAssessment) {
			report.Issues = append(report.Issues, "assessment_research units_per_month must cover exa_requests_per_assessment")
		}
		if artifact.ModelRequestsPerAssessment > 0 && (modelInputUnits <= 0 || modelOutputUnits <= 0) {
			report.Issues = append(report.Issues, "assessment_model input and output units_per_month must be positive when model requests are recorded")
		}
	}

	if len(report.Issues) > 0 {
		report.Passed = false
	}
	return report
}

func receiptRoleTokens(raw string) (map[string]struct{}, bool) {
	allowed := map[string]struct{}{
		llm.CostRoleCodebackedProvider: {},
		llm.CostRoleAssessmentModel:    {},
		llm.CostRoleAssessmentResearch: {},
	}
	tokens := make(map[string]struct{})
	for _, token := range strings.Split(strings.ToLower(strings.TrimSpace(raw)), "_and_") {
		if _, exists := allowed[token]; !exists {
			return tokens, false
		}
		if _, duplicate := tokens[token]; duplicate {
			return tokens, false
		}
		tokens[token] = struct{}{}
	}
	return tokens, len(tokens) > 0
}

func hasReceiptRole(tokens map[string]struct{}, role string) bool {
	_, exists := tokens[role]
	return exists
}

func resolveAssessmentRoot(artifactFile, rawPath string) (string, error) {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return "", fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(rawPath) {
		return requireDirectory(rawPath)
	}
	if root, err := requireDirectory(rawPath); err == nil {
		return root, nil
	}

	starts := []string{}
	if cwd, err := os.Getwd(); err == nil {
		starts = append(starts, cwd)
	}
	if absoluteFile, err := filepath.Abs(artifactFile); err == nil {
		starts = append(starts, filepath.Dir(absoluteFile))
	}
	for _, start := range starts {
		if repoRoot, ok := findGoModuleRoot(start); ok {
			if root, err := requireDirectory(filepath.Join(repoRoot, rawPath)); err == nil {
				return root, nil
			}
		}
	}
	return "", fmt.Errorf("%q is not a directory", rawPath)
}

func requireDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", path)
	}
	return absolute, nil
}

func findGoModuleRoot(start string) (string, bool) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		if info, err := os.Stat(filepath.Join(current, "go.mod")); err == nil && !info.IsDir() {
			return current, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

func readAssessmentEvidence(root, evidenceFile string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(evidenceFile))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("evidence path must stay within the assessed path")
	}
	candidate := filepath.Join(root, clean)
	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("evidence path must stay within the assessed path")
	}
	data, err := os.ReadFile(candidate)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func urlPathMatchesExact(rawURL, expectedPath string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	actual := "/" + strings.Trim(strings.ToLower(parsed.Path), "/")
	expected := "/" + strings.Trim(strings.ToLower(strings.TrimSpace(expectedPath)), "/")
	return actual == expected
}

func urlPathMatchesFragments(rawURL string, fragments []string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	actual := "/" + strings.Trim(strings.ToLower(parsed.Path), "/")
	for _, fragment := range fragments {
		fragment = "/" + strings.Trim(strings.ToLower(strings.TrimSpace(fragment)), "/")
		if actual == fragment || strings.HasSuffix(actual, fragment) || strings.Contains(actual, fragment+"/") {
			return true
		}
	}
	return false
}

func urlMatchesDomain(rawURL, domain string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	domain = strings.ToLower(strings.TrimSpace(domain))
	return domain != "" && (host == domain || strings.HasSuffix(host, "."+domain))
}

type detectedProvider struct {
	Provider string
	Aliases  []string
	File     string
}

type providerSignature struct {
	Provider string
	Aliases  []string
	Markers  []string
}

var codeBackedProviderSignatures = []providerSignature{
	{Provider: "anthropic", Aliases: []string{"anthropic"}, Markers: []string{"api.anthropic.com", "anthropic_api_key", "@anthropic-ai/", "github.com/anthropics/"}},
	{Provider: "aws", Aliases: []string{"aws"}, Markers: []string{"github.com/aws/", "@aws-sdk/", "aws_access_key_id", "amazonaws.com", `provider "aws"`}},
	{Provider: "azure", Aliases: []string{"azure"}, Markers: []string{"github.com/azure/", "@azure/", "azure.microsoft.com", "azure_client_id", `provider "azurerm"`}},
	{Provider: "buildkite", Aliases: []string{"buildkite"}, Markers: []string{"api.buildkite.com", "buildkite_agent_token", "@buildkite/", "github.com/buildkite/"}},
	{Provider: "clerk", Aliases: []string{"clerk"}, Markers: []string{"@clerk/", "clerk.com", "github.com/clerk/"}},
	{Provider: "cloudflare", Aliases: []string{"cloudflare"}, Markers: []string{"api.cloudflare.com", "cloudflare_api_token", "@cloudflare/", "github.com/cloudflare/", "workers.dev", `provider "cloudflare"`}},
	{Provider: "datadog", Aliases: []string{"datadog"}, Markers: []string{"api.datadoghq.com", "datadog_api_key", "dd_api_key", "@datadog/", "github.com/datadog/"}},
	{Provider: "exa", Aliases: []string{"exa"}, Markers: []string{"exa.ai", "exa_api_key", "exa-py", "exa-js", "github.com/exa-labs/"}},
	{Provider: "gcp", Aliases: []string{"gcp", "google-cloud"}, Markers: []string{"cloud.google.com", "google.golang.org/api", "@google-cloud/", "google-cloud-", `provider "google"`, "google_application_credentials"}},
	{Provider: "github", Aliases: []string{"github"}, Markers: []string{"api.github.com", "github_token", "github_app", "github_repository", "google/go-github", "@octokit/", "@actions/"}},
	{Provider: "mongodb-atlas", Aliases: []string{"mongodb-atlas"}, Markers: []string{"mongodb+srv", "mongodb-atlas", "mongodbatlas", "go.mongodb.org/mongo-driver", "@mongodb-js/"}},
	{Provider: "neon", Aliases: []string{"neon"}, Markers: []string{"neon.tech", "@neondatabase/", "neon_database", "neon_api_key"}},
	{Provider: "openai", Aliases: []string{"openai"}, Markers: []string{"api.openai.com", "openai_api_key", "@openai/", "github.com/openai/", "github.com/sashabaranov/go-openai"}},
	{Provider: "openrouter", Aliases: []string{"openrouter"}, Markers: []string{"openrouter.ai", "openrouter_api_key", "github.com/revrost/go-openrouter", "@openrouter/"}},
	{Provider: "pinecone", Aliases: []string{"pinecone"}, Markers: []string{"api.pinecone.io", "pinecone_api_key", "@pinecone-database/", "github.com/pinecone-io/"}},
	{Provider: "resend", Aliases: []string{"resend"}, Markers: []string{"api.resend.com", "resend_api_key", "@resend/", "github.com/resend/"}},
	{Provider: "sentry", Aliases: []string{"sentry"}, Markers: []string{"sentry.io", "sentry_dsn", "@sentry/", "github.com/getsentry/"}},
	{Provider: "stripe", Aliases: []string{"stripe"}, Markers: []string{"api.stripe.com", "stripe_api_key", "stripe_secret_key", "@stripe/", "github.com/stripe/"}},
	{Provider: "supabase", Aliases: []string{"supabase"}, Markers: []string{"supabase.co", "supabase_url", "supabase_key", "@supabase/", "github.com/supabase/"}},
	{Provider: "twilio", Aliases: []string{"twilio"}, Markers: []string{"api.twilio.com", "twilio_account_sid", "@twilio/", "github.com/twilio/"}},
	{Provider: "upstash", Aliases: []string{"upstash"}, Markers: []string{"upstash.com", "upstash_redis_rest_url", "@upstash/", "github.com/upstash/"}},
	{Provider: "vercel", Aliases: []string{"vercel"}, Markers: []string{"api.vercel.com", "vercel_url", "vercel_token", "@vercel/", "github.com/vercel/"}},
}

var directNPMProviderPackages = map[string][]string{
	"openai": {"openai"},
	"resend": {"resend"},
	"stripe": {"stripe"},
	"twilio": {"twilio"},
}

func detectCodeBackedProviders(files map[string]string) []detectedProvider {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var detected []detectedProvider
	for _, signature := range codeBackedProviderSignatures {
		foundFile := ""
		for _, path := range paths {
			rawContent := files[path]
			if packageManifestDeclaresProvider(path, rawContent, signature.Provider) {
				foundFile = path
				break
			}
			content := strings.ToLower(providerDetectionContent(path, rawContent))
			for _, marker := range signature.Markers {
				if strings.Contains(content, marker) {
					foundFile = path
					break
				}
			}
			if foundFile != "" {
				break
			}
		}
		if foundFile != "" {
			detected = append(detected, detectedProvider{
				Provider: signature.Provider,
				Aliases:  signature.Aliases,
				File:     foundFile,
			})
		}
	}
	return detected
}

func packageManifestDeclaresProvider(path, content, provider string) bool {
	if strings.ToLower(filepath.Base(path)) != "package.json" {
		return false
	}
	packageNames := directNPMProviderPackages[provider]
	if len(packageNames) == 0 {
		return false
	}
	var manifest struct {
		Dependencies         map[string]json.RawMessage `json:"dependencies"`
		OptionalDependencies map[string]json.RawMessage `json:"optionalDependencies"`
	}
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		return false
	}
	for _, packageName := range packageNames {
		if _, exists := manifest.Dependencies[packageName]; exists {
			return true
		}
		if _, exists := manifest.OptionalDependencies[packageName]; exists {
			return true
		}
	}
	return false
}

func providerDetectionContent(path, content string) string {
	switch strings.ToLower(filepath.Base(path)) {
	case "go.sum":
		return ""
	case "go.mod":
		return directGoModuleRequirements(content)
	default:
		return content
	}
}

func directGoModuleRequirements(content string) string {
	var direct []string
	inRequireBlock := false
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if inRequireBlock {
			if line == ")" {
				inRequireBlock = false
				continue
			}
			if line == "" || strings.HasPrefix(line, "//") || strings.Contains(line, "// indirect") {
				continue
			}
			direct = append(direct, line)
			continue
		}
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "// indirect") {
			direct = append(direct, line)
		}
	}
	return strings.Join(direct, "\n")
}

func declaresAnyProvider(declared map[string]struct{}, aliases []string) bool {
	for _, alias := range aliases {
		if _, exists := declared[alias]; exists {
			return true
		}
	}
	return false
}

func judgeRecommendationFile(file string) FileReport {
	report := FileReport{File: file, Passed: true}
	data, err := os.ReadFile(file)
	if err != nil {
		report.Passed = false
		report.Issues = append(report.Issues, fmt.Sprintf("recommendation artifact cannot be read: %v", err))
		return report
	}

	body := strings.ToLower(string(data))
	sections := recommendationSections(string(data))
	requireRecommendationSection(&report, sections, "recommendation")
	requireRecommendationSection(&report, sections, "assumptions")
	requireRecommendationSection(&report, sections, "economics")
	requireRecommendationSection(&report, sections, "alternative", "tradeoff")
	requireAny(&report, "recommendation artifact needs an explicit recommendation", body, "recommendation")
	requireAny(&report, "recommendation artifact needs explicit assumptions", body, "assumption", "assumptions")
	requireAny(&report, "recommendation artifact needs monthly fixed cost", body, "fixed cost", "fixed_cost")
	requireAny(&report, "recommendation artifact needs variable cost drivers", body, "variable cost", "variable_cost", "cost driver")
	requireAny(&report, "recommendation artifact needs gross margin evidence", body, "gross margin", "modeled margin", "modelled margin")
	requireAny(&report, "recommendation artifact needs p95 stress evidence", body, "p95")
	requireAny(&report, "recommendation artifact needs cost per user evidence", body, "cost/user", "cost per user", "cost_per_user", "cost/active user", "cost per active user", "cost_per_active_user")
	requireAny(&report, "recommendation artifact needs covenant status", body, "covenant")
	requireAny(&report, "recommendation artifact needs cheaper alternative analysis", body, "alternative", "tradeoff")
	requireAny(&report, "recommendation artifact needs source provenance", body, "source", "provenance")
	requireAny(&report, "recommendation artifact needs confidence level", body, "confidence")

	if (strings.Contains(body, "$") || strings.Contains(body, "%")) && (!hasAny(body, "source", "provenance") || !hasAny(body, "confidence")) {
		report.Issues = append(report.Issues, "precise cost or margin claims need provenance and confidence")
	}
	if hasAffirmedHighCertaintyClaim(body) && !hasMeasuredRecommendationEvidence(body) {
		report.Issues = append(report.Issues, "high-certainty recommendation language needs telemetry or invoice evidence")
	}
	if !hasAny(body, "template", "repo_detected", "user_supplied", "provider_catalog", "telemetry", "invoice") {
		report.Warnings = append(report.Warnings, "recommendation does not name a supported source type")
	}

	if len(report.Issues) > 0 {
		report.Passed = false
	}
	return report
}

func recommendationSections(body string) map[string]string {
	sections := make(map[string]string)
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "#-* \t")
		line = strings.ReplaceAll(line, "**", "")
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		label := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if label != "" && value != "" {
			sections[label] = value
		}
	}
	return sections
}

func requireRecommendationSection(report *FileReport, sections map[string]string, label string, aliases ...string) {
	for _, candidate := range append([]string{label}, aliases...) {
		value, exists := sections[candidate]
		normalized := strings.ToLower(strings.Trim(strings.TrimSpace(value), "."))
		if exists && !recommendationSectionPlaceholder(candidate, normalized) && len(strings.Fields(value)) >= 2 {
			return
		}
	}
	report.Issues = append(report.Issues, fmt.Sprintf("recommendation artifact needs a labeled %s section with a non-placeholder value", label))
}

func recommendationSectionPlaceholder(label, value string) bool {
	for _, placeholder := range []string{"n/a", "none", "not available", "not provided", "unknown"} {
		if value == placeholder || strings.HasPrefix(value, placeholder+" ") {
			return true
		}
	}
	labels := []string{label}
	if label == "assumptions" {
		labels = append(labels, "assumption")
	}
	for _, candidate := range labels {
		for _, suffix := range []string{"", " available", " identified", " provided"} {
			if value == "no "+candidate+suffix {
				return true
			}
		}
	}
	return label == "economics" && strings.HasPrefix(value, "no fixed cost")
}

func hasAffirmedHighCertaintyClaim(body string) bool {
	for _, line := range strings.Split(strings.ToLower(body), "\n") {
		for _, term := range []string{"guaranteed", "always", "invoice-grade"} {
			for offset := 0; offset < len(line); {
				index := strings.Index(line[offset:], term)
				if index < 0 {
					break
				}
				index += offset
				prefixStart := index - 24
				if prefixStart < 0 {
					prefixStart = 0
				}
				prefix := strings.TrimSpace(line[prefixStart:index])
				negated := false
				for _, phrase := range []string{"not", "never", "no longer", "isn't", "is not", "cannot be", "can't be"} {
					if prefix == phrase || strings.HasSuffix(prefix, " "+phrase) {
						negated = true
						break
					}
				}
				if !negated {
					return true
				}
				offset = index + len(term)
			}
		}
	}
	return false
}

func hasMeasuredRecommendationEvidence(body string) bool {
	for _, line := range strings.Split(strings.ToLower(body), "\n") {
		for _, sourceType := range []string{"telemetry", "invoice"} {
			if !hasAny(
				line,
				"source.type: "+sourceType,
				"source.type="+sourceType,
				"source type: "+sourceType,
				"source: "+sourceType,
				"provenance is "+sourceType,
				sourceType+" evidence",
				sourceType+"-backed",
				sourceType+" backed",
				sourceType+" shows",
				sourceType+" confirms",
				"from "+sourceType,
				"`"+sourceType+"`",
			) {
				continue
			}
			if !recommendationEvidenceIsNegated(line, sourceType) {
				return true
			}
		}
	}
	return false
}

func recommendationEvidenceIsNegated(line, sourceType string) bool {
	otherSourceType := "invoice"
	if sourceType == "invoice" {
		otherSourceType = "telemetry"
	}
	if hasAny(
		line,
		"no "+sourceType,
		"without "+sourceType,
		"missing "+sourceType,
		"lack of "+sourceType,
		"lacks "+sourceType,
		"neither "+sourceType,
		"no "+otherSourceType+" or "+sourceType,
		"without "+otherSourceType+" or "+sourceType,
		"neither "+otherSourceType+" nor "+sourceType,
	) {
		return true
	}
	return hasAny(
		line,
		sourceType+" not available",
		sourceType+" not provided",
		sourceType+" unavailable",
		sourceType+" absent",
		sourceType+" missing",
	)
}

func pricedUnitScale(unit string) float64 {
	unit = strings.ToLower(strings.TrimSpace(unit))
	switch {
	case hasAny(unit, "million", "1,000,000", "1000000"):
		return 1_000_000
	case hasAny(unit, "thousand", "1,000", "1000"):
		return 1_000
	default:
		return 1
	}
}

func requireAny(report *FileReport, issue string, body string, terms ...string) {
	if !hasAny(body, terms...) {
		report.Issues = append(report.Issues, issue)
	}
}

func hasAny(body string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(body, term) {
			return true
		}
	}
	return false
}
