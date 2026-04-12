package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONResult represents the structured JSON output format
type JSONResult struct {
	Scenario struct {
		Users         int     `json:"users"`
		BillableUsers int     `json:"billable_users,omitempty"`
		Months        int     `json:"months"`
		GrowthFactor  float64 `json:"growth_factor,omitempty"`
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
		Mode           string  `json:"mode,omitempty"`
		Total          float64 `json:"total"`
		RecurringTotal float64 `json:"recurring_total,omitempty"`
		OneTimeTotal   float64 `json:"one_time_total,omitempty"`
		MinimumUplift  float64 `json:"minimum_uplift,omitempty"`
		Components     []struct {
			Name   string  `json:"name"`
			Amount float64 `json:"amount"`
		} `json:"components,omitempty"`
		ByPlan []struct {
			PlanName string   `json:"plan_name"`
			Price    float64  `json:"price"`
			Share    *float64 `json:"share,omitempty"`
			Users    int      `json:"users"`
			Revenue  float64  `json:"revenue"`
		} `json:"by_plan,omitempty"`
	} `json:"revenue,omitempty"`
	PaymentFees struct {
		Processor        string  `json:"processor,omitempty"`
		Currency         string  `json:"currency,omitempty"`
		Monthly          float64 `json:"monthly,omitempty"`
		AnnualAmortized  float64 `json:"annual_amortized,omitempty"`
		OperatingTotal   float64 `json:"operating_total,omitempty"`
		OneTimeAmount    float64 `json:"one_time_amount,omitempty"`
		FreeUserAmount   float64 `json:"free_user_amount,omitempty"`
		PaidUserAmount   float64 `json:"paid_user_amount,omitempty"`
		FreeUsers        float64 `json:"free_users,omitempty"`
		PaidMonthlyUsers float64 `json:"paid_monthly_users,omitempty"`
		PaidAnnualUsers  float64 `json:"paid_annual_users,omitempty"`
		PercentageAmount float64 `json:"percentage_amount,omitempty"`
		FixedAmount      float64 `json:"fixed_amount,omitempty"`
		Total            float64 `json:"total,omitempty"`
	} `json:"payment_fees,omitempty"`
	Calibration struct {
		Period             string             `json:"period,omitempty"`
		Source             string             `json:"source,omitempty"`
		RevenueActual      float64            `json:"revenue_actual,omitempty"`
		RevenueModeled     float64            `json:"revenue_modeled,omitempty"`
		RevenueDelta       float64            `json:"revenue_delta,omitempty"`
		PaymentFeesActual  float64            `json:"payment_fees_actual,omitempty"`
		PaymentFeesModeled float64            `json:"payment_fees_modeled,omitempty"`
		PaymentFeesDelta   float64            `json:"payment_fees_delta,omitempty"`
		FreeUsersActual    int                `json:"free_users_actual,omitempty"`
		FreeUsersModeled   float64            `json:"free_users_modeled,omitempty"`
		FreeUsersDelta     float64            `json:"free_users_delta,omitempty"`
		PaidMonthlyActual  int                `json:"paid_monthly_actual,omitempty"`
		PaidMonthlyModeled float64            `json:"paid_monthly_modeled,omitempty"`
		PaidMonthlyDelta   float64            `json:"paid_monthly_delta,omitempty"`
		PaidAnnualActual   int                `json:"paid_annual_actual,omitempty"`
		PaidAnnualModeled  float64            `json:"paid_annual_modeled,omitempty"`
		PaidAnnualDelta    float64            `json:"paid_annual_delta,omitempty"`
		PlanMixDeltas      map[string]float64 `json:"plan_mix_deltas,omitempty"`
	} `json:"calibration,omitempty"`
	Margin struct {
		Gross                float64 `json:"gross"`
		OperatingGross       float64 `json:"operating_gross,omitempty"`
		CostPerUser          float64 `json:"cost_per_user"`
		OperatingCostPerUser float64 `json:"operating_cost_per_user,omitempty"`
		ByLayer              struct {
			Infrastructure float64 `json:"infrastructure,omitempty"`
			Application    float64 `json:"application,omitempty"`
			Service        float64 `json:"service,omitempty"`
		} `json:"by_layer,omitempty"`
	} `json:"margin,omitempty"`
	StressTest struct {
		P95 struct {
			Margin          float64 `json:"margin,omitempty"`
			OperatingMargin float64 `json:"operating_margin,omitempty"`
			CostPerUser     float64 `json:"cost_per_user"`
		} `json:"p95"`
		P99 struct {
			Margin          float64 `json:"margin,omitempty"`
			OperatingMargin float64 `json:"operating_margin,omitempty"`
			CostPerUser     float64 `json:"cost_per_user"`
		} `json:"p99"`
		Mean struct {
			CostPerUser float64 `json:"cost_per_user"`
		} `json:"mean"`
		WorstCaseTotalCost float64 `json:"worst_case_total_cost"`
	} `json:"stress_test,omitempty"`
	Covenants struct {
		Passed     bool `json:"passed"`
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
	if result.BillableUsers > 0 && result.BillableUsers != result.Users {
		jsonResult.Scenario.BillableUsers = result.BillableUsers
	}
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
		jsonResult.Revenue.Mode = result.Revenue.Mode
		jsonResult.Revenue.Total = result.Revenue.Total
		jsonResult.Revenue.RecurringTotal = result.Revenue.RecurringTotal
		jsonResult.Revenue.OneTimeTotal = result.Revenue.OneTimeTotal
		jsonResult.Revenue.MinimumUplift = result.Revenue.MinimumUplift
		jsonResult.Revenue.Components = make([]struct {
			Name   string  `json:"name"`
			Amount float64 `json:"amount"`
		}, len(result.Revenue.Components))
		for i, component := range result.Revenue.Components {
			jsonResult.Revenue.Components[i].Name = component.Name
			jsonResult.Revenue.Components[i].Amount = component.Amount
		}
		jsonResult.Revenue.ByPlan = make([]struct {
			PlanName string   `json:"plan_name"`
			Price    float64  `json:"price"`
			Share    *float64 `json:"share,omitempty"`
			Users    int      `json:"users"`
			Revenue  float64  `json:"revenue"`
		}, len(result.Revenue.ByPlan))

		for i, plan := range result.Revenue.ByPlan {
			jsonResult.Revenue.ByPlan[i].PlanName = plan.PlanName
			jsonResult.Revenue.ByPlan[i].Price = plan.Price
			jsonResult.Revenue.ByPlan[i].Share = plan.Share
			jsonResult.Revenue.ByPlan[i].Users = plan.Users
			jsonResult.Revenue.ByPlan[i].Revenue = plan.Revenue
		}
	}

	if result.PaymentFees.Total > 0 {
		jsonResult.PaymentFees.Processor = result.PaymentFees.Processor
		jsonResult.PaymentFees.Currency = result.PaymentFees.Currency
		jsonResult.PaymentFees.Monthly = result.PaymentFees.MonthlyAmount
		jsonResult.PaymentFees.AnnualAmortized = result.PaymentFees.AnnualAmortizedAmount
		jsonResult.PaymentFees.OperatingTotal = result.PaymentFees.OperatingAmount
		jsonResult.PaymentFees.OneTimeAmount = result.PaymentFees.OneTimeAmount
		jsonResult.PaymentFees.FreeUserAmount = result.PaymentFees.FreeUserAmount
		jsonResult.PaymentFees.PaidUserAmount = result.PaymentFees.PaidUserAmount
		jsonResult.PaymentFees.FreeUsers = result.PaymentFees.FreeUsers
		jsonResult.PaymentFees.PaidMonthlyUsers = result.PaymentFees.PaidMonthlyUsers
		jsonResult.PaymentFees.PaidAnnualUsers = result.PaymentFees.PaidAnnualUsers
		jsonResult.PaymentFees.PercentageAmount = result.PaymentFees.PercentageAmount
		jsonResult.PaymentFees.FixedAmount = result.PaymentFees.FixedAmount
		jsonResult.PaymentFees.Total = result.PaymentFees.Total
	}

	if result.Calibration.Period != "" || result.Calibration.Source != "" || result.Calibration.RevenueActual != 0 || result.Calibration.PaymentFeesActual != 0 {
		jsonResult.Calibration.Period = result.Calibration.Period
		jsonResult.Calibration.Source = result.Calibration.Source
		jsonResult.Calibration.RevenueActual = result.Calibration.RevenueActual
		jsonResult.Calibration.RevenueModeled = result.Calibration.RevenueModeled
		jsonResult.Calibration.RevenueDelta = result.Calibration.RevenueDelta
		jsonResult.Calibration.PaymentFeesActual = result.Calibration.PaymentFeesActual
		jsonResult.Calibration.PaymentFeesModeled = result.Calibration.PaymentFeesModeled
		jsonResult.Calibration.PaymentFeesDelta = result.Calibration.PaymentFeesDelta
		jsonResult.Calibration.FreeUsersActual = result.Calibration.FreeUsersActual
		jsonResult.Calibration.FreeUsersModeled = result.Calibration.FreeUsersModeled
		jsonResult.Calibration.FreeUsersDelta = result.Calibration.FreeUsersDelta
		jsonResult.Calibration.PaidMonthlyActual = result.Calibration.PaidMonthlyActual
		jsonResult.Calibration.PaidMonthlyModeled = result.Calibration.PaidMonthlyModeled
		jsonResult.Calibration.PaidMonthlyDelta = result.Calibration.PaidMonthlyDelta
		jsonResult.Calibration.PaidAnnualActual = result.Calibration.PaidAnnualActual
		jsonResult.Calibration.PaidAnnualModeled = result.Calibration.PaidAnnualModeled
		jsonResult.Calibration.PaidAnnualDelta = result.Calibration.PaidAnnualDelta
		jsonResult.Calibration.PlanMixDeltas = result.Calibration.PlanMixDeltas
	}

	// Margin
	if result.Margin.GrossMargin != 0 {
		jsonResult.Margin.Gross = result.Margin.GrossMargin
		jsonResult.Margin.OperatingGross = result.OperatingMargin.GrossMargin
		jsonResult.Margin.CostPerUser = result.Margin.CostPerUser
		jsonResult.Margin.OperatingCostPerUser = result.OperatingMargin.CostPerUser
		jsonResult.Margin.ByLayer.Infrastructure = result.Margin.LayerCostPercent.Infrastructure
		jsonResult.Margin.ByLayer.Application = result.Margin.LayerCostPercent.Application
		jsonResult.Margin.ByLayer.Service = result.Margin.LayerCostPercent.Service
	}

	// Stress test
	if result.StressResult.P95CostPerUser > 0 {
		jsonResult.StressTest.P95.CostPerUser = result.StressResult.P95CostPerUser
		if result.Revenue.Total > 0 {
			jsonResult.StressTest.P95.Margin = bookedStressMargin(result, result.StressResult.P95CostPerUser)
			jsonResult.StressTest.P95.OperatingMargin = operatingStressMargin(result, result.StressResult.P95CostPerUser)
		}

		jsonResult.StressTest.P99.CostPerUser = result.StressResult.P99CostPerUser
		if result.Revenue.Total > 0 {
			jsonResult.StressTest.P99.Margin = bookedStressMargin(result, result.StressResult.P99CostPerUser)
			jsonResult.StressTest.P99.OperatingMargin = operatingStressMargin(result, result.StressResult.P99CostPerUser)
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
