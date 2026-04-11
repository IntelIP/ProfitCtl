package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunCompare_RequiresTwoFiles(t *testing.T) {
	cmd := &cobra.Command{}

	err := runCompare(cmd, []string{"one.yml"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least two config files")
}

func TestRunCompare_ConflictingOutputFlags(t *testing.T) {
	oldJSON, oldMD := compareJSONOutput, compareMarkdownOutput
	defer func() {
		compareJSONOutput = oldJSON
		compareMarkdownOutput = oldMD
	}()

	compareJSONOutput = true
	compareMarkdownOutput = true

	err := runCompare(&cobra.Command{}, []string{"one.yml", "two.yml"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be used together")
}

func TestRunCompare_ValidConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	fileA := filepath.Join(tmpDir, "a.yml")
	fileB := filepath.Join(tmpDir, "b.yml")
	err := os.WriteFile(fileA, []byte(createValidConfig()), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(fileB, []byte(createValidConfig()), 0644)
	assert.NoError(t, err)

	err = runCompare(&cobra.Command{}, []string{fileA, fileB})
	assert.NoError(t, err)
}

func TestCompareCmd_Flags(t *testing.T) {
	assert.NotNil(t, compareCmd)
	flags := compareCmd.Flags()
	assert.NotNil(t, flags.Lookup("json"))
	assert.NotNil(t, flags.Lookup("markdown"))
}

func TestRunCompare_PropagatesConfigErrors(t *testing.T) {
	err := runCompare(&cobra.Command{}, []string{"/nonexistent/a.yml", "/nonexistent/b.yml"})
	assert.Error(t, err)

	var exitErr *ExitError
	assert.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
}
