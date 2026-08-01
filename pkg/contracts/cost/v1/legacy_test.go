package v1_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/internal/cost"
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
	before := cost.NewCostEngine(cfg).CalculateTotalCosts(100, 1)

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
	}

	after := cost.NewCostEngine(cfg).CalculateTotalCosts(100, 1)
	require.Equal(t, before, after)
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

func TestCompatibilityFixtureIsStableJSON(t *testing.T) {
	path := filepath.Join(repoRoot(t), "test", "fixtures", "cost_contract", "v1", "legacy_valid_config_mapping.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.True(t, json.Valid(data))
}
