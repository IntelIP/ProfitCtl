package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/scanner/llm"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubProvider struct {
	response string
	err      error
}

func (s *stubProvider) Chat(ctx context.Context, messages []llm.Message) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.response, nil
}

func (s *stubProvider) IsAvailable(ctx context.Context) bool {
	return s.err == nil
}

func TestRunDetect_MissingAPIKey(t *testing.T) {
	oldKey := os.Getenv("OPENROUTER_API_KEY")
	require.NoError(t, os.Unsetenv("OPENROUTER_API_KEY"))
	t.Cleanup(func() { _ = os.Setenv("OPENROUTER_API_KEY", oldKey) })

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	err := runDetect(cmd, nil)

	assert.Error(t, err)
	var exitErr *ExitError
	assert.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 2, exitErr.Code)
	assert.Contains(t, err.Error(), "OPENROUTER_API_KEY")
}

func TestRunDetect_InvalidProviderResponse(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\ngo 1.21\n"), 0644))

	oldPath, oldProvider, oldModel, oldOut := detectPath, detectProvider, detectModel, detectOut
	oldNewProvider := newProvider
	oldKey := os.Getenv("OPENROUTER_API_KEY")
	defer func() {
		detectPath, detectProvider, detectModel, detectOut = oldPath, oldProvider, oldModel, oldOut
		newProvider = oldNewProvider
		_ = os.Setenv("OPENROUTER_API_KEY", oldKey)
	}()

	detectPath = tmpDir
	detectProvider = "openrouter"
	detectModel = "test-model"
	detectOut = filepath.Join(tmpDir, "report.json")
	require.NoError(t, os.Setenv("OPENROUTER_API_KEY", "test-key"))

	newProvider = func(apiKey, model string) (llm.LLMProvider, error) {
		return &stubProvider{response: "not-json"}, nil
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	err := runDetect(cmd, nil)

	assert.Error(t, err)
	var exitErr *ExitError
	assert.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 3, exitErr.Code)
	assert.Contains(t, err.Error(), "invalid provider response")
}

func TestRunDetect_WritesReport(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\ngo 1.21\n"), 0644))

	oldPath, oldProvider, oldModel, oldOut := detectPath, detectProvider, detectModel, detectOut
	oldNewProvider := newProvider
	oldKey := os.Getenv("OPENROUTER_API_KEY")
	defer func() {
		detectPath, detectProvider, detectModel, detectOut = oldPath, oldProvider, oldModel, oldOut
		newProvider = oldNewProvider
		_ = os.Setenv("OPENROUTER_API_KEY", oldKey)
	}()

	reportPath := filepath.Join(tmpDir, "report.json")
	detectPath = tmpDir
	detectProvider = "openrouter"
	detectModel = "test-model"
	detectOut = reportPath
	require.NoError(t, os.Setenv("OPENROUTER_API_KEY", "test-key"))

	newProvider = func(apiKey, model string) (llm.LLMProvider, error) {
		return &stubProvider{response: `{"services":[],"dependencies":[],"patterns":[]}`}, nil
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	err := runDetect(cmd, nil)

	assert.NoError(t, err)
	contents, readErr := os.ReadFile(reportPath)
	require.NoError(t, readErr)
	assert.Contains(t, string(contents), `"provider": "openrouter"`)
	assert.Contains(t, string(contents), `"analysis"`)
}
