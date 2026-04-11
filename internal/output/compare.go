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
	Name             string  `json:"name"`
	File             string  `json:"file"`
	PricingMode      string  `json:"pricing_mode,omitempty"`
	Users            int     `json:"users"`
	Revenue          float64 `json:"revenue"`
	RecurringRevenue float64 `json:"recurring_revenue,omitempty"`
	PaymentFees      float64 `json:"payment_fees,omitempty"`
	TotalCosts       float64 `json:"total_costs"`
	Margin           float64 `json:"margin,omitempty"`
	CostPerUser      float64 `json:"cost_per_user,omitempty"`
	CovenantsPassed  bool    `json:"covenants_passed"`
	ViolationCount   int     `json:"violation_count"`
	RevenueDelta     float64 `json:"revenue_delta_vs_baseline,omitempty"`
	MarginDelta      float64 `json:"margin_delta_vs_baseline,omitempty"`
	CostPerUserDelta float64 `json:"cost_per_user_delta_vs_baseline,omitempty"`
}

type ComparisonLeaders struct {
	HighestRevenue     string `json:"highest_revenue"`
	HighestMargin      string `json:"highest_margin"`
	LowestCostPerUser  string `json:"lowest_cost_per_user"`
	BestCovenantHealth string `json:"best_covenant_health"`
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
	bestMargin := baseline.Name
	bestCostPerUser := baseline.Name
	bestCovenant := baseline.Name

	bestRevenueValue := baseline.Result.Revenue.Total
	bestMarginValue := baseline.Result.Margin.GrossMargin
	bestCostPerUserValue := baseline.Result.Margin.CostPerUser
	bestCovenantValue := covenantHealthScore(baseline.Result)

	for _, item := range items {
		scenario := ComparisonScenario{
			Name:             item.Name,
			File:             item.File,
			PricingMode:      item.Result.Revenue.Mode,
			Users:            item.Result.Users,
			Revenue:          item.Result.Revenue.Total,
			RecurringRevenue: item.Result.Revenue.RecurringTotal,
			PaymentFees:      item.Result.PaymentFees.Total,
			TotalCosts:       item.Result.TotalCosts.GrandTotal + item.Result.PaymentFees.Total,
			Margin:           item.Result.Margin.GrossMargin,
			CostPerUser:      item.Result.Margin.CostPerUser,
			CovenantsPassed:  item.Result.Covenants.Passed,
			ViolationCount:   len(item.Result.Covenants.Violations),
			RevenueDelta:     item.Result.Revenue.Total - baseline.Result.Revenue.Total,
			MarginDelta:      item.Result.Margin.GrossMargin - baseline.Result.Margin.GrossMargin,
			CostPerUserDelta: item.Result.Margin.CostPerUser - baseline.Result.Margin.CostPerUser,
		}
		result.Scenarios = append(result.Scenarios, scenario)

		if scenario.Revenue > bestRevenueValue {
			bestRevenueValue = scenario.Revenue
			bestRevenue = scenario.Name
		}
		if scenario.Margin > bestMarginValue {
			bestMarginValue = scenario.Margin
			bestMargin = scenario.Name
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
		HighestRevenue:     bestRevenue,
		HighestMargin:      bestMargin,
		LowestCostPerUser:  bestCostPerUser,
		BestCovenantHealth: bestCovenant,
	}

	return result
}

func FormatCLIComparisonResult(result ComparisonResult) string {
	var b strings.Builder

	b.WriteString("=== profitctl Scenario Comparison ===\n\n")
	if result.Baseline != "" {
		fmt.Fprintf(&b, "Baseline: %s\n\n", result.Baseline)
	}
	b.WriteString("Scenario            Mode     Revenue    Fees       Total Cost  Margin   CPU     Covenants\n")
	b.WriteString("----------------------------------------------------------------------------------------\n")
	for _, scenario := range result.Scenarios {
		status := "PASS"
		if !scenario.CovenantsPassed {
			status = fmt.Sprintf("FAIL(%d)", scenario.ViolationCount)
		}
		fmt.Fprintf(&b, "%-18s %-8s $%-9.2f $%-9.2f $%-10.2f %-7.2f $%-7.2f %s\n",
			truncateRight(scenario.Name, 18),
			truncateRight(emptyFallback(scenario.PricingMode, "-"), 8),
			scenario.Revenue,
			scenario.PaymentFees,
			scenario.TotalCosts,
			scenario.Margin,
			scenario.CostPerUser,
			status,
		)
	}

	b.WriteString("\nDelta vs baseline:\n")
	for _, scenario := range result.Scenarios {
		if scenario.Name == result.Baseline {
			continue
		}
		fmt.Fprintf(&b, "  %s: revenue %+0.2f, margin %+0.2f pts, cost/user %+0.2f\n",
			scenario.Name, scenario.RevenueDelta, scenario.MarginDelta, scenario.CostPerUserDelta)
	}

	b.WriteString("\nLeaders:\n")
	fmt.Fprintf(&b, "  Highest revenue: %s\n", result.Leaders.HighestRevenue)
	fmt.Fprintf(&b, "  Highest margin: %s\n", result.Leaders.HighestMargin)
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
	b.WriteString("| Scenario | Mode | Revenue | Payment Fees | Total Cost | Margin | Cost/User | Covenants |\n")
	b.WriteString("|----------|------|---------|--------------|------------|--------|-----------|-----------|\n")
	for _, scenario := range result.Scenarios {
		status := "PASS"
		if !scenario.CovenantsPassed {
			status = fmt.Sprintf("FAIL (%d)", scenario.ViolationCount)
		}
		fmt.Fprintf(&b, "| %s | %s | $%.2f | $%.2f | $%.2f | %.2f%% | $%.2f | %s |\n",
			scenario.Name,
			emptyFallback(scenario.PricingMode, "-"),
			scenario.Revenue,
			scenario.PaymentFees,
			scenario.TotalCosts,
			scenario.Margin,
			scenario.CostPerUser,
			status,
		)
	}

	b.WriteString("\n### Delta vs Baseline\n\n")
	b.WriteString("| Scenario | Revenue Delta | Margin Delta | Cost/User Delta |\n")
	b.WriteString("|----------|---------------|--------------|-----------------|\n")
	for _, scenario := range result.Scenarios {
		if scenario.Name == result.Baseline {
			continue
		}
		fmt.Fprintf(&b, "| %s | $%+.2f | %+.2f pts | $%+.2f |\n",
			scenario.Name, scenario.RevenueDelta, scenario.MarginDelta, scenario.CostPerUserDelta)
	}

	b.WriteString("\n### Leaders\n\n")
	b.WriteString(fmt.Sprintf("- Highest revenue: **%s**\n", result.Leaders.HighestRevenue))
	b.WriteString(fmt.Sprintf("- Highest margin: **%s**\n", result.Leaders.HighestMargin))
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
