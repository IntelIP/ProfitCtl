package systemmodel

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/IntelIP/ProfitCtl/internal/scanner"
)

type FileSignal struct {
	Path      string   `json:"path"`
	SHA256    string   `json:"sha256"`
	Providers []string `json:"provider_signals"`
}
type Inventory struct {
	SchemaVersion string       `json:"schema_version"`
	Files         []FileSignal `json:"files"`
	Limitations   []string     `json:"limitations"`
}

func Inspect(root string) (Inventory, error) {
	c := scanner.NewCollector()
	c.IncludeSource = true
	c.Strict = true
	c.MaxFiles = 2048
	c.MaxTotalSize = 32 * 1024 * 1024
	files, err := c.Collect(root)
	if err != nil {
		return Inventory{}, err
	}
	result := Inventory{SchemaVersion: SchemaVersion, Files: []FileSignal{}, Limitations: []string{
		"Signals are source evidence, not confirmed deployed services or dependency edges.",
		"Development and retired configuration require review against deployment authority.",
		"Only configuration and Python source are inspected; docs, generated trees and symlinks are excluded.",
		"No file contents or credential values are emitted. Do not send raw collected content to a provider without separate review.",
	}}
	// Deliberately lexical: names are candidates, not inferred runtime topology.
	markers := map[string][]string{"gcp": {"cloud.google.com", "gcloud", "google.cloud"}, "neon": {"neon.tech", "neon.com", "neon"}, "openrouter": {"openrouter"}, "exa": {"exa.ai", "exa_api", "exapublic"}, "modal": {"modal"}, "upstash": {"upstash"}, "vercel": {"vercel"}, "cloudflare": {"cloudflare"}}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		content := strings.ToLower(files[path])
		row := FileSignal{Path: path, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(files[path]))), Providers: []string{}}
		for provider, terms := range markers {
			for _, term := range terms {
				if strings.Contains(content, term) {
					row.Providers = append(row.Providers, provider)
					break
				}
			}
		}
		sort.Strings(row.Providers)
		result.Files = append(result.Files, row)
	}
	return result, nil
}
