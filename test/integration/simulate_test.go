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
