package cmd

import (
 "bytes"
 "context"
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

// Diff compares two observations without inventing driver evidence. Use Compare
// with a variancev1.Input when driver pairs, exchange rates or receipts exist.
func Diff(ctx context.Context, forecast, actual *costv1.CostObservation) (*variancev1.Result, error) {
 if err := ctx.Err(); err != nil { return nil, err }
 if forecast == nil { return nil, errors.New("forecast observation is required") }
 in := variancev1.Input{Forecast: *forecast, Actual: actual}
 if actual == nil { in.UnavailableReason = "actual observation not supplied" }
 result, err := variancev1.Compare(in)
 if err != nil { return nil, err }
 return &result, nil
}

func newCostDiffCommand() *cobra.Command {
 command := &cobra.Command{Use: "diff", Short: "Compare local forecast and actual costs", Args: cobra.NoArgs}
 command.Flags().StringP("input", "i", "", "Variance input JSON (observations, driver pairs, exchange rates and receipts)")
 command.Flags().String("forecast", "", "Forecast CostObservation JSON (without driver pairs)")
 command.Flags().String("actual", "", "Actual CostObservation JSON (without driver pairs)")
 command.Flags().StringP("output", "o", "", "Optional JSON output file (also prints to stdout)")
 command.RunE = func(cmd *cobra.Command, _ []string) error {
  input, _ := cmd.Flags().GetString("input")
  forecastPath, _ := cmd.Flags().GetString("forecast")
  actualPath, _ := cmd.Flags().GetString("actual")
  output, _ := cmd.Flags().GetString("output")
  if (input == "" && (forecastPath == "" || actualPath == "")) || (input != "" && (forecastPath != "" || actualPath != "")) {
   return wrapExit(2, errors.New("provide --input or both --forecast and --actual"))
  }
  if err := rejectVarianceOutputAlias(output, input, forecastPath, actualPath); err != nil { return wrapExit(2, err) }
  var in variancev1.Input
  if input != "" {
   if err := readVarianceJSON(input, &in); err != nil { return wrapExit(2, fmt.Errorf("input: %w", err)) }
  } else {
   if err := readVarianceJSON(forecastPath, &in.Forecast); err != nil { return wrapExit(2, fmt.Errorf("forecast: %w", err)) }
   var actual costv1.CostObservation
   if err := readVarianceJSON(actualPath, &actual); err != nil { return wrapExit(2, fmt.Errorf("actual: %w", err)) }
   in.Actual = &actual
  }
  if err := cmd.Context().Err(); err != nil { return wrapExit(2, err) }
  report, err := variancev1.Compare(in)
  if err != nil { return wrapExit(2, fmt.Errorf("diff: %w", err)) }
  return emitVarianceJSON(cmd, output, report)
 }
 return command
}

func readVarianceJSON(path string, target any) error {
 data, err := os.ReadFile(path)
 if err != nil { return err }
 return strictVarianceJSON(data, target)
}

func strictVarianceJSON(data []byte, target any) error {
 decoder := json.NewDecoder(bytes.NewReader(data))
 decoder.DisallowUnknownFields()
 if err := decoder.Decode(target); err != nil { return err }
 if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
  if err == nil { return errors.New("multiple JSON values are not allowed") }
  return err
 }
 return nil
}

func rejectVarianceOutputAlias(output string, inputs ...string) error {
 if strings.TrimSpace(output) == "" { return nil }
 for _, input := range inputs {
  if input == "" { continue }
  alias, err := pathsAlias(output, input)
  if err != nil { return fmt.Errorf("compare output paths: %w", err) }
  if alias { return fmt.Errorf("--output must not reference input %q", input) }
 }
 return nil
}

func emitVarianceJSON(cmd *cobra.Command, output string, report any) error {
 payload, err := json.MarshalIndent(report, "", "  ")
 if err != nil { return wrapExit(3, fmt.Errorf("encode variance report: %w", err)) }
 payload = append(payload, '\n')
 if strings.TrimSpace(output) != "" {
  if err := os.WriteFile(output, payload, 0600); err != nil { return wrapExit(3, fmt.Errorf("write output: %w", err)) }
 }
 if _, err := cmd.OutOrStdout().Write(payload); err != nil { return wrapExit(3, fmt.Errorf("write stdout: %w", err)) }
 return nil
}
