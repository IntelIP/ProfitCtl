package v1_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/config"
	v1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
	"github.com/IntelIP/ProfitCtl/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestLegacyScenarioMapsForwardWithoutChangingCosts(t *testing.T) {
	var fixture struct {
		SourceScenario string           `json:"source_scenario"`
		Mapping        v1.LegacyMapping `json:"mapping"`
		Expected       []struct {
			Name    string        `json:"name"`
			Kind    v1.DriverKind `json:"kind"`
			Unit    string        `json:"unit"`
			PerUnit string        `json:"per_unit"`
		} `json:"expected"`
	}
	decodeFixture(t, "legacy_valid_config_mapping.json", &fixture)

	root := repoRoot(t)
	cfg, err := config.ParseConfig(filepath.Join(root, fixture.SourceScenario))
	require.NoError(t, err)
	before := *cfg

	drivers, err := v1.MapLegacyCosts(cfg.FixedCosts, cfg.VariableCosts, fixture.Mapping)
	require.NoError(t, err)
	require.Len(t, drivers, len(fixture.Expected))

	for i, expected := range fixture.Expected {
		require.NoError(t, drivers[i].Validate())
		require.Equal(t, expected.Name, drivers[i].Name)
		require.Equal(t, expected.Kind, drivers[i].Kind)
		require.Equal(t, expected.Unit, drivers[i].Quantity.Unit)
		require.NotNil(t, drivers[i].Per)
		require.Equal(t, expected.PerUnit, drivers[i].Per.Unit)
		if expected.Kind == v1.DriverVariable {
			require.NotNil(t, drivers[i].Distribution)
		}
	}

	require.Equal(t, before, *cfg)
}

func TestLegacyMappingPreservesStochasticVariableBasis(t *testing.T) {
	rate := 0.001
	drivers, err := v1.MapLegacyCosts(nil, []types.VariableCost{{
		Name:         "LLM Tokens",
		CostPerUnit:  0.01,
		UnitsPerUser: 5000,
		Distribution: types.DistExponential,
		Rate:         &rate,
	}}, v1.LegacyMapping{
		ArtifactIdentity: "stress.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
		VariableUnits: map[string]string{"LLM Tokens": "token"},
	})
	require.NoError(t, err)
	require.NotNil(t, drivers[0].Distribution)
	require.Equal(t, v1.DistributionExponential, drivers[0].Distribution.Type)
	require.Equal(t, rate, *drivers[0].Distribution.Rate)
	require.NotEqual(t, drivers[0].Quantity.Value, 1/rate)
}

func TestLegacyMappingPreservesNormalNonNegativeFloor(t *testing.T) {
	mean, stddev := 1.0, 10.0
	drivers, err := v1.MapLegacyCosts(nil, []types.VariableCost{{
		Name:         "Bursty requests",
		CostPerUnit:  0.01,
		UnitsPerUser: 1,
		Distribution: types.DistNormal,
		Mean:         &mean,
		StdDev:       &stddev,
	}}, v1.LegacyMapping{
		ArtifactIdentity: "normal.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
		VariableUnits: map[string]string{"Bursty requests": "request"},
	})
	require.NoError(t, err)
	require.NotNil(t, drivers[0].Distribution.Floor)
	require.Zero(t, *drivers[0].Distribution.Floor)
	*drivers[0].Distribution.Mean = 99
	require.Equal(t, 1.0, mean)
}

func TestLegacyMappingPreservesUniformNonNegativeFloor(t *testing.T) {
	minimum, maximum := -10.0, 10.0
	drivers, err := v1.MapLegacyCosts(nil, []types.VariableCost{{
		Name:         "Burst range",
		CostPerUnit:  0.01,
		UnitsPerUser: 1,
		Distribution: types.DistUniform,
		Min:          &minimum,
		Max:          &maximum,
	}}, v1.LegacyMapping{
		ArtifactIdentity: "uniform.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
		VariableUnits: map[string]string{"Burst range": "request"},
	})
	require.NoError(t, err)
	require.Equal(t, minimum, *drivers[0].Distribution.Min)
	require.NotNil(t, drivers[0].Distribution.Floor)
	require.Zero(t, *drivers[0].Distribution.Floor)
}

func TestLegacyMappingUsesStableIDsAcrossReordering(t *testing.T) {
	first := types.FixedCost{Name: "Database", Amount: 20, Period: types.PeriodMonthly}
	second := types.FixedCost{Name: "Worker", Amount: 30, Period: types.PeriodMonthly}
	mapping := v1.LegacyMapping{
		ArtifactIdentity: "stable.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
	}
	original, err := v1.MapLegacyCosts([]types.FixedCost{first, second}, nil, mapping)
	require.NoError(t, err)
	reordered, err := v1.MapLegacyCosts([]types.FixedCost{second, first}, nil, mapping)
	require.NoError(t, err)
	require.Equal(t, original[0].ID, reordered[1].ID)
	require.Equal(t, original[1].ID, reordered[0].ID)
}

func TestLegacyMappingCapsHighConfidenceTemplate(t *testing.T) {
	drivers, err := v1.MapLegacyCosts([]types.FixedCost{{
		Name:   "Template database",
		Amount: 20,
		Period: types.PeriodMonthly,
		Source: &types.CostSource{
			Type:       types.CostSourceTemplate,
			Confidence: types.CostSourceConfidenceHigh,
		},
	}}, nil, v1.LegacyMapping{
		ArtifactIdentity: "template.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
	})
	require.NoError(t, err)
	require.Equal(t, v1.SourceTemplate, drivers[0].Evidence.Source.Type)
	require.Equal(t, v1.ConfidenceMedium, drivers[0].Evidence.Confidence)
}

func TestLegacyMappingCapsHighConfidenceRepoDetected(t *testing.T) {
	drivers, err := v1.MapLegacyCosts([]types.FixedCost{{
		Name:   "Detected worker",
		Amount: 20,
		Period: types.PeriodMonthly,
		Source: &types.CostSource{
			Type:       types.CostSourceRepoDetected,
			Confidence: types.CostSourceConfidenceHigh,
		},
	}}, nil, v1.LegacyMapping{
		ArtifactIdentity: "detected.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
	})
	require.NoError(t, err)
	require.Equal(t, v1.ConfidenceMedium, drivers[0].Evidence.Confidence)
}

func TestLegacyMappingFallsBackFromFreeFormCaptureText(t *testing.T) {
	drivers, err := v1.MapLegacyCosts([]types.FixedCost{{
		Name:   "Database",
		Amount: 20,
		Period: types.PeriodMonthly,
		Source: &types.CostSource{
			Type:       types.CostSourceUserSupplied,
			Confidence: types.CostSourceConfidenceMedium,
			CapturedAt: "May 2026",
		},
	}}, nil, v1.LegacyMapping{
		ArtifactIdentity: "capture.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "2026-08-01", drivers[0].Evidence.Source.CapturedAt)
	require.Contains(t, drivers[0].Evidence.ConfidenceRationale, "May 2026")
}

func TestLegacyMappingFailsClosedWithoutVariableUnit(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "test", "fixtures", "valid_config.yml"))
	require.NoError(t, err)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "profit.yml")
	require.NoError(t, os.WriteFile(configPath, data, 0o600))
	cfg, err := config.ParseConfig(configPath)
	require.NoError(t, err)

	_, err = v1.MapLegacyCosts(cfg.FixedCosts, cfg.VariableCosts, v1.LegacyMapping{
		ArtifactIdentity: "profit.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
	})
	require.ErrorContains(t, err, "variable cost")
	require.ErrorContains(t, err, "unit")
}

func TestLegacyMappingPreservesPaidUserScope(t *testing.T) {
	drivers, err := v1.MapLegacyCosts(nil, []types.VariableCost{{
		Name:         "Payment processing",
		CostPerUnit:  0.30,
		UnitsPerUser: 1,
		UserScope:    types.UserScopePaidUsers,
	}}, v1.LegacyMapping{
		ArtifactIdentity: "paid.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
		VariableUnits: map[string]string{"Payment processing": "payment"},
	})
	require.NoError(t, err)
	require.Equal(t, "paid_user", drivers[0].Per.Unit)
}

func TestLegacyMappingPreservesThirtyDayDailyNormalization(t *testing.T) {
	drivers, err := v1.MapLegacyCosts([]types.FixedCost{{
		Name:   "Daily worker",
		Amount: 2,
		Period: types.PeriodDaily,
	}}, nil, v1.LegacyMapping{
		ArtifactIdentity: "daily.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-07-01T00:00:00Z",
			End:   "2026-08-01T00:00:00Z",
		},
	})
	require.NoError(t, err)
	require.Equal(t, float64(30), drivers[0].Quantity.Value)
	require.Equal(t, "month", drivers[0].Per.Unit)
	require.Equal(t, float64(60), drivers[0].Quantity.Value*drivers[0].UnitPrice.Amount.Amount)
}

func TestLegacyMappingDowngradesIncompleteProviderCatalogProvenance(t *testing.T) {
	drivers, err := v1.MapLegacyCosts([]types.FixedCost{{
		Name:   "Catalog-backed database",
		Amount: 20,
		Period: types.PeriodMonthly,
		Source: &types.CostSource{
			Type:       types.CostSourceProviderCatalog,
			URL:        "https://example.com/pricing",
			CapturedAt: "2026-07-31",
			Confidence: types.CostSourceConfidenceMedium,
		},
	}}, nil, v1.LegacyMapping{
		ArtifactIdentity: "legacy.yml",
		CapturedAt:       "2026-08-01",
		Currency:         "USD",
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
	})
	require.NoError(t, err)
	require.Equal(t, v1.SourceLegacyScenario, drivers[0].Evidence.Source.Type)
	require.Equal(t, "https://example.com/pricing", drivers[0].Evidence.Source.URL)
	require.Equal(t, v1.ConfidenceLow, drivers[0].Evidence.Confidence)
	require.Contains(t, drivers[0].Evidence.ConfidenceRationale, "without v1 refresh policy")
}

func TestCompatibilityFixtureIsStableJSON(t *testing.T) {
	path := filepath.Join(repoRoot(t), "test", "fixtures", "cost_contract", "v1", "legacy_valid_config_mapping.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.True(t, json.Valid(data))
}
