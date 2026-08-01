package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDoctorPassesWithValidLocalInputs(t *testing.T) {
	oldCatalog := doctorCatalog
	defer func() { doctorCatalog = oldCatalog }()
	doctorCatalog = filepath.Join("..", "provider_catalog", "ai_saas_defaults.yml")

	command := doctorTestCommand(filepath.Join("..", "examples", "valid_profit.yml"))
	var output bytes.Buffer
	command.SetOut(&output)

	err := runDoctor(command, nil)
	require.NoError(t, err)
	assert.Contains(t, output.String(), "[ok] binary:")
	assert.Contains(t, output.String(), "[ok] config:")
	assert.Contains(t, output.String(), "[warn] catalog:")
	assert.Contains(t, output.String(), "stale catalog cannot support current-price claims")
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

func TestCatalogFreshness(t *testing.T) {
	now := time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC)

	fresh, detail := catalogFreshness("2026-07-15", now)
	assert.True(t, fresh)
	assert.Contains(t, detail, "16 days old")

	fresh, detail = catalogFreshness("2026-05-29", now)
	assert.False(t, fresh)
	assert.Contains(t, detail, "30-day freshness window exceeded")
}

func doctorTestCommand(configPath string) *cobra.Command {
	command := &cobra.Command{}
	command.Flags().String("file", configPath, "")
	return command
}
