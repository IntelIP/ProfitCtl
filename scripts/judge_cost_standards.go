package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/IntelIP/ProfitCtl/internal/config"
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
		paths = []string{"skills/profitctl-cost-aware/references/templates"}
	}

	files, err := collectScenarioFiles(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect scenario files: %v\n", err)
		os.Exit(2)
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no scenario files found")
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

func collectScenarioFiles(paths []string) ([]string, error) {
	var files []string
	for _, raw := range paths {
		path := filepath.Clean(raw)
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if isYAML(path) {
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

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yml" || ext == ".yaml"
}

func judgeFile(file string) FileReport {
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
