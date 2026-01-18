package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// This will be called to initialize configuration
	// For now, just set up the config file flag
}

func AddCommands(rootCmd *cobra.Command) {
	var cfgFile string

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "file", "f", "profit.yml", "Config file (default \"./profit.yml\")")

	// init command
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize profitctl configuration from template",
		Long: `Initialize a new profit.yml configuration file from a stack template.
Supported stacks: go, node, python, rust`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Init command not yet implemented - coming in Task 15a")
		},
	}

	// simulate command is defined in cmd/simulate.go

	// validate command
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate profit configuration",
		Long:  `Validate profit.yml configuration file for syntax errors and covenant rules.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Validate command not yet implemented - coming in Task 17")
		},
	}

	rootCmd.AddCommand(initCmd, simulateCmd, validateCmd)
}
