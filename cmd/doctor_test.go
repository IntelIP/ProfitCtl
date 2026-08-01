package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDoctorPassesWithValidLocalInputs(t *testing.T) {
	oldCatalog := doctorCatalog
	defer func() { doctorCatalog = oldCatalog }()
	doctorCatalog = filepath.Join("..", "test", "fixtures", "provider_catalog_valid.yml")

	command := doctorTestCommand(filepath.Join("..", "examples", "valid_profit.yml"))
	var output bytes.Buffer
	command.SetOut(&output)

	err := runDoctor(command, nil)
	require.NoError(t, err)
	assert.Contains(t, output.String(), "[warn] binary:")
	assert.Contains(t, output.String(), "[ok] config:")
	assert.Contains(t, output.String(), "[ok] catalog:")
	assert.Contains(t, output.String(), "doctor passed")
}

func TestRunDoctorReportsMissingInputsWithoutFallback(t *testing.T) {
	oldCatalog := doctorCatalog
	defer func() { doctorCatalog = oldCatalog }()
	doctorCatalog = filepath.Join(t.TempDir(), "missing-catalog.yml")

	command := doctorTestCommand(filepath.Join(t.TempDir(), "missing-profit.yml"))
	var output bytes.Buffer
	command.SetOut(&output)

	err := runDoctor(command, nil)
	require.Error(t, err)
	var exitErr *ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
	assert.True(t, exitErr.Silent)
	assert.Contains(t, output.String(), "[fail] config:")
	assert.Contains(t, output.String(), "profitctl init --file")
	assert.Contains(t, output.String(), "[fail] catalog:")
	assert.Contains(t, output.String(), "--catalog PATH")
	assert.Contains(t, output.String(), "no fallback or local state change was performed")
}

func TestReadCatalogRejectsIncompleteEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.yml")
	err := os.WriteFile(path, []byte("catalog_version: today\nstatus: test\nentries:\n  - id: incomplete\n"), 0644)
	require.NoError(t, err)

	_, err = readCatalog(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires id, provider, service, and unit")
}

func TestReadCatalogRequiresProvenanceFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.yml")
	err := os.WriteFile(path, []byte(`
catalog_version: test-v1
status: test
entries:
  - id: missing-provenance
    provider: synthetic
    service: test
    unit: request
`), 0644)
	require.NoError(t, err)

	_, err = readCatalog(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires value and currency")
}

func TestReadCatalogRejectsInvalidValue(t *testing.T) {
	for _, value := range []string{"-0.001", ".nan", ".inf"} {
		t.Run(value, func(t *testing.T) {
			path := writeCatalogVariant(t, "value: 0.001", "value: "+value)

			_, err := readCatalog(path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "value must be finite and non-negative")
		})
	}
}

func TestReadCatalogRejectsInvalidProvenanceEnums(t *testing.T) {
	sourcePath := writeCatalogVariant(t, "type: user_supplied", "type: provider-catalg")
	_, err := readCatalog(sourcePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported source type")

	confidencePath := writeCatalogVariant(t, "confidence: medium", "confidence: medum")
	_, err = readCatalog(confidencePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported source confidence")
}

func TestReadCatalogRejectsFutureCaptureDate(t *testing.T) {
	path := writeCatalogVariant(t, "captured_at: 2026-07-31", "captured_at: 2099-01-01")

	_, err := readCatalog(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "captured_at cannot be in the future")
}

func TestCatalogFreshness(t *testing.T) {
	now := time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC)
	catalog, err := readCatalog(filepath.Join("..", "test", "fixtures", "provider_catalog_valid.yml"))
	require.NoError(t, err)

	fresh, detail := catalogFreshness(catalog, now)
	assert.True(t, fresh)
	assert.Contains(t, detail, "within declared stale-after dates")

	catalog.Entries[0].StaleAfter = "2026-07-01"
	fresh, detail = catalogFreshness(catalog, now)
	assert.False(t, fresh)
	assert.Contains(t, detail, "1 of 1 entries are stale")
}

func TestSupportedExecutableName(t *testing.T) {
	assert.True(t, supportedExecutableName("/tmp/profitctl", "linux"))
	assert.True(t, supportedExecutableName(`C:\bin\profitctl.exe`, "windows"))
	assert.False(t, supportedExecutableName("/tmp/ProfitCtl", "linux"))
	assert.False(t, supportedExecutableName("/tmp/profitctl-old", "linux"))
}

func doctorTestCommand(configPath string) *cobra.Command {
	command := &cobra.Command{}
	command.Flags().String("file", configPath, "")
	return command
}

func writeCatalogVariant(t *testing.T, oldValue, newValue string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "test", "fixtures", "provider_catalog_valid.yml"))
	require.NoError(t, err)
	require.Contains(t, string(data), oldValue)

	path := filepath.Join(t.TempDir(), "catalog.yml")
	err = os.WriteFile(path, []byte(strings.Replace(string(data), oldValue, newValue, 1)), 0644)
	require.NoError(t, err)
	return path
}
