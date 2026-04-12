package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	calibrationInput  string
	calibrationOutput string
	calibrationJSON   bool
)

var calibrateCmd = &cobra.Command{
	Use:   "calibrate",
	Short: "Normalize a calibration artifact into ProfitCtl's calibration schema",
	Long:  "Read a YAML, JSON, or CSV calibration artifact and emit a normalized calibration document that can be referenced via calibration_file.",
	RunE:  runCalibrate,
}

func init() {
	calibrateCmd.Flags().StringVarP(&calibrationInput, "input", "i", "", "Calibration input file (yaml, json, or csv)")
	calibrateCmd.Flags().StringVarP(&calibrationOutput, "out", "o", "", "Write normalized calibration to a file instead of stdout")
	calibrateCmd.Flags().BoolVar(&calibrationJSON, "json", false, "Output normalized calibration as JSON")
}

func runCalibrate(cmd *cobra.Command, args []string) error {
	if calibrationInput == "" {
		return wrapExit(2, fmt.Errorf("--input is required"))
	}

	calibration, err := config.LoadCalibrationFile(calibrationInput)
	if err != nil {
		return wrapExit(2, err)
	}

	var payload []byte
	if calibrationJSON {
		payload, err = json.MarshalIndent(calibration, "", "  ")
		if err != nil {
			return wrapExit(3, fmt.Errorf("failed to format calibration JSON: %w", err))
		}
	} else {
		payload, err = yaml.Marshal(calibration)
		if err != nil {
			return wrapExit(3, fmt.Errorf("failed to format calibration YAML: %w", err))
		}
	}

	if calibrationOutput == "" {
		fmt.Print(string(payload))
		if len(payload) == 0 || payload[len(payload)-1] != '\n' {
			fmt.Println()
		}
		return nil
	}

	if err := os.WriteFile(calibrationOutput, payload, 0644); err != nil {
		return wrapExit(3, fmt.Errorf("failed to write calibration output: %w", err))
	}

	fmt.Printf("Wrote %s\n", calibrationOutput)
	return nil
}
