package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:           "profitctl",
	Short:         "Profit-first unit economics as code",
	Long:          "profitctl simulates profitability, validates covenant constraints, and detects cost-related services from repository configuration.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	AddCommands(rootCmd)
}
