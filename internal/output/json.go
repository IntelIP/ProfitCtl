package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONResult represents the structured JSON output format
type JSONResult struct {
	Scenario struct {
		Users        int     `json:"users"`
		Months       int     `json:"months"`
		GrowthFactor float64 `json:"growth_factor,omitempty"`
	} `json:"scenario"`
	Costs struct {
		Fixed struct {
			Total   float64 `json:"total"`
			Monthly float64 `json:"monthly"`
			ByLayer struct {
				Infrastructure float64 `json:"infrastructure"`
				Application    float64 `json:"application"`
				Service        float64 `json:"service"`
			} `json:"by_layer"`
		} `json:"fixed"`
		Variable struct {
			Total   float64 `json:"total"`
			PerUser float64 `json:"per_user"`
			ByLayer struct {
				Infrastructure float64 `json:"infrastructure"`
				Application    float64 `json:"application"`
				Service        float64 `json:"service"`
			} `json:"by_layer"`
		} `json:"variable"`
		Total float64 `json:"total"`
	} `json:"costs"`
	Revenue struct {
		Total  float64 `json:"total"`
		ByPlan []struct {
			PlanName string  `json:"plan_name"`
			Price    float64 `json:"price"`
			Users    int     `json:"users"`
			Revenue  float64 `json:"revenue"`
		} `json:"by_plan,omitempty"`
	} `json:"revenue,omitempty"`
	Margin struct {
		Gross        float64 `json:"gross"`
		CostPerUser  float64 `json:"cost_per_user"`
		ByLayer      struct {
			Infrastructure float64 `json:"infrastructure,omitempty"`
			Application    float64 `json:"application,omitempty"`
			Service        float64 `json:"service,omitempty"`
		} `json:"by_layer,omitempty"`
	} `json:"margin,omitempty"`
	StressTest struct {
		P95 struct {
			Margin       float64 `json:"margin,omitempty"`
			CostPerUser  float64 `json:"cost_per_user"`
		} `json:"p95"`
		P99 struct {
			Margin       float64 `json:"margin,omitempty"`
			CostPerUser  float64 `json:"cost_per_user"`
		} `json:"p99"`
		Mean struct {
			CostPerUser float64 `json:"cost_per_user"`
		} `json:"mean"`
		WorstCaseTotalCost float64 `json:"worst_case_total_cost"`
	} `json:"stress_test,omitempty"`
	Covenants struct {
		Passed    bool `json:"passed"`
		Violations []struct {
			Field    string  `json:"field"`
			Operator string  `json:"operator"`
			Value    float64 `json:"value"`
			Actual   float64 `json:"actual"`
			Message  string  `json:"message"`
		} `json:"violations"`
	} `json:"covenants"`
}

// FormatJSONResult converts SimulationResult to JSON format
func FormatJSONResult(result SimulationResult) ([]byte, error) {
	jsonResult := JSONResult{}

	// Scenario
	jsonResult.Scenario.Users = result.Users
	jsonResult.Scenario.Months = result.Months
	jsonResult.Scenario.GrowthFactor = result.GrowthFactor

	// Costs
	jsonResult.Costs.Fixed.Total = result.FixedCosts.Total
	jsonResult.Costs.Fixed.Monthly = result.FixedCosts.Monthly
	jsonResult.Costs.Fixed.ByLayer.Infrastructure = result.FixedCosts.ByLayer.Infrastructure
	jsonResult.Costs.Fixed.ByLayer.Application = result.FixedCosts.ByLayer.Application
	jsonResult.Costs.Fixed.ByLayer.Service = result.FixedCosts.ByLayer.Service

	jsonResult.Costs.Variable.Total = result.VariableCosts.Total
	jsonResult.Costs.Variable.PerUser = result.VariableCosts.PerUser
	jsonResult.Costs.Variable.ByLayer.Infrastructure = result.VariableCosts.ByLayer.Infrastructure
	jsonResult.Costs.Variable.ByLayer.Application = result.VariableCosts.ByLayer.Application
	jsonResult.Costs.Variable.ByLayer.Service = result.VariableCosts.ByLayer.Service

	jsonResult.Costs.Total = result.TotalCosts.GrandTotal

	// Revenue
	if result.Revenue.Total > 0 {
		jsonResult.Revenue.Total = result.Revenue.Total
		jsonResult.Revenue.ByPlan = make([]struct {
			PlanName string  `json:"plan_name"`
			Price    float64 `json:"price"`
			Users    int     `json:"users"`
			Revenue  float64 `json:"revenue"`
		}, len(result.Revenue.ByPlan))

		for i, plan := range result.Revenue.ByPlan {
			jsonResult.Revenue.ByPlan[i].PlanName = plan.PlanName
			jsonResult.Revenue.ByPlan[i].Price = plan.Price
			jsonResult.Revenue.ByPlan[i].Users = plan.Users
			jsonResult.Revenue.ByPlan[i].Revenue = plan.Revenue
		}
	}

	// Margin
	if result.Margin.GrossMargin != 0 {
		jsonResult.Margin.Gross = result.Margin.GrossMargin
		jsonResult.Margin.CostPerUser = result.Margin.CostPerUser
		jsonResult.Margin.ByLayer.Infrastructure = result.Margin.LayerCostPercent.Infrastructure
		jsonResult.Margin.ByLayer.Application = result.Margin.LayerCostPercent.Application
		jsonResult.Margin.ByLayer.Service = result.Margin.LayerCostPercent.Service
	}

	// Stress test
	if result.StressResult.P95CostPerUser > 0 {
		jsonResult.StressTest.P95.CostPerUser = result.StressResult.P95CostPerUser
		if result.Revenue.Total > 0 {
			p95TotalCost := result.StressResult.P95CostPerUser * float64(result.Users)
			p95Margin := ((result.Revenue.Total - p95TotalCost) / result.Revenue.Total) * 100
			jsonResult.StressTest.P95.Margin = p95Margin
		}

		jsonResult.StressTest.P99.CostPerUser = result.StressResult.P99CostPerUser
		if result.Revenue.Total > 0 {
			p99TotalCost := result.StressResult.P99CostPerUser * float64(result.Users)
			p99Margin := ((result.Revenue.Total - p99TotalCost) / result.Revenue.Total) * 100
			jsonResult.StressTest.P99.Margin = p99Margin
		}

		jsonResult.StressTest.Mean.CostPerUser = result.StressResult.MeanCostPerUser
		jsonResult.StressTest.WorstCaseTotalCost = result.StressResult.WorstCaseTotalCost
	}

	// Covenants
	jsonResult.Covenants.Passed = result.Covenants.Passed
	jsonResult.Covenants.Violations = make([]struct {
		Field    string  `json:"field"`
		Operator string  `json:"operator"`
		Value    float64 `json:"value"`
		Actual   float64 `json:"actual"`
		Message  string  `json:"message"`
	}, len(result.Covenants.Violations))

	for i, violation := range result.Covenants.Violations {
		jsonResult.Covenants.Violations[i].Field = violation.Field
		jsonResult.Covenants.Violations[i].Operator = violation.Operator
		jsonResult.Covenants.Violations[i].Value = violation.Value
		jsonResult.Covenants.Violations[i].Actual = violation.Actual
		jsonResult.Covenants.Violations[i].Message = violation.Message
	}

	return json.MarshalIndent(jsonResult, "", "  ")
}

// PrintJSONResult writes JSON output to stdout
func PrintJSONResult(result SimulationResult) error {
	jsonBytes, err := FormatJSONResult(result)
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// WriteJSONResult writes JSON output to a file
func WriteJSONResult(result SimulationResult, filename string) error {
	jsonBytes, err := FormatJSONResult(result)
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	return os.WriteFile(filename, jsonBytes, 0644)
}