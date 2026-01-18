package covenant

import (
	"testing"

	"github.com/profitctl/profitctl/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestValidateCovenants_Passed(t *testing.T) {
	covenants := []config.Covenant{
		{
			Type:     "threshold",
			Field:    "margin",
			Operator: "gte",
			Value:    20,
			Message:  "Gross margin must be >= 20%",
		},
		{
			Type:     "threshold",
			Field:    "cost_per_user",
			Operator: "lte",
			Value:    5,
			Message:  "Cost per user must be <= $5",
		},
	}

	results := SimulationResults{
		Margin:      25.0, // 25% margin (meets >= 20%)
		CostPerUser: 3.0,  // $3 per user (meets <= $5)
	}

	validation := ValidateCovenants(covenants, results)

	assert.True(t, validation.Passed, "Validation should pass")
	assert.Empty(t, validation.Violations, "No violations should exist")
}

func TestValidateCovenants_Failed_Margin(t *testing.T) {
	covenants := []config.Covenant{
		{
			Type:     "threshold",
			Field:    "margin",
			Operator: "gte",
			Value:    20,
			Message:  "Gross margin must be >= 20%",
		},
	}

	results := SimulationResults{
		Margin:      15.0, // 15% margin (fails >= 20%)
		CostPerUser: 3.0,
	}

	validation := ValidateCovenants(covenants, results)

	assert.False(t, validation.Passed, "Validation should fail")
	assert.Len(t, validation.Violations, 1, "Should have one violation")
	assert.Equal(t, "margin", validation.Violations[0].Field)
	assert.Equal(t, "gte", validation.Violations[0].Operator)
	assert.Equal(t, 20.0, validation.Violations[0].Value)
	assert.Equal(t, 15.0, validation.Violations[0].Actual)
}

func TestValidateCovenants_Failed_CostPerUser(t *testing.T) {
	covenants := []config.Covenant{
		{
			Type:     "threshold",
			Field:    "cost_per_user",
			Operator: "lte",
			Value:    5,
			Message:  "Cost per user must be <= $5",
		},
	}

	results := SimulationResults{
		Margin:      25.0,
		CostPerUser: 6.0, // $6 per user (fails <= $5)
	}

	validation := ValidateCovenants(covenants, results)

	assert.False(t, validation.Passed, "Validation should fail")
	assert.Len(t, validation.Violations, 1, "Should have one violation")
	assert.Equal(t, "cost_per_user", validation.Violations[0].Field)
	assert.Equal(t, 6.0, validation.Violations[0].Actual)
}

func TestValidateCovenants_MultipleViolations(t *testing.T) {
	covenants := []config.Covenant{
		{
			Type:     "threshold",
			Field:    "margin",
			Operator: "gte",
			Value:    20,
			Message:  "Gross margin must be >= 20%",
		},
		{
			Type:     "threshold",
			Field:    "cost_per_user",
			Operator: "lte",
			Value:    5,
			Message:  "Cost per user must be <= $5",
		},
	}

	results := SimulationResults{
		Margin:      15.0, // Fails >= 20%
		CostPerUser: 6.0,  // Fails <= $5
	}

	validation := ValidateCovenants(covenants, results)

	assert.False(t, validation.Passed, "Validation should fail")
	assert.Len(t, validation.Violations, 2, "Should have two violations")
}

func TestValidateCovenants_AllOperators(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		value    float64
		actual   float64
		expected bool // true = should pass, false = should fail
	}{
		{"gt - passes", "gt", 10, 15, true},
		{"gt - fails", "gt", 10, 5, false},
		{"lt - passes", "lt", 10, 5, true},
		{"lt - fails", "lt", 10, 15, false},
		{"gte - passes (equal)", "gte", 10, 10, true},
		{"gte - passes (greater)", "gte", 10, 15, true},
		{"gte - fails", "gte", 10, 5, false},
		{"lte - passes (equal)", "lte", 10, 10, true},
		{"lte - passes (less)", "lte", 10, 5, true},
		{"lte - fails", "lte", 10, 15, false},
		{"eq - passes", "eq", 10, 10, true},
		{"eq - fails", "eq", 10, 15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			covenants := []config.Covenant{
				{
					Type:     "threshold",
					Field:    "margin",
					Operator: tt.operator,
					Value:    tt.value,
					Message:  "Test covenant",
				},
			}

			results := SimulationResults{
				Margin: tt.actual,
			}

			validation := ValidateCovenants(covenants, results)

			if tt.expected {
				assert.True(t, validation.Passed, "Should pass")
				assert.Empty(t, validation.Violations, "No violations")
			} else {
				assert.False(t, validation.Passed, "Should fail")
				assert.Len(t, validation.Violations, 1, "Should have one violation")
			}
		})
	}
}

func TestValidateCovenants_P95Fields(t *testing.T) {
	covenants := []config.Covenant{
		{
			Type:     "threshold",
			Field:    "p95_margin",
			Operator: "gte",
			Value:    15,
			Message:  "p95 margin must be >= 15%",
		},
		{
			Type:     "threshold",
			Field:    "p95_cost_per_user",
			Operator: "lte",
			Value:    10,
			Message:  "p95 cost per user must be <= $10",
		},
	}

	results := SimulationResults{
		P95Margin:      18.0, // Passes >= 15%
		P95CostPerUser: 8.0,  // Passes <= $10
	}

	validation := ValidateCovenants(covenants, results)

	assert.True(t, validation.Passed, "Validation should pass")
	assert.Empty(t, validation.Violations, "No violations")
}

func TestValidateCovenants_P99Fields(t *testing.T) {
	covenants := []config.Covenant{
		{
			Type:     "threshold",
			Field:    "p99_margin",
			Operator: "gte",
			Value:    10,
			Message:  "p99 margin must be >= 10%",
		},
		{
			Type:     "threshold",
			Field:    "p99_cost_per_user",
			Operator: "lte",
			Value:    15,
			Message:  "p99 cost per user must be <= $15",
		},
	}

	results := SimulationResults{
		P99Margin:      12.0, // Passes >= 10%
		P99CostPerUser: 14.0, // Passes <= $15
	}

	validation := ValidateCovenants(covenants, results)

	assert.True(t, validation.Passed, "Validation should pass")
	assert.Empty(t, validation.Violations, "No violations")
}

func TestValidateCovenants_UnknownField(t *testing.T) {
	covenants := []config.Covenant{
		{
			Type:     "threshold",
			Field:    "unknown_field",
			Operator: "gte",
			Value:    10,
			Message:  "Unknown field test",
		},
	}

	results := SimulationResults{
		Margin: 25.0,
	}

	validation := ValidateCovenants(covenants, results)

	// Unknown fields should be skipped (not cause violations)
	assert.True(t, validation.Passed, "Unknown fields should not cause failures")
	assert.Empty(t, validation.Violations, "No violations for unknown fields")
}

func TestValidateCovenants_UnknownOperator(t *testing.T) {
	covenants := []config.Covenant{
		{
			Type:     "threshold",
			Field:    "margin",
			Operator: "unknown_operator",
			Value:    10,
			Message:  "Unknown operator test",
		},
	}

	results := SimulationResults{
		Margin: 25.0,
	}

	validation := ValidateCovenants(covenants, results)

	// Unknown operators should be skipped
	assert.True(t, validation.Passed, "Unknown operators should not cause failures")
	assert.Empty(t, validation.Violations, "No violations for unknown operators")
}

func TestValidateCovenants_EmptyCovenants(t *testing.T) {
	covenants := []config.Covenant{}

	results := SimulationResults{
		Margin: 25.0,
	}

	validation := ValidateCovenants(covenants, results)

	assert.True(t, validation.Passed, "Empty covenants should pass")
	assert.Empty(t, validation.Violations, "No violations")
}

func TestValidateCovenants_NonThresholdType(t *testing.T) {
	covenants := []config.Covenant{
		{
			Type:     "expression", // Not threshold type
			Field:    "margin",
			Operator: "gte",
			Value:    20,
			Message:  "Expression-based (deferred to v2)",
		},
	}

	results := SimulationResults{
		Margin: 15.0, // Would fail if evaluated
	}

	validation := ValidateCovenants(covenants, results)

	// Non-threshold types should be skipped in MVP
	assert.True(t, validation.Passed, "Non-threshold types should be skipped")
	assert.Empty(t, validation.Violations, "No violations for non-threshold types")
}