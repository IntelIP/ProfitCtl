package v1_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
	"github.com/stretchr/testify/require"
)

func TestAllDriverKindsValidate(t *testing.T) {
	for _, kind := range []v1.DriverKind{
		v1.DriverFixed,
		v1.DriverVariable,
		v1.DriverCadence,
		v1.DriverConcurrency,
		v1.DriverUptime,
	} {
		t.Run(string(kind), func(t *testing.T) {
			driver := validDriver()
			driver.Kind = kind
			if kind == v1.DriverCadence {
				driver.Per = &v1.Quantity{Value: 1, Unit: "minute"}
			}
			require.NoError(t, driver.Validate())
		})
	}
}

func TestDriverRejectsMissingMalformedAndAmbiguousUnits(t *testing.T) {
	for _, unit := range []string{"", "commands / month", "GB", "gb", "units"} {
		t.Run(unit, func(t *testing.T) {
			driver := validDriver()
			driver.Quantity.Unit = unit
			driver.UnitPrice.Per.Unit = unit
			require.Error(t, driver.Validate())
		})
	}
}

func TestDriverRejectsMissingTimeWindow(t *testing.T) {
	driver := validDriver()
	driver.Window = v1.TimeWindow{}
	require.ErrorContains(t, driver.Validate(), "window")
}

func TestScalingDriversRequirePerBasis(t *testing.T) {
	for _, kind := range []v1.DriverKind{v1.DriverVariable, v1.DriverCadence} {
		driver := validDriver()
		driver.Kind = kind
		driver.Per = nil
		require.ErrorContains(t, driver.Validate(), "requires per scale basis")

		driver.Per = &v1.Quantity{Value: 0, Unit: "user"}
		require.ErrorContains(t, driver.Validate(), "per.value must be greater than zero")
	}
}

func TestEveryPresentPerBasisMustBePositive(t *testing.T) {
	for _, kind := range []v1.DriverKind{
		v1.DriverFixed,
		v1.DriverVariable,
		v1.DriverCadence,
		v1.DriverConcurrency,
		v1.DriverUptime,
	} {
		driver := validDriver()
		driver.Kind = kind
		driver.Per = &v1.Quantity{Value: 0, Unit: "user"}
		require.ErrorContains(t, driver.Validate(), "per.value must be greater than zero")
	}
}

func TestCadenceRequiresTimeBasis(t *testing.T) {
	driver := validDriver()
	driver.Kind = v1.DriverCadence
	driver.Per = &v1.Quantity{Value: 1, Unit: "user"}
	require.ErrorContains(t, driver.Validate(), "canonical time unit")

	driver.Per.Unit = "minute"
	require.NoError(t, driver.Validate())
}

func TestUniformDistributionAllowsZeroMinimum(t *testing.T) {
	minimum, maximum := 0.0, 10.0
	driver := validDriver()
	driver.Distribution = &v1.Distribution{
		Type: v1.DistributionUniform,
		Min:  &minimum,
		Max:  &maximum,
	}
	require.NoError(t, driver.Validate())
}

func TestEvidenceDoesNotTreatProvenanceAsMeasurement(t *testing.T) {
	evidence := validEvidence()
	evidence.Measurement = v1.MeasurementMeasured
	evidence.Source.Type = v1.SourceProviderCatalog
	evidence.Source.RefreshOwner = "catalog-maintainer"
	evidence.Source.RefreshCadence = "30d"
	evidence.Source.StaleAfter = "2026-09-01"
	require.ErrorContains(t, evidence.Validate(), "measured evidence requires")
}

func TestBilledEvidenceRequiresInvoice(t *testing.T) {
	evidence := validEvidence()
	evidence.Kind = v1.EvidenceBilled
	require.ErrorContains(t, evidence.Validate(), "invoice")

	evidence.Source.Type = v1.SourceInvoice
	evidence.Measurement = v1.MeasurementDeclared
	require.ErrorContains(t, evidence.Validate(), "must be measured")

	evidence.Measurement = v1.MeasurementMeasured
	require.NoError(t, evidence.Validate())
}

func TestObservationEnforcesClaimSpecificAuthority(t *testing.T) {
	var fixture struct {
		Observation v1.CostObservation `json:"observation"`
	}
	decodeFixture(t, "upstash_idle_polling.json", &fixture)
	fixture.Observation.Evidence.UnitPrice = v1.Evidence{
		Kind:        v1.EvidenceObserved,
		Measurement: v1.MeasurementMeasured,
		Source: v1.SourceReference{
			Type:             v1.SourceTelemetry,
			ArtifactIdentity: "telemetry",
			CapturedAt:       "2026-08-01",
		},
		Confidence:          v1.ConfidenceHigh,
		ConfidenceRationale: "Measured usage only.",
	}
	require.ErrorContains(t, fixture.Observation.Validate(), "not authoritative for unit_price")
}

func TestObservationValidatesDerivedTotalArithmetic(t *testing.T) {
	var fixture struct {
		Observation v1.CostObservation `json:"observation"`
	}
	decodeFixture(t, "upstash_idle_polling.json", &fixture)
	fixture.Observation.Evidence.TotalCost.Source.Type = v1.SourceProfitCtlDerived
	fixture.Observation.Evidence.TotalCost.Source.ArtifactIdentity = "profitctl://derived/upstash-total"
	require.NoError(t, fixture.Observation.Validate())
	fixture.Observation.TotalCost.Amount = 999
	require.ErrorContains(t, fixture.Observation.Validate(), "derived total_cost")
}

func TestProviderCatalogRequiresStalePolicyAndCapsConfidence(t *testing.T) {
	evidence := validEvidence()
	evidence.Source.Type = v1.SourceProviderCatalog
	require.ErrorContains(t, evidence.Validate(), "refresh_owner")

	evidence.Source.RefreshOwner = "catalog-maintainer"
	evidence.Source.RefreshCadence = "30d"
	evidence.Source.StaleAfter = "2026-09-01"
	evidence.Confidence = v1.ConfidenceHigh
	require.ErrorContains(t, evidence.Validate(), "cannot claim high confidence")

	evidence.Confidence = v1.ConfidenceMedium
	require.NoError(t, evidence.Validate())

	evidence.Source.StaleAfter = "2026-07-31"
	require.ErrorContains(t, evidence.Validate(), "must be after")
}

func TestTemplateEvidenceCannotClaimHighConfidence(t *testing.T) {
	evidence := validEvidence()
	evidence.Source.Type = v1.SourceTemplate
	evidence.Confidence = v1.ConfidenceHigh
	require.ErrorContains(t, evidence.Validate(), "template evidence cannot claim high confidence")
}

func TestObservationRejectsDuplicateDriverIDs(t *testing.T) {
	var fixture struct {
		Observation v1.CostObservation `json:"observation"`
	}
	decodeFixture(t, "upstash_idle_polling.json", &fixture)
	fixture.Observation.DriverIDs = append(fixture.Observation.DriverIDs, fixture.Observation.DriverIDs[0])
	require.ErrorContains(t, fixture.Observation.Validate(), "duplicate driver id")
}

func TestUpstashForecastToObservationFixture(t *testing.T) {
	var fixture struct {
		Drivers     []v1.CostDriver    `json:"drivers"`
		Observation v1.CostObservation `json:"observation"`
	}
	decodeFixture(t, "upstash_idle_polling.json", &fixture)

	require.Len(t, fixture.Drivers, 1)
	require.NoError(t, fixture.Drivers[0].Validate())
	require.NoError(t, fixture.Observation.Validate())
	require.Equal(t, fixture.Drivers[0].ID, fixture.Observation.DriverIDs[0])
	require.Equal(t, v1.MeasurementSynthetic, fixture.Observation.Evidence.Quantity.Measurement)
	require.Equal(t, v1.SourceUserSupplied, fixture.Observation.Evidence.UnitPrice.Source.Type)
}

func TestSchemasContainValidJSON(t *testing.T) {
	root := repoRoot(t)
	schemaDir := filepath.Join(root, "schemas", "cost-contract", "v1")
	entries, err := os.ReadDir(schemaDir)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	ids := map[string]bool{}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
		require.NoError(t, err)
		var schema map[string]any
		require.NoError(t, json.Unmarshal(data, &schema), entry.Name())
		id, ok := schema["$id"].(string)
		require.True(t, ok, entry.Name())
		require.False(t, ids[id], entry.Name())
		ids[id] = true
	}
}

func validDriver() v1.CostDriver {
	return v1.CostDriver{
		SchemaVersion: v1.SchemaVersion,
		ID:            "driver-1",
		Name:          "Requests",
		Kind:          v1.DriverVariable,
		Quantity:      v1.Quantity{Value: 10, Unit: "request"},
		Per:           &v1.Quantity{Value: 1, Unit: "user"},
		UnitPrice: v1.UnitPrice{
			Amount: v1.Money{Amount: 0.001, Currency: "USD"},
			Per:    v1.Quantity{Value: 1, Unit: "request"},
		},
		Window: v1.TimeWindow{
			Start: "2026-08-01T00:00:00Z",
			End:   "2026-09-01T00:00:00Z",
		},
		Dimensions: v1.Dimensions{Workload: "api_requests"},
		Evidence:   validEvidence(),
	}
}

func validEvidence() v1.Evidence {
	return v1.Evidence{
		Kind:        v1.EvidencePredicted,
		Measurement: v1.MeasurementDeclared,
		Source: v1.SourceReference{
			Type:             v1.SourceUserSupplied,
			ArtifactIdentity: "test",
			CapturedAt:       "2026-08-01",
		},
		Confidence:          v1.ConfidenceMedium,
		ConfidenceRationale: "Explicit test assumption.",
	}
}

func decodeFixture(t *testing.T, name string, target any) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "test", "fixtures", "cost_contract", "v1", name))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, target))
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	require.NoError(t, err)
	return root
}
