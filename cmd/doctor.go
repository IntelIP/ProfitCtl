package cmd

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/pkg/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type catalogProbe struct {
	CatalogVersion string `yaml:"catalog_version"`
	Status         string `yaml:"status"`
	Entries        []struct {
		ID                  string   `yaml:"id"`
		Provider            string   `yaml:"provider"`
		Service             string   `yaml:"service"`
		Unit                string   `yaml:"unit"`
		Value               *float64 `yaml:"value"`
		Currency            string   `yaml:"currency"`
		ConfidenceRationale string   `yaml:"confidence_rationale"`
		RefreshOwner        string   `yaml:"refresh_owner"`
		RefreshCadence      string   `yaml:"refresh_cadence"`
		StaleAfter          string   `yaml:"stale_after"`
		Note                string   `yaml:"note"`
		Source              struct {
			Type             string `yaml:"type"`
			URL              string `yaml:"url"`
			ArtifactIdentity string `yaml:"artifact_identity"`
			Confidence       string `yaml:"confidence"`
			CapturedAt       string `yaml:"captured_at"`
		} `yaml:"source"`
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
	doctorCmd.Flags().StringVar(&doctorCatalog, "catalog", "", "Provider catalog file path (required)")
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
	} else if !supportedExecutableName(executable, runtime.GOOS) {
		report("warn", "binary", fmt.Sprintf("%s has unexpected executable name; supported name is profitctl", executable))
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

	parsedConfig, configErr := config.ParseConfig(cfgFile)
	if configErr != nil {
		report("fail", "config", fmt.Sprintf("%s: %v; create or select one with `profitctl init --file %s`", cfgFile, configErr, cfgFile))
	} else if parsedConfig.Simulation == nil {
		report("fail", "config", fmt.Sprintf("%s: simulation configuration is required", cfgFile))
	} else {
		report("ok", "config", cfgFile)
	}

	if strings.TrimSpace(doctorCatalog) == "" {
		report("fail", "catalog", "catalog path is required; pass a provenance-complete catalog with `--catalog PATH`")
	} else {
		catalog, catalogErr := readCatalog(doctorCatalog)
		if catalogErr != nil {
			report("fail", "catalog", fmt.Sprintf("%s: %v; pass a provenance-complete catalog with `--catalog PATH`", doctorCatalog, catalogErr))
		} else {
			fresh, freshness := catalogFreshness(catalog, time.Now())
			detail := fmt.Sprintf("%s (%s, %d entries; %s)", doctorCatalog, catalog.CatalogVersion, len(catalog.Entries), freshness)
			if fresh {
				report("ok", "catalog", detail)
			} else {
				report("fail", "catalog", detail+"; stale catalog cannot support current-price or release-readiness claims")
			}
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

func supportedExecutableName(path, goos string) bool {
	name := filepath.Base(strings.ReplaceAll(path, `\`, "/"))
	if goos == "windows" {
		return strings.EqualFold(name, "profitctl.exe")
	}
	return name == "profitctl"
}

func catalogFreshness(catalog *catalogProbe, now time.Time) (bool, string) {
	for _, entry := range catalog.Entries {
		staleAfter, err := time.Parse("2006-01-02", entry.StaleAfter)
		if err != nil {
			return false, fmt.Sprintf("entry %s stale_after is not YYYY-MM-DD", entry.ID)
		}
		if now.UTC().After(staleAfter.Add(24*time.Hour - time.Nanosecond)) {
			return false, fmt.Sprintf("entry %s exceeded stale_after %s", entry.ID, entry.StaleAfter)
		}

		capturedAt, err := parseCapturedAt(entry.Source.CapturedAt)
		if err != nil {
			return false, fmt.Sprintf("entry %s source captured_at is invalid", entry.ID)
		}
		age := now.UTC().Sub(capturedAt)
		switch types.CostSourceType(entry.Source.Type) {
		case types.CostSourceProviderCatalog:
			if age > 7*24*time.Hour {
				return false, fmt.Sprintf("entry %s provider catalog capture exceeds the 7-day decision-grade window", entry.ID)
			}
		case types.CostSourceTemplate:
			if age > 90*24*time.Hour {
				return false, fmt.Sprintf("entry %s template capture exceeds the 90-day review window", entry.ID)
			}
		}
	}
	return true, fmt.Sprintf("%d entries within declared stale-after dates", len(catalog.Entries))
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
		if entry.Value == nil || strings.TrimSpace(entry.Currency) == "" {
			return nil, fmt.Errorf("entry %s requires value and currency", entry.ID)
		}
		if *entry.Value < 0 || math.IsNaN(*entry.Value) || math.IsInf(*entry.Value, 0) {
			return nil, fmt.Errorf("entry %s value must be finite and non-negative", entry.ID)
		}
		if strings.TrimSpace(entry.Source.Type) == "" || strings.TrimSpace(entry.Source.Confidence) == "" ||
			strings.TrimSpace(entry.Source.CapturedAt) == "" {
			return nil, fmt.Errorf("entry %s requires source type, confidence, and captured_at", entry.ID)
		}
		if !supportedSourceType(entry.Source.Type) {
			return nil, fmt.Errorf("entry %s has unsupported source type %q", entry.ID, entry.Source.Type)
		}
		if !supportedSourceConfidence(entry.Source.Confidence) {
			return nil, fmt.Errorf("entry %s has unsupported source confidence %q", entry.ID, entry.Source.Confidence)
		}
		if entry.Source.Type == string(types.CostSourceTemplate) && entry.Source.Confidence == "high" {
			return nil, fmt.Errorf("entry %s template source cannot use high confidence", entry.ID)
		}
		if strings.TrimSpace(entry.Source.URL) == "" && strings.TrimSpace(entry.Source.ArtifactIdentity) == "" {
			return nil, fmt.Errorf("entry %s requires source url or artifact_identity", entry.ID)
		}
		if strings.TrimSpace(entry.Source.URL) != "" && !validSourceURI(entry.Source.URL) {
			return nil, fmt.Errorf("entry %s source url must be an absolute URI", entry.ID)
		}
		capturedAt, err := parseCapturedAt(entry.Source.CapturedAt)
		if err != nil {
			return nil, fmt.Errorf("entry %s source captured_at must be YYYY-MM-DD or RFC3339", entry.ID)
		}
		if capturedAt.After(time.Now().UTC()) {
			return nil, fmt.Errorf("entry %s source captured_at cannot be in the future", entry.ID)
		}
		if strings.TrimSpace(entry.ConfidenceRationale) == "" || strings.TrimSpace(entry.Note) == "" {
			return nil, fmt.Errorf("entry %s requires confidence_rationale and note", entry.ID)
		}
		if strings.TrimSpace(entry.RefreshOwner) == "" || strings.TrimSpace(entry.RefreshCadence) == "" ||
			strings.TrimSpace(entry.StaleAfter) == "" {
			return nil, fmt.Errorf("entry %s requires refresh_owner, refresh_cadence, and stale_after", entry.ID)
		}
		if _, err := time.Parse("2006-01-02", entry.StaleAfter); err != nil {
			return nil, fmt.Errorf("entry %s stale_after must be YYYY-MM-DD", entry.ID)
		}
	}
	return &catalog, nil
}

func validSourceURI(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || !parsed.IsAbs() {
		return false
	}
	if (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host == "" {
		return false
	}
	return true
}

func parseCapturedAt(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func supportedSourceType(value string) bool {
	switch types.CostSourceType(value) {
	case types.CostSourceTemplate,
		types.CostSourceUserSupplied,
		types.CostSourceRepoDetected,
		types.CostSourceTelemetry,
		types.CostSourceInvoice,
		types.CostSourceProviderCatalog:
		return true
	default:
		return false
	}
}

func supportedSourceConfidence(value string) bool {
	switch types.CostSourceConfidence(value) {
	case types.CostSourceConfidenceLow,
		types.CostSourceConfidenceMedium,
		types.CostSourceConfidenceHigh:
		return true
	default:
		return false
	}
}
