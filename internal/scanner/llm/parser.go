package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrEmptyResponse = errors.New("empty analysis response")
	ErrInvalidJSON   = errors.New("analysis response is not valid JSON")
	ErrInvalidSchema = errors.New("analysis response does not satisfy required schema")
)

var fencedJSONPattern = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*\\})\\s*```")

// ParseAnalysisResponse parses and validates provider output into AnalysisResponse.
func ParseAnalysisResponse(raw string) (AnalysisResponse, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return AnalysisResponse{}, ErrEmptyResponse
	}

	candidates := []string{clean}

	if match := fencedJSONPattern.FindStringSubmatch(clean); len(match) == 2 {
		candidates = append(candidates, strings.TrimSpace(match[1]))
	}

	if open := strings.Index(clean, "{"); open >= 0 {
		if close := strings.LastIndex(clean, "}"); close > open {
			candidates = append(candidates, strings.TrimSpace(clean[open:close+1]))
		}
	}

	seen := make(map[string]struct{}, len(candidates))
	var lastErr error
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}

		var analysis AnalysisResponse
		if err := json.Unmarshal([]byte(candidate), &analysis); err != nil {
			lastErr = err
			continue
		}
		normalizeAnalysis(&analysis)
		if err := validateAnalysis(analysis); err != nil {
			return AnalysisResponse{}, fmt.Errorf("%w: %v", ErrInvalidSchema, err)
		}
		return analysis, nil
	}

	if lastErr != nil {
		return AnalysisResponse{}, fmt.Errorf("%w: %v", ErrInvalidJSON, lastErr)
	}
	return AnalysisResponse{}, ErrInvalidJSON
}

func normalizeAnalysis(analysis *AnalysisResponse) {
	if analysis.Services == nil {
		analysis.Services = []DetectedService{}
	}
	if analysis.Dependencies == nil {
		analysis.Dependencies = []DetectedDependency{}
	}
	if analysis.Patterns == nil {
		analysis.Patterns = []UsagePattern{}
	}
}

func validateAnalysis(analysis AnalysisResponse) error {
	for i, service := range analysis.Services {
		if strings.TrimSpace(service.Name) == "" {
			return fmt.Errorf("services[%d].name is required", i)
		}
		if strings.TrimSpace(service.Type) == "" {
			return fmt.Errorf("services[%d].type is required", i)
		}
		if strings.TrimSpace(service.Provider) == "" {
			return fmt.Errorf("services[%d].provider is required", i)
		}
	}

	for i, dep := range analysis.Dependencies {
		if strings.TrimSpace(dep.Name) == "" {
			return fmt.Errorf("dependencies[%d].name is required", i)
		}
		if strings.TrimSpace(dep.Type) == "" {
			return fmt.Errorf("dependencies[%d].type is required", i)
		}
	}

	for i, pattern := range analysis.Patterns {
		if strings.TrimSpace(pattern.Description) == "" {
			return fmt.Errorf("patterns[%d].description is required", i)
		}
		if pattern.Confidence < 0 || pattern.Confidence > 1 {
			return fmt.Errorf("patterns[%d].confidence must be between 0 and 1", i)
		}
	}

	return nil
}
