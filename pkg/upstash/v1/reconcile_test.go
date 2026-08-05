package v1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ledgerv1 "github.com/IntelIP/ProfitCtl/pkg/ledger/v1"
	"github.com/stretchr/testify/require"
)

func TestReconcileFixtureCases(t *testing.T) {
	tests := []struct {
		name               string
		fixture            string
		expectedPredicted  float64
		expectedObserved   float64
		expectedDuplicates int
		expectedBilling    BillingEvidenceStatus
		expectedDriver     string
	}{
		{
			name:               "normal load",
			fixture:            "normal_load.json",
			expectedPredicted:  360,
			expectedObserved:   360,
			expectedDuplicates: 0,
			expectedBilling:    BillingEvidenceAvailable,
			expectedDriver:     "active_workload",
		},
		{
			name:               "idle polling",
			fixture:            "idle_polling.json",
			expectedPredicted:  103680,
			expectedObserved:   103680,
			expectedDuplicates: 0,
			expectedBilling:    BillingEvidenceAvailable,
			expectedDriver:     "idle_polling",
		},
		{
			name:               "missing billing",
			fixture:            "missing_billing.json",
			expectedPredicted:  360,
			expectedObserved:   360,
			expectedDuplicates: 0,
			expectedBilling:    BillingEvidenceUnavailable,
			expectedDriver:     "active_workload",
		},
		{
			name:               "duplicate export",
			fixture:            "duplicate_export.json",
			expectedPredicted:  360,
			expectedObserved:   360,
			expectedDuplicates: 1,
			expectedBilling:    BillingEvidenceAvailable,
			expectedDriver:     "active_workload",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reconciliation, normalized, err := ReconcileFile(upstashFixturePath(t, test.fixture))
			require.NoError(t, err)
			require.Equal(t, ReconciliationSchemaVersion, reconciliation.SchemaVersion)
			require.Equal(t, test.expectedPredicted, reconciliation.Forecast.PredictedCommandCount)
			require.Equal(t, test.expectedObserved, reconciliation.Observed.CommandCount)
			require.Equal(t, test.expectedDuplicates, reconciliation.Observed.DuplicateRecordsIgnored)
			require.Equal(t, test.expectedBilling, reconciliation.Billing.EvidenceStatus)
			require.Equal(t, test.expectedDriver, reconciliation.DominantOperationalDriver.Classification)
			require.Equal(t, "declared_fixture_input", reconciliation.DominantOperationalDriver.Basis)
			require.Regexp(t, "^[a-f0-9]{64}$", normalized.RawSource.SHA256)
			require.NoError(t, normalized.Drivers[0].Validate())
			require.NoError(t, normalized.Observation.Validate())

			ledger := ledgerv1.New()
			first, err := ledger.Ingest(normalized.Drivers, normalized.Observation, normalized.RawSource)
			require.NoError(t, err)
			require.Equal(t, ledgerv1.Ingested, first.Status)
			second, err := ledger.Ingest(normalized.Drivers, normalized.Observation, normalized.RawSource)
			require.NoError(t, err)
			require.Equal(t, ledgerv1.AlreadyPresent, second.Status)

			artifact, err := NewArtifact(reconciliation, first)
			require.NoError(t, err)
			payload, err := MarshalArtifact(artifact)
			require.NoError(t, err)
			require.NotContains(t, strings.ToLower(string(payload)), "user demand")
			var decoded map[string]any
			require.NoError(t, json.Unmarshal(payload, &decoded))
			require.Equal(t, ReconciliationSchemaVersion, decoded["schema_version"])

			if test.expectedBilling == BillingEvidenceUnavailable {
				require.Nil(t, reconciliation.Billing.TotalCost)
				require.Nil(t, reconciliation.Variance.BilledMinusObservedCost)
				require.NotContains(t, string(payload), `"total_cost": 0`)
			}
		})
	}
}

func TestIdlePollingForecastIsMateriallyHigherThanNormalLoad(t *testing.T) {
	normal, _, err := ReconcileFile(upstashFixturePath(t, "normal_load.json"))
	require.NoError(t, err)
	idle, _, err := ReconcileFile(upstashFixturePath(t, "idle_polling.json"))
	require.NoError(t, err)
	require.Greater(t, idle.Forecast.PredictedCommandCount, normal.Forecast.PredictedCommandCount*100)
	require.Greater(t, idle.Forecast.PredictedCost.Amount, normal.Forecast.PredictedCost.Amount*100)
	require.Equal(t, "idle_polling", idle.DominantOperationalDriver.Classification)
	require.Contains(t, strings.ToLower(strings.Join(idle.Limitations, " ")), "does not infer causality")
}

func TestReconcileRejectsConflictingDuplicateTelemetry(t *testing.T) {
	data, err := os.ReadFile(upstashFixturePath(t, "duplicate_export.json"))
	require.NoError(t, err)
	var fixture Fixture
	require.NoError(t, json.Unmarshal(data, &fixture))
	fixture.Telemetry[1].Commands = 361

	_, _, err = Reconcile(fixture, ledgerv1.RawSourceReference{
		ArtifactIdentity: fixture.Source.ArtifactIdentity,
		SHA256:           strings.Repeat("a", 64),
	})
	require.ErrorContains(t, err, "conflicts with prior data")
}

func TestReconcileRejectsAccountIdentifierFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "account-id.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"schema_version":"profitctl.upstash-fixture/v1","account_id":"must-not-be-stored"}`), 0600))
	_, _, err := ReconcileFile(path)
	require.ErrorContains(t, err, "unknown field")
}

func TestReconciliationSchemaIsValidJSON(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(root, "schemas", "upstash-reconciliation", "v1", "reconciliation.schema.json"))
	require.NoError(t, err)
	var schema map[string]any
	require.NoError(t, json.Unmarshal(data, &schema))
	require.Equal(t, "https://github.com/IntelIP/ProfitCtl/schemas/upstash-reconciliation/v1/reconciliation.schema.json", schema["$id"])
}

func upstashFixturePath(t *testing.T, name string) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err)
	return filepath.Join(root, "test", "fixtures", "upstash", "v1", name)
}
