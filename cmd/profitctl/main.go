package main

import (
	"fmt"
	"os"

	"github.com/IntelIP/ProfitCtl/cmd"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		if cmd.ShouldPrintError(err) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(cmd.ExitCode(err))
	}
}
