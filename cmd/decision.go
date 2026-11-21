package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/IntelIP/ProfitCtl/internal/decision"
	"github.com/IntelIP/ProfitCtl/internal/systemmodel"
	"github.com/spf13/cobra"
)

func decisionCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "decision",
		Short: "Save and revisit a local economic decision with versioned evidence",
		Long:  "Save private case revisions in .profitctl/decisions in the selected workspace. ProfitCtl checks pinned input digests and records calculations as data. On Windows, select a workspace with a private ACL.",
	}
	var workspace string
	root.PersistentFlags().StringVar(&workspace, "workspace", ".", "Existing directory shared by the CLI and MCP clients")
	for _, action := range []string{"create", "revise"} {
		var input string
		var expected int
		command := &cobra.Command{
			Use:   action + " <id>",
			Short: "Save a complete decision case from JSON",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if action == "revise" && expected <= 0 {
					return wrapExit(2, errors.New("--expected must name the positive revision being edited"))
				}
				value, err := decisionInput(cmd, input)
				if err != nil {
					return wrapExit(2, err)
				}
				saved, err := decision.Save(workspace, args[0], expected, value)
				return writeDecision(cmd, saved, err)
			},
		}
		command.Flags().StringVar(&input, "input", "", "Case JSON file or - for stdin (required)")
		if action == "revise" {
			command.Flags().IntVar(&expected, "expected", 0, "Revision being edited (required)")
		}
		root.AddCommand(command)
	}
	var revision int
	read := &cobra.Command{
		Use: "read <id>", Short: "Return the current or a historical decision revision", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			record, err := decision.Read(workspace, args[0], revision)
			return writeDecision(cmd, record, err)
		},
	}
	read.Flags().IntVar(&revision, "revision", 0, "Revision to read (zero selects latest)")
	var query, after string
	var limit int
	list := &cobra.Command{
		Use: "list", Short: "Find saved decisions by project, question or ID", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			page, err := decision.List(workspace, query, after, limit)
			return writeDecision(cmd, page, err)
		},
	}
	list.Flags().StringVar(&query, "query", "", "Case-insensitive text in the project, question or case ID")
	list.Flags().StringVar(&after, "after", "", "Continuation cursor from the previous page")
	list.Flags().IntVar(&limit, "limit", 20, "Maximum summaries to return per page (1–50)")
	root.AddCommand(read, list)
	return root
}

func decisionInput(cmd *cobra.Command, path string) (decision.Case, error) {
	if path == "" {
		return decision.Case{}, errors.New("--input needs a case JSON file or - for stdin")
	}
	r := cmd.InOrStdin()
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return decision.Case{}, err
		}
		defer f.Close()
		r = f
	}
	data, err := io.ReadAll(io.LimitReader(r, decision.MaxCaseBytes+1))
	if err != nil {
		return decision.Case{}, err
	}
	if len(data) > decision.MaxCaseBytes {
		return decision.Case{}, errors.New("decision input exceeds 64 KiB")
	}
	var c decision.Case
	err = systemmodel.Decode(bytes.NewReader(data), &c)
	return c, err
}

func writeDecision(cmd *cobra.Command, r any, err error) error {
	switch {
	case errors.Is(err, decision.ErrInvalid):
		return wrapExit(2, err)
	case errors.Is(err, decision.ErrConflict):
		return wrapExit(4, err)
	case errors.Is(err, fs.ErrNotExist):
		return wrapExit(1, fmt.Errorf("decision workspace, case, or revision not found: %w", err))
	case err != nil:
		return wrapExit(3, err)
	}
	result := json.NewEncoder(cmd.OutOrStdout())
	result.SetIndent("", "  ")
	return wrapExit(3, result.Encode(r))
}
