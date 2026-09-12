package systemmodel

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/scanner"
)

func TestInspectCoverageAndPrivacy(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"pyproject.toml", "deployment/Dockerfile.agentos", "src/runtime.py", ".next/package.json", ".venv/provider/config.yml", "docs/retired.yaml"} {
		path := filepath.Join(root, p)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
		require.NoError(t, os.WriteFile(path, []byte("modal = 'DO_NOT_EMIT_SECRET'"), 0600))
	}
	outside := filepath.Join(t.TempDir(), "secret.yml")
	require.NoError(t, os.WriteFile(outside, []byte("sensitive"), 0600))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "linked.yml")))
	r, err := Inspect(root)
	require.NoError(t, err)
	require.Len(t, r.Files, 3)
	for _, f := range r.Files {
		require.Equal(t, []string{"modal"}, f.Providers)
		require.Len(t, f.SHA256, 64)
	}
	out, err := json.Marshal(r)
	require.NoError(t, err)
	require.NotContains(t, string(out), "DO_NOT_EMIT_SECRET")
	require.NotContains(t, string(out), "linked.yml")
}

func TestInspectionBounds(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.yml"), []byte("1234"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "b.yml"), []byte("5678"), 0600))
	for _, configure := range []func(*scanner.Collector){
		func(c *scanner.Collector) { c.MaxFiles = 1 },
		func(c *scanner.Collector) { c.MaxTotalSize = 7 },
		func(c *scanner.Collector) { c.MaxFileSize = 3 },
	} {
		c := scanner.NewCollector()
		c.Strict = true
		configure(c)
		_, err := c.Collect(root)
		require.Error(t, err)
	}
	linked := filepath.Join(t.TempDir(), "linked")
	require.NoError(t, os.Symlink(root, linked))
	_, err := Inspect(linked)
	require.Error(t, err)
}
