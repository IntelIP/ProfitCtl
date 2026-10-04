package v1

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
)

// DriverContribution is an accounting decomposition, not a causal delivery claim.
type DriverContribution struct {
	Driver          string            `json:"driver"`
	Amount          costv1.Money      `json:"amount"`
	Confidence      string            `json:"confidence"`
	ConfidenceScore float64           `json:"confidence_score"`
	Attribution     string            `json:"attribution"`
	Evidence        []costv1.Evidence `json:"evidence"`
	MissingEvidence []string          `json:"missing_evidence"`
	RemedialActions []string          `json:"remedial_actions"`
}

type DiffPeriod struct {
	Window              costv1.TimeWindow      `json:"window"`
	Forecast            costv1.CostObservation `json:"forecast"`
	Actual              costv1.CostObservation `json:"actual"`
	VarianceAbsolute    costv1.Money           `json:"variance_absolute"`
	VariancePercent     *float64               `json:"variance_percent"`
	DriverContributions []DriverContribution   `json:"driver_contributions"`
	Residual            costv1.Money           `json:"residual"`
	Confidence          string                 `json:"confidence"`
}

// DiffReport retains the original observations so that explain never needs to
// reconstruct provenance from totals. Percent uses the sum of forecasts.
type DiffReport struct {
	SchemaVersion       string               `json:"schema_version"`
	Forecast            costv1.Money         `json:"forecast"`
	Actual              costv1.Money         `json:"actual"`
	VarianceAbsolute    costv1.Money         `json:"variance_absolute"`
	VariancePercent     *float64             `json:"variance_percent"`
	DriverContributions []DriverContribution `json:"driver_contributions"`
	Residual            costv1.Money         `json:"residual"`
	Confidence          string               `json:"confidence"`
	Periods             []DiffPeriod         `json:"periods"`
}

type ExplainReport struct {
	SchemaVersion       string               `json:"schema_version"`
	Forecast            costv1.Money         `json:"forecast"`
	Actual              costv1.Money         `json:"actual"`
	VarianceAbsolute    costv1.Money         `json:"variance_absolute"`
	VariancePercent     *float64             `json:"variance_percent"`
	DriverContributions []DriverContribution `json:"driver_contributions"`
	Residual            costv1.Money         `json:"residual"`
	Confidence          string               `json:"confidence"`
	Periods             []DiffPeriod         `json:"periods"`
}

// Diff pairs observations by window and full workload dimensions, rather than
// pairing by list position. An unmatched or ambiguous period is an error.
func Diff(forecast, actual []costv1.CostObservation) (*DiffReport, error) {
	if len(forecast) == 0 || len(forecast) != len(actual) {
		return nil, errors.New("forecast and actual require the same nonzero number of observations")
	}
	result := &DiffReport{SchemaVersion: SchemaVersion, DriverContributions: []DriverContribution{}, Periods: []DiffPeriod{}}
	used := make([]bool, len(actual))
	for i, f := range forecast {
		if err := f.Validate(); err != nil {
			return nil, fmt.Errorf("forecast[%d]: %w", i, err)
		}
		match := -1
		for j, a := range actual {
			if err := a.Validate(); err != nil {
				return nil, fmt.Errorf("actual[%d]: %w", j, err)
			}
			if f.Window == a.Window && f.Dimensions == a.Dimensions {
				if match >= 0 {
					return nil, fmt.Errorf("ambiguous actual for forecast[%d]", i)
				}
				match = j
			}
		}
		if match < 0 || used[match] {
			return nil, fmt.Errorf("unmatched forecast[%d] window or dimensions", i)
		}
		used[match] = true
		a := actual[match]
		if f.TotalCost.Currency != a.TotalCost.Currency || f.Quantity.Unit != a.Quantity.Unit || f.UnitPrice.Per.Unit != a.UnitPrice.Per.Unit {
			return nil, fmt.Errorf("forecast[%d]: currency or quantity units differ", i)
		}
		delta := a.TotalCost.Amount - f.TotalCost.Amount
		fq, aq := f.Quantity.Value, a.Quantity.Value
		fp, ap := f.UnitPrice.Amount.Amount/f.UnitPrice.Per.Value, a.UnitPrice.Amount.Amount/a.UnitPrice.Per.Value
		quantity := (aq - fq) * (fp + ap) / 2
		price := (ap - fp) * (fq + aq) / 2
		residual := delta - quantity - price
		if !finite(delta) || !finite(quantity) || !finite(price) || !finite(residual) {
			return nil, fmt.Errorf("forecast[%d]: variance overflow", i)
		}
		currency := f.TotalCost.Currency
		contributions := []DriverContribution{
			makeContribution("usage_volume", quantity, currency, f.Evidence.Quantity, a.Evidence.Quantity),
			makeContribution("price_change", price, currency, f.Evidence.UnitPrice, a.Evidence.UnitPrice),
		}
		confidence := "high"
		if math.Abs(residual) > Rounding {
			confidence = "unknown"
		}
		for _, c := range contributions {
			if c.Confidence != "high" {
				confidence = "unknown"
			}
		}
		period := DiffPeriod{Window: f.Window, Forecast: f, Actual: a, VarianceAbsolute: costv1.Money{Amount: delta, Currency: currency}, DriverContributions: contributions, Residual: costv1.Money{Amount: residual, Currency: currency}, Confidence: confidence}
		if f.TotalCost.Amount != 0 {
			percentage := delta / f.TotalCost.Amount * 100
			if !finite(percentage) {
				return nil, fmt.Errorf("forecast[%d]: percentage overflow", i)
			}
			period.VariancePercent = &percentage
		}
		result.Periods = append(result.Periods, period)
		result.DriverContributions = append(result.DriverContributions, contributions...)
		result.Forecast.Amount += f.TotalCost.Amount
		result.Actual.Amount += a.TotalCost.Amount
		result.Residual.Amount += residual
		result.Forecast.Currency = currency
		result.Actual.Currency = currency
		result.Residual.Currency = currency
		if result.Confidence == "" {
			result.Confidence = confidence
		} else if confidence != "high" {
			result.Confidence = "unknown"
		}
		if i > 0 && forecast[0].TotalCost.Currency != currency {
			return nil, errors.New("period currencies differ")
		}
	}
	result.VarianceAbsolute = costv1.Money{Amount: result.Actual.Amount - result.Forecast.Amount, Currency: result.Forecast.Currency}
	if !finite(result.Forecast.Amount) || !finite(result.Actual.Amount) || !finite(result.Residual.Amount) || !finite(result.VarianceAbsolute.Amount) {
		return nil, errors.New("aggregate variance overflow")
	}
	if result.Forecast.Amount != 0 {
		percentage := result.VarianceAbsolute.Amount / result.Forecast.Amount * 100
		if !finite(percentage) {
			return nil, errors.New("aggregate percentage overflow")
		}
		result.VariancePercent = &percentage
	}
	sort.Slice(result.Periods, func(i, j int) bool { return result.Periods[i].Window.Start < result.Periods[j].Window.Start })
	sort.SliceStable(result.DriverContributions, func(i, j int) bool {
		x, y := result.DriverContributions[i], result.DriverContributions[j]
		if math.Abs(x.Amount.Amount) != math.Abs(y.Amount.Amount) {
			return math.Abs(x.Amount.Amount) > math.Abs(y.Amount.Amount)
		}
		if x.Driver != y.Driver {
			return x.Driver < y.Driver
		}
		return x.Evidence[0].Source.CapturedAt < y.Evidence[0].Source.CapturedAt
	})
	return result, nil
}

func makeContribution(name string, amount float64, currency string, f, a costv1.Evidence) DriverContribution {
	c := DriverContribution{Driver: name, Amount: costv1.Money{Amount: amount, Currency: currency}, Evidence: []costv1.Evidence{f, a}, Confidence: "low", ConfidenceScore: .35, Attribution: "unknown", MissingEvidence: []string{"exact delivery receipt"}, RemedialActions: []string{}}
	// Evidence of a measured change supports the accounting difference, but not
	// the unsupported assertion that a particular deployment caused it.
	if (a.Measurement == costv1.MeasurementMeasured || a.Kind == costv1.EvidenceBilled) && f.Confidence != costv1.ConfidenceLow && a.Confidence != costv1.ConfidenceLow {
		c.Confidence = "medium"
		c.ConfidenceScore = .7
	}
	return c
}

// Explain checks arithmetic even for untrusted JSON input; it never silently
// converts an unexplained residual into a named causal driver.
func Explain(report DiffReport) (*ExplainReport, error) {
	if report.SchemaVersion != SchemaVersion || len(report.Periods) == 0 {
		return nil, errors.New("invalid diff report")
	}
	recomputedForecast := 0.0
	recomputedActual := 0.0
	recomputedResidual := 0.0
	for i, p := range report.Periods {
		if err := p.Forecast.Validate(); err != nil {
			return nil, fmt.Errorf("period[%d] forecast: %w", i, err)
		}
		if err := p.Actual.Validate(); err != nil {
			return nil, fmt.Errorf("period[%d] actual: %w", i, err)
		}
		if p.Forecast.Window != p.Actual.Window || p.Window != p.Forecast.Window || p.Forecast.Dimensions != p.Actual.Dimensions || p.Forecast.TotalCost.Currency != report.Forecast.Currency || p.Actual.TotalCost.Currency != report.Forecast.Currency {
			return nil, fmt.Errorf("period[%d]: unmatched observation", i)
		}
		sum := p.Residual.Amount
		for _, c := range p.DriverContributions {
			sum += c.Amount.Amount
		}
		if !finite(sum) || math.Abs(sum-(p.Actual.TotalCost.Amount-p.Forecast.TotalCost.Amount)) > Rounding {
			return nil, fmt.Errorf("period[%d]: contributions do not reconcile", i)
		}
		recomputedForecast += p.Forecast.TotalCost.Amount
		recomputedActual += p.Actual.TotalCost.Amount
		recomputedResidual += p.Residual.Amount
	}
	if !finite(recomputedForecast) || !finite(recomputedActual) || !finite(recomputedResidual) || !finite(report.Forecast.Amount) || !finite(report.Actual.Amount) || !finite(report.Residual.Amount) || !finite(report.VarianceAbsolute.Amount) || math.Abs(recomputedForecast-report.Forecast.Amount) > Rounding || math.Abs(recomputedActual-report.Actual.Amount) > Rounding || math.Abs(recomputedResidual-report.Residual.Amount) > Rounding || math.Abs(report.VarianceAbsolute.Amount-(recomputedActual-recomputedForecast)) > Rounding {
		return nil, errors.New("report totals do not reconcile")
	}
	if (recomputedForecast == 0 && report.VariancePercent != nil) || (recomputedForecast != 0 && (report.VariancePercent == nil || !finite(*report.VariancePercent) || math.Abs(*report.VariancePercent-100*(recomputedActual-recomputedForecast)/recomputedForecast) > Rounding)) {
		return nil, errors.New("report totals do not reconcile")
	}
	forecasts := make([]costv1.CostObservation, len(report.Periods))
	actuals := make([]costv1.CostObservation, len(report.Periods))
	for i, p := range report.Periods {
		forecasts[i], actuals[i] = p.Forecast, p.Actual
	}
	canonical, err := Diff(forecasts, actuals)
	if err != nil || !reflect.DeepEqual(*canonical, report) {
		return nil, errors.New("diff report differs from validated observations")
	}
	contributions := append([]DriverContribution{}, report.DriverContributions...)
	for i := range contributions {
		c := &contributions[i]
		if c.Attribution != "unknown" || len(c.MissingEvidence) == 0 || len(c.Evidence) != 2 || c.ConfidenceScore > .7 {
			return nil, errors.New("unsupported causal attribution")
		}
		switch c.Driver {
		case "usage_volume":
			c.RemedialActions = []string{"Inspect workload demand and polling cadence before changing configuration"}
		case "price_change":
			c.RemedialActions = []string{"Verify the billed unit price against the maintained price catalog"}
		default:
			c.RemedialActions = []string{"Investigate the unexplained driver"}
		}
	}
	// Explanations cannot claim high confidence when residual or receipts are missing.
	confidence := "unknown"
	return &ExplainReport{SchemaVersion: SchemaVersion, Forecast: report.Forecast, Actual: report.Actual, VarianceAbsolute: report.VarianceAbsolute, VariancePercent: report.VariancePercent, DriverContributions: contributions, Residual: report.Residual, Confidence: confidence, Periods: report.Periods}, nil
}
