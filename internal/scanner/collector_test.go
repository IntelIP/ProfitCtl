package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()
	assert.NotNil(t, collector)
	assert.Equal(t, int64(1024*1024), collector.MaxFileSize)
}

func TestCollector_isConfigFile(t *testing.T) {
	collector := NewCollector()

	tests := []struct {
		filename string
		expected bool
	}{
		{"go.mod", true},
		{"go.sum", true},
		{"package.json", true},
		{"requirements.txt", true},
		{"Dockerfile", true},
		{"docker-compose.yml", true},
		{"docker-compose.yaml", true},
		{"main.tf", true},
		{"config.yaml", true},
		{"settings.yml", true},
		{"main.go", false},
		{"README.md", false},
		{"script.sh", false},
		{".gitignore", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := collector.isConfigFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCollector_Collect(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"go.mod":                "module test\n\ngo 1.21\n\nrequire github.com/test v1.0.0",
		"package.json":          `{"name": "test", "version": "1.0.0"}`,
		"config.yaml":           "key: value\n",
		"main.go":               "package main\n\nfunc main() {}",
		"README.md":             "# Test Project",
		"subdir/nested.json":    `{"nested": true}`,
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	collector := NewCollector()
	files, err := collector.Collect(tempDir)

	assert.NoError(t, err)
	assert.NotEmpty(t, files)

	// Should collect config files but not source files
	expectedFiles := map[string]bool{
		"go.mod":       true,
		"package.json": true,
		"config.yaml":  true,
	}

	for relPath := range files {
		if expectedFiles[relPath] {
			continue
		}
		// Should not have collected main.go, README.md
		assert.NotContains(t, relPath, "main.go")
		assert.NotContains(t, relPath, "README.md")
	}

	// Verify content
	assert.Equal(t, testFiles["go.mod"], files["go.mod"])
	assert.Equal(t, testFiles["package.json"], files["package.json"])
	assert.Equal(t, testFiles["config.yaml"], files["config.yaml"])
}

func TestCollector_Collect_FileSizeLimit(t *testing.T) {
	tempDir := t.TempDir()

	collector := NewCollector()
	collector.MaxFileSize = 10 // Very small limit

	// Create a file larger than the limit
	largeContent := strings.Repeat("x", 20)
	largeFile := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(largeFile, []byte(largeContent), 0644))

	// Create a small file
	smallFile := filepath.Join(tempDir, "go.mod")
	require.NoError(t, os.WriteFile(smallFile, []byte("module test"), 0644))

	files, err := collector.Collect(tempDir)

	assert.NoError(t, err)
	// Should collect the small file but not the large one
	assert.Contains(t, files, "go.mod")
	assert.NotContains(t, files, "config.yaml")
}

func TestCollector_Collect_NonExistentDirectory(t *testing.T) {
	collector := NewCollector()
	files, err := collector.Collect("/non/existent/path")

	assert.Error(t, err)
	assert.Empty(t, files)
}

func TestCollector_readFile(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello World"
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0644))

	collector := NewCollector()
	content, err := collector.readFile(testFile)

	assert.NoError(t, err)
	assert.Equal(t, testContent, content)
}

func TestCollector_readFile_SizeLimit(t *testing.T) {
	tempDir := t.TempDir()

	collector := NewCollector()
	collector.MaxFileSize = 5

	testFile := filepath.Join(tempDir, "large.txt")
	largeContent := "This is too long"
	require.NoError(t, os.WriteFile(testFile, []byte(largeContent), 0644))

	_, err := collector.readFile(testFile)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file too large")
}

func BenchmarkCollector_Collect(b *testing.B) {
	// Create a test directory with multiple files
	tempDir := b.TempDir()

	// Create test files
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("config%d.yaml", i))
		content := fmt.Sprintf("key%d: value%d\n", i, i)
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			b.Fatal(err)
		}
	}

	collector := NewCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = collector.Collect(tempDir)
	}
}