package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/IntelIP/ProfitCtl/internal/systemmodel"
	"github.com/spf13/cobra"
)

func systemCommand() *cobra.Command {
	root := &cobra.Command{Use: "system", Short: "Inspect and compare sourced service economics locally", SilenceUsage: true, SilenceErrors: true}
	inspect := &cobra.Command{Use: "inspect <repository>", Short: "List local source evidence and provider signals; no network", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		r, err := systemmodel.Inspect(args[0])
		if err != nil {
			return wrapExit(2, err)
		}
		return writeSystemJSON(cmd, r)
	}}
	evaluate := &cobra.Command{Use: "evaluate <model.json>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		var m systemmodel.Model
		if err := readSystemJSON(args[0], &m); err != nil {
			return wrapExit(2, err)
		}
		r, err := systemmodel.Evaluate(m)
		if err != nil {
			return wrapExit(2, err)
		}
		if err := writeSystemJSON(cmd, r); err != nil {
			return err
		}
		if r.Status == "incomplete" {
			return wrapExitSilent(1, fmt.Errorf("system inputs incomplete"))
		}
		return nil
	}}
	compare := &cobra.Command{Use: "compare <baseline.json> <proposed.json>", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		var a, b systemmodel.Model
		if err := readSystemJSON(args[0], &a); err != nil {
			return wrapExit(2, err)
		}
		if err := readSystemJSON(args[1], &b); err != nil {
			return wrapExit(2, err)
		}
		r, err := systemmodel.Compare(a, b)
		if err != nil {
			return wrapExit(2, err)
		}
		if err := writeSystemJSON(cmd, r); err != nil {
			return err
		}
		if r.Baseline.Status == "incomplete" || r.Proposed.Status == "incomplete" {
			return wrapExitSilent(1, fmt.Errorf("system inputs incomplete"))
		}
		return nil
	}}
	receipts := &cobra.Command{Use: "receipts <sanitized-export.json>", Short: "Summarize a sanitized Condere receipt export without changing model inputs", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return wrapExit(2, err)
		}
		defer f.Close()
		raw, err := io.ReadAll(io.LimitReader(f, systemmodel.MaxInputBytes+1))
		if err != nil {
			return wrapExit(2, err)
		}
		var input systemmodel.ReceiptExport
		if err := systemmodel.Decode(bytes.NewReader(raw), &input); err != nil {
			return wrapExit(2, err)
		}
		result, err := systemmodel.SummarizeReceipts(input, raw)
		if err != nil {
			return wrapExit(2, err)
		}
		if err := writeSystemJSON(cmd, result); err != nil {
			return err
		}
		if result.Total == nil {
			return wrapExitSilent(1, fmt.Errorf("receipt costs incomplete"))
		}
		return nil
	}}
	root.AddCommand(inspect, evaluate, compare, receipts)
	return root
}
func readSystemJSON(path string, target any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return systemmodel.Decode(f, target)
}
func writeSystemJSON(cmd *cobra.Command, value any) error {
	e := json.NewEncoder(cmd.OutOrStdout())
	e.SetIndent("", "  ")
	if err := e.Encode(value); err != nil {
		return wrapExit(3, fmt.Errorf("write system result: %w", err))
	}
	return nil
}
