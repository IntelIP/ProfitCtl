package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
	variancev1 "github.com/IntelIP/ProfitCtl/pkg/variance/v1"
	"github.com/spf13/cobra"
)

var costDiffCmd = newCostDiffCommand()

func init() { rootCmd.AddCommand(costDiffCmd) }

func newCostDiffCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "cost_diff", Short: "Compare forecast and actual cost observations as JSON",
		Args: cobra.NoArgs,
	}
	command.Flags().String("forecast", "", "Forecast CostObservation JSON file")
	command.Flags().String("actual", "", "Actual CostObservation JSON file")
	command.Flags().StringP("output", "o", "", "Optional JSON output file (also writes to stdout)")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		forecastPath, _ := cmd.Flags().GetString("forecast")
		actualPath, _ := cmd.Flags().GetString("actual")
		output, _ := cmd.Flags().GetString("output")
		if strings.TrimSpace(forecastPath) == "" || strings.TrimSpace(actualPath) == "" {
			return wrapExit(2, errors.New("--forecast and --actual are required"))
		}
		if err := rejectVarianceOutputAlias(output, forecastPath, actualPath); err != nil {
			return wrapExit(2, err)
		}
		forecast, err := readCostObservations(forecastPath)
		if err != nil {
			return wrapExit(2, fmt.Errorf("forecast: %w", err))
		}
		for i, observation := range forecast {
			if observation.TotalCost.Amount == 0 {
				return wrapExit(2, fmt.Errorf("forecast observation[%d]: zero total cost denominator for percent variance", i))
			}
		}
		actual, err := readCostObservations(actualPath)
		if err != nil {
			return wrapExit(2, fmt.Errorf("actual: %w", err))
		}
		report, err := variancev1.Diff(forecast, actual)
		if err != nil {
			return wrapExit(2, fmt.Errorf("diff: %w", err))
		}
		return emitVarianceJSON(cmd, output, report)
	}
	return command
}

// A single observation is convenient for one-period comparisons; an array
// allows the same command to compare multiple windows without dropping data.
func readCostObservations(path string) ([]costv1.CostObservation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var observations []costv1.CostObservation
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, errors.New("empty JSON document")
	}
	if bytes.TrimSpace(data)[0] == '[' {
		if err := strictVarianceJSON(data, &observations); err != nil {
			return nil, err
		}
	} else {
		var observation costv1.CostObservation
		if err := strictVarianceJSON(data, &observation); err != nil {
			return nil, err
		}
		observations = []costv1.CostObservation{observation}
	}
	if len(observations) == 0 {
		return nil, errors.New("at least one observation is required")
	}
	for i, observation := range observations {
		if err := observation.Validate(); err != nil {
			return nil, fmt.Errorf("observation[%d]: %w", i, err)
		}
	}
	return observations, nil
}

func strictVarianceJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}

func rejectVarianceOutputAlias(output string, inputs ...string) error {
	if strings.TrimSpace(output) == "" {
		return nil
	}
	for _, input := range inputs {
		alias, err := pathsAlias(output, input)
		if err != nil {
			return fmt.Errorf("compare output paths: %w", err)
		}
		if alias {
			return fmt.Errorf("--output must not reference input %q", input)
		}
	}
	return nil
}

func emitVarianceJSON(cmd *cobra.Command, output string, report any) error {
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return wrapExit(3, fmt.Errorf("encode variance report: %w", err))
	}
	payload = append(payload, '\n')
	if strings.TrimSpace(output) != "" {
		if err := os.WriteFile(output, payload, 0600); err != nil {
			return wrapExit(3, fmt.Errorf("write output: %w", err))
		}
	}
	if _, err := cmd.OutOrStdout().Write(payload); err != nil {
		return wrapExit(3, fmt.Errorf("write stdout: %w", err))
	}
	return nil
}
