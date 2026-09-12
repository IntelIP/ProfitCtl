// Package mcpserver exposes a bounded local MCP wrapper around the ProfitCtl CLI.
package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// ResultSchemaVersion is the stable outer schema returned by every tool.
	ResultSchemaVersion = "profitctl-plugin-result/v1"

	defaultTimeout        = 30 * time.Second
	defaultMaxOutputBytes = 1 << 20
)

// Outcome classifies the result of a supported tool call.
type Outcome string

const (
	OutcomePassed              Outcome = "passed"
	OutcomeCovenantFailed      Outcome = "covenant_failed"
	OutcomeStandardsFailed     Outcome = "standards_failed"
	OutcomeInvalidInput        Outcome = "invalid_input"
	OutcomeInvalidScenario     Outcome = "invalid_scenario"
	OutcomeRuntimeFailed       Outcome = "runtime_failed"
	OutcomeDeadlineExceeded    Outcome = "deadline_exceeded"
	OutcomeOutputLimitExceeded Outcome = "output_limit_exceeded"
)

// Config fixes the local execution boundary when the MCP server starts.
// WorkspaceRoot is intentionally not accepted from model-provided tool input.
type Config struct {
	WorkspaceRoot   string
	Binary          string
	StandardsBinary string
	Timeout         time.Duration
	MaxOutputBytes  int
}

// InputEvidence identifies the local files used for a completed tool call.
type InputEvidence struct {
	Paths  []string `json:"paths"`
	SHA256 []string `json:"sha256"`
}

// ErrorInfo contains only safe, stable failure information.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// ResultEnvelope is returned by every ProfitCtl MCP tool.
type ResultEnvelope struct {
	SchemaVersion string        `json:"schema_version"`
	Tool          string        `json:"tool"`
	Outcome       Outcome       `json:"outcome"`
	Input         InputEvidence `json:"input"`
	DurationMS    int64         `json:"duration_ms"`
	Result        any           `json:"result"`
	Error         *ErrorInfo    `json:"error"`
}

type processResult struct {
	stdout              []byte
	exitCode            int
	deadlineExceeded    bool
	outputLimitExceeded bool
	err                 error
}

type processRunner interface {
	Run(ctx context.Context, binary string, args []string, dir string, maxOutput int) processResult
}

// Server owns the fixed workspace and local process boundary for tool calls.
type Server struct {
	workspaceRoot   string
	binary          string
	standardsBinary string
	timeout         time.Duration
	maxOutput       int
	runner          processRunner
}

// New creates a server with validated fixed local paths.
func New(cfg Config) (*Server, error) {
	return newWithRunner(cfg, execRunner{})
}

func newWithRunner(cfg Config, runner processRunner) (*Server, error) {
	if runner == nil {
		return nil, fmt.Errorf("profitctl MCP runner is required")
	}

	workspaceRoot, err := resolveWorkspaceRoot(cfg.WorkspaceRoot)
	if err != nil {
		return nil, err
	}

	binary := cfg.Binary
	if binary == "" {
		binary, err = os.Executable()
		if err != nil {
			return nil, fmt.Errorf("resolve profitctl executable: %w", err)
		}
	}
	binary, err = resolveExecutable(binary)
	if err != nil {
		return nil, fmt.Errorf("resolve profitctl executable: %w", err)
	}

	standardsBinary := cfg.StandardsBinary
	if standardsBinary == "" {
		standardsBinary = filepath.Join(filepath.Dir(binary), standardsExecutableName(runtime.GOOS))
	}
	standardsBinary, err = resolveExecutable(standardsBinary)
	if err != nil {
		return nil, fmt.Errorf("resolve ProfitCtl standards executable: %w", err)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	if timeout < 0 {
		return nil, fmt.Errorf("profitctl MCP timeout must be positive")
	}

	maxOutput := cfg.MaxOutputBytes
	if maxOutput == 0 {
		maxOutput = defaultMaxOutputBytes
	}
	if maxOutput < 1 {
		return nil, fmt.Errorf("profitctl MCP max output must be positive")
	}

	return &Server{
		workspaceRoot:   workspaceRoot,
		binary:          binary,
		standardsBinary: standardsBinary,
		timeout:         timeout,
		maxOutput:       maxOutput,
		runner:          runner,
	}, nil
}

func resolveWorkspaceRoot(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("workspace root is required")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root: %w", err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("stat workspace root: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace root must be a directory")
	}
	return resolved, nil
}

func standardsExecutableName(goos string) string {
	if goos == "windows" {
		return "profitctl-standards.exe"
	}
	return "profitctl-standards"
}

func resolveExecutable(raw string) (string, error) {
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("must be a regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		return "", fmt.Errorf("must be executable")
	}
	return resolved, nil
}

func emptyInput() InputEvidence {
	return InputEvidence{Paths: []string{}, SHA256: []string{}}
}

func newEnvelope(tool string, started time.Time, input InputEvidence) ResultEnvelope {
	return ResultEnvelope{
		SchemaVersion: ResultSchemaVersion,
		Tool:          tool,
		Input:         input,
		DurationMS:    time.Since(started).Milliseconds(),
	}
}

func failEnvelope(tool string, started time.Time, outcome Outcome, code, message, field string) ResultEnvelope {
	envelope := newEnvelope(tool, started, emptyInput())
	envelope.Outcome = outcome
	envelope.Error = &ErrorInfo{Code: code, Message: message, Field: field}
	return envelope
}
