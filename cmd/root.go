package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var buildVersion = "dev"

var rootCmd = &cobra.Command{
	Use:           "profitctl",
	Short:         "Profit-first unit economics as code",
	Long:          "profitctl simulates profitability, validates covenant constraints, and detects cost-related services from repository configuration.",
	Version:       buildVersion,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion supplies the build-time version used by --version, version, and doctor.
func SetVersion(version string) {
	version = strings.TrimSpace(version)
	if version == "" {
		version = "dev"
	}
	buildVersion = version
	rootCmd.Version = version
}

func init() {
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	AddCommands(rootCmd)
}
