package cmd

import (
	"fmt"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  "Parse and validate profitctl configuration rules without running a simulation.",
	RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfgFile, _ := cmd.Flags().GetString("file")
	if _, err := config.ParseConfig(cfgFile); err != nil {
		return wrapExit(2, fmt.Errorf("invalid configuration: %w", err))
	}

	fmt.Printf("Configuration is valid: %s\n", cfgFile)
	return nil
}
