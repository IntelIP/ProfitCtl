package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	ledgerv1 "github.com/IntelIP/ProfitCtl/pkg/ledger/v1"
	upstashv1 "github.com/IntelIP/ProfitCtl/pkg/upstash/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRunUpstashReconcilePersistsArtifactAndLedger(t *testing.T) {
	oldInput, oldLedger, oldOutput := upstashInput, upstashLedger, upstashOutput
	defer func() {
		upstashInput, upstashLedger, upstashOutput = oldInput, oldLedger, oldOutput
	}()

	temporary := t.TempDir()
	upstashInput = filepath.Join("..", "test", "fixtures", "upstash", "v1", "idle_polling.json")
	upstashLedger = filepath.Join(temporary, "ledger.json")
	upstashOutput = filepath.Join(temporary, "reconciliation.json")

	var output bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&output)
	require.NoError(t, runUpstashReconcile(command, nil))

	var artifact upstashv1.Artifact
	require.NoError(t, json.Unmarshal(output.Bytes(), &artifact))
	require.Equal(t, upstashv1.ReconciliationSchemaVersion, artifact.SchemaVersion)
	require.Equal(t, ledgerv1.Ingested, artifact.LedgerNormalization.Status)
	require.Equal(t, "idle_polling", artifact.DominantOperationalDriver.Classification)
	stored, err := os.ReadFile(upstashOutput)
	require.NoError(t, err)
	require.Equal(t, output.Bytes(), stored)

	var repeated bytes.Buffer
	repeatCommand := &cobra.Command{}
	repeatCommand.SetOut(&repeated)
	require.NoError(t, runUpstashReconcile(repeatCommand, nil))
	require.NoError(t, json.Unmarshal(repeated.Bytes(), &artifact))
	require.Equal(t, ledgerv1.AlreadyPresent, artifact.LedgerNormalization.Status)

	ledger, err := ledgerv1.Load(upstashLedger)
	require.NoError(t, err)
	require.Len(t, ledger.Observations, 1)
	require.Equal(t, "upstash-idle-polling-observed", ledger.Observations[0].Observation.ID)
}

func TestRunUpstashReconcileFixtureCases(t *testing.T) {
	oldInput, oldLedger, oldOutput := upstashInput, upstashLedger, upstashOutput
	defer func() {
		upstashInput, upstashLedger, upstashOutput = oldInput, oldLedger, oldOutput
	}()

	cases := []struct {
		fixture           string
		billing           upstashv1.BillingEvidenceStatus
		duplicateRecords  int
		operationalDriver string
	}{
		{"normal_load.json", upstashv1.BillingEvidenceAvailable, 0, "active_workload"},
		{"idle_polling.json", upstashv1.BillingEvidenceAvailable, 0, "idle_polling"},
		{"missing_billing.json", upstashv1.BillingEvidenceUnavailable, 0, "active_workload"},
		{"duplicate_export.json", upstashv1.BillingEvidenceAvailable, 1, "active_workload"},
	}

	for _, test := range cases {
		t.Run(test.fixture, func(t *testing.T) {
			temporary := t.TempDir()
			upstashInput = filepath.Join("..", "test", "fixtures", "upstash", "v1", test.fixture)
			upstashLedger = filepath.Join(temporary, "ledger.json")
			upstashOutput = ""

			var output bytes.Buffer
			command := &cobra.Command{}
			command.SetOut(&output)
			require.NoError(t, runUpstashReconcile(command, nil))

			var artifact upstashv1.Artifact
			require.NoError(t, json.Unmarshal(output.Bytes(), &artifact))
			require.Equal(t, ledgerv1.Ingested, artifact.LedgerNormalization.Status)
			require.Equal(t, test.billing, artifact.Billing.EvidenceStatus)
			require.Equal(t, test.duplicateRecords, artifact.Observed.DuplicateRecordsIgnored)
			require.Equal(t, test.operationalDriver, artifact.DominantOperationalDriver.Classification)
			if test.billing == upstashv1.BillingEvidenceUnavailable {
				require.Nil(t, artifact.Billing.TotalCost)
			}
		})
	}
}

func TestRunUpstashReconcileRequiresInputAndLedger(t *testing.T) {
	oldInput, oldLedger := upstashInput, upstashLedger
	defer func() {
		upstashInput, upstashLedger = oldInput, oldLedger
	}()

	upstashInput = ""
	upstashLedger = "ledger.json"
	err := runUpstashReconcile(&cobra.Command{}, nil)
	require.ErrorContains(t, err, "--input is required")

	upstashInput = "fixture.json"
	upstashLedger = ""
	err = runUpstashReconcile(&cobra.Command{}, nil)
	require.ErrorContains(t, err, "--ledger is required")
}

func TestRunUpstashReconcileRejectsOutputAliasingLedger(t *testing.T) {
	oldInput, oldLedger, oldOutput := upstashInput, upstashLedger, upstashOutput
	defer func() {
		upstashInput, upstashLedger, upstashOutput = oldInput, oldLedger, oldOutput
	}()

	temporary := t.TempDir()
	upstashInput = filepath.Join("..", "test", "fixtures", "upstash", "v1", "idle_polling.json")
	upstashLedger = filepath.Join(temporary, "ledger.json")
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	relativeLedger, err := filepath.Rel(workingDirectory, upstashLedger)
	require.NoError(t, err)
	upstashOutput = relativeLedger

	err = runUpstashReconcile(&cobra.Command{}, nil)
	require.ErrorContains(t, err, "--output must not reference --ledger")
	_, err = os.Stat(upstashLedger)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestUpstashCommandFlags(t *testing.T) {
	require.NotNil(t, upstashReconcileCmd.Flags().Lookup("input"))
	require.NotNil(t, upstashReconcileCmd.Flags().Lookup("ledger"))
	require.NotNil(t, upstashReconcileCmd.Flags().Lookup("output"))
}
