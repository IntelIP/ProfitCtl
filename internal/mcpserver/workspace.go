package mcpserver

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	maxScenarioBytes        = 1 << 20
	maxSimulationIterations = 10000
)

var scenarioExtensions = map[string]bool{
	".yml":  true,
	".yaml": true,
}

var standardsExtensions = map[string]bool{
	".yml":  true,
	".yaml": true,
	".md":   true,
	".txt":  true,
}

var calibrationExtensions = map[string]bool{
	".yml":  true,
	".yaml": true,
	".json": true,
	".csv":  true,
}

type loadedScenario struct {
	relative string
	digest   string
}

type preflightConfig struct {
	CalibrationFile string `yaml:"calibration_file"`
	Simulation      *struct {
		Iterations int `yaml:"iterations"`
	} `yaml:"simulation"`
}

type scenarioError struct {
	outcome Outcome
	code    string
	message string
	field   string
}

func (s *Server) loadScenario(raw string) (loadedScenario, *scenarioError) {
	resolved, relative, err := s.resolveToolPath(raw, scenarioExtensions, "scenario_path")
	if err != nil {
		return loadedScenario{}, err
	}

	data, readErr := readBoundedFile(resolved, maxScenarioBytes)
	if readErr != nil {
		return loadedScenario{}, &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "scenario_unreadable",
			message: "The scenario file cannot be read within the local tool boundary.",
			field:   "scenario_path",
		}
	}

	var preflight preflightConfig
	if err := yaml.Unmarshal(data, &preflight); err != nil {
		return loadedScenario{}, &scenarioError{
			outcome: OutcomeInvalidScenario,
			code:    "scenario_yaml_invalid",
			message: "The scenario YAML cannot be parsed.",
			field:   "scenario_path",
		}
	}
	if preflight.Simulation != nil && preflight.Simulation.Iterations > maxSimulationIterations {
		return loadedScenario{}, &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "simulation_iterations_exceed_limit",
			message: "The scenario exceeds the local simulation iteration limit of 10000.",
			field:   "scenario_path",
		}
	}
	if strings.TrimSpace(preflight.CalibrationFile) != "" {
		if calibrationErr := s.validateCalibrationReference(resolved, preflight.CalibrationFile); calibrationErr != nil {
			return loadedScenario{}, calibrationErr
		}
	}

	digest := sha256.Sum256(data)
	return loadedScenario{relative: relative, digest: hex.EncodeToString(digest[:])}, nil
}

func (s *Server) loadStandardsTarget(raw string) (loadedScenario, *scenarioError) {
	resolved, relative, err := s.resolveToolPath(raw, standardsExtensions, "target_path")
	if err != nil {
		return loadedScenario{}, err
	}
	data, readErr := readBoundedFile(resolved, maxScenarioBytes)
	if readErr != nil {
		return loadedScenario{}, &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "target_unreadable",
			message: "The standards target cannot be read within the local tool boundary.",
			field:   "target_path",
		}
	}

	if scenarioExtensions[strings.ToLower(filepath.Ext(relative))] {
		var preflight preflightConfig
		if err := yaml.Unmarshal(data, &preflight); err != nil {
			return loadedScenario{}, &scenarioError{
				outcome: OutcomeInvalidScenario,
				code:    "scenario_yaml_invalid",
				message: "The scenario YAML cannot be parsed.",
				field:   "target_path",
			}
		}
		if strings.TrimSpace(preflight.CalibrationFile) != "" {
			if calibrationErr := s.validateCalibrationReference(resolved, preflight.CalibrationFile); calibrationErr != nil {
				return loadedScenario{}, calibrationErr
			}
		}
	}

	digest := sha256.Sum256(data)
	return loadedScenario{relative: relative, digest: hex.EncodeToString(digest[:])}, nil
}

func (s *Server) resolveToolPath(raw string, allowedExtensions map[string]bool, field string) (string, string, *scenarioError) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "path_required",
			message: "A non-empty root-relative file path is required.",
		}
	}
	if strings.ContainsRune(path, 0) || filepath.IsAbs(path) {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "path_must_be_root_relative",
			message: "The file path must be relative to the configured workspace root.",
		}
	}
	if hasParentReference(path) {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "path_escape_rejected",
			message: "Parent-directory path references are not allowed.",
		}
	}

	clean := filepath.Clean(path)
	if clean == "." {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "path_must_reference_file",
			message: "The file path must reference a supported regular file.",
		}
	}
	if !allowedExtensions[strings.ToLower(filepath.Ext(clean))] {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "extension_not_allowed",
			message: "The file extension is not allowed for this tool.",
		}
	}

	return s.resolveExistingWorkspaceFile(filepath.Join(s.workspaceRoot, clean), allowedExtensions, field)
}

func (s *Server) validateCalibrationReference(scenarioPath, raw string) *scenarioError {
	path := strings.TrimSpace(raw)
	if path == "" || strings.ContainsRune(path, 0) || filepath.IsAbs(path) {
		return &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "calibration_path_outside_workspace",
			message: "The referenced calibration file must stay inside the configured workspace root.",
			field:   "calibration_file",
		}
	}
	clean := filepath.Clean(path)
	if clean == "." || !calibrationExtensions[strings.ToLower(filepath.Ext(clean))] {
		return &scenarioError{
			outcome: OutcomeInvalidScenario,
			code:    "calibration_file_invalid",
			message: "The referenced calibration file is not a supported local calibration artifact.",
			field:   "calibration_file",
		}
	}

	_, _, err := s.resolveExistingWorkspaceFile(filepath.Join(filepath.Dir(scenarioPath), clean), calibrationExtensions, "calibration_file")
	if err == nil {
		return nil
	}
	if err.code == "path_escape_rejected" || err.code == "path_must_be_root_relative" {
		err.outcome = OutcomeInvalidInput
		err.code = "calibration_path_outside_workspace"
		err.message = "The referenced calibration file must stay inside the configured workspace root."
		err.field = "calibration_file"
		return err
	}
	err.outcome = OutcomeInvalidScenario
	err.code = "calibration_file_unavailable"
	err.message = "The referenced calibration file is not available inside the configured workspace root."
	err.field = "calibration_file"
	return err
}

func (s *Server) resolveExistingWorkspaceFile(candidate string, allowedExtensions map[string]bool, field string) (string, string, *scenarioError) {
	if !isWithinRoot(s.workspaceRoot, candidate) {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "path_escape_rejected",
			message: "The requested file resolves outside the configured workspace root.",
			field:   field,
		}
	}
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "path_unavailable",
			message: "The requested file is not available inside the configured workspace root.",
			field:   field,
		}
	}
	relative, err := filepath.Rel(s.workspaceRoot, resolved)
	if err != nil || !isWithinRoot(s.workspaceRoot, resolved) {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "path_escape_rejected",
			message: "The requested file resolves outside the configured workspace root.",
			field:   field,
		}
	}
	if !allowedExtensions[strings.ToLower(filepath.Ext(resolved))] {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "extension_not_allowed",
			message: "The resolved file extension is not allowed for this tool.",
			field:   field,
		}
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() {
		return "", "", &scenarioError{
			outcome: OutcomeInvalidInput,
			code:    "path_must_reference_file",
			message: "The requested path must resolve to a regular file inside the configured workspace root.",
			field:   field,
		}
	}
	return resolved, filepath.ToSlash(relative), nil
}

func isWithinRoot(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

func hasParentReference(path string) bool {
	parts := strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	for _, part := range parts {
		if part == ".." {
			return true
		}
	}
	return false
}

func readBoundedFile(path string, maxBytes int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, int64(maxBytes)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxBytes {
		return nil, errors.New("file exceeds local input limit")
	}
	return data, nil
}
