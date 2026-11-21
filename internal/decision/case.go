// Package decision stores the economic decisions made by a human and their
// supporting evidence. It does not run commands or contact providers.
package decision

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	SchemaVersion = "profitctl.decision/v1"
	MaxCaseBytes  = 64 << 10
)

var (
	ErrInvalid  = errors.New("invalid decision")
	ErrConflict = errors.New("decision revision conflict; read the current revision before revising")
	slug        = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
)

type Case struct {
	Project           string        `json:"project" jsonschema:"Product or repository that owns this decision."`
	Question          string        `json:"question"`
	Workload          string        `json:"workload" jsonschema:"Unit of useful work and the environment being evaluated."`
	SourceRevision    string        `json:"source_revision" jsonschema:"Source revision or identified snapshot for this workload."`
	Window            string        `json:"window" jsonschema:"Time horizon and scope of the comparison."`
	Alternatives      []Alternative `json:"alternatives"`
	Evidence          []Evidence    `json:"evidence"`
	Recommendation    string        `json:"recommendation,omitempty" jsonschema:"Recommendation or break-even threshold supported by the evidence."`
	ReversalCondition string        `json:"reversal_condition,omitempty" jsonschema:"What would change the recommendation."`
	Missing           []string      `json:"missing,omitempty" jsonschema:"Missing evidence and how it would resolve the question."`
	Calculations      []Calculation `json:"calculations,omitempty"`
	Choice            string        `json:"choice,omitempty" jsonschema:"Exact name of the option actually chosen by the decision owner."`
	Outcome           *Outcome      `json:"outcome,omitempty"`
}

type Alternative struct {
	Name    string `json:"name"`
	Cost    string `json:"cost" jsonschema:"Estimated or observed economics with units and scope, or why cost is unknown."`
	Quality string `json:"quality" jsonschema:"Quality, latency, reliability, or human effort constraints."`
}

type Evidence struct {
	Kind    string `json:"kind" jsonschema:"Use observed for a measurement from a named run or receipt in this workload. Use assumed for supplied fixture values, modeled for calculated scenarios, and unknown for missing evidence."`
	Source  string `json:"source" jsonschema:"Safe source identity or acquisition route for missing evidence."`
	Finding string `json:"finding" jsonschema:"What this source establishes."`
}

// Calculation is a replay instruction. Save checks the digest of each listed
// input. It does not execute Args or certify the interpretation of a result.
type Calculation struct {
	Args   []string        `json:"args" jsonschema:"Read-only CLI argument tokens. Use simulate --file SCENARIO or validate --file SCENARIO for one pricing file and compare LEFT RIGHT for two. Use system evaluate, receipts, or outcomes with one input, system compare with two, and diff or explain --input FILE for variance."`
	Inputs []ArtifactInput `json:"inputs" jsonschema:"All input and calibration files needed by this calculation."`
}

type ArtifactInput struct {
	Path   string `json:"path" jsonschema:"Workspace-relative calculation input path."`
	SHA256 string `json:"sha256" jsonschema:"Lowercase SHA-256 of the complete input file."`
}

type Outcome struct {
	ForecastRevision int    `json:"forecast_revision" jsonschema:"Earlier revision of this case containing the forecast."`
	Observed         string `json:"observed" jsonschema:"What actually happened over the comparable observation window."`
	EvidenceSource   string `json:"evidence_source" jsonschema:"Source identity of an observed evidence entry supporting the actual outcome."`
	Conclusion       string `json:"conclusion" jsonschema:"How the outcome compares to the forecast, including unresolved differences."`
}

type Revision struct {
	SchemaVersion string    `json:"schema_version"`
	ID            string    `json:"id"`
	Revision      int       `json:"revision"`
	RecordedAt    time.Time `json:"recorded_at"`
	Case          Case      `json:"case"`
	RecordSHA256  string    `json:"record_sha256,omitempty" jsonschema:"Checksum of the canonical revision, excluding the checksum field. Early saved revisions have no checksum."`
}

func invalid(message string) error { return fmt.Errorf("%w: %s", ErrInvalid, message) }

func blank(value string) bool { return strings.TrimSpace(value) == "" }

func (c Case) validate() error {
	payload, err := json.Marshal(c)
	if err != nil || len(payload) > MaxCaseBytes {
		return invalid("case exceeds 64 KiB")
	}
	if blank(c.Project) || blank(c.Question) || blank(c.Workload) || blank(c.SourceRevision) || blank(c.Window) {
		return invalid("project, question, workload, source_revision and window are required")
	}
	if len(c.Alternatives) < 2 || len(c.Alternatives) > 16 {
		return invalid("supply 2 to 16 alternatives")
	}
	names := make(map[string]bool)
	for _, option := range c.Alternatives {
		if blank(option.Name) || blank(option.Cost) || blank(option.Quality) || names[option.Name] {
			return invalid("each alternative needs a unique name, cost and quality")
		}
		names[option.Name] = true
	}
	if c.Choice != "" && !names[c.Choice] {
		return invalid("choice must name a listed alternative")
	}
	if len(c.Evidence) == 0 || len(c.Evidence) > 64 {
		return invalid("supply 1 to 64 evidence entries")
	}
	for _, e := range c.Evidence {
		switch e.Kind {
		case "observed", "modeled", "assumed", "unknown":
		default:
			return invalid("evidence kind must be observed, modeled, assumed, or unknown")
		}
		if blank(e.Source) || blank(e.Finding) {
			return invalid("each evidence entry requires a source and finding")
		}
	}
	if len(c.Missing) > 32 || slices.ContainsFunc(c.Missing, blank) {
		return invalid("missing must contain at most 32 nonempty evidence requests")
	}
	if blank(c.Recommendation) && len(c.Missing) == 0 {
		return invalid("provide a recommendation or the missing evidence required to decide")
	}
	if !blank(c.Recommendation) && blank(c.ReversalCondition) {
		return invalid("a recommendation requires a reversal_condition")
	}
	if outcome := c.Outcome; outcome != nil {
		if c.Choice == "" || outcome.ForecastRevision <= 0 || blank(outcome.Observed) || blank(outcome.Conclusion) {
			return invalid("outcome requires the chosen alternative, forecast_revision, observed result, and conclusion")
		}
		if !slices.ContainsFunc(c.Evidence, func(e Evidence) bool {
			return e.Kind == "observed" && e.Source == outcome.EvidenceSource
		}) {
			return invalid("outcome.evidence_source must identify an observed evidence entry")
		}
	}
	if len(c.Calculations) > 8 {
		return invalid("supply at most 8 calculations")
	}
	for _, calc := range c.Calculations {
		if err := calc.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c Calculation) validate() error {
	if len(c.Args) == 0 || len(c.Args) > 32 || slices.ContainsFunc(c.Args, blank) || len(c.Inputs) == 0 || len(c.Inputs) > 8 {
		return invalid("each calculation requires an argument vector and 1 to 8 pinned inputs")
	}
	pinned := make(map[string]bool)
	for _, input := range c.Inputs {
		if !validInputPath(input.Path) || pinned[input.Path] {
			return invalid("supply unique, nonhidden workspace-relative JSON, YAML, or CSV input paths")
		}
		if decoded, err := hex.DecodeString(input.SHA256); err != nil || len(decoded) != 32 || strings.ToLower(input.SHA256) != input.SHA256 {
			return invalid("input sha256 must be a lowercase SHA-256 digest")
		}
		pinned[input.Path] = true
	}
	paths, err := calculationPaths(c.Args)
	if err != nil {
		return err
	}
	for _, name := range paths {
		if !pinned[name] {
			return invalid("every calculation argument naming a file must match an explicitly pinned input")
		}
	}
	return nil
}

// Accept the documented, local read-only forms and identify files the CLI
// would actually open. In particular, simulate reads --file; an unrelated
// positional argument must never make its default profit.yml look pinned.
func calculationPaths(args []string) ([]string, error) {
	if args[0] == "profitctl" {
		args = args[1:]
	}
	if len(args) == 0 {
		return nil, invalid("calculation command is missing")
	}
	command := args[0]
	var values []string
	formatted := false
	for _, arg := range args[1:] {
		switch arg {
		case "--json", "--markdown", "--quiet", "-q":
			if formatted || (command != "simulate" && command != "compare") || (command == "compare" && (arg == "--quiet" || arg == "-q")) {
				return nil, invalid("unsupported calculation output flags")
			}
			formatted = true
		default:
			values = append(values, arg)
		}
	}
	switch command {
	case "simulate", "validate":
		if len(values) == 2 && (values[0] == "--file" || values[0] == "-f") {
			return values[1:], nil
		}
		return nil, invalid("use " + command + " --file SCENARIO with the pinned workspace-relative path")
	case "compare":
		if len(values) >= 2 {
			return values, nil
		}
	case "system":
		if len(values) == 2 && slices.Contains([]string{"evaluate", "receipts", "outcomes"}, values[0]) {
			return values[1:], nil
		}
		if len(values) == 3 && values[0] == "compare" {
			return values[1:], nil
		}
	case "diff", "explain":
		if len(values) == 2 && (values[0] == "--input" || values[0] == "-i") {
			return values[1:], nil
		}
		if command == "diff" && len(values) == 4 &&
			((values[0] == "--actual" && values[2] == "--forecast") || (values[0] == "--forecast" && values[2] == "--actual")) {
			return []string{values[1], values[3]}, nil
		}
	}
	return nil, invalid("calculation must use a supported read-only ProfitCtl command and explicit input arguments")
}

func validInputPath(name string) bool {
	// Use portable slash-relative paths; refuse drive specifiers and hidden
	// directories even on hosts where they are legal filename characters.
	if len(name) > 512 || !fs.ValidPath(name) || strings.ContainsAny(name, ":\\\x00") || !filepath.IsLocal(filepath.FromSlash(name)) {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if strings.HasPrefix(part, ".") || strings.HasPrefix(part, "-") {
			return false
		}
	}
	return slices.Contains([]string{".json", ".yaml", ".yml", ".csv"}, strings.ToLower(path.Ext(name)))
}
