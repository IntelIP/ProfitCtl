package integration

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getBinaryPath(t *testing.T) string {
	binaryPath := filepath.Join(t.TempDir(), "profitctl")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../..")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	return binaryPath
}

func getFixturePath(fixture string) string {
	return filepath.Join("..", "fixtures", fixture)
}

func exitCodeFromErr(err error) int {
	if err == nil {
		return 0
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		return -1
	}
	return ee.ExitCode()
}

func TestSimulate_ValidConfig(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("valid_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Logf("Error: %v", err)
	}

	outputStr := string(output)
	assert.Contains(t, outputStr, "=== profitctl Simulation Results ===")
	assert.Contains(t, outputStr, "Scenario:")
	assert.Contains(t, outputStr, "Fixed COGS:")
	assert.Contains(t, outputStr, "Covenant Status:")
}

func TestSimulate_JSONOutput(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("valid_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Logf("Error: %v", err)
	}

	outputStr := string(output)
	var jsonResult map[string]interface{}
	err = json.Unmarshal([]byte(outputStr), &jsonResult)
	assert.NoError(t, err, "Output should be valid JSON")

	scenario, ok := jsonResult["scenario"].(map[string]interface{})
	assert.True(t, ok, "Should have scenario field")
	users, ok := scenario["users"].(float64)
	assert.True(t, ok, "Should have users field")
	assert.Equal(t, float64(100), users)

	costs, ok := jsonResult["costs"].(map[string]interface{})
	assert.True(t, ok, "Should have costs field")
	totalCost, ok := costs["total"].(float64)
	assert.True(t, ok, "Should have total cost")
	assert.Greater(t, totalCost, 0.0)
}

func TestSimulate_MarkdownOutput(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("valid_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--markdown")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Logf("Error: %v", err)
	}

	outputStr := string(output)
	assert.Contains(t, outputStr, "## profitctl Results")
	assert.Contains(t, outputStr, "| Metric | Value |")
	assert.Contains(t, outputStr, "### Cost Breakdown")
}

func TestSimulate_CovenantBreach(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("covenant_breach_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath)
	output, err := cmd.CombinedOutput()

	require.Error(t, err, "Should fail with covenant breach")
	assert.Equal(t, 1, exitCodeFromErr(err))

	outputStr := string(output)
	assert.Contains(t, outputStr, "❌ FAILED")
	assert.Contains(t, outputStr, "Violations:")
}

func TestSimulate_MinimalConfig(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("minimal_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath)
	output, _ := cmd.CombinedOutput()

	outputStr := string(output)
	assert.Contains(t, outputStr, "=== profitctl Simulation Results ===")
}

func TestSimulate_NoPricingConfig(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("no_pricing_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath)
	output, _ := cmd.CombinedOutput()

	outputStr := string(output)
	assert.Contains(t, outputStr, "=== profitctl Simulation Results ===")
}

func TestSimulate_ConfigNotFound(t *testing.T) {
	binaryPath := getBinaryPath(t)

	cmd := exec.Command(binaryPath, "simulate", "-f", "/nonexistent/file.yml")
	output, err := cmd.CombinedOutput()

	require.Error(t, err)
	assert.Equal(t, 2, exitCodeFromErr(err))
	assert.Contains(t, string(output), "failed to parse config")
}

func TestSimulate_VerboseMode(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("valid_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--verbose")
	output, _ := cmd.CombinedOutput()

	outputStr := string(output)
	assert.Contains(t, outputStr, "Scale Simulation:")
	assert.Contains(t, outputStr, "Stress Test Details")
}

func TestSimulate_QuietMode(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("valid_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--quiet")
	output, _ := cmd.CombinedOutput()

	outputStr := string(output)
	assert.Contains(t, outputStr, "violations=")
	assert.NotContains(t, outputStr, "=== profitctl Simulation Results ===")
}

func TestSimulate_UsesMonthlyMarginForSummaryAndCovenants(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("monthly_margin_regression.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--quiet")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "monthly pricing should be compared against monthly costs")
	outputStr := string(output)
	assert.Contains(t, outputStr, "PASSED")
	assert.Contains(t, outputStr, "margin=80.00%")
	assert.Contains(t, outputStr, "cost_per_user=2.00")
}

func TestSimulate_MixModeJSONOutput(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("mix_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--json")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)

	var jsonResult map[string]interface{}
	err = json.Unmarshal(output, &jsonResult)
	require.NoError(t, err, "mix mode output should be valid JSON")

	revenue, ok := jsonResult["revenue"].(map[string]interface{})
	require.True(t, ok, "Should have revenue field")
	assert.Equal(t, "mix", revenue["mode"])

	byPlan, ok := revenue["by_plan"].([]interface{})
	require.True(t, ok, "Should have revenue by_plan")
	require.Len(t, byPlan, 2)

	firstPlan, ok := byPlan[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "free", firstPlan["plan_name"])
	assert.Equal(t, 0.7, firstPlan["share"])
	assert.Equal(t, float64(7), firstPlan["users"])

	secondPlan, ok := byPlan[1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "pro", secondPlan["plan_name"])
	assert.Equal(t, 0.3, secondPlan["share"])
	assert.Equal(t, float64(3), secondPlan["users"])
}

func TestSimulate_HybridModeJSONOutput(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("hybrid_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--json")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)

	var jsonResult map[string]interface{}
	err = json.Unmarshal(output, &jsonResult)
	require.NoError(t, err, "hybrid mode output should be valid JSON")

	revenue, ok := jsonResult["revenue"].(map[string]interface{})
	require.True(t, ok, "Should have revenue field")
	assert.Equal(t, "hybrid", revenue["mode"])
	assert.Equal(t, 2000.0, revenue["total"])
	assert.Equal(t, 2000.0, revenue["recurring_total"])
	assert.Equal(t, 500.0, revenue["minimum_uplift"])

	components, ok := revenue["components"].([]interface{})
	require.True(t, ok, "Should have hybrid revenue components")
	require.NotEmpty(t, components)
}

func TestSimulate_HybridPilotJSONOutput(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("hybrid_pilot_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--json")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)

	var jsonResult map[string]interface{}
	err = json.Unmarshal(output, &jsonResult)
	require.NoError(t, err, "hybrid pilot output should be valid JSON")

	revenue, ok := jsonResult["revenue"].(map[string]interface{})
	require.True(t, ok, "Should have revenue field")
	assert.Equal(t, "hybrid", revenue["mode"])
	assert.Equal(t, 6500.0, revenue["total"])
	assert.Equal(t, 2000.0, revenue["recurring_total"])
	assert.Equal(t, 5000.0, revenue["one_time_total"])
}

func TestCompare_JSONOutput(t *testing.T) {
	binaryPath := getBinaryPath(t)
	baselinePath := getFixturePath("hybrid_config.yml")
	mixPath := getFixturePath("hybrid_pilot_config.yml")

	cmd := exec.Command(binaryPath, "compare", baselinePath, mixPath, "--json")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)

	var jsonResult map[string]interface{}
	err = json.Unmarshal(output, &jsonResult)
	require.NoError(t, err, "compare output should be valid JSON")

	assert.Equal(t, "hybrid_config", jsonResult["baseline"])

	scenarios, ok := jsonResult["scenarios"].([]interface{})
	require.True(t, ok, "Should have scenarios field")
	require.Len(t, scenarios, 2)

	leaders, ok := jsonResult["leaders"].(map[string]interface{})
	require.True(t, ok, "Should have leaders field")
	assert.NotEmpty(t, leaders["highest_revenue"])
}

func TestSimulate_PaymentFeesJSONOutput(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("payment_fees_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--json")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)

	var jsonResult map[string]interface{}
	err = json.Unmarshal(output, &jsonResult)
	require.NoError(t, err, "payment fee output should be valid JSON")

	paymentFees, ok := jsonResult["payment_fees"].(map[string]interface{})
	require.True(t, ok, "Should have payment_fees field")
	assert.Equal(t, 32.0, paymentFees["monthly"])
	assert.Equal(t, 29.0, paymentFees["percentage_amount"])
	assert.Equal(t, 3.0, paymentFees["fixed_amount"])
	assert.Equal(t, 32.0, paymentFees["total"])
}

func TestSimulate_RichPaymentFeesAndCalibrationJSONOutput(t *testing.T) {
	binaryPath := getBinaryPath(t)
	fixturePath := getFixturePath("payment_fees_calibrated_config.yml")

	cmd := exec.Command(binaryPath, "simulate", "-f", fixturePath, "--json")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)

	var jsonResult map[string]interface{}
	err = json.Unmarshal(output, &jsonResult)
	require.NoError(t, err, "calibrated payment fee output should be valid JSON")

	paymentFees, ok := jsonResult["payment_fees"].(map[string]interface{})
	require.True(t, ok, "Should have payment_fees field")
	assert.Equal(t, "stripe", paymentFees["processor"])
	assert.Equal(t, "usd", paymentFees["currency"])
	assert.Equal(t, 700.0, paymentFees["free_users"])
	assert.Equal(t, 225.0, paymentFees["paid_monthly_users"])
	assert.Equal(t, 75.0, paymentFees["paid_annual_users"])
	assert.Greater(t, paymentFees["monthly"].(float64), 0.0)
	assert.Greater(t, paymentFees["annual_amortized"].(float64), 0.0)

	calibration, ok := jsonResult["calibration"].(map[string]interface{})
	require.True(t, ok, "Should have calibration field")
	assert.Equal(t, "2026-03", calibration["period"])
	assert.Equal(t, "stripe_export", calibration["source"])
	assert.Contains(t, calibration, "revenue_delta")
	assert.Contains(t, calibration, "payment_fees_delta")
}
