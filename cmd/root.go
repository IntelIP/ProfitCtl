package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "profitctl",
	Short: "Profit-first unit economics as code",
	Long: `profitctl is a CLI tool that helps developers simulate, stress-test, 
and enforce profitability for software products.`,
}

// Execute adds all child commands to the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func init() {
	AddCommands(rootCmd)
}
