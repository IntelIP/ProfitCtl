package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const defaultCatalogPath = "provider_catalog/ai_saas_defaults.yml"

type catalogProbe struct {
	CatalogVersion string `yaml:"catalog_version"`
	Status         string `yaml:"status"`
	Entries        []struct {
		ID       string `yaml:"id"`
		Provider string `yaml:"provider"`
		Service  string `yaml:"service"`
		Unit     string `yaml:"unit"`
	} `yaml:"entries"`
}

var doctorCatalog string

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check local profitctl prerequisites",
	Long:  "Check binary identity, runtime support, configuration, and provider catalog readiness without changing local state.",
	Args:  cobra.NoArgs,
	RunE:  runDoctor,
}

func init() {
	doctorCmd.Flags().StringVar(&doctorCatalog, "catalog", defaultCatalogPath, "Provider catalog file path")
}

func runDoctor(cmd *cobra.Command, args []string) error {
	cfgFile, err := cmd.Flags().GetString("file")
	if err != nil {
		return wrapExit(2, fmt.Errorf("read --file: %w", err))
	}
	if strings.TrimSpace(cfgFile) == "" {
		cfgFile = "profit.yml"
	}

	failures := 0
	warnings := 0
	report := func(status, name, detail string) {
		switch status {
		case "fail":
			failures++
		case "warn":
			warnings++
		}
		fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %s\n", status, name, detail)
	}

	executable, executableErr := os.Executable()
	if executableErr != nil {
		report("fail", "binary", executableDetail(executable, executableErr))
	} else {
		report("ok", "binary", executable)
	}

	versionDetail := buildVersion
	if buildVersion == "dev" {
		versionDetail += " (developer build; release builds inject the package tag)"
	}
	if strings.TrimSpace(buildVersion) == "" {
		report("fail", "version", "missing build identity")
	} else if buildVersion == "dev" {
		report("warn", "version", versionDetail)
	} else {
		report("ok", "version", versionDetail)
	}

	runtimeDetail := runtime.GOOS + "/" + runtime.GOARCH
	if !supportedRuntime(runtime.GOOS, runtime.GOARCH) {
		runtimeDetail += " is not in the supported release matrix"
		report("fail", "runtime", runtimeDetail)
	} else {
		report("ok", "runtime", runtimeDetail)
	}

	_, configErr := config.ParseConfig(cfgFile)
	if configErr != nil {
		report("fail", "config", fmt.Sprintf("%s: %v; create or select one with `profitctl init --file %s`", cfgFile, configErr, cfgFile))
	} else {
		report("ok", "config", cfgFile)
	}

	catalog, catalogErr := readCatalog(doctorCatalog)
	if catalogErr != nil {
		report("fail", "catalog", fmt.Sprintf("%s: %v; pass a readable catalog with `--catalog PATH`", doctorCatalog, catalogErr))
	} else {
		fresh, freshness := catalogFreshness(catalog.CatalogVersion, time.Now())
		detail := fmt.Sprintf("%s (%s, %d entries; %s)", doctorCatalog, catalog.CatalogVersion, len(catalog.Entries), freshness)
		if fresh {
			report("ok", "catalog", detail)
		} else {
			report("warn", "catalog", detail+"; stale catalog cannot support current-price claims")
		}
	}

	if failures > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "doctor found %d failed check(s); no fallback or local state change was performed\n", failures)
		return wrapExitSilent(2, fmt.Errorf("doctor found %d failed check(s)", failures))
	}

	if warnings > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "doctor passed with %d warning(s)\n", warnings)
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "doctor passed")
	}
	return nil
}

func executableDetail(path string, err error) string {
	if err != nil {
		return err.Error()
	}
	return path
}

func supportedRuntime(goos, goarch string) bool {
	switch goos + "/" + goarch {
	case "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64":
		return true
	default:
		return false
	}
}

func catalogFreshness(catalogVersion string, now time.Time) (bool, string) {
	capturedAt, err := time.Parse("2006-01-02", catalogVersion)
	if err != nil {
		return false, "capture date is not YYYY-MM-DD"
	}
	ageDays := int(now.UTC().Sub(capturedAt).Hours() / 24)
	if ageDays < 0 {
		return false, "capture date is in the future"
	}
	if ageDays > 30 {
		return false, fmt.Sprintf("%d days old; 30-day freshness window exceeded", ageDays)
	}
	return true, fmt.Sprintf("%d days old", ageDays)
}

func readCatalog(path string) (*catalogProbe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var catalog catalogProbe
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	if strings.TrimSpace(catalog.CatalogVersion) == "" {
		return nil, errors.New("catalog_version is required")
	}
	if strings.TrimSpace(catalog.Status) == "" {
		return nil, errors.New("status is required")
	}
	if len(catalog.Entries) == 0 {
		return nil, errors.New("at least one entry is required")
	}
	for i, entry := range catalog.Entries {
		if strings.TrimSpace(entry.ID) == "" || strings.TrimSpace(entry.Provider) == "" ||
			strings.TrimSpace(entry.Service) == "" || strings.TrimSpace(entry.Unit) == "" {
			return nil, fmt.Errorf("entry %d requires id, provider, service, and unit", i+1)
		}
	}
	return &catalog, nil
}
