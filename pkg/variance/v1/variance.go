// Package v1 compares validated cost observations without inferring causal delivery claims.
package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	cost "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
)

const SchemaVersion = "profitctl.cost-variance/v1"
const Rounding = 1e-6 // Currency amounts are reconciled before display rounding.

type ExchangeRate struct {
	From   string               `json:"from"`
	To     string               `json:"to"`
	Rate   float64              `json:"rate"` // Units of To per one From; supplied snapshot, never fetched.
	Source cost.SourceReference `json:"source"`
}

type DriverPair struct {
	Forecast cost.CostDriver `json:"forecast"`
	Actual   cost.CostDriver `json:"actual"`
}

// Receipt identifies the exact pair of observations supporting a delivery attribution.
// A source label, commit string or timestamp alone is not an exact receipt.
type Receipt struct {
	DriverID              string               `json:"driver_id"`
	Kind                  string               `json:"kind"` // commit, pull_request, release, or deployment
	ForecastObservationID string               `json:"forecast_observation_id"`
	ActualObservationID   string               `json:"actual_observation_id"`
	Source                cost.SourceReference `json:"source"`
}

type Input struct {
	Forecast          cost.CostObservation  `json:"forecast"`
	Actual            *cost.CostObservation `json:"actual,omitempty"`
	UnavailableReason string                `json:"unavailable_reason,omitempty"`
	Drivers           []DriverPair          `json:"drivers"`
	ExchangeRates     []ExchangeRate        `json:"exchange_rates,omitempty"`
	Receipts          []Receipt             `json:"receipts,omitempty"`
}

type Contribution struct {
	DriverID        string                 `json:"driver_id"`
	Kind            cost.DriverKind        `json:"kind"`
	Quantity        float64                `json:"quantity"`
	UnitPrice       float64                `json:"unit_price"`
	Distribution    float64                `json:"distribution"`
	Amount          cost.Money             `json:"amount"`
	Confidence      cost.Confidence        `json:"confidence"`
	ConfidenceScore float64                `json:"confidence_score"`
	Attribution     string                 `json:"attribution"` // modeled or unknown, never a causal assertion
	MissingEvidence []string               `json:"missing_evidence"`
	Sources         []cost.SourceReference `json:"sources"`
	Receipts        []Receipt              `json:"receipts"`
}

type Result struct {
	SchemaVersion         string                 `json:"schema_version"`
	Status                string                 `json:"status"` // available or unavailable
	Window                cost.TimeWindow        `json:"window"`
	ForecastObservationID string                 `json:"forecast_observation_id"`
	ActualObservationID   string                 `json:"actual_observation_id,omitempty"`
	Forecast              cost.Money             `json:"forecast"`
	Actual                *cost.Money            `json:"actual"`
	AbsoluteVariance      *cost.Money            `json:"absolute_variance"`
	PercentageVariance    *float64               `json:"percentage_variance"`
	Contributions         []Contribution         `json:"contributions"`
	Residual              *cost.Money            `json:"residual"`
	MissingEvidence       []string               `json:"missing_evidence"`
	UnavailableReason     string                 `json:"unavailable_reason,omitempty"`
	RoundingTolerance     float64                `json:"rounding_tolerance"`
	Sources               []cost.SourceReference `json:"sources"`
}

// Compare solves the multiplicative quantity/price/distribution cost equation
// via symmetric (Shapley) linear contributions. Residual always includes any
// ledger amount the supplied drivers cannot explain. No live exchange rates are used.
func Compare(in Input) (Result, error) {
	if err := in.Forecast.Validate(); err != nil {
		return Result{}, fmt.Errorf("forecast: %w", err)
	}
	result := Result{SchemaVersion: SchemaVersion, Status: "unavailable", Window: in.Forecast.Window, ForecastObservationID: in.Forecast.ID, Forecast: in.Forecast.TotalCost, RoundingTolerance: Rounding, Contributions: []Contribution{}, MissingEvidence: []string{}, Sources: []cost.SourceReference{in.Forecast.Evidence.TotalCost.Source}}
	if in.Actual == nil {
		if in.UnavailableReason == "" {
			return Result{}, errors.New("actual is unavailable: unavailable_reason is required")
		}
		if len(in.Drivers) > 0 || len(in.Receipts) > 0 {
			return Result{}, errors.New("unavailable actual cannot have driver pairs or receipts")
		}
		result.UnavailableReason = in.UnavailableReason
		result.MissingEvidence = append(result.MissingEvidence, "actual cost: "+in.UnavailableReason)
		return result, nil
	}
	actual := *in.Actual
	if err := actual.Validate(); err != nil {
		return Result{}, fmt.Errorf("actual: %w", err)
	}
	if in.UnavailableReason != "" {
		return Result{}, errors.New("unavailable_reason must be empty when actual is present")
	}
	if in.Forecast.Window != actual.Window || in.Forecast.Dimensions.Workload != actual.Dimensions.Workload {
		return Result{}, errors.New("forecast and actual must have identical windows and workloads")
	}
	convert, err := converter(in.Forecast.TotalCost.Currency, in.ExchangeRates)
	if err != nil {
		return Result{}, err
	}
	actualTotal, err := convert(actual.TotalCost)
	if err != nil {
		return Result{}, fmt.Errorf("actual total: %w", err)
	}
	delta := actualTotal - in.Forecast.TotalCost.Amount
	if !finite(delta) {
		return Result{}, errors.New("variance overflow")
	}
	result.Status = "available"
	result.ActualObservationID = actual.ID
	result.Actual = &cost.Money{Amount: actualTotal, Currency: result.Forecast.Currency}
	result.AbsoluteVariance = &cost.Money{Amount: delta, Currency: result.Forecast.Currency}
	result.Residual = &cost.Money{Amount: delta, Currency: result.Forecast.Currency}
	result.Sources = append(result.Sources, actual.Evidence.TotalCost.Source)
	for _, rate := range in.ExchangeRates {
		if rate.To == result.Forecast.Currency {
			result.Sources = append(result.Sources, rate.Source)
		}
	}
	if in.Forecast.TotalCost.Amount != 0 {
		p := 100 * delta / in.Forecast.TotalCost.Amount
		if !finite(p) {
			return Result{}, errors.New("percentage overflow")
		}
		result.PercentageVariance = &p
	}
	forecastIDs := make(map[string]bool)
	actualIDs := make(map[string]bool)
	for _, id := range in.Forecast.DriverIDs {
		forecastIDs[id] = true
	}
	for _, id := range actual.DriverIDs {
		actualIDs[id] = true
	}
	seen := make(map[string]bool)
	for _, pair := range in.Drivers {
		f, a := pair.Forecast, pair.Actual
		if err := f.Validate(); err != nil {
			return Result{}, fmt.Errorf("forecast driver %q: %w", f.ID, err)
		}
		if err := a.Validate(); err != nil {
			return Result{}, fmt.Errorf("actual driver %q: %w", a.ID, err)
		}
		if f.ID != a.ID || seen[f.ID] || !forecastIDs[f.ID] || !actualIDs[a.ID] {
			return Result{}, fmt.Errorf("driver pair %q must be unique and referenced by both observations", f.ID)
		}
		seen[f.ID] = true
		if f.Window != result.Window || a.Window != result.Window || f.Kind != a.Kind || f.Dimensions.Workload != in.Forecast.Dimensions.Workload || a.Dimensions.Workload != actual.Dimensions.Workload {
			return Result{}, fmt.Errorf("driver %q window, workload or kind mismatch", f.ID)
		}
		fq, fp, fd, err := factors(f, convert)
		if err != nil {
			return Result{}, fmt.Errorf("forecast driver %q: %w", f.ID, err)
		}
		aq, ap, ad, err := factors(a, convert)
		if err != nil {
			return Result{}, fmt.Errorf("actual driver %q: %w", a.ID, err)
		}
		if f.Quantity.Unit != a.Quantity.Unit && !(timeUnits[f.Quantity.Unit] > 0 && timeUnits[a.Quantity.Unit] > 0) {
			return Result{}, fmt.Errorf("driver %q incompatible quantity units", f.ID)
		}
		// The per-unit currency amount was normalized to one canonical quantity unit.
		parts := shapley([3]float64{fq, fp, fd}, [3]float64{aq, ap, ad})
		amount := parts[0] + parts[1] + parts[2]
		if !finite(amount) {
			return Result{}, fmt.Errorf("driver %q contribution overflow", f.ID)
		}
		c := Contribution{DriverID: f.ID, Kind: f.Kind, Quantity: parts[0], UnitPrice: parts[1], Distribution: parts[2], Amount: cost.Money{Amount: amount, Currency: result.Forecast.Currency}, Attribution: "modeled", MissingEvidence: []string{}, Sources: []cost.SourceReference{f.Evidence.Quantity.Source, f.Evidence.UnitPrice.Source, a.Evidence.Quantity.Source, a.Evidence.UnitPrice.Source}, Receipts: []Receipt{}}
		for _, r := range in.Receipts {
			if r.DriverID == f.ID {
				if r.ForecastObservationID != in.Forecast.ID || r.ActualObservationID != actual.ID || r.Source.ArtifactIdentity == "" || r.Source.Type != cost.SourceRuntimeLedger || !validCapture(r.Source.CapturedAt) || !receiptKind(r.Kind) {
					return Result{}, fmt.Errorf("driver %q has invalid exact receipt", f.ID)
				}
				c.Receipts = append(c.Receipts, r)
			}
		}
		confidence(&c, f, a, delta)
		result.Contributions = append(result.Contributions, c)
		result.Residual.Amount -= amount
	}
	for _, r := range in.Receipts {
		if !seen[r.DriverID] {
			return Result{}, fmt.Errorf("receipt references unknown driver %q", r.DriverID)
		}
	}
	sort.Slice(result.Contributions, func(i, j int) bool {
		x, y := result.Contributions[i], result.Contributions[j]
		if math.Abs(x.Amount.Amount) == math.Abs(y.Amount.Amount) {
			return x.DriverID < y.DriverID
		}
		return math.Abs(x.Amount.Amount) > math.Abs(y.Amount.Amount)
	})
	if math.Abs(result.Residual.Amount) <= Rounding {
		result.Residual.Amount = 0
	} else {
		result.MissingEvidence = append(result.MissingEvidence, "unexplained ledger variance: missing or incomplete driver evidence")
	}
	return result, nil
}

func receiptKind(k string) bool {
	switch k {
	case "commit", "pull_request", "release", "deployment":
		return true
	}
	return false
}

// Rates are direct, non-ambiguous snapshots. Same-currency conversion is identity.
func converter(target string, rates []ExchangeRate) (func(cost.Money) (float64, error), error) {
	table := map[string]float64{}
	for _, r := range rates {
		if r.From == r.To || r.From == "" || r.To == "" || !finite(r.Rate) || r.Rate <= 0 {
			return nil, errors.New("invalid exchange rate")
		}
		if err := (cost.Money{Amount: 0, Currency: r.From}).Validate(); err != nil {
			return nil, err
		}
		if err := (cost.Money{Amount: 0, Currency: r.To}).Validate(); err != nil {
			return nil, err
		}
		if r.Source.ArtifactIdentity == "" || !validCapture(r.Source.CapturedAt) || r.Source.Type == "" {
			return nil, errors.New("exchange rate requires source identity and timestamp")
		}
		key := r.From + "/" + r.To
		if _, ok := table[key]; ok {
			return nil, fmt.Errorf("duplicate exchange rate %s", key)
		}
		table[key] = r.Rate
	}
	return func(m cost.Money) (float64, error) {
		rate := 1.0
		if m.Currency != target {
			var ok bool
			rate, ok = table[m.Currency+"/"+target]
			if !ok {
				return 0, fmt.Errorf("missing exchange rate %s/%s", m.Currency, target)
			}
		}
		n := m.Amount * rate
		if !finite(n) {
			return 0, errors.New("currency conversion overflow")
		}
		return n, nil
	}, nil
}

var timeUnits = map[string]float64{"second": 1, "minute": 60, "hour": 3600, "day": 86400, "week": 604800}

func factors(d cost.CostDriver, convert func(cost.Money) (float64, error)) (float64, float64, float64, error) {
	q := d.Quantity.Value
	per := d.UnitPrice.Per.Value
	if scale := timeUnits[d.Quantity.Unit]; scale > 0 {
		q *= scale
		per *= scale
	}
	p, err := convert(d.UnitPrice.Amount)
	if err != nil {
		return 0, 0, 0, err
	}
	p /= per
	dist := 1.0
	if d.Distribution != nil {
		switch d.Distribution.Type {
		case cost.DistributionNormal:
			dist = *d.Distribution.Mean
		case cost.DistributionUniform:
			dist = (*d.Distribution.Min + *d.Distribution.Max) / 2
		case cost.DistributionExponential:
			dist = 1 / *d.Distribution.Rate
		}
		if d.Distribution.Floor != nil {
			dist = math.Max(dist, *d.Distribution.Floor)
		}
	}
	if !finite(q) || !finite(p) || !finite(dist) {
		return 0, 0, 0, errors.New("non-finite normalized factors")
	}
	return q, p, dist, nil
}
func shapley(f, a [3]float64) (out [3]float64) {
	// Solve the 3-factor multilinear equation by averaging all six update orders.
	for i := 0; i < 3; i++ {
		for mask := 0; mask < 8; mask++ {
			if mask&(1<<i) != 0 {
				continue
			}
			n := 0
			before, after := 1.0, 1.0
			for j := 0; j < 3; j++ {
				v := f[j]
				if mask&(1<<j) != 0 {
					v = a[j]
					n++
				}
				before *= v
				if j == i {
					v = a[j]
				}
				after *= v
			}
			weight := 1.0 / 3
			if n == 1 {
				weight = 1.0 / 6
			}
			out[i] += weight * (after - before)
		}
	}
	return
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func validCapture(s string) bool {
	if _, err := time.Parse(time.RFC3339, s); err == nil {
		return true
	}
	_, err := time.Parse(time.DateOnly, s)
	return err == nil
}
func confidence(c *Contribution, f, a cost.CostDriver, total float64) {
	score := 1.0
	for _, e := range []cost.Evidence{f.Evidence.Quantity, f.Evidence.UnitPrice, a.Evidence.Quantity, a.Evidence.UnitPrice} {
		switch e.Confidence {
		case cost.ConfidenceLow:
			score = math.Min(score, .35)
		case cost.ConfidenceMedium:
			score = math.Min(score, .7)
		}
	}
	if math.Abs(total) > Rounding && math.Abs(c.Amount.Amount) < .01*math.Abs(total) {
		score = math.Min(score, .7)
	}
	if (f.Kind == cost.DriverCadence || f.Kind == cost.DriverConcurrency || f.Kind == cost.DriverUptime) && len(c.Receipts) == 0 {
		c.MissingEvidence = append(c.MissingEvidence, "exact delivery receipt")
		c.Attribution = "unknown"
		score = math.Min(score, .35)
	}
	if f.Distribution != nil || a.Distribution != nil {
		c.MissingEvidence = append(c.MissingEvidence, "distribution expectation is modeled, not observed causation")
		score = math.Min(score, .7)
	}
	c.ConfidenceScore = score
	switch {
	case score >= .85:
		c.Confidence = cost.ConfidenceHigh
	case score >= .55:
		c.Confidence = cost.ConfidenceMedium
	default:
		c.Confidence = cost.ConfidenceLow
	}
}

// MarshalJSONResult returns deterministic, indented JSON with a final newline.
func MarshalJSONResult(r Result) ([]byte, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
