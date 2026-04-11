package pricing

import (
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCalculatePaymentFees_Legacy(t *testing.T) {
	fees := &config.PaymentFeesConfig{
		PercentOfRevenue:    2.9,
		FixedPerPayment:     0.30,
		MonthlyTransactions: 10,
	}

	revenue := RevenueResult{Total: 1000}
	result := CalculatePaymentFees(fees, nil, revenue, 10)

	assert.Equal(t, 29.0, result.PercentageAmount)
	assert.Equal(t, 3.0, result.FixedAmount)
	assert.Equal(t, 32.0, result.MonthlyAmount)
	assert.Equal(t, 32.0, result.Total)
}

func TestCalculatePaymentFees_CohortAndBillingAware(t *testing.T) {
	fees := &config.PaymentFeesConfig{
		Processor: "stripe",
		Currency:  "usd",
		FreeUserFee: &config.PaymentFeeProfile{
			MonthlyPercent: 0,
			PerTransaction: 0,
		},
		PaidUserFee: &config.PaymentFeeProfile{
			MonthlyPercent: 2.9,
			AnnualPercent:  2.0,
			PerTransaction: 0.30,
		},
		BillingMix: &config.BillingMixConfig{
			MonthlyShare: 0.75,
			AnnualShare:  0.25,
		},
		PlanBurden: []config.PlanBurdenConfig{
			{Plan: "pro", AppliesTo: "annual", FeeMultiplier: 0.5},
		},
		Assumptions: &config.PaymentFeeAssumptions{
			AnnualDiscountPercent: 10,
			AnnualPrepaidMonths:   12,
		},
	}

	pricingCfg := &config.PricingConfig{
		Mode: "mix",
		Plans: []config.PricingPlan{
			{Name: "free", Price: 0, Share: sharePtr(0.7), Cohort: "free"},
			{Name: "pro", Price: 29, Share: sharePtr(0.3), Cohort: "paid"},
		},
	}

	revenue := RevenueResult{
		Mode:  "mix",
		Total: 8700,
		ByPlan: []PlanRevenue{
			{PlanName: "free", Price: 0, Cohort: "free", Share: sharePtr(0.7), Users: 700, Revenue: 0},
			{PlanName: "pro", Price: 29, Cohort: "paid", Share: sharePtr(0.3), Users: 300, Revenue: 8700},
		},
	}

	result := CalculatePaymentFees(fees, pricingCfg, revenue, 1000)

	assert.Equal(t, "stripe", result.Processor)
	assert.Equal(t, "usd", result.Currency)
	assert.Equal(t, 700.0, result.FreeUsers)
	assert.Equal(t, 225.0, result.PaidMonthlyUsers)
	assert.Equal(t, 75.0, result.PaidAnnualUsers)
	assert.Greater(t, result.MonthlyAmount, 0.0)
	assert.Greater(t, result.AnnualAmortizedAmount, 0.0)
	assert.Equal(t, 0.0, result.FreeUserAmount)
	assert.Greater(t, result.PaidUserAmount, 0.0)
	assert.Equal(t, roundCurrency(result.MonthlyAmount+result.AnnualAmortizedAmount), result.Total)
}

func TestCalculatePaymentFees_HybridUsesPaidProfile(t *testing.T) {
	fees := &config.PaymentFeesConfig{
		PaidUserFee: &config.PaymentFeeProfile{
			MonthlyPercent: 3,
			AnnualPercent:  2,
			PerTransaction: 0.25,
		},
		BillingMix: &config.BillingMixConfig{
			MonthlyShare: 0.5,
			AnnualShare:  0.5,
		},
		Assumptions: &config.PaymentFeeAssumptions{
			AnnualPrepaidMonths: 12,
		},
	}

	revenue := RevenueResult{
		Mode:           "hybrid",
		Total:          6500,
		RecurringTotal: 2000,
		OneTimeTotal:   5000,
	}

	result := CalculatePaymentFees(fees, &config.PricingConfig{Mode: "hybrid"}, revenue, 50)

	assert.Equal(t, 25.0, result.PaidMonthlyUsers)
	assert.Equal(t, 25.0, result.PaidAnnualUsers)
	assert.Greater(t, result.OneTimeAmount, 0.0)
	assert.Less(t, result.OperatingAmount, result.Total)
	assert.Greater(t, result.Total, 0.0)
}

func TestCalculatePaymentFees_NoRevenue(t *testing.T) {
	fees := &config.PaymentFeesConfig{PercentOfRevenue: 2.9}
	result := CalculatePaymentFees(fees, nil, RevenueResult{}, 0)
	assert.Equal(t, 0.0, result.Total)
}
