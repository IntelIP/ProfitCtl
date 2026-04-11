package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/IntelIP/ProfitCtl/internal/output"
	"github.com/spf13/cobra"
)

var (
	compareJSONOutput     bool
	compareMarkdownOutput bool
)

var compareCmd = &cobra.Command{
	Use:   "compare <config-a.yml> <config-b.yml> [config-c.yml...]",
	Short: "Compare multiple profitability scenarios side by side",
	Long:  "Run multiple scenarios and compare revenue, fees, costs, margin, and covenant outcomes side by side.",
	RunE:  runCompare,
}

func init() {
	compareCmd.Flags().BoolVar(&compareJSONOutput, "json", false, "Output comparison as JSON")
	compareCmd.Flags().BoolVar(&compareMarkdownOutput, "markdown", false, "Output comparison as markdown")
}

func runCompare(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return wrapExit(2, fmt.Errorf("compare requires at least two config files"))
	}

	if compareJSONOutput && compareMarkdownOutput {
		return wrapExit(2, fmt.Errorf("--json and --markdown cannot be used together"))
	}

	items := make([]output.ComparisonItem, 0, len(args))
	anyFailed := false
	for _, cfgFile := range args {
		result, err := buildSimulationResult(cfgFile)
		if err != nil {
			return wrapExit(2, err)
		}
		if !result.Covenants.Passed {
			anyFailed = true
		}

		items = append(items, output.ComparisonItem{
			Name:   scenarioNameFromPath(cfgFile),
			File:   cfgFile,
			Result: result,
		})
	}

	comparison := output.BuildComparisonResult(items)

	switch {
	case compareJSONOutput:
		if err := output.PrintJSONComparisonResult(comparison); err != nil {
			return wrapExit(3, fmt.Errorf("failed to output JSON: %w", err))
		}
	case compareMarkdownOutput:
		output.PrintMarkdownComparisonResult(comparison)
	default:
		output.PrintCLIComparisonResult(comparison)
	}

	if anyFailed {
		return wrapExitSilent(1, fmt.Errorf("one or more compared scenarios failed covenant validation"))
	}

	return nil
}

func scenarioNameFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
