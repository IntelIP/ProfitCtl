package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunCalibrate_RequiresInput(t *testing.T) {
	oldInput, oldOutput, oldJSON := calibrationInput, calibrationOutput, calibrationJSON
	defer func() {
		calibrationInput = oldInput
		calibrationOutput = oldOutput
		calibrationJSON = oldJSON
	}()

	calibrationInput = ""
	calibrationOutput = ""
	calibrationJSON = false

	err := runCalibrate(&cobra.Command{}, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--input is required")
}

func TestRunCalibrate_WritesNormalizedYAML(t *testing.T) {
	oldInput, oldOutput, oldJSON := calibrationInput, calibrationOutput, calibrationJSON
	defer func() {
		calibrationInput = oldInput
		calibrationOutput = oldOutput
		calibrationJSON = oldJSON
	}()

	tmpDir := t.TempDir()
	input := filepath.Join(tmpDir, "calibration.csv")
	output := filepath.Join(tmpDir, "calibration.yml")
	err := os.WriteFile(input, []byte(`field,value
period,2026-03
source,stripe_export
gross_revenue,125000
payment_fees,4380
free_users,12000
paid_users.monthly,1800
paid_users.annual,400
plan_mix.starter,0.65
`), 0644)
	assert.NoError(t, err)

	calibrationInput = input
	calibrationOutput = output
	calibrationJSON = false

	err = runCalibrate(&cobra.Command{}, []string{})
	assert.NoError(t, err)

	data, err := os.ReadFile(output)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "period: 2026-03")
	assert.Contains(t, string(data), "gross_revenue: 125000")
	assert.Contains(t, string(data), "paid_users:")
}

func TestCalibrateCmd_Flags(t *testing.T) {
	assert.NotNil(t, calibrateCmd)
	flags := calibrateCmd.Flags()
	assert.NotNil(t, flags.Lookup("input"))
	assert.NotNil(t, flags.Lookup("out"))
	assert.NotNil(t, flags.Lookup("json"))
}

func TestRunCalibrate_InvalidInput(t *testing.T) {
	oldInput, oldOutput, oldJSON := calibrationInput, calibrationOutput, calibrationJSON
	defer func() {
		calibrationInput = oldInput
		calibrationOutput = oldOutput
		calibrationJSON = oldJSON
	}()

	calibrationInput = "/nonexistent/calibration.csv"
	calibrationOutput = ""
	calibrationJSON = false

	err := runCalibrate(&cobra.Command{}, []string{})
	assert.Error(t, err)
	var exitErr *ExitError
	assert.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
}
