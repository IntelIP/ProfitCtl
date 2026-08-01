package v1

import (
	"errors"
	"fmt"
	"strings"

	"github.com/IntelIP/ProfitCtl/pkg/types"
)

type LegacyMapping struct {
	ArtifactIdentity string            `json:"artifact_identity" yaml:"artifact_identity"`
	CapturedAt       string            `json:"captured_at" yaml:"captured_at"`
	Currency         string            `json:"currency" yaml:"currency"`
	Window           TimeWindow        `json:"window" yaml:"window"`
	VariableUnits    map[string]string `json:"variable_units" yaml:"variable_units"`
}

// MapLegacyCosts maps current scenario inputs into v1 drivers without changing
// simulator behavior. Legacy variable costs have no unit field, so callers must
// provide an explicit name-to-unit mapping.
func MapLegacyCosts(fixed []types.FixedCost, variable []types.VariableCost, mapping LegacyMapping) ([]CostDriver, error) {
	if strings.TrimSpace(mapping.ArtifactIdentity) == "" {
		return nil, errors.New("artifact_identity is required")
	}
	if !validCaptureTime(mapping.CapturedAt) {
		return nil, errors.New("captured_at must be an RFC3339 timestamp or ISO date")
	}
	if err := (Money{Currency: mapping.Currency}).Validate(); err != nil {
		return nil, err
	}
	if err := mapping.Window.Validate(); err != nil {
		return nil, fmt.Errorf("window: %w", err)
	}

	drivers := make([]CostDriver, 0, len(fixed)+len(variable))
	for i, cost := range fixed {
		quantity, periodUnit, err := legacyFixedBasis(cost.Period)
		if err != nil {
			return nil, fmt.Errorf("fixed cost %q: %w", cost.Name, err)
		}
		driver := CostDriver{
			SchemaVersion: SchemaVersion,
			ID:            fmt.Sprintf("legacy-fixed-%d", i+1),
			Name:          cost.Name,
			Kind:          DriverFixed,
			Quantity:      Quantity{Value: quantity, Unit: "commitment"},
			Per:           &Quantity{Value: 1, Unit: periodUnit},
			UnitPrice: UnitPrice{
				Amount: Money{Amount: cost.Amount, Currency: mapping.Currency},
				Per:    Quantity{Value: 1, Unit: "commitment"},
			},
			Window:     mapping.Window,
			Dimensions: Dimensions{Workload: cost.Name},
			Evidence:   legacyEvidence(cost.Source, mapping),
		}
		if err := driver.Validate(); err != nil {
			return nil, fmt.Errorf("fixed cost %q: %w", cost.Name, err)
		}
		drivers = append(drivers, driver)
	}

	for i, cost := range variable {
		unit := mapping.VariableUnits[cost.Name]
		if err := validateUnit(unit); err != nil {
			return nil, fmt.Errorf("variable cost %q unit: %w", cost.Name, err)
		}
		subjectUnit := "user"
		if types.NormalizeVariableCostUserScope(cost.UserScope) == types.UserScopePaidUsers {
			subjectUnit = "paid_user"
		}
		driver := CostDriver{
			SchemaVersion: SchemaVersion,
			ID:            fmt.Sprintf("legacy-variable-%d", i+1),
			Name:          cost.Name,
			Kind:          DriverVariable,
			Quantity:      Quantity{Value: cost.UnitsPerUser, Unit: unit},
			Per:           &Quantity{Value: 1, Unit: subjectUnit},
			UnitPrice: UnitPrice{
				Amount: Money{Amount: cost.CostPerUnit, Currency: mapping.Currency},
				Per:    Quantity{Value: 1, Unit: unit},
			},
			Window:     mapping.Window,
			Dimensions: Dimensions{Workload: cost.Name},
			Evidence:   legacyEvidence(cost.Source, mapping),
		}
		if err := driver.Validate(); err != nil {
			return nil, fmt.Errorf("variable cost %q: %w", cost.Name, err)
		}
		drivers = append(drivers, driver)
	}
	return drivers, nil
}

func legacyEvidence(source *types.CostSource, mapping LegacyMapping) Evidence {
	evidence := Evidence{
		Kind:        EvidencePredicted,
		Measurement: MeasurementDeclared,
		Source: SourceReference{
			Type:             SourceLegacyScenario,
			ArtifactIdentity: mapping.ArtifactIdentity,
			CapturedAt:       mapping.CapturedAt,
		},
		Confidence:          ConfidenceLow,
		ConfidenceRationale: "Legacy scenario assumption mapped without claiming runtime measurement.",
	}
	if source == nil {
		return evidence
	}

	// Legacy provider_catalog entries predate the required refresh policy. Keep
	// them mappable without upgrading incomplete provenance into v1 catalog
	// evidence.
	if SourceType(source.Type) == SourceProviderCatalog {
		evidence.Source.URL = source.URL
		if source.CapturedAt != "" {
			evidence.Source.CapturedAt = source.CapturedAt
		}
		evidence.ConfidenceRationale = "Legacy provider catalog assumption mapped without v1 refresh policy or current-price claim."
		return evidence
	}

	evidence.Source.Type = SourceType(source.Type)
	evidence.Source.URL = source.URL
	evidence.Source.CapturedAt = source.CapturedAt
	if evidence.Source.CapturedAt == "" {
		evidence.Source.CapturedAt = mapping.CapturedAt
	}
	evidence.Confidence = Confidence(source.Confidence)
	if strings.TrimSpace(source.Note) != "" {
		evidence.ConfidenceRationale = source.Note
	}
	return evidence
}

func legacyFixedBasis(period types.CostPeriod) (float64, string, error) {
	switch period {
	case types.PeriodDaily:
		return 30, "month", nil
	case types.PeriodMonthly:
		return 1, "month", nil
	case types.PeriodYearly:
		return 1, "year", nil
	default:
		return 0, "", fmt.Errorf("unsupported period %q", period)
	}
}
