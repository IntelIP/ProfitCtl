package covenant

import (
	"fmt"

	"github.com/profitctl/profitctl/internal/config"
)

// ValidationResult represents the result of covenant validation
type ValidationResult struct {
	Passed    bool
	Violations []Violation
}

// Violation represents a single covenant violation
type Violation struct {
	Field    string
	Operator string
	Value    float64
	Actual   float64
	Message  string
}

// SimulationResults contains all simulation data needed for covenant validation
type SimulationResults struct {
	Margin        float64 // Gross margin percentage
	CostPerUser   float64 // Cost per user
	P95Margin     float64 // p95 margin percentage (optional)
	P95CostPerUser float64 // p95 cost per user (optional)
	P99Margin     float64 // p99 margin percentage (optional)
	P99CostPerUser float64 // p99 cost per user (optional)
}

// ValidateCovenants validates all covenants against simulation results
func ValidateCovenants(covenants []config.Covenant, results SimulationResults) ValidationResult {
	validation := ValidationResult{
		Passed:    true,
		Violations: []Violation{},
	}

	for _, covenant := range covenants {
		if covenant.Type != "threshold" {
			// Skip non-threshold covenants (expression-based deferred to v2)
			continue
		}

		violation := validateThresholdCovenant(covenant, results)
		if violation != nil {
			validation.Passed = false
			validation.Violations = append(validation.Violations, *violation)
		}
	}

	return validation
}

// validateThresholdCovenant validates a single threshold-based covenant
func validateThresholdCovenant(covenant config.Covenant, results SimulationResults) *Violation {
	// Get the actual value based on field name
	actualValue, err := getFieldValue(covenant.Field, results)
	if err != nil {
		// Field not found or not supported - skip this covenant
		return nil
	}

	// Apply the operator
	isValid := false
	switch covenant.Operator {
	case "gt":
		isValid = actualValue > covenant.Value
	case "lt":
		isValid = actualValue < covenant.Value
	case "gte":
		isValid = actualValue >= covenant.Value
	case "lte":
		isValid = actualValue <= covenant.Value
	case "eq":
		isValid = actualValue == covenant.Value
	default:
		// Unknown operator - skip this covenant
		return nil
	}

	if !isValid {
		return &Violation{
			Field:    covenant.Field,
			Operator: covenant.Operator,
			Value:    covenant.Value,
			Actual:   actualValue,
			Message:  covenant.Message,
		}
	}

	return nil
}

// getFieldValue extracts the actual value from simulation results based on field name
func getFieldValue(field string, results SimulationResults) (float64, error) {
	switch field {
	case "margin":
		return results.Margin, nil
	case "cost_per_user":
		return results.CostPerUser, nil
	case "p95_margin":
		return results.P95Margin, nil
	case "p95_cost_per_user":
		return results.P95CostPerUser, nil
	case "p99_margin":
		return results.P99Margin, nil
	case "p99_cost_per_user":
		return results.P99CostPerUser, nil
	default:
		return 0, fmt.Errorf("unknown field: %s", field)
	}
}