package pricing

import (
	"strings"

	"github.com/IntelIP/ProfitCtl/internal/config"
)

type PaymentFeeResult struct {
	Processor             string
	Currency              string
	MonthlyAmount         float64
	AnnualAmortizedAmount float64
	FreeUserAmount        float64
	PaidUserAmount        float64
	FreeUsers             float64
	PaidMonthlyUsers      float64
	PaidAnnualUsers       float64
	PercentageAmount      float64
	FixedAmount           float64
	Total                 float64
}

func CalculatePaymentFees(fees *config.PaymentFeesConfig, pricingCfg *config.PricingConfig, revenue RevenueResult, totalUsers int) PaymentFeeResult {
	if fees == nil || revenue.Total <= 0 {
		return PaymentFeeResult{}
	}

	if usesLegacyPaymentFeeConfig(fees) {
		return calculateLegacyPaymentFees(fees, revenue)
	}

	result := PaymentFeeResult{
		Processor: fees.Processor,
		Currency:  fees.Currency,
	}

	monthlyShare := 1.0
	annualShare := 0.0
	if fees.BillingMix != nil {
		monthlyShare = fees.BillingMix.MonthlyShare
		annualShare = fees.BillingMix.AnnualShare
	}

	annualDiscount := 0.0
	annualPrepaidMonths := 12
	if fees.Assumptions != nil {
		annualDiscount = fees.Assumptions.AnnualDiscountPercent
		annualPrepaidMonths = fees.Assumptions.AnnualPrepaidMonths
	}

	if revenue.Mode == "hybrid" {
		applyHybridPaymentFees(&result, fees, revenue, totalUsers, monthlyShare, annualShare, annualDiscount, annualPrepaidMonths)
		return finalizePaymentFeeResult(result)
	}

	planConfigByName := map[string]config.PricingPlan{}
	if pricingCfg != nil {
		for _, plan := range pricingCfg.Plans {
			planConfigByName[plan.Name] = plan
		}
	}

	for _, plan := range revenue.ByPlan {
		cohort := inferPlanCohort(plan, planConfigByName[plan.PlanName])
		monthlyUsers := float64(plan.Users) * monthlyShare
		annualUsers := float64(plan.Users) * annualShare

		if cohort == "free" {
			result.FreeUsers += float64(plan.Users)
		} else {
			result.PaidMonthlyUsers += monthlyUsers
			result.PaidAnnualUsers += annualUsers
		}

		profile := paymentProfileForCohort(fees, cohort)
		if profile == nil {
			continue
		}

		monthlyRevenue := plan.Revenue * monthlyShare
		annualCashCollected := plan.Price * float64(annualPrepaidMonths) * annualUsers * (1 - annualDiscount/100)
		monthlyFee := monthlyRevenue*profile.MonthlyPercent/100 + monthlyUsers*profile.PerTransaction
		annualPercent := profile.AnnualPercent
		if annualPercent == 0 {
			annualPercent = profile.MonthlyPercent
		}
		annualAmortizedFee := 0.0
		if annualPrepaidMonths > 0 {
			annualFeeCash := annualCashCollected*annualPercent/100 + annualUsers*profile.PerTransaction
			annualAmortizedFee = annualFeeCash / float64(annualPrepaidMonths)
		}

		monthlyMultiplier := feeMultiplier(fees.PlanBurden, plan.PlanName, "monthly")
		annualMultiplier := feeMultiplier(fees.PlanBurden, plan.PlanName, "annual")
		monthlyFee = roundCurrency(monthlyFee * monthlyMultiplier)
		annualAmortizedFee = roundCurrency(annualAmortizedFee * annualMultiplier)

		result.MonthlyAmount += monthlyFee
		result.AnnualAmortizedAmount += annualAmortizedFee
		result.PercentageAmount += roundCurrency(monthlyRevenue*profile.MonthlyPercent/100*monthlyMultiplier + annualCashCollected*annualPercent/100/float64(maxInt(annualPrepaidMonths, 1))*annualMultiplier)
		result.FixedAmount += roundCurrency(monthlyUsers*profile.PerTransaction*monthlyMultiplier + annualUsers*profile.PerTransaction/float64(maxInt(annualPrepaidMonths, 1))*annualMultiplier)

		if cohort == "free" {
			result.FreeUserAmount += monthlyFee + annualAmortizedFee
		} else {
			result.PaidUserAmount += monthlyFee + annualAmortizedFee
		}
	}

	return finalizePaymentFeeResult(result)
}

func calculateLegacyPaymentFees(fees *config.PaymentFeesConfig, revenue RevenueResult) PaymentFeeResult {
	percentageAmount := roundCurrency(revenue.Total * fees.PercentOfRevenue / 100)
	fixedAmount := roundCurrency(fees.FixedPerPayment * float64(fees.MonthlyTransactions))

	return PaymentFeeResult{
		PercentageAmount: percentageAmount,
		FixedAmount:      fixedAmount,
		MonthlyAmount:    roundCurrency(percentageAmount + fixedAmount),
		Total:            roundCurrency(percentageAmount + fixedAmount),
	}
}

func applyHybridPaymentFees(result *PaymentFeeResult, fees *config.PaymentFeesConfig, revenue RevenueResult, totalUsers int, monthlyShare, annualShare, annualDiscount float64, annualPrepaidMonths int) {
	paidProfile := paymentProfileForCohort(fees, "paid")
	if paidProfile == nil {
		return
	}

	monthlyUsers := float64(totalUsers) * monthlyShare
	annualUsers := float64(totalUsers) * annualShare
	result.PaidMonthlyUsers = monthlyUsers
	result.PaidAnnualUsers = annualUsers

	monthlyRevenue := revenue.Total * monthlyShare
	annualRevenueEquivalent := revenue.RecurringTotal
	if annualRevenueEquivalent == 0 {
		annualRevenueEquivalent = revenue.Total
	}
	annualCashCollected := annualRevenueEquivalent * annualShare * float64(annualPrepaidMonths) * (1 - annualDiscount/100)

	monthlyFee := monthlyRevenue*paidProfile.MonthlyPercent/100 + monthlyUsers*paidProfile.PerTransaction
	annualPercent := paidProfile.AnnualPercent
	if annualPercent == 0 {
		annualPercent = paidProfile.MonthlyPercent
	}
	annualAmortizedFee := 0.0
	if annualPrepaidMonths > 0 {
		annualFeeCash := annualCashCollected*annualPercent/100 + annualUsers*paidProfile.PerTransaction
		annualAmortizedFee = annualFeeCash / float64(annualPrepaidMonths)
	}

	result.MonthlyAmount = roundCurrency(monthlyFee)
	result.AnnualAmortizedAmount = roundCurrency(annualAmortizedFee)
	result.PaidUserAmount = roundCurrency(result.MonthlyAmount + result.AnnualAmortizedAmount)
	result.PercentageAmount = roundCurrency(monthlyRevenue*paidProfile.MonthlyPercent/100 + annualCashCollected*annualPercent/100/float64(maxInt(annualPrepaidMonths, 1)))
	result.FixedAmount = roundCurrency(monthlyUsers*paidProfile.PerTransaction + annualUsers*paidProfile.PerTransaction/float64(maxInt(annualPrepaidMonths, 1)))
}

func finalizePaymentFeeResult(result PaymentFeeResult) PaymentFeeResult {
	result.MonthlyAmount = roundCurrency(result.MonthlyAmount)
	result.AnnualAmortizedAmount = roundCurrency(result.AnnualAmortizedAmount)
	result.FreeUserAmount = roundCurrency(result.FreeUserAmount)
	result.PaidUserAmount = roundCurrency(result.PaidUserAmount)
	result.PercentageAmount = roundCurrency(result.PercentageAmount)
	result.FixedAmount = roundCurrency(result.FixedAmount)
	result.Total = roundCurrency(result.MonthlyAmount + result.AnnualAmortizedAmount)
	return result
}

func usesLegacyPaymentFeeConfig(fees *config.PaymentFeesConfig) bool {
	return fees.FreeUserFee == nil &&
		fees.PaidUserFee == nil &&
		len(fees.PlanBurden) == 0 &&
		(fees.BillingMix == nil || (fees.BillingMix.MonthlyShare == 1 && fees.BillingMix.AnnualShare == 0)) &&
		(fees.Assumptions == nil || (fees.Assumptions.AnnualDiscountPercent == 0 && fees.Assumptions.AnnualPrepaidMonths == 12)) &&
		(fees.PercentOfRevenue != 0 || fees.FixedPerPayment != 0 || fees.MonthlyTransactions != 0)
}

func inferPlanCohort(plan PlanRevenue, cfgPlan config.PricingPlan) string {
	if cfgPlan.Cohort != "" {
		return cfgPlan.Cohort
	}
	if plan.Cohort != "" {
		return plan.Cohort
	}
	if plan.Price == 0 {
		return "free"
	}
	return "paid"
}

func paymentProfileForCohort(fees *config.PaymentFeesConfig, cohort string) *config.PaymentFeeProfile {
	switch cohort {
	case "free":
		return fees.FreeUserFee
	default:
		if fees.PaidUserFee != nil {
			return fees.PaidUserFee
		}
		if fees.PercentOfRevenue != 0 || fees.FixedPerPayment != 0 {
			return &config.PaymentFeeProfile{
				MonthlyPercent: fees.PercentOfRevenue,
				PerTransaction: fees.FixedPerPayment,
			}
		}
		return nil
	}
}

func feeMultiplier(planBurden []config.PlanBurdenConfig, planName, appliesTo string) float64 {
	multiplier := 1.0
	for _, burden := range planBurden {
		if !strings.EqualFold(burden.Plan, planName) {
			continue
		}
		if burden.AppliesTo == "" || burden.AppliesTo == "all" || burden.AppliesTo == appliesTo {
			multiplier *= burden.FeeMultiplier
		}
	}
	return multiplier
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
