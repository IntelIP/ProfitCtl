package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	ledgerv1 "github.com/IntelIP/ProfitCtl/pkg/ledger/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRunLedgerIngestAndQuery(t *testing.T) {
	oldPath, oldInput, oldID, oldWorkload := ledgerPath, ledgerInput, ledgerObservationID, ledgerWorkload
	defer func() {
		ledgerPath, ledgerInput, ledgerObservationID, ledgerWorkload = oldPath, oldInput, oldID, oldWorkload
	}()

	ledgerPath = filepath.Join(t.TempDir(), "ledger.json")
	ledgerInput = filepath.Join("..", "test", "fixtures", "cost_contract", "v1", "upstash_idle_polling.json")
	ledgerObservationID = ""
	ledgerWorkload = ""

	var ingestOutput bytes.Buffer
	ingestCommand := &cobra.Command{}
	ingestCommand.SetOut(&ingestOutput)
	require.NoError(t, runLedgerIngest(ingestCommand, nil))

	var ingest ledgerv1.IngestResult
	require.NoError(t, json.Unmarshal(ingestOutput.Bytes(), &ingest))
	require.Equal(t, ledgerv1.Ingested, ingest.Status)

	var repeatOutput bytes.Buffer
	repeatCommand := &cobra.Command{}
	repeatCommand.SetOut(&repeatOutput)
	require.NoError(t, runLedgerIngest(repeatCommand, nil))
	require.Contains(t, repeatOutput.String(), `"status": "already_present"`)

	ledgerWorkload = "idle_worker_polling"
	var queryOutput bytes.Buffer
	queryCommand := &cobra.Command{}
	queryCommand.SetOut(&queryOutput)
	require.NoError(t, runLedgerQuery(queryCommand, nil))

	var query ledgerv1.QueryResult
	require.NoError(t, json.Unmarshal(queryOutput.Bytes(), &query))
	require.Len(t, query.Observations, 1)
	require.Equal(t, "upstash-idle-polling-synthetic-observation", query.Observations[0].Observation.ID)
}

func TestRunLedgerCommandsRequireLedgerPath(t *testing.T) {
	oldPath, oldInput := ledgerPath, ledgerInput
	defer func() {
		ledgerPath, ledgerInput = oldPath, oldInput
	}()
	ledgerPath = ""
	ledgerInput = "fixture.json"

	err := runLedgerIngest(&cobra.Command{}, nil)
	require.ErrorContains(t, err, "--ledger is required")
	err = runLedgerQuery(&cobra.Command{}, nil)
	require.ErrorContains(t, err, "--ledger is required")
}

func TestLedgerCommandFlags(t *testing.T) {
	require.NotNil(t, ledgerIngestCmd.Flags().Lookup("ledger"))
	require.NotNil(t, ledgerIngestCmd.Flags().Lookup("input"))
	require.NotNil(t, ledgerQueryCmd.Flags().Lookup("ledger"))
	require.NotNil(t, ledgerQueryCmd.Flags().Lookup("observation-id"))
	require.NotNil(t, ledgerQueryCmd.Flags().Lookup("workload"))
}
