package cmd

import (
	"fmt"
	"os"
	"strings"

	variancev1 "github.com/IntelIP/ProfitCtl/pkg/variance/v1"
	"github.com/spf13/cobra"
)

var costExplainCmd = newCostExplainCommand()

func init() { rootCmd.AddCommand(costExplainCmd) }

func newCostExplainCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "cost_explain", Short: "Rank evidenced drivers of a cost difference as JSON",
		Args: cobra.NoArgs,
	}
	command.Flags().StringP("input", "i", "", "Diff report JSON file")
	command.Flags().StringP("output", "o", "", "Optional JSON output file (also writes to stdout)")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		if strings.TrimSpace(input) == "" {
			return wrapExit(2, fmt.Errorf("--input is required"))
		}
		if err := rejectVarianceOutputAlias(output, input); err != nil {
			return wrapExit(2, err)
		}
		data, err := os.ReadFile(input)
		if err != nil {
			return wrapExit(2, fmt.Errorf("read diff report: %w", err))
		}
		var diff variancev1.DiffReport
		if err := strictVarianceJSON(data, &diff); err != nil {
			return wrapExit(2, fmt.Errorf("decode diff report: %w", err))
		}
		report, err := variancev1.Explain(diff)
		if err != nil {
			return wrapExit(2, fmt.Errorf("explain: %w", err))
		}
		return emitVarianceJSON(cmd, output, report)
	}
	return command
}
