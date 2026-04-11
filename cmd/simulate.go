package cmd

import (
	"fmt"

	"github.com/IntelIP/ProfitCtl/internal/output"
	"github.com/spf13/cobra"
)

var (
	jsonOutput     bool
	markdownOutput bool
	quietOutput    bool
	verboseOutput  bool
)

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Run cost and profit simulations",
	Long: `Run scale simulations, p95/p99 stress testing, and calculate margins.
Requires a valid profit.yml configuration file.`,
	RunE: runSimulate,
}

func init() {
	simulateCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	simulateCmd.Flags().BoolVar(&markdownOutput, "markdown", false, "Output results as markdown")
	simulateCmd.Flags().BoolVarP(&quietOutput, "quiet", "q", false, "Minimal output")
	simulateCmd.Flags().BoolVarP(&verboseOutput, "verbose", "v", false, "Verbose output")
}

func runSimulate(cmd *cobra.Command, args []string) error {
	cfgFile, _ := cmd.Flags().GetString("file")

	if jsonOutput && markdownOutput {
		return wrapExit(2, fmt.Errorf("--json and --markdown cannot be used together"))
	}
	if quietOutput && (jsonOutput || markdownOutput) {
		return wrapExit(2, fmt.Errorf("--quiet cannot be combined with --json or --markdown"))
	}

	outputResult, err := buildSimulationResult(cfgFile)
	if err != nil {
		return wrapExit(2, err)
	}

	switch {
	case jsonOutput:
		if err := output.PrintJSONResult(outputResult); err != nil {
			return wrapExit(3, fmt.Errorf("failed to output JSON: %w", err))
		}
	case markdownOutput:
		output.PrintMarkdownResult(outputResult)
	case quietOutput:
		output.PrintCLIQuietResult(outputResult)
	default:
		output.PrintCLIResult(outputResult, verboseOutput)
	}

	if !outputResult.Covenants.Passed {
		return wrapExitSilent(1, fmt.Errorf("covenant validation failed"))
	}

	return nil
}
