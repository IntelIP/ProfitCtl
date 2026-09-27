package cmd

import (
 "context"
 "errors"
 "fmt"
 "math"
 "strings"

 costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
 variancev1 "github.com/IntelIP/ProfitCtl/pkg/variance/v1"
 "github.com/spf13/cobra"
)

// DriverContribution preserves the modeled decomposition and adds non-causal
// suggestions. Receipts and source references remain on the embedded contribution.
type DriverContribution struct {
 variancev1.Contribution
 RemedialActions []string `json:"remedial_actions"`
}

type ExplainReport struct {
 SchemaVersion string `json:"schema_version"`
 Status string `json:"status"`
 Window costv1.TimeWindow `json:"window"`
 ForecastObservationID string `json:"forecast_observation_id"`
 ActualObservationID string `json:"actual_observation_id,omitempty"`
 UnavailableReason string `json:"unavailable_reason,omitempty"`
 Confidence costv1.Confidence `json:"confidence"`
 ConfidenceScore float64 `json:"confidence_score"`
 Forecast costv1.Money `json:"forecast"`
 Actual *costv1.Money `json:"actual"`
 VarianceAbsolute *costv1.Money `json:"variance_absolute"`
 VariancePercent *float64 `json:"variance_percent"`
 DriverContributions []DriverContribution `json:"driver_contributions"`
 Residual *costv1.Money `json:"residual"`
 MissingEvidence []string `json:"missing_evidence"`
 Sources []costv1.SourceReference `json:"sources"`
}

// Explain ranks existing modeled evidence; it never asserts delivery causation.
func Explain(ctx context.Context, report *variancev1.Result) (*ExplainReport, error) {
 if err := ctx.Err(); err != nil { return nil, err }
 if report == nil || report.SchemaVersion != variancev1.SchemaVersion || (report.Status != "available" && report.Status != "unavailable") { return nil, errors.New("valid cost variance result is required") }
 if report.Status == "available" && (report.Actual == nil || report.AbsoluteVariance == nil || report.Residual == nil) { return nil, errors.New("available result lacks actual, variance or residual") }
 if report.Status == "available" {
  for _, field := range []struct { name string; money *costv1.Money }{{"actual", report.Actual}, {"absolute_variance", report.AbsoluteVariance}, {"residual", report.Residual}} {
   if field.money.Currency != report.Forecast.Currency { return nil, fmt.Errorf("%s currency %q differs from forecast currency %q", field.name, field.money.Currency, report.Forecast.Currency) }
  }
 }
 if report.Status == "available" && (math.IsNaN(report.Residual.Amount) || math.IsInf(report.Residual.Amount, 0) || math.IsNaN(report.AbsoluteVariance.Amount) || math.IsInf(report.AbsoluteVariance.Amount, 0)) { return nil, errors.New("non-finite variance or residual") }
 if report.Status == "unavailable" && (report.Actual != nil || report.AbsoluteVariance != nil || report.Residual != nil || len(report.Contributions) != 0) { return nil, errors.New("unavailable result cannot contain measured variance") }
 out := &ExplainReport{SchemaVersion: report.SchemaVersion, Status: report.Status, Window: report.Window, ForecastObservationID: report.ForecastObservationID, ActualObservationID: report.ActualObservationID, UnavailableReason: report.UnavailableReason, Forecast: report.Forecast, Actual: report.Actual, VarianceAbsolute: report.AbsoluteVariance, VariancePercent: report.PercentageVariance, Residual: report.Residual, DriverContributions: []DriverContribution{}, MissingEvidence: report.MissingEvidence, Sources: report.Sources, Confidence: costv1.ConfidenceLow, ConfidenceScore: 0}
 if out.MissingEvidence == nil { out.MissingEvidence = []string{} }
 if report.Status == "available" {
  out.ConfidenceScore = 1
  // Result carries total-cost sources, not the observations' claim evidence.
  // Provenance can limit confidence but cannot prove a high-confidence total.
  for i, name := range []string{"forecast", "actual"} {
   score := .35
   if i < len(report.Sources) {
    switch report.Sources[i].Type {
    case costv1.SourceInvoice, costv1.SourceProfitCtlDerived: score = .7
    }
   }
   out.MissingEvidence = append(out.MissingEvidence, name+" total cost: high-confidence claim evidence unavailable")
   out.ConfidenceScore = math.Min(out.ConfidenceScore, score)
  }
  sum := report.Residual.Amount
  for _, c := range report.Contributions {
   if c.Amount.Currency != report.Forecast.Currency || math.IsNaN(c.Amount.Amount) || math.IsInf(c.Amount.Amount, 0) || c.ConfidenceScore < 0 || c.ConfidenceScore > 1 || math.IsNaN(c.ConfidenceScore) || (c.Attribution != "unknown" && c.Attribution != "modeled") { return nil, errors.New("invalid driver contribution") }
   sum += c.Amount.Amount
   out.ConfidenceScore = math.Min(out.ConfidenceScore, c.ConfidenceScore)
   action := "Review workload measurements and cost evidence before adjusting this driver"
   switch c.Kind {
   case costv1.DriverCadence: action = "Review polling interval and measure command volume before changing cadence"
   case costv1.DriverConcurrency: action = "Review replica count and utilization before resizing"
   case costv1.DriverVariable: action = "Review request volume and per-unit price before tuning usage"
   case costv1.DriverUptime: action = "Review measured uptime and scheduling before changing runtime"
   case costv1.DriverFixed: action = "Review contracted unit price before renegotiating"
   }
   out.DriverContributions = append(out.DriverContributions, DriverContribution{Contribution: c, RemedialActions: []string{action}})
  }
  if math.Abs(sum-report.AbsoluteVariance.Amount) > variancev1.Rounding { return nil, errors.New("driver contributions and residual do not reconcile") }
  if math.Abs(report.Residual.Amount) > variancev1.Rounding || len(report.Contributions) == 0 { out.ConfidenceScore = math.Min(out.ConfidenceScore, .35) }
  switch { case out.ConfidenceScore >= .85: out.Confidence = costv1.ConfidenceHigh; case out.ConfidenceScore >= .55: out.Confidence = costv1.ConfidenceMedium; default: out.Confidence = costv1.ConfidenceLow }
 }
 return out, nil
}

var costExplainCmd = newCostExplainCommand()

func init() { rootCmd.AddCommand(costExplainCmd) }

func newCostExplainCommand() *cobra.Command {
 command := &cobra.Command{Use: "explain", Short: "Rank evidenced cost variance drivers", Args: cobra.NoArgs}
 command.Flags().StringP("input", "i", "", "Diff result JSON file")
 command.Flags().StringP("output", "o", "", "Optional JSON output file (also prints to stdout)")
 command.RunE = func(cmd *cobra.Command, _ []string) error {
  input, _ := cmd.Flags().GetString("input")
  output, _ := cmd.Flags().GetString("output")
  if strings.TrimSpace(input) == "" { return wrapExit(2, errors.New("--input is required")) }
  if err := rejectVarianceOutputAlias(output, input); err != nil { return wrapExit(2, err) }
  var diff variancev1.Result
  if err := readVarianceJSON(input, &diff); err != nil { return wrapExit(2, fmt.Errorf("decode diff result: %w", err)) }
  report, err := Explain(cmd.Context(), &diff)
  if err != nil { return wrapExit(2, fmt.Errorf("explain: %w", err)) }
  return emitVarianceJSON(cmd, output, report)
 }
 return command
}
