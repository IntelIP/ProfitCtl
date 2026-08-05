package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/IntelIP/ProfitCtl/internal/scanner"
	"github.com/IntelIP/ProfitCtl/internal/scanner/llm"
	"github.com/spf13/cobra"
)

const (
	defaultDetectPath     = "."
	defaultDetectProvider = "openrouter"
	defaultDetectModel    = llm.DefaultOpenRouterModel
)

var (
	detectPath     string
	detectOut      string
	detectProvider string
	detectModel    string

	newCollector = scanner.NewCollector
	newProvider  = func(apiKey, model string) (llm.LLMProvider, error) {
		return llm.NewOpenRouterProvider(apiKey, model)
	}
)

type detectReport struct {
	Path         string               `json:"path"`
	Provider     string               `json:"provider"`
	Model        string               `json:"model"`
	AnalyzedFile int                  `json:"analyzed_files"`
	Analysis     llm.AnalysisResponse `json:"analysis"`
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect services and dependencies from repository config files",
	Long:  "Collect configuration files recursively, analyze them with an LLM provider, and emit a JSON report.",
	RunE:  runDetect,
}

func init() {
	detectCmd.Flags().StringVar(&detectPath, "path", defaultDetectPath, "Directory to scan recursively")
	detectCmd.Flags().StringVar(&detectOut, "out", "", "Write JSON report to file instead of stdout")
	detectCmd.Flags().StringVar(&detectProvider, "provider", defaultDetectProvider, "LLM provider (openrouter)")
	detectCmd.Flags().StringVar(&detectModel, "model", defaultDetectModel, "Model id used by the provider")
}

func runDetect(cmd *cobra.Command, args []string) error {
	provider := strings.ToLower(strings.TrimSpace(detectProvider))
	if provider != defaultDetectProvider {
		return wrapExit(2, fmt.Errorf("unsupported provider %q (supported: openrouter)", detectProvider))
	}

	apiKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if apiKey == "" {
		return wrapExit(2, errors.New("OPENROUTER_API_KEY is required for detect"))
	}

	collector := newCollector()
	files, err := collector.Collect(detectPath)
	if err != nil {
		return wrapExit(2, fmt.Errorf("failed to collect files from %q: %w", detectPath, err))
	}
	if len(files) == 0 {
		return wrapExit(2, fmt.Errorf("no supported configuration files found under %q", detectPath))
	}

	prompt := llm.BuildDetectionPrompt(llm.CodeContext{Files: files})

	llmProvider, err := newProvider(apiKey, detectModel)
	if err != nil {
		return wrapExit(3, fmt.Errorf("failed to initialize provider: %w", err))
	}

	rawResponse, err := llmProvider.Chat(cmd.Context(), []llm.Message{{Role: "user", Content: prompt}})
	if err != nil {
		return wrapExit(3, fmt.Errorf("provider request failed: %w", err))
	}

	analysis, err := llm.ParseAnalysisResponse(rawResponse)
	if err != nil {
		return wrapExit(3, fmt.Errorf("invalid provider response: %w", err))
	}

	report := detectReport{
		Path:         detectPath,
		Provider:     provider,
		Model:        detectModel,
		AnalyzedFile: len(files),
		Analysis:     analysis,
	}

	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return wrapExit(3, fmt.Errorf("failed to marshal detect report: %w", err))
	}

	if detectOut != "" {
		if err := os.WriteFile(detectOut, append(jsonBytes, '\n'), 0600); err != nil {
			return wrapExit(3, fmt.Errorf("failed to write detect report to %s: %w", detectOut, err))
		}
		return nil
	}

	fmt.Println(string(jsonBytes))
	return nil
}
