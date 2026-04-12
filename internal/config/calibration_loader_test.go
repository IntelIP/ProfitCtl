package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCalibrationFile_YAMLWrapper(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "calibration.yml")
	err := os.WriteFile(filename, []byte(`
calibration:
  period: 2026-03
  source: stripe_export
  gross_revenue: 125000
  payment_fees: 4380
  free_users: 12000
  paid_users:
    monthly: 1800
    annual: 400
  plan_mix:
    starter: 0.65
    pro: 0.35
  usage:
    api_calls: 3200000
`), 0644)
	require.NoError(t, err)

	calibration, err := LoadCalibrationFile(filename)
	require.NoError(t, err)
	require.NotNil(t, calibration)
	assert.Equal(t, "2026-03", calibration.Period)
	assert.Equal(t, 125000.0, calibration.GrossRevenue)
	assert.Equal(t, 1800, calibration.PaidUsers.Monthly)
	assert.Equal(t, 0.65, calibration.PlanMix["starter"])
	assert.Equal(t, 3200000.0, calibration.Usage["api_calls"])
}

func TestLoadCalibrationFile_CSV(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "calibration.csv")
	err := os.WriteFile(filename, []byte(`field,value
period,2026-03
source,stripe_export
gross_revenue,125000
payment_fees,4380
free_users,12000
paid_users.monthly,1800
paid_users.annual,400
plan_mix.starter,0.65
plan_mix.pro,0.35
usage.api_calls,3200000
`), 0644)
	require.NoError(t, err)

	calibration, err := LoadCalibrationFile(filename)
	require.NoError(t, err)
	require.NotNil(t, calibration)
	assert.Equal(t, "stripe_export", calibration.Source)
	assert.Equal(t, 4380.0, calibration.PaymentFees)
	assert.Equal(t, 12000, calibration.FreeUsers)
	assert.Equal(t, 400, calibration.PaidUsers.Annual)
	assert.Equal(t, 0.35, calibration.PlanMix["pro"])
	assert.Equal(t, 3200000.0, calibration.Usage["api_calls"])
}
