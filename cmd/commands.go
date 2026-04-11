package cmd

import "github.com/spf13/cobra"

func AddCommands(root *cobra.Command) {
	var cfgFile string
	root.PersistentFlags().StringVarP(&cfgFile, "file", "f", "profit.yml", "Configuration file path")
	root.AddCommand(initCmd, simulateCmd, compareCmd, validateCmd, detectCmd)
}
