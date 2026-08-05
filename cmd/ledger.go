package cmd

import (
	"fmt"
	"strings"

	ledgerv1 "github.com/IntelIP/ProfitCtl/pkg/ledger/v1"
	"github.com/spf13/cobra"
)

var (
	ledgerPath          string
	ledgerInput         string
	ledgerObservationID string
	ledgerWorkload      string
)

var ledgerCmd = &cobra.Command{
	Use:   "ledger",
	Short: "Ingest and query a local normalized cost ledger",
	Long:  "Persist normalized profitctl.cost/v1 observations locally without provider calls or background services.",
}

var ledgerIngestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest one deterministic cost-contract fixture into a local ledger",
	Args:  cobra.NoArgs,
	RunE:  runLedgerIngest,
}

var ledgerQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Emit matching local ledger observations as stable JSON",
	Args:  cobra.NoArgs,
	RunE:  runLedgerQuery,
}

func init() {
	ledgerIngestCmd.Flags().StringVar(&ledgerPath, "ledger", "", "Local ledger JSON file path")
	ledgerIngestCmd.Flags().StringVarP(&ledgerInput, "input", "i", "", "PCTL-18 cost-contract fixture JSON path")
	ledgerQueryCmd.Flags().StringVar(&ledgerPath, "ledger", "", "Local ledger JSON file path")
	ledgerQueryCmd.Flags().StringVar(&ledgerObservationID, "observation-id", "", "Exact observation ID")
	ledgerQueryCmd.Flags().StringVar(&ledgerWorkload, "workload", "", "Exact workload dimension")
	ledgerCmd.AddCommand(ledgerIngestCmd, ledgerQueryCmd)
}

func runLedgerIngest(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(ledgerPath) == "" {
		return wrapExit(2, fmt.Errorf("--ledger is required"))
	}
	if strings.TrimSpace(ledgerInput) == "" {
		return wrapExit(2, fmt.Errorf("--input is required"))
	}

	ledger, err := ledgerv1.Open(ledgerPath)
	if err != nil {
		return wrapExit(2, fmt.Errorf("open ledger: %w", err))
	}
	result, err := ledger.IngestFixture(ledgerInput)
	if err != nil {
		return wrapExit(2, fmt.Errorf("ingest fixture: %w", err))
	}
	if result.Status == ledgerv1.Ingested {
		if err := ledgerv1.Save(ledgerPath, ledger); err != nil {
			return wrapExit(3, fmt.Errorf("save ledger: %w", err))
		}
	}
	payload, err := ledgerv1.MarshalIngest(result)
	if err != nil {
		return wrapExit(3, fmt.Errorf("format ingest result: %w", err))
	}
	_, err = cmd.OutOrStdout().Write(payload)
	if err != nil {
		return wrapExit(3, fmt.Errorf("write ingest result: %w", err))
	}
	return nil
}

func runLedgerQuery(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(ledgerPath) == "" {
		return wrapExit(2, fmt.Errorf("--ledger is required"))
	}

	ledger, err := ledgerv1.Load(ledgerPath)
	if err != nil {
		return wrapExit(2, fmt.Errorf("load ledger: %w", err))
	}
	payload, err := ledgerv1.MarshalQuery(ledger.Query(ledgerv1.QueryOptions{
		ObservationID: strings.TrimSpace(ledgerObservationID),
		Workload:      strings.TrimSpace(ledgerWorkload),
	}))
	if err != nil {
		return wrapExit(3, fmt.Errorf("format query result: %w", err))
	}
	_, err = cmd.OutOrStdout().Write(payload)
	if err != nil {
		return wrapExit(3, fmt.Errorf("write query result: %w", err))
	}
	return nil
}
