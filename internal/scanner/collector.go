package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var excludedConfigDirectories = map[string]struct{}{
	".codex":              {},
	".entire":             {},
	".git":                {},
	"benchmark_scenarios": {},
	"docs":                {},
	"examples":            {},
	"fixtures":            {},
	"graphify-out":        {},
	"node_modules":        {},
	"provider_catalog":    {},
	"test":                {},
	"tests":               {},
	"vendor":              {},
}

// Collector gathers configuration files from a project directory
type Collector struct {
	MaxFileSize int64 // Maximum file size in bytes (default 1MB)
}

// NewCollector creates a new file collector with default settings
func NewCollector() *Collector {
	return &Collector{
		MaxFileSize: 1024 * 1024, // 1MB default
	}
}

// Collect scans the directory and returns configuration files
func (c *Collector) Collect(rootPath string) (map[string]string, error) {
	files := make(map[string]string)

	rootInfo, err := os.Stat(rootPath)
	if err != nil {
		return files, fmt.Errorf("invalid root path: %w", err)
	}
	if !rootInfo.IsDir() {
		return files, fmt.Errorf("invalid root path: not a directory: %s", rootPath)
	}

	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Continue on errors but log them
			return nil
		}

		if d.IsDir() {
			if path != rootPath && c.shouldSkipDirectory(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a configuration file we want
		if c.isConfigFile(d.Name()) {
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return nil // Skip files we can't make relative
			}

			content, err := c.readFile(path)
			if err != nil {
				// Continue on read errors, return partial results
				return nil
			}

			files[relPath] = content
		}

		return nil
	})

	if err != nil {
		return files, fmt.Errorf("error walking directory: %w", err)
	}

	return files, nil
}

func (c *Collector) shouldSkipDirectory(name string) bool {
	_, excluded := excludedConfigDirectories[name]
	return excluded
}

// isConfigFile checks if the filename matches our target configuration files
func (c *Collector) isConfigFile(filename string) bool {
	configFiles := []string{
		"go.mod",
		"go.sum",
		"package.json",
		"requirements.txt",
		"Dockerfile",
		"docker-compose.yml",
		"docker-compose.yaml",
	}

	// Check exact matches
	for _, config := range configFiles {
		if filename == config {
			return true
		}
	}

	// Check extensions
	extensions := []string{".tf", ".yaml", ".yml"}
	for _, ext := range extensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}

	return false
}

// readFile reads a file with size limit checking
func (c *Collector) readFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.Size() > c.MaxFileSize {
		return "", fmt.Errorf("file too large: %s (%d bytes > %d limit)",
			path, info.Size(), c.MaxFileSize)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
