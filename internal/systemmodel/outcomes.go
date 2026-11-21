package systemmodel

import (
	"fmt"
	"strings"
	"unicode"
)

// OutcomeComparison describes one comparable cohort per option. All options
// share the caller's workload, observation window, acceptance rule and cost scope.
type OutcomeComparison struct {
	SchemaVersion string          `json:"schema_version" jsonschema:"Set to profitctl.outcomes/v1."`
	Workload      string          `json:"workload"`
	Window        string          `json:"window"`
	Acceptance    string          `json:"acceptance"`
	CostScope     string          `json:"cost_scope"`
	Options       []OutcomeOption `json:"options" jsonschema:"Baseline first. Each option includes all attempts including failed work."`
}

type OutcomeOption struct {
	Name        string          `json:"name"`
	Attempts    *int64          `json:"attempts" jsonschema:"Number of attempts in this cohort. Null if unavailable."`
	Accepted    *int64          `json:"accepted" jsonschema:"Outcomes meeting the stated acceptance rule. Null if unmeasured."`
	CountSource string          `json:"count_source"`
	Costs       []OutcomeCharge `json:"costs" jsonschema:"Nonoverlapping net costs for the entire cohort in USD. Include a line for each missing charge."`
}

type OutcomeCharge struct {
	ID     string   `json:"id"`
	USD    *float64 `json:"usd"`
	Basis  string   `json:"basis" jsonschema:"Preserve the source basis. Use modeled for scenario and illustrative fixture costs, reported for charges on a real usage record, invoice for billed amounts, and unknown with a null amount for missing charges."`
	Source string   `json:"source" jsonschema:"Source of the amount, or the evidence needed when unknown."`
}

type OutcomeResult struct {
	OutcomeOption
	KnownSubtotal    *float64 `json:"known_subtotal_usd"`
	ScopeTotal       *float64 `json:"scope_total_usd"`
	PerAccepted      *float64 `json:"usd_per_accepted"`
	PerAttempt       *float64 `json:"usd_per_attempt"`
	AcceptanceRate   *float64 `json:"acceptance_rate"`
	BreakEvenRate    *float64 `json:"break_even_acceptance_rate"`
	BreakEvenMeaning string   `json:"break_even_meaning,omitempty"`
	Limitations      []string `json:"limitations"`
}

type OutcomeReport struct {
	SchemaVersion string          `json:"schema_version"`
	Status        string          `json:"status"`
	Workload      string          `json:"workload"`
	Window        string          `json:"window"`
	Acceptance    string          `json:"acceptance"`
	CostScope     string          `json:"cost_scope"`
	Options       []OutcomeResult `json:"options"`
}

func outcomeText(s string) bool {
	return len(s) <= 2048 && strings.TrimSpace(s) != "" && strings.IndexFunc(s, unicode.IsControl) == -1
}

func summarizeOption(option OutcomeOption) (OutcomeResult, error) {
	r := OutcomeResult{OutcomeOption: option, Limitations: []string{}}
	if !outcomeText(option.Name) || !outcomeText(option.CountSource) || len(option.Costs) < 1 || len(option.Costs) > 128 {
		return r, fmt.Errorf("each option requires a name, count source, and 1..128 cost lines")
	}
	for _, count := range []*int64{option.Attempts, option.Accepted} {
		if count != nil && (*count < 0 || *count > 1_000_000_000_000) {
			return r, fmt.Errorf("count outside supported range")
		}
	}
	if option.Attempts != nil && option.Accepted != nil && *option.Accepted > *option.Attempts {
		return r, fmt.Errorf("accepted count exceeds attempts")
	}
	seen := make(map[string]bool)
	subtotal := 0.0
	for _, charge := range option.Costs {
		if !idPattern.MatchString(charge.ID) || seen[charge.ID] || !outcomeText(charge.Source) {
			return r, fmt.Errorf("each cost requires a unique charge id and a source")
		}
		seen[charge.ID] = true
		switch charge.Basis {
		case "unknown":
			if charge.USD != nil {
				return r, fmt.Errorf("unknown cost must have a null amount")
			}
			r.Limitations = append(r.Limitations, "Missing cost: "+charge.ID)
		case "invoice", "reported", "modeled":
			if charge.USD == nil || !finite(*charge.USD) {
				return r, fmt.Errorf("known cost requires a finite nonnegative USD amount")
			}
			subtotal += *charge.USD
			if !finite(subtotal) {
				return r, fmt.Errorf("cost sum overflow")
			}
			r.KnownSubtotal = &subtotal
		default:
			return r, fmt.Errorf("unsupported cost basis")
		}
	}
	if len(r.Limitations) == 0 {
		r.ScopeTotal = &subtotal
	}
	if option.Attempts == nil {
		r.Limitations = append(r.Limitations, "Attempt count is unknown.")
	} else if *option.Attempts > 0 {
		if r.ScopeTotal != nil {
			r.PerAttempt = number(subtotal / float64(*option.Attempts))
		}
		if option.Accepted != nil {
			r.AcceptanceRate = number(float64(*option.Accepted) / float64(*option.Attempts))
		}
	}
	if option.Accepted == nil {
		r.Limitations = append(r.Limitations, "Accepted count is unknown.")
	} else if *option.Accepted == 0 {
		r.Limitations = append(r.Limitations, "No accepted outcomes; unit cost is undefined.")
	} else if r.ScopeTotal != nil {
		r.PerAccepted = number(subtotal / float64(*option.Accepted))
	}
	return r, nil
}

// CompareOutcomes solves candidate spend per attempt / baseline spend per
// accepted outcome. This threshold assumes candidate spend per attempt stays
// fixed as acceptance changes. It does not predict acceptance or quality.
func CompareOutcomes(in OutcomeComparison) (OutcomeReport, error) {
	report := OutcomeReport{SchemaVersion: "profitctl.outcome-comparison/v1", Status: "complete",
		Workload: in.Workload, Window: in.Window, Acceptance: in.Acceptance, CostScope: in.CostScope}
	if in.SchemaVersion != "profitctl.outcomes/v1" || !outcomeText(in.Workload) || !outcomeText(in.Window) ||
		!outcomeText(in.Acceptance) || !outcomeText(in.CostScope) || len(in.Options) < 1 || len(in.Options) > 4 {
		return report, fmt.Errorf("expected profitctl.outcomes/v1, workload, window, acceptance rule, cost scope, and 1..4 options")
	}
	names := make(map[string]bool)
	for i, option := range in.Options {
		result, err := summarizeOption(option)
		if err != nil {
			return report, fmt.Errorf("option %d: %w", i+1, err)
		}
		name := strings.ToLower(strings.TrimSpace(option.Name))
		if names[name] {
			return report, fmt.Errorf("duplicate option name")
		}
		names[name] = true
		if len(result.Limitations) > 0 {
			report.Status = "incomplete"
		}
		if i > 0 {
			baseline := report.Options[0].PerAccepted
			switch {
			case baseline == nil || result.PerAttempt == nil:
				result.BreakEvenMeaning = "Needs baseline cost per accepted outcome and candidate cost per attempt."
			case *result.PerAttempt == 0:
				result.BreakEvenMeaning = "Any positive acceptance rate matches or beats the baseline at zero scoped cost."
			case *baseline == 0:
				result.BreakEvenMeaning = "Positive candidate spend cannot match zero baseline unit cost."
			default:
				rate := *result.PerAttempt / *baseline
				if !finite(rate) {
					return report, fmt.Errorf("break-even rate overflow")
				}
				result.BreakEvenRate = number(rate)
				result.BreakEvenMeaning = "Fraction of attempts that must be accepted to match baseline unit cost at this spend per attempt."
				if rate > 1 {
					result.BreakEvenMeaning = "Even accepting every attempt would exceed baseline unit cost."
				}
			}
		}
		report.Options = append(report.Options, result)
	}
	return report, nil
}
