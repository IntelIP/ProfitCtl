package pricing

import "github.com/IntelIP/ProfitCtl/internal/config"

type CalibrationResult struct {
	Period             string
	Source             string
	RevenueActual      float64
	RevenueModeled     float64
	RevenueDelta       float64
	PaymentFeesActual  float64
	PaymentFeesModeled float64
	PaymentFeesDelta   float64
	FreeUsersActual    int
	FreeUsersModeled   float64
	FreeUsersDelta     float64
	PaidMonthlyActual  int
	PaidMonthlyModeled float64
	PaidMonthlyDelta   float64
	PaidAnnualActual   int
	PaidAnnualModeled  float64
	PaidAnnualDelta    float64
	PlanMixDeltas      map[string]float64
}

func CalculateCalibration(calibration *config.CalibrationConfig, revenue RevenueResult, paymentFees PaymentFeeResult) CalibrationResult {
	if calibration == nil {
		return CalibrationResult{}
	}

	result := CalibrationResult{
		Period:             calibration.Period,
		Source:             calibration.Source,
		RevenueActual:      calibration.GrossRevenue,
		RevenueModeled:     revenue.Total,
		RevenueDelta:       roundCurrency(revenue.Total - calibration.GrossRevenue),
		PaymentFeesActual:  calibration.PaymentFees,
		PaymentFeesModeled: paymentFees.Total,
		PaymentFeesDelta:   roundCurrency(paymentFees.Total - calibration.PaymentFees),
		FreeUsersActual:    calibration.FreeUsers,
		FreeUsersModeled:   roundCurrency(paymentFees.FreeUsers),
		FreeUsersDelta:     roundCurrency(paymentFees.FreeUsers - float64(calibration.FreeUsers)),
		PlanMixDeltas:      map[string]float64{},
	}

	if calibration.PaidUsers != nil {
		result.PaidMonthlyActual = calibration.PaidUsers.Monthly
		result.PaidMonthlyModeled = roundCurrency(paymentFees.PaidMonthlyUsers)
		result.PaidMonthlyDelta = roundCurrency(paymentFees.PaidMonthlyUsers - float64(calibration.PaidUsers.Monthly))
		result.PaidAnnualActual = calibration.PaidUsers.Annual
		result.PaidAnnualModeled = roundCurrency(paymentFees.PaidAnnualUsers)
		result.PaidAnnualDelta = roundCurrency(paymentFees.PaidAnnualUsers - float64(calibration.PaidUsers.Annual))
	}

	if len(calibration.PlanMix) > 0 {
		modeledMix := modeledPaidPlanMix(revenue)
		for plan, actualShare := range calibration.PlanMix {
			result.PlanMixDeltas[plan] = roundCurrency(modeledMix[plan] - actualShare)
		}
	}

	return result
}

func modeledPaidPlanMix(revenue RevenueResult) map[string]float64 {
	result := map[string]float64{}
	totalPaidUsers := 0.0
	for _, plan := range revenue.ByPlan {
		if inferPlanCohort(plan, config.PricingPlan{}) == "free" {
			continue
		}
		totalPaidUsers += float64(plan.Users)
	}
	if totalPaidUsers == 0 {
		return result
	}
	for _, plan := range revenue.ByPlan {
		if inferPlanCohort(plan, config.PricingPlan{}) == "free" {
			continue
		}
		result[plan.PlanName] = roundCurrency(float64(plan.Users) / totalPaidUsers)
	}
	return result
}
