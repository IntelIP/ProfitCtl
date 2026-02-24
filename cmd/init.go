package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultConfigTemplate = `project:
  name: Example SaaS Product

fixed_costs:
  - name: Base Infrastructure
    amount: 500
    period: monthly
    layer: infrastructure

variable_costs:
  - name: API Calls
    cost_per_unit: 0.0001
    units_per_user: 10000
    distribution: normal
    mean: 10000
    stddev: 2000
    layer: application

pricing:
  plans:
    - name: basic
      price: 10
      limits:
        users: 1000

covenants:
  - type: threshold
    field: margin
    operator: gte
    value: 20
    message: Gross margin must be >= 20%

simulation:
  base_users: 100
  growth_factor: 1.5
  iterations: 10000
`

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a starter profit.yml configuration",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing file")
}

func runInit(cmd *cobra.Command, args []string) error {
	cfgFile, _ := cmd.Flags().GetString("file")
	if cfgFile == "" {
		cfgFile = "profit.yml"
	}

	if !initForce {
		if _, err := os.Stat(cfgFile); err == nil {
			return wrapExit(2, fmt.Errorf("config file already exists: %s (use --force to overwrite)", cfgFile))
		}
	}

	if err := os.WriteFile(cfgFile, []byte(defaultConfigTemplate), 0644); err != nil {
		return wrapExit(3, fmt.Errorf("failed to write starter config: %w", err))
	}

	fmt.Printf("Created %s\n", cfgFile)
	return nil
}
