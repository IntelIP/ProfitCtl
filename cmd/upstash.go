package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ledgerv1 "github.com/IntelIP/ProfitCtl/pkg/ledger/v1"
	upstashv1 "github.com/IntelIP/ProfitCtl/pkg/upstash/v1"
	"github.com/spf13/cobra"
)

var (
	upstashInput  string
	upstashLedger string
	upstashOutput string
)

var upstashCmd = &cobra.Command{
	Use:   "upstash",
	Short: "Reconcile deterministic Upstash-shaped Redis cost fixtures",
	Long:  "Reconcile synthetic Redis polling telemetry and billing evidence locally without provider calls or credentials.",
}

var upstashReconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Reconcile one synthetic Upstash fixture and normalize it into the local ledger",
	Args:  cobra.NoArgs,
	RunE:  runUpstashReconcile,
}

func init() {
	upstashReconcileCmd.Flags().StringVarP(&upstashInput, "input", "i", "", "Synthetic Upstash fixture JSON path")
	upstashReconcileCmd.Flags().StringVar(&upstashLedger, "ledger", "", "Local ledger JSON file path")
	upstashReconcileCmd.Flags().StringVarP(&upstashOutput, "output", "o", "", "Optional reconciliation artifact JSON path")
	upstashCmd.AddCommand(upstashReconcileCmd)
}

func runUpstashReconcile(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(upstashInput) == "" {
		return wrapExit(2, fmt.Errorf("--input is required"))
	}
	if strings.TrimSpace(upstashLedger) == "" {
		return wrapExit(2, fmt.Errorf("--ledger is required"))
	}
	if strings.TrimSpace(upstashOutput) != "" {
		aliases, err := pathsAlias(upstashLedger, upstashOutput)
		if err != nil {
			return wrapExit(2, fmt.Errorf("compare --ledger and --output paths: %w", err))
		}
		if aliases {
			return wrapExit(2, fmt.Errorf("--output must not reference --ledger"))
		}
	}

	reconciliation, normalized, err := upstashv1.ReconcileFile(upstashInput)
	if err != nil {
		return wrapExit(2, fmt.Errorf("reconcile fixture: %w", err))
	}
	ledger, err := ledgerv1.Open(upstashLedger)
	if err != nil {
		return wrapExit(2, fmt.Errorf("open ledger: %w", err))
	}
	ingest, err := ledger.Ingest(normalized.Drivers, normalized.Observation, normalized.RawSource)
	if err != nil {
		return wrapExit(2, fmt.Errorf("normalize ledger: %w", err))
	}
	if ingest.Status == ledgerv1.Ingested {
		if err := ledgerv1.Save(upstashLedger, ledger); err != nil {
			return wrapExit(3, fmt.Errorf("save ledger: %w", err))
		}
	}
	artifact, err := upstashv1.NewArtifact(reconciliation, ingest)
	if err != nil {
		return wrapExit(3, fmt.Errorf("build reconciliation artifact: %w", err))
	}
	payload, err := upstashv1.MarshalArtifact(artifact)
	if err != nil {
		return wrapExit(3, fmt.Errorf("format reconciliation artifact: %w", err))
	}
	if strings.TrimSpace(upstashOutput) != "" {
		if err := upstashv1.WriteArtifact(upstashOutput, payload); err != nil {
			return wrapExit(3, fmt.Errorf("write reconciliation artifact: %w", err))
		}
	}
	if _, err := cmd.OutOrStdout().Write(payload); err != nil {
		return wrapExit(3, fmt.Errorf("write reconciliation artifact: %w", err))
	}
	return nil
}

// pathsAlias rejects equivalent local paths before either file is written. It
// compares cleaned absolute paths, resolved parent directories, and existing
// filesystem identities so a reconciliation artifact cannot replace its ledger.
func pathsAlias(left, right string) (bool, error) {
	leftPath, err := canonicalLocalPath(left)
	if err != nil {
		return false, err
	}
	rightPath, err := canonicalLocalPath(right)
	if err != nil {
		return false, err
	}
	if leftPath == rightPath {
		return true, nil
	}

	leftInfo, leftErr := os.Stat(leftPath)
	rightInfo, rightErr := os.Stat(rightPath)
	if leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo) {
		return true, nil
	}
	return false, nil
}

func canonicalLocalPath(path string) (string, error) {
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	if realDirectory, err := filepath.EvalSymlinks(filepath.Dir(absPath)); err == nil {
		return filepath.Join(realDirectory, filepath.Base(absPath)), nil
	}
	return absPath, nil
}
