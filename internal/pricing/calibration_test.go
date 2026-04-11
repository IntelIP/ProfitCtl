package pricing

import (
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCalculateCalibration(t *testing.T) {
	calibration := &config.CalibrationConfig{
		Period:       "2026-03",
		Source:       "stripe_export",
		GrossRevenue: 9000,
		PaymentFees:  250,
		FreeUsers:    700,
		PaidUsers: &config.CalibrationPaidUsers{
			Monthly: 225,
			Annual:  75,
		},
		PlanMix: map[string]float64{
			"pro": 1.0,
		},
	}

	revenue := RevenueResult{
		Mode:  "mix",
		Total: 8700,
		ByPlan: []PlanRevenue{
			{PlanName: "free", Price: 0, Cohort: "free", Users: 700, Revenue: 0},
			{PlanName: "pro", Price: 29, Cohort: "paid", Users: 300, Revenue: 8700},
		},
	}
	fees := PaymentFeeResult{
		Total:            200,
		FreeUsers:        700,
		PaidMonthlyUsers: 225,
		PaidAnnualUsers:  75,
	}

	result := CalculateCalibration(calibration, revenue, fees)

	assert.Equal(t, "2026-03", result.Period)
	assert.Equal(t, "stripe_export", result.Source)
	assert.Equal(t, -300.0, result.RevenueDelta)
	assert.Equal(t, -50.0, result.PaymentFeesDelta)
	assert.Equal(t, 0.0, result.FreeUsersDelta)
	assert.Equal(t, 0.0, result.PaidMonthlyDelta)
	assert.Equal(t, 0.0, result.PaidAnnualDelta)
	assert.Equal(t, 0.0, result.PlanMixDeltas["pro"])
}
