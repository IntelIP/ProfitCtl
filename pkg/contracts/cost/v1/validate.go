package v1

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	canonicalUnitPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	currencyPattern      = regexp.MustCompile(`^[A-Z]{3}$`)
	idPattern            = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,127}$`)
)

func (d CostDriver) Validate() error {
	if d.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version must be %q", SchemaVersion)
	}
	if err := validateID(d.ID); err != nil {
		return fmt.Errorf("id: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return errors.New("name is required")
	}
	if !validDriverKind(d.Kind) {
		return fmt.Errorf("unsupported driver kind %q", d.Kind)
	}
	if (d.Kind == DriverVariable || d.Kind == DriverCadence) && d.Per == nil {
		return fmt.Errorf("%s driver requires per scale basis", d.Kind)
	}
	if err := d.Quantity.Validate(); err != nil {
		return fmt.Errorf("quantity: %w", err)
	}
	if d.Per != nil {
		if err := d.Per.Validate(); err != nil {
			return fmt.Errorf("per: %w", err)
		}
	}
	if err := d.UnitPrice.Validate(); err != nil {
		return fmt.Errorf("unit_price: %w", err)
	}
	if d.UnitPrice.Per.Unit != d.Quantity.Unit {
		return fmt.Errorf("unit_price.per unit %q must match quantity unit %q", d.UnitPrice.Per.Unit, d.Quantity.Unit)
	}
	if err := d.Window.Validate(); err != nil {
		return fmt.Errorf("window: %w", err)
	}
	if err := d.Dimensions.Validate(); err != nil {
		return fmt.Errorf("dimensions: %w", err)
	}
	if err := d.Evidence.Validate(); err != nil {
		return fmt.Errorf("evidence: %w", err)
	}
	return nil
}

func (o CostObservation) Validate() error {
	if o.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version must be %q", SchemaVersion)
	}
	if err := validateID(o.ID); err != nil {
		return fmt.Errorf("id: %w", err)
	}
	if len(o.DriverIDs) == 0 {
		return errors.New("driver_ids must contain at least one driver")
	}
	seenDriverIDs := make(map[string]struct{}, len(o.DriverIDs))
	for i, driverID := range o.DriverIDs {
		if err := validateID(driverID); err != nil {
			return fmt.Errorf("driver_ids[%d]: %w", i, err)
		}
		if _, exists := seenDriverIDs[driverID]; exists {
			return fmt.Errorf("driver_ids[%d]: duplicate driver id %q", i, driverID)
		}
		seenDriverIDs[driverID] = struct{}{}
	}
	if err := o.Window.Validate(); err != nil {
		return fmt.Errorf("window: %w", err)
	}
	if err := o.Quantity.Validate(); err != nil {
		return fmt.Errorf("quantity: %w", err)
	}
	if err := o.UnitPrice.Validate(); err != nil {
		return fmt.Errorf("unit_price: %w", err)
	}
	if o.UnitPrice.Per.Unit != o.Quantity.Unit {
		return fmt.Errorf("unit_price.per unit %q must match quantity unit %q", o.UnitPrice.Per.Unit, o.Quantity.Unit)
	}
	if err := o.TotalCost.Validate(); err != nil {
		return fmt.Errorf("total_cost: %w", err)
	}
	if o.UnitPrice.Amount.Currency != o.TotalCost.Currency {
		return errors.New("unit_price and total_cost currencies must match")
	}
	if err := o.Dimensions.Validate(); err != nil {
		return fmt.Errorf("dimensions: %w", err)
	}
	claims := []struct {
		name     string
		evidence Evidence
	}{
		{"quantity", o.Evidence.Quantity},
		{"unit_price", o.Evidence.UnitPrice},
		{"total_cost", o.Evidence.TotalCost},
	}
	for _, claim := range claims {
		if err := claim.evidence.Validate(); err != nil {
			return fmt.Errorf("evidence.%s: %w", claim.name, err)
		}
		if err := validateClaimAuthority(claim.name, claim.evidence.Source.Type); err != nil {
			return fmt.Errorf("evidence.%s: %w", claim.name, err)
		}
	}
	if o.Evidence.TotalCost.Measurement == MeasurementDerived {
		expected := o.Quantity.Value * o.UnitPrice.Amount.Amount / o.UnitPrice.Per.Value
		tolerance := math.Max(1e-9, math.Abs(expected)*1e-9)
		if math.Abs(o.TotalCost.Amount-expected) > tolerance {
			return fmt.Errorf("derived total_cost amount %.12g must equal quantity times unit price %.12g", o.TotalCost.Amount, expected)
		}
	}
	return nil
}

func (q Quantity) Validate() error {
	if math.IsNaN(q.Value) || math.IsInf(q.Value, 0) || q.Value < 0 {
		return errors.New("value must be a finite number greater than or equal to zero")
	}
	return validateUnit(q.Unit)
}

func (m Money) Validate() error {
	if math.IsNaN(m.Amount) || math.IsInf(m.Amount, 0) || m.Amount < 0 {
		return errors.New("amount must be a finite number greater than or equal to zero")
	}
	if !currencyPattern.MatchString(m.Currency) {
		return errors.New("currency must be an uppercase ISO 4217 code")
	}
	return nil
}

func (p UnitPrice) Validate() error {
	if err := p.Amount.Validate(); err != nil {
		return fmt.Errorf("amount: %w", err)
	}
	if err := p.Per.Validate(); err != nil {
		return fmt.Errorf("per: %w", err)
	}
	if p.Per.Value <= 0 {
		return errors.New("per.value must be greater than zero")
	}
	return nil
}

func (w TimeWindow) Validate() error {
	start, err := time.Parse(time.RFC3339, w.Start)
	if err != nil {
		return errors.New("start must be an RFC3339 timestamp")
	}
	end, err := time.Parse(time.RFC3339, w.End)
	if err != nil {
		return errors.New("end must be an RFC3339 timestamp")
	}
	if !start.Before(end) {
		return errors.New("start must be before end")
	}
	return nil
}

func (d Dimensions) Validate() error {
	if strings.TrimSpace(d.Workload) == "" {
		return errors.New("workload is required")
	}
	return nil
}

func (e Evidence) Validate() error {
	if !validEvidenceKind(e.Kind) {
		return fmt.Errorf("unsupported evidence kind %q", e.Kind)
	}
	if !validMeasurementKind(e.Measurement) {
		return fmt.Errorf("unsupported measurement kind %q", e.Measurement)
	}
	if !validSourceType(e.Source.Type) {
		return fmt.Errorf("unsupported source type %q", e.Source.Type)
	}
	if !validConfidence(e.Confidence) {
		return fmt.Errorf("unsupported confidence %q", e.Confidence)
	}
	if strings.TrimSpace(e.ConfidenceRationale) == "" {
		return errors.New("confidence_rationale is required")
	}
	if strings.TrimSpace(e.Source.ArtifactIdentity) == "" && strings.TrimSpace(e.Source.URL) == "" {
		return errors.New("source requires artifact_identity or url")
	}
	if e.Source.URL != "" {
		parsed, err := url.ParseRequestURI(e.Source.URL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return errors.New("source.url must be an absolute URI")
		}
	}
	if !validCaptureTime(e.Source.CapturedAt) {
		return errors.New("source.captured_at must be an RFC3339 timestamp or ISO date")
	}
	if e.Source.Type == SourceProviderCatalog {
		if strings.TrimSpace(e.Source.RefreshOwner) == "" ||
			strings.TrimSpace(e.Source.RefreshCadence) == "" ||
			strings.TrimSpace(e.Source.StaleAfter) == "" {
			return errors.New("provider_catalog source requires refresh_owner, refresh_cadence, and stale_after")
		}
		if !validCaptureTime(e.Source.StaleAfter) {
			return errors.New("source.stale_after must be an RFC3339 timestamp or ISO date")
		}
		if e.Confidence == ConfidenceHigh {
			return errors.New("provider_catalog evidence cannot claim high confidence")
		}
	}
	if e.Measurement == MeasurementMeasured && e.Source.Type != SourceTelemetry &&
		e.Source.Type != SourceRuntimeLedger && e.Source.Type != SourceInvoice {
		return errors.New("measured evidence requires telemetry, runtime_ledger, or invoice source")
	}
	if e.Measurement == MeasurementSynthetic && e.Source.Type != SourceSyntheticFixture {
		return errors.New("synthetic measurement requires synthetic_fixture source")
	}
	if e.Kind == EvidenceBilled && e.Source.Type != SourceInvoice {
		return errors.New("billed evidence requires invoice source")
	}
	if e.Kind == EvidenceUserSupplied && e.Source.Type != SourceUserSupplied {
		return errors.New("user_supplied evidence requires user_supplied source")
	}
	if e.Kind == EvidenceObserved && e.Measurement != MeasurementMeasured && e.Measurement != MeasurementSynthetic {
		return errors.New("observed evidence must be measured or synthetic")
	}
	if e.Kind == EvidencePredicted && e.Measurement == MeasurementMeasured {
		return errors.New("predicted evidence cannot claim measured status")
	}
	return nil
}

func validateClaimAuthority(claim string, source SourceType) error {
	var allowed bool
	switch claim {
	case "quantity":
		switch source {
		case SourceTemplate, SourceUserSupplied, SourceRepoDetected, SourceTelemetry,
			SourceRuntimeLedger, SourceSyntheticFixture, SourceLegacyScenario:
			allowed = true
		}
	case "unit_price":
		switch source {
		case SourceTemplate, SourceUserSupplied, SourceRepoDetected, SourceInvoice,
			SourceProviderCatalog, SourceSyntheticFixture, SourceLegacyScenario:
			allowed = true
		}
	case "total_cost":
		switch source {
		case SourceTemplate, SourceUserSupplied, SourceInvoice,
			SourceProfitCtlDerived, SourceSyntheticFixture, SourceLegacyScenario:
			allowed = true
		}
	default:
		return fmt.Errorf("unsupported claim %q", claim)
	}
	if !allowed {
		return fmt.Errorf("source type %q is not authoritative for %s claim", source, claim)
	}
	return nil
}

func validateUnit(unit string) error {
	if !canonicalUnitPattern.MatchString(unit) {
		return errors.New("unit must be a non-empty canonical lowercase identifier")
	}
	switch unit {
	case "unit", "units", "gb", "mb", "kb":
		return fmt.Errorf("ambiguous unit %q is not allowed", unit)
	}
	return nil
}

func validateID(id string) error {
	if !idPattern.MatchString(id) {
		return errors.New("must be a non-empty canonical identifier")
	}
	return nil
}

func validCaptureTime(value string) bool {
	if _, err := time.Parse(time.RFC3339, value); err == nil {
		return true
	}
	_, err := time.Parse(time.DateOnly, value)
	return err == nil
}

func validDriverKind(value DriverKind) bool {
	switch value {
	case DriverFixed, DriverVariable, DriverCadence, DriverConcurrency, DriverUptime:
		return true
	default:
		return false
	}
}

func validEvidenceKind(value EvidenceKind) bool {
	switch value {
	case EvidencePredicted, EvidenceObserved, EvidenceBilled, EvidenceUserSupplied:
		return true
	default:
		return false
	}
}

func validMeasurementKind(value MeasurementKind) bool {
	switch value {
	case MeasurementMeasured, MeasurementSynthetic, MeasurementDerived, MeasurementDeclared:
		return true
	default:
		return false
	}
}

func validSourceType(value SourceType) bool {
	switch value {
	case SourceTemplate, SourceUserSupplied, SourceRepoDetected, SourceTelemetry,
		SourceRuntimeLedger, SourceInvoice, SourceProviderCatalog,
		SourceProfitCtlDerived, SourceSyntheticFixture, SourceLegacyScenario:
		return true
	default:
		return false
	}
}

func validConfidence(value Confidence) bool {
	switch value {
	case ConfidenceLow, ConfidenceMedium, ConfidenceHigh:
		return true
	default:
		return false
	}
}
