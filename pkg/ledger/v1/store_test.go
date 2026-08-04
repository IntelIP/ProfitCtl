package v1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
	"github.com/stretchr/testify/require"
)

func TestIngestFixtureIsIdempotentAndTraceable(t *testing.T) {
	ledger := New()
	path := fixturePath(t)

	first, err := ledger.IngestFixture(path)
	require.NoError(t, err)
	require.Equal(t, Ingested, first.Status)
	require.Equal(t, "upstash-idle-polling-synthetic-observation", first.ObservationID)
	require.Equal(t, "test/fixtures/cost_contract/v1/upstash_idle_polling.json", first.RawSource.ArtifactIdentity)
	require.Regexp(t, "^[a-f0-9]{64}$", first.RawSource.SHA256)

	second, err := ledger.IngestFixture(path)
	require.NoError(t, err)
	require.Equal(t, AlreadyPresent, second.Status)
	require.Len(t, ledger.Drivers, 1)
	require.Len(t, ledger.Observations, 1)
	require.NoError(t, ledger.Validate())
}

func TestIngestRejectsSameObservationIDWithChangedData(t *testing.T) {
	ledger := New()
	require.NoError(t, ingestFixture(t, &ledger))

	record := ledger.Observations[0]
	changed := record.Observation
	changed.TotalCost.Amount = 1.5
	changed.UnitPrice.Amount.Amount = changed.TotalCost.Amount / changed.Quantity.Value
	changed.Evidence.TotalCost.Source.Type = costv1.SourceProfitCtlDerived
	changed.Evidence.TotalCost.Source.ArtifactIdentity = "profitctl://derived/changed-total"

	_, err := ledger.Ingest(ledger.Drivers, changed, record.RawSource)
	require.ErrorContains(t, err, "conflicts with existing ledger data")
	require.Len(t, ledger.Observations, 1)
}

func TestSaveLoadAndQueryAreStable(t *testing.T) {
	ledger := New()
	require.NoError(t, ingestFixture(t, &ledger))
	second := ledger.Observations[0].Observation
	second.ID = "aardvark-idle-polling-synthetic-observation"
	_, err := ledger.Ingest(nil, second, ledger.Observations[0].RawSource)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "ledger.json")
	require.NoError(t, Save(path, ledger))
	loaded, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, ledger, loaded)

	result := loaded.Query(QueryOptions{Workload: "idle_worker_polling"})
	first, err := MarshalQuery(result)
	require.NoError(t, err)
	secondOutput, err := MarshalQuery(loaded.Query(QueryOptions{Workload: "idle_worker_polling"}))
	require.NoError(t, err)
	require.Equal(t, string(first), string(secondOutput))
	require.Contains(t, string(first), `"schema_version": "profitctl.ledger/v1"`)
	require.Contains(t, string(first), `"raw_source"`)

	var decoded QueryResult
	require.NoError(t, json.Unmarshal(first, &decoded))
	require.Len(t, decoded.Drivers, 1)
	require.Equal(t, "upstash-idle-poll-cadence", decoded.Drivers[0].ID)
	require.Len(t, decoded.Observations, 2)
	require.Equal(t, "aardvark-idle-polling-synthetic-observation", decoded.Observations[0].Observation.ID)
	require.Equal(t, "upstash-idle-polling-synthetic-observation", decoded.Observations[1].Observation.ID)
}

func TestLoadRejectsUnknownLedgerFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"schema_version":"profitctl.ledger/v1","drivers":[],"observations":[],"unexpected":true}`), 0600))
	_, err := Load(path)
	require.ErrorContains(t, err, "unknown field")
}

func TestQueryFiltersByObservationID(t *testing.T) {
	ledger := New()
	require.NoError(t, ingestFixture(t, &ledger))
	result := ledger.Query(QueryOptions{ObservationID: "not-present"})
	require.Empty(t, result.Observations)
}

func TestLedgerSchemaIsValidJSON(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(root, "schemas", "cost-ledger", "v1", "ledger.schema.json"))
	require.NoError(t, err)
	var schema map[string]any
	require.NoError(t, json.Unmarshal(data, &schema))
	require.Equal(t, "https://github.com/IntelIP/ProfitCtl/schemas/cost-ledger/v1/ledger.schema.json", schema["$id"])
}

func ingestFixture(t *testing.T, ledger *Ledger) error {
	t.Helper()
	_, err := ledger.IngestFixture(fixturePath(t))
	return err
}

func fixturePath(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err)
	return filepath.Join(root, "test", "fixtures", "cost_contract", "v1", "upstash_idle_polling.json")
}
