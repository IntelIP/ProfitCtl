package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJudgeAssessmentFilePassesValidArtifact(t *testing.T) {
	report := judgeAssessmentFile(filepath.Join("..", "test", "fixtures", "assessment_valid.json"))

	if !report.Passed {
		t.Fatalf("expected valid assessment artifact to pass, issues: %v", report.Issues)
	}
}

func TestJudgeAssessmentFileRejectsInvalidEvidenceAndMath(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(map[string]any)
		wantIssue string
	}{
		{
			name: "unofficial receipt URL",
			mutate: func(artifact map[string]any) {
				receipts := artifact["pricing_receipts"].([]any)
				receipts[0].(map[string]any)["url"] = "https://openrouter.ai.evil.example/openai/gpt-5.6-terra"
			},
			wantIssue: "url does not match domain",
		},
		{
			name: "untrusted detected domain",
			mutate: func(artifact map[string]any) {
				providers := artifact["providers"].([]any)
				providers[0].(map[string]any)["official_domain"] = "aws.example"
			},
			wantIssue: "official_domain does not match the trusted provider registry",
		},
		{
			name: "provider evidence does not identify provider",
			mutate: func(artifact map[string]any) {
				providers := artifact["providers"].([]any)
				evidence := providers[0].(map[string]any)["evidence"].([]any)
				evidence[0].(map[string]any)["excerpt"] = "github.com/example/unrelated v1.0.0"
			},
			wantIssue: "excerpt does not identify provider",
		},
		{
			name: "provider evidence is not exact assessed code",
			mutate: func(artifact map[string]any) {
				providers := artifact["providers"].([]any)
				evidence := providers[0].(map[string]any)["evidence"].([]any)
				evidence[0].(map[string]any)["excerpt"] = "github.com/aws/aws-sdk-go v9.99.0"
			},
			wantIssue: "excerpt is not exact assessed code",
		},
		{
			name: "provider evidence file is missing",
			mutate: func(artifact map[string]any) {
				providers := artifact["providers"].([]any)
				evidence := providers[0].(map[string]any)["evidence"].([]any)
				evidence[0].(map[string]any)["file"] = "missing.mod"
			},
			wantIssue: "file cannot be read from assessed path",
		},
		{
			name: "model receipt does not match selected model",
			mutate: func(artifact map[string]any) {
				receipts := artifact["pricing_receipts"].([]any)
				receipts[0].(map[string]any)["url"] = "https://openrouter.ai/openai/gpt-5.6-terra-pro"
				draft := artifact["draft"].(map[string]any)
				lines := draft["cost_lines"].([]any)
				lines[1].(map[string]any)["source_url"] = "https://openrouter.ai/openai/gpt-5.6-terra-pro"
				lines[2].(map[string]any)["source_url"] = "https://openrouter.ai/openai/gpt-5.6-terra-pro"
			},
			wantIssue: "url does not match selected assessment model",
		},
		{
			name: "receipt does not use trusted pricing path",
			mutate: func(artifact map[string]any) {
				receipts := artifact["pricing_receipts"].([]any)
				receipts[2].(map[string]any)["url"] = "https://aws.amazon.com/blogs/example"
				draft := artifact["draft"].(map[string]any)
				lines := draft["cost_lines"].([]any)
				lines[0].(map[string]any)["source_url"] = "https://aws.amazon.com/blogs/example"
			},
			wantIssue: "url does not match a trusted pricing path",
		},
		{
			name: "incorrect monthly cost",
			mutate: func(artifact map[string]any) {
				draft := artifact["draft"].(map[string]any)
				lines := draft["cost_lines"].([]any)
				lines[0].(map[string]any)["monthly_cost_usd"] = 0.02
			},
			wantIssue: "must equal price_per_unit_usd",
		},
		{
			name: "unit does not match pricing excerpt",
			mutate: func(artifact map[string]any) {
				draft := artifact["draft"].(map[string]any)
				lines := draft["cost_lines"].([]any)
				lines[0].(map[string]any)["unit"] = "million output tokens"
			},
			wantIssue: "unit is not fully stated in source_excerpt",
		},
		{
			name: "missing inference label",
			mutate: func(artifact map[string]any) {
				draft := artifact["draft"].(map[string]any)
				assumptions := draft["assumptions"].([]any)
				assumptions[0].(map[string]any)["source"] = ""
			},
			wantIssue: "source must be repo_detected or inferred",
		},
		{
			name: "missing assessment runtime role",
			mutate: func(artifact map[string]any) {
				receipts := artifact["pricing_receipts"].([]any)
				receipts[1].(map[string]any)["role"] = "codebacked_provider"
			},
			wantIssue: "requires OpenRouter model and Exa research runtime receipts",
		},
		{
			name: "assessment role uses substring instead of token",
			mutate: func(artifact map[string]any) {
				receipts := artifact["pricing_receipts"].([]any)
				receipts[0].(map[string]any)["role"] = "not_assessment_model"
			},
			wantIssue: "role is invalid",
		},
		{
			name: "assessment research understates request count",
			mutate: func(artifact map[string]any) {
				draft := artifact["draft"].(map[string]any)
				lines := draft["cost_lines"].([]any)
				lines[3].(map[string]any)["units_per_month"] = 1
				lines[3].(map[string]any)["monthly_cost_usd"] = 0.005
				artifact["estimated_monthly_cost_usd"] = 6.027
			},
			wantIssue: "must cover exa_requests_per_assessment",
		},
		{
			name: "assessment model reports zero runtime units",
			mutate: func(artifact map[string]any) {
				draft := artifact["draft"].(map[string]any)
				lines := draft["cost_lines"].([]any)
				lines[1].(map[string]any)["units_per_month"] = 0
				lines[1].(map[string]any)["monthly_cost_usd"] = 0
				lines[2].(map[string]any)["units_per_month"] = 0
				lines[2].(map[string]any)["monthly_cost_usd"] = 0
				artifact["estimated_monthly_cost_usd"] = 6.02
			},
			wantIssue: "assessment_model input and output units_per_month must be positive",
		},
		{
			name: "code-backed provider is omitted",
			mutate: func(artifact map[string]any) {
				providers := artifact["providers"].([]any)
				artifact["providers"] = providers[:1]
				receipts := artifact["pricing_receipts"].([]any)
				artifact["pricing_receipts"] = receipts[:3]
				draft := artifact["draft"].(map[string]any)
				lines := draft["cost_lines"].([]any)
				draft["cost_lines"] = lines[:4]
				artifact["exa_requests_per_assessment"] = 3
				artifact["estimated_monthly_cost_usd"] = 1.042
			},
			wantIssue: `assessed files contain code-backed provider "stripe"`,
		},
		{
			name: "analyzed file count is stale",
			mutate: func(artifact map[string]any) {
				artifact["analyzed_files"] = 2
			},
			wantIssue: "analyzed_files must equal the recollected supported file count",
		},
		{
			name: "draft is wrapped in an array",
			mutate: func(artifact map[string]any) {
				artifact["draft"] = []any{artifact["draft"]}
			},
			wantIssue: "assessment draft must be a JSON object",
		},
		{
			name: "missing assessment model output line",
			mutate: func(artifact map[string]any) {
				draft := artifact["draft"].(map[string]any)
				lines := draft["cost_lines"].([]any)
				draft["cost_lines"] = append(lines[:2], lines[3:]...)
			},
			wantIssue: "requires a cost line for provider \"openrouter\" role \"assessment_model\" usage \"output\"",
		},
		{
			name: "wrong reported total",
			mutate: func(artifact map[string]any) {
				artifact["estimated_monthly_cost_usd"] = 9.99
			},
			wantIssue: "must equal the sum of cost lines",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := writeMutatedAssessment(t, test.mutate)

			report := judgeAssessmentFile(path)

			if report.Passed {
				t.Fatalf("expected invalid assessment artifact to fail")
			}
			if !issuesContain(report.Issues, test.wantIssue) {
				t.Fatalf("expected issue containing %q, got %v", test.wantIssue, report.Issues)
			}
		})
	}
}

func TestDetectCodeBackedProvidersIgnoresTransitiveGoEntries(t *testing.T) {
	files := map[string]string{
		"go.sum": "github.com/aws/aws-sdk-go v1.48.0 h1:checksum",
		"go.mod": `module example.com/app

require github.com/stripe/stripe-go/v76 v76.15.0 // indirect
`,
	}

	if detected := detectCodeBackedProviders(files); len(detected) != 0 {
		t.Fatalf("expected transitive Go entries to be ignored, got %v", detected)
	}

	files["go.mod"] = `module example.com/app

require github.com/stripe/stripe-go/v76 v76.15.0
`
	detected := detectCodeBackedProviders(files)
	if len(detected) != 1 || detected[0].Provider != "stripe" {
		t.Fatalf("expected direct Stripe requirement, got %v", detected)
	}
}

func TestJudgeAssessmentFileRejectsTrailingJSON(t *testing.T) {
	source := filepath.Join("..", "test", "fixtures", "assessment_valid.json")
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read assessment fixture: %v", err)
	}
	path := filepath.Join(t.TempDir(), "assessment.json")
	if err := os.WriteFile(path, append(data, []byte("\n{}\n")...), 0o600); err != nil {
		t.Fatalf("write trailing assessment fixture: %v", err)
	}

	report := judgeAssessmentFile(path)

	if report.Passed || !issuesContain(report.Issues, "trailing JSON data") {
		t.Fatalf("expected trailing JSON rejection, got passed=%v issues=%v", report.Passed, report.Issues)
	}
}

func TestJudgeRecommendationFilePassesValidArtifact(t *testing.T) {
	report := judgeRecommendationFile(filepath.Join("..", "test", "fixtures", "agent_recommendation_valid.md"))

	if !report.Passed {
		t.Fatalf("expected valid recommendation artifact to pass, issues: %v", report.Issues)
	}
}

func TestJudgeRecommendationFileFailsMissingEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recommendation.md")
	if err := os.WriteFile(path, []byte("Recommendation: use the cheaper option because it costs $10."), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	report := judgeRecommendationFile(path)

	if report.Passed {
		t.Fatalf("expected incomplete recommendation artifact to fail")
	}
	if len(report.Issues) == 0 {
		t.Fatalf("expected failure issues")
	}
}

func TestJudgeRecommendationFileRejectsKeywordOnlyArtifact(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recommendation.md")
	body := "No recommendation, assumptions, fixed cost, variable cost, margin, p95, cost per user, covenant, alternative, source, confidence."
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	report := judgeRecommendationFile(path)

	if report.Passed {
		t.Fatal("expected keyword-only recommendation artifact to fail")
	}
	if !issuesContain(report.Issues, "labeled recommendation section") {
		t.Fatalf("expected labeled-section failure, got %v", report.Issues)
	}
}

func TestCollectStandardsFilesDirectoryOnlyCollectsYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("Recommendation: no"), 0o600); err != nil {
		t.Fatalf("write markdown fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assessment.json"), []byte(`{"schema_version":"no"}`), 0o600); err != nil {
		t.Fatalf("write JSON fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scenario.yml"), []byte("name: test"), 0o600); err != nil {
		t.Fatalf("write YAML fixture: %v", err)
	}

	files, err := collectStandardsFiles([]string{dir})
	if err != nil {
		t.Fatalf("collect standards files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected only YAML file from directory, got %d: %v", len(files), files)
	}
	if filepath.Base(files[0]) != "scenario.yml" {
		t.Fatalf("expected scenario.yml, got %s", files[0])
	}
}

func TestURLMatchesDomainRejectsLookalike(t *testing.T) {
	if urlMatchesDomain("https://openrouter.ai.evil.example/pricing", "openrouter.ai") {
		t.Fatal("expected lookalike domain to be rejected")
	}
}

func writeMutatedAssessment(t *testing.T, mutate func(map[string]any)) string {
	t.Helper()
	source := filepath.Join("..", "test", "fixtures", "assessment_valid.json")
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read assessment fixture: %v", err)
	}
	var artifact map[string]any
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fatalf("parse assessment fixture: %v", err)
	}
	mutate(artifact)
	encoded, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		t.Fatalf("marshal assessment fixture: %v", err)
	}
	path := filepath.Join(t.TempDir(), "assessment.json")
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		t.Fatalf("write assessment fixture: %v", err)
	}
	return path
}

func issuesContain(issues []string, want string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, want) {
			return true
		}
	}
	return false
}
