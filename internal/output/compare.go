package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ComparisonItem struct {
	Name   string
	File   string
	Result SimulationResult
}

type ComparisonScenario struct {
	Name                 string  `json:"name"`
	File                 string  `json:"file"`
	PricingMode          string  `json:"pricing_mode,omitempty"`
	Users                int     `json:"users"`
	Revenue              float64 `json:"revenue"`
	RecurringRevenue     float64 `json:"recurring_revenue,omitempty"`
	OneTimeRevenue       float64 `json:"one_time_revenue,omitempty"`
	PaymentFees          float64 `json:"payment_fees,omitempty"`
	OperatingFees        float64 `json:"operating_fees,omitempty"`
	TotalCosts           float64 `json:"total_costs"`
	Margin               float64 `json:"margin,omitempty"`
	OperatingMargin      float64 `json:"operating_margin,omitempty"`
	CostPerUser          float64 `json:"cost_per_user,omitempty"`
	CovenantsPassed      bool    `json:"covenants_passed"`
	ViolationCount       int     `json:"violation_count"`
	RevenueDelta         float64 `json:"revenue_delta_vs_baseline,omitempty"`
	MarginDelta          float64 `json:"margin_delta_vs_baseline,omitempty"`
	OperatingMarginDelta float64 `json:"operating_margin_delta_vs_baseline,omitempty"`
	CostPerUserDelta     float64 `json:"cost_per_user_delta_vs_baseline,omitempty"`
}

type ComparisonLeaders struct {
	HighestRevenue         string `json:"highest_revenue"`
	HighestOperatingMargin string `json:"highest_operating_margin"`
	LowestCostPerUser      string `json:"lowest_cost_per_user"`
	BestCovenantHealth     string `json:"best_covenant_health"`
}

type ComparisonResult struct {
	Baseline  string               `json:"baseline"`
	Scenarios []ComparisonScenario `json:"scenarios"`
	Leaders   ComparisonLeaders    `json:"leaders"`
}

func BuildComparisonResult(items []ComparisonItem) ComparisonResult {
	result := ComparisonResult{
		Scenarios: make([]ComparisonScenario, 0, len(items)),
	}
	if len(items) == 0 {
		return result
	}

	baseline := items[0]
	result.Baseline = baseline.Name

	bestRevenue := baseline.Name
	bestOperatingMargin := baseline.Name
	bestCostPerUser := baseline.Name
	bestCovenant := baseline.Name

	bestRevenueValue := baseline.Result.Revenue.Total
	bestOperatingMarginValue := baseline.Result.OperatingMargin.GrossMargin
	bestCostPerUserValue := baseline.Result.OperatingMargin.CostPerUser
	bestCovenantValue := covenantHealthScore(baseline.Result)

	for _, item := range items {
		scenario := ComparisonScenario{
			Name:                 item.Name,
			File:                 item.File,
			PricingMode:          item.Result.Revenue.Mode,
			Users:                item.Result.Users,
			Revenue:              item.Result.Revenue.Total,
			RecurringRevenue:     item.Result.Revenue.RecurringTotal,
			OneTimeRevenue:       item.Result.Revenue.OneTimeTotal,
			PaymentFees:          item.Result.PaymentFees.Total,
			OperatingFees:        operatingPaymentFeeAmount(item.Result),
			TotalCosts:           item.Result.TotalCosts.GrandTotal + item.Result.PaymentFees.Total,
			Margin:               item.Result.Margin.GrossMargin,
			OperatingMargin:      item.Result.OperatingMargin.GrossMargin,
			CostPerUser:          item.Result.OperatingMargin.CostPerUser,
			CovenantsPassed:      item.Result.Covenants.Passed,
			ViolationCount:       len(item.Result.Covenants.Violations),
			RevenueDelta:         item.Result.Revenue.Total - baseline.Result.Revenue.Total,
			MarginDelta:          item.Result.Margin.GrossMargin - baseline.Result.Margin.GrossMargin,
			OperatingMarginDelta: item.Result.OperatingMargin.GrossMargin - baseline.Result.OperatingMargin.GrossMargin,
			CostPerUserDelta:     item.Result.OperatingMargin.CostPerUser - baseline.Result.OperatingMargin.CostPerUser,
		}
		result.Scenarios = append(result.Scenarios, scenario)

		if scenario.Revenue > bestRevenueValue {
			bestRevenueValue = scenario.Revenue
			bestRevenue = scenario.Name
		}
		if scenario.OperatingMargin > bestOperatingMarginValue {
			bestOperatingMarginValue = scenario.OperatingMargin
			bestOperatingMargin = scenario.Name
		}
		if scenario.CostPerUser < bestCostPerUserValue {
			bestCostPerUserValue = scenario.CostPerUser
			bestCostPerUser = scenario.Name
		}
		score := covenantHealthScore(item.Result)
		if score > bestCovenantValue {
			bestCovenantValue = score
			bestCovenant = scenario.Name
		}
	}

	result.Leaders = ComparisonLeaders{
		HighestRevenue:         bestRevenue,
		HighestOperatingMargin: bestOperatingMargin,
		LowestCostPerUser:      bestCostPerUser,
		BestCovenantHealth:     bestCovenant,
	}

	return result
}

func FormatCLIComparisonResult(result ComparisonResult) string {
	var b strings.Builder

	b.WriteString("=== profitctl Scenario Comparison ===\n\n")
	if result.Baseline != "" {
		fmt.Fprintf(&b, "Baseline: %s\n\n", result.Baseline)
	}
	b.WriteString("Scenario            Mode     Revenue    Recurring  Fees       Booked   Op Marg  CPU     Covenants\n")
	b.WriteString("------------------------------------------------------------------------------------------------\n")
	for _, scenario := range result.Scenarios {
		status := "PASS"
		if !scenario.CovenantsPassed {
			status = fmt.Sprintf("FAIL(%d)", scenario.ViolationCount)
		}
		fmt.Fprintf(&b, "%-18s %-8s $%-9.2f $%-9.2f $%-9.2f %-8.2f %-8.2f $%-7.2f %s\n",
			truncateRight(scenario.Name, 18),
			truncateRight(emptyFallback(scenario.PricingMode, "-"), 8),
			scenario.Revenue,
			scenario.RecurringRevenue,
			scenario.PaymentFees,
			scenario.Margin,
			scenario.OperatingMargin,
			scenario.CostPerUser,
			status,
		)
	}

	b.WriteString("\nDelta vs baseline:\n")
	for _, scenario := range result.Scenarios {
		if scenario.Name == result.Baseline {
			continue
		}
		fmt.Fprintf(&b, "  %s: revenue %+0.2f, booked %+0.2f pts, operating %+0.2f pts, cost/user %+0.2f\n",
			scenario.Name, scenario.RevenueDelta, scenario.MarginDelta, scenario.OperatingMarginDelta, scenario.CostPerUserDelta)
	}

	b.WriteString("\nLeaders:\n")
	fmt.Fprintf(&b, "  Highest revenue: %s\n", result.Leaders.HighestRevenue)
	fmt.Fprintf(&b, "  Highest operating margin: %s\n", result.Leaders.HighestOperatingMargin)
	fmt.Fprintf(&b, "  Lowest cost/user: %s\n", result.Leaders.LowestCostPerUser)
	fmt.Fprintf(&b, "  Best covenant health: %s\n", result.Leaders.BestCovenantHealth)

	return b.String()
}

func PrintCLIComparisonResult(result ComparisonResult) {
	fmt.Print(FormatCLIComparisonResult(result))
}

func FormatMarkdownComparisonResult(result ComparisonResult) string {
	var b strings.Builder

	b.WriteString("## profitctl Scenario Comparison\n\n")
	if result.Baseline != "" {
		fmt.Fprintf(&b, "Baseline: **%s**\n\n", result.Baseline)
	}
	b.WriteString("| Scenario | Mode | Revenue | Recurring Revenue | Payment Fees | Booked Margin | Operating Margin | Cost/User | Covenants |\n")
	b.WriteString("|----------|------|---------|-------------------|--------------|---------------|------------------|-----------|-----------|\n")
	for _, scenario := range result.Scenarios {
		status := "PASS"
		if !scenario.CovenantsPassed {
			status = fmt.Sprintf("FAIL (%d)", scenario.ViolationCount)
		}
		fmt.Fprintf(&b, "| %s | %s | $%.2f | $%.2f | $%.2f | %.2f%% | %.2f%% | $%.2f | %s |\n",
			scenario.Name,
			emptyFallback(scenario.PricingMode, "-"),
			scenario.Revenue,
			scenario.RecurringRevenue,
			scenario.PaymentFees,
			scenario.Margin,
			scenario.OperatingMargin,
			scenario.CostPerUser,
			status,
		)
	}

	b.WriteString("\n### Delta vs Baseline\n\n")
	b.WriteString("| Scenario | Revenue Delta | Booked Margin Delta | Operating Margin Delta | Cost/User Delta |\n")
	b.WriteString("|----------|---------------|---------------------|------------------------|-----------------|\n")
	for _, scenario := range result.Scenarios {
		if scenario.Name == result.Baseline {
			continue
		}
		fmt.Fprintf(&b, "| %s | $%+.2f | %+.2f pts | %+.2f pts | $%+.2f |\n",
			scenario.Name, scenario.RevenueDelta, scenario.MarginDelta, scenario.OperatingMarginDelta, scenario.CostPerUserDelta)
	}

	b.WriteString("\n### Leaders\n\n")
	b.WriteString(fmt.Sprintf("- Highest revenue: **%s**\n", result.Leaders.HighestRevenue))
	b.WriteString(fmt.Sprintf("- Highest operating margin: **%s**\n", result.Leaders.HighestOperatingMargin))
	b.WriteString(fmt.Sprintf("- Lowest cost/user: **%s**\n", result.Leaders.LowestCostPerUser))
	b.WriteString(fmt.Sprintf("- Best covenant health: **%s**\n", result.Leaders.BestCovenantHealth))

	return b.String()
}

func PrintMarkdownComparisonResult(result ComparisonResult) {
	fmt.Print(FormatMarkdownComparisonResult(result))
}

func FormatJSONComparisonResult(result ComparisonResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

func PrintJSONComparisonResult(result ComparisonResult) error {
	jsonBytes, err := FormatJSONComparisonResult(result)
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

func WriteJSONComparisonResult(result ComparisonResult, filename string) error {
	jsonBytes, err := FormatJSONComparisonResult(result)
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	return os.WriteFile(filename, jsonBytes, 0644)
}

func covenantHealthScore(result SimulationResult) float64 {
	score := 0.0
	if result.Covenants.Passed {
		score += 1000
	}
	score += result.Margin.GrossMargin
	score -= float64(len(result.Covenants.Violations)) * 10
	return score
}

func truncateRight(v string, max int) string {
	if len(v) <= max {
		return v
	}
	if max <= 3 {
		return v[:max]
	}
	return v[:max-3] + "..."
}

func emptyFallback(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
