package v1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
	ledgerv1 "github.com/IntelIP/ProfitCtl/pkg/ledger/v1"
)

var canonicalIDPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,127}$`)

// ReconcileFile reads a strict synthetic fixture and derives both the
// reconciliation result and the local ledger input. It makes no network call.
func ReconcileFile(path string) (Reconciliation, NormalizedLedgerInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Reconciliation{}, NormalizedLedgerInput{}, err
	}

	var fixture Fixture
	if err := decodeStrictJSON(data, &fixture); err != nil {
		return Reconciliation{}, NormalizedLedgerInput{}, fmt.Errorf("decode Upstash fixture %q: %w", path, err)
	}
	digest := sha256.Sum256(data)
	return Reconcile(fixture, ledgerv1.RawSourceReference{
		ArtifactIdentity: fixture.Source.ArtifactIdentity,
		SHA256:           hex.EncodeToString(digest[:]),
	})
}

// Reconcile calculates one deterministic fixture without reading a provider.
func Reconcile(fixture Fixture, rawSource ledgerv1.RawSourceReference) (Reconciliation, NormalizedLedgerInput, error) {
	if err := fixture.Validate(); err != nil {
		return Reconciliation{}, NormalizedLedgerInput{}, err
	}
	if err := rawSource.Validate(); err != nil {
		return Reconciliation{}, NormalizedLedgerInput{}, fmt.Errorf("raw source: %w", err)
	}
	if rawSource.ArtifactIdentity != fixture.Source.ArtifactIdentity {
		return Reconciliation{}, NormalizedLedgerInput{}, fmt.Errorf("raw source artifact_identity must match fixture source")
	}

	pollsPerWorker := math.Ceil(float64(fixture.Polling.ActiveSeconds) / float64(fixture.Polling.IntervalSeconds))
	pollingEvents := pollsPerWorker * float64(fixture.Polling.WorkerReplicas)
	predictedCommands := pollingEvents * float64(fixture.Polling.CommandsPerPoll)
	predictedMisses := pollingEvents * float64(fixture.Polling.MissesPerPoll)
	if !finiteNonNegative(pollsPerWorker) || !finiteNonNegative(pollingEvents) || !finiteNonNegative(predictedCommands) || !finiteNonNegative(predictedMisses) {
		return Reconciliation{}, NormalizedLedgerInput{}, errors.New("polling forecast must remain finite and non-negative")
	}
	predictedCost, err := costForCommands(predictedCommands, fixture.Pricing.UnitPrice)
	if err != nil {
		return Reconciliation{}, NormalizedLedgerInput{}, err
	}

	observed, err := aggregateTelemetry(fixture.Telemetry)
	if err != nil {
		return Reconciliation{}, NormalizedLedgerInput{}, err
	}
	observed.DerivedCost, err = costForCommands(observed.CommandCount, fixture.Pricing.UnitPrice)
	if err != nil {
		return Reconciliation{}, NormalizedLedgerInput{}, err
	}

	billing := BillingSummary{
		EvidenceStatus:    BillingEvidenceUnavailable,
		UnavailableReason: "synthetic fixture does not include billing evidence",
	}
	if fixture.Billing != nil {
		cost := fixture.Billing.TotalCost
		window := fixture.Billing.Window
		billing = BillingSummary{
			EvidenceStatus:   BillingEvidenceAvailable,
			TotalCost:        &cost,
			Window:           &window,
			ArtifactIdentity: fixture.Billing.ArtifactIdentity,
			CapturedAt:       fixture.Billing.CapturedAt,
		}
	}

	variance := Variance{
		ObservedMinusPredictedCommands: observed.CommandCount - predictedCommands,
		ObservedMinusPredictedCost: costv1.Money{
			Amount:   observed.DerivedCost.Amount - predictedCost.Amount,
			Currency: predictedCost.Currency,
		},
	}
	if fixture.Billing != nil {
		billedMinusObserved := costv1.Money{
			Amount:   fixture.Billing.TotalCost.Amount - observed.DerivedCost.Amount,
			Currency: fixture.Billing.TotalCost.Currency,
		}
		variance.BilledMinusObservedCost = &billedMinusObserved
	}

	forecast := Forecast{
		PollingIntervalSeconds: fixture.Polling.IntervalSeconds,
		WorkerReplicas:         fixture.Polling.WorkerReplicas,
		ActiveSeconds:          fixture.Polling.ActiveSeconds,
		CommandsPerPoll:        fixture.Polling.CommandsPerPoll,
		MissesPerPoll:          fixture.Polling.MissesPerPoll,
		PollsPerWorker:         pollsPerWorker,
		PollingEvents:          pollingEvents,
		PredictedCommandCount:  predictedCommands,
		PredictedMissCount:     predictedMisses,
		UnitPrice:              fixture.Pricing.UnitPrice,
		PredictedCost:          predictedCost,
	}

	driver, observation := normalizedLedgerRecords(fixture, observed)
	if err := driver.Validate(); err != nil {
		return Reconciliation{}, NormalizedLedgerInput{}, fmt.Errorf("normalize driver: %w", err)
	}
	if err := observation.Validate(); err != nil {
		return Reconciliation{}, NormalizedLedgerInput{}, fmt.Errorf("normalize observation: %w", err)
	}

	reconciliation := Reconciliation{
		SchemaVersion: ReconciliationSchemaVersion,
		FixtureStatus: fixture.FixtureStatus,
		ScenarioID:    fixture.ScenarioID,
		Source:        fixture.Source,
		Window:        fixture.Window,
		Workload:      fixture.Workload,
		Forecast:      forecast,
		Observed:      observed,
		Billing:       billing,
		Variance:      variance,
		DominantOperationalDriver: OperationalDriver{
			Classification:       fixture.Polling.DeclaredDriver,
			Basis:                "declared_fixture_input",
			ModeledPollingEvents: pollingEvents,
			ModeledCommandCount:  predictedCommands,
		},
		Limitations: []string{
			"Synthetic fixture only; no provider API, export, credential, or current-price claim was used.",
			"Operational-driver classification is declared by the fixture and does not infer causality from temporal correlation.",
		},
		LedgerObservationID: observation.ID,
	}
	return reconciliation, NormalizedLedgerInput{
		Drivers:     []costv1.CostDriver{driver},
		Observation: observation,
		RawSource:   rawSource,
	}, nil
}

// MarshalArtifact encodes a stable machine-readable output document.
func MarshalArtifact(artifact Artifact) ([]byte, error) {
	payload, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

// WriteArtifact atomically persists an already-marshaled artifact with owner-
// only permissions. It never creates a remote resource.
func WriteArtifact(path string, payload []byte) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("artifact path is required")
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(path)+".")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

// Validate rejects ambiguous or live-shaped input before reconciliation.
func (fixture Fixture) Validate() error {
	if fixture.SchemaVersion != FixtureSchemaVersion {
		return fmt.Errorf("schema_version must be %q", FixtureSchemaVersion)
	}
	if fixture.FixtureStatus != "synthetic" {
		return errors.New("fixture_status must be synthetic for the provider-free vertical slice")
	}
	if !canonicalIDPattern.MatchString(fixture.ScenarioID) {
		return errors.New("scenario_id must be a canonical identifier")
	}
	if err := validateSyntheticSource(fixture.Source); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if err := fixture.Window.Validate(); err != nil {
		return fmt.Errorf("window: %w", err)
	}
	if err := (costv1.Dimensions{Workload: fixture.Workload}).Validate(); err != nil {
		return fmt.Errorf("workload: %w", err)
	}
	if err := fixture.Pricing.UnitPrice.Validate(); err != nil {
		return fmt.Errorf("pricing.unit_price: %w", err)
	}
	if fixture.Pricing.UnitPrice.Per.Value != 1 || fixture.Pricing.UnitPrice.Per.Unit != "command" {
		return errors.New("pricing.unit_price must be expressed per one command")
	}
	if err := validateDeclaredPriceEvidence(fixture.Source); err != nil {
		return fmt.Errorf("pricing: %w", err)
	}
	if err := fixture.Polling.Validate(); err != nil {
		return fmt.Errorf("polling: %w", err)
	}
	if len(fixture.Telemetry) == 0 {
		return errors.New("telemetry is required")
	}
	for index, record := range fixture.Telemetry {
		if err := record.Validate(fixture.Window); err != nil {
			return fmt.Errorf("telemetry[%d]: %w", index, err)
		}
	}
	if fixture.Billing != nil {
		if err := fixture.Billing.Validate(fixture.Window, fixture.Pricing.UnitPrice.Amount.Currency); err != nil {
			return fmt.Errorf("billing: %w", err)
		}
	}
	return nil
}

// Validate proves polling inputs produce a bounded, deterministic forecast.
func (polling PollingWorkload) Validate() error {
	if polling.IntervalSeconds <= 0 {
		return errors.New("interval_seconds must be greater than zero")
	}
	if polling.WorkerReplicas <= 0 {
		return errors.New("worker_replicas must be greater than zero")
	}
	if polling.ActiveSeconds <= 0 {
		return errors.New("active_seconds must be greater than zero")
	}
	if polling.CommandsPerPoll <= 0 {
		return errors.New("commands_per_poll must be greater than zero")
	}
	if polling.MissesPerPoll < 0 {
		return errors.New("misses_per_poll must be greater than or equal to zero")
	}
	switch polling.DeclaredDriver {
	case "idle_polling", "active_workload":
		return nil
	default:
		return fmt.Errorf("declared_driver %q is not supported", polling.DeclaredDriver)
	}
}

// Validate ensures telemetry belongs to the reconciled billing window.
func (record TelemetryRecord) Validate(window costv1.TimeWindow) error {
	if !canonicalIDPattern.MatchString(record.RecordID) {
		return errors.New("record_id must be a canonical identifier")
	}
	if err := record.Window.Validate(); err != nil {
		return fmt.Errorf("window: %w", err)
	}
	if record.Window != window {
		return errors.New("window must exactly match fixture window")
	}
	if !finiteNonNegative(record.Commands) {
		return errors.New("commands must be finite and non-negative")
	}
	if !finiteNonNegative(record.Misses) {
		return errors.New("misses must be finite and non-negative")
	}
	return validateSyntheticSource(FixtureSource{
		ArtifactIdentity: record.ArtifactIdentity,
		CapturedAt:       record.CapturedAt,
	})
}

// Validate ensures billing evidence has matching window, currency, and
// synthetic provenance. It does not treat the value as a live invoice claim.
func (billing BillingExport) Validate(window costv1.TimeWindow, currency string) error {
	if !canonicalIDPattern.MatchString(billing.ExportID) {
		return errors.New("export_id must be a canonical identifier")
	}
	if err := billing.Window.Validate(); err != nil {
		return fmt.Errorf("window: %w", err)
	}
	if billing.Window != window {
		return errors.New("window must exactly match fixture window")
	}
	if err := billing.TotalCost.Validate(); err != nil {
		return fmt.Errorf("total_cost: %w", err)
	}
	if billing.TotalCost.Currency != currency {
		return fmt.Errorf("total_cost currency %q must match pricing currency %q", billing.TotalCost.Currency, currency)
	}
	return validateSyntheticSource(FixtureSource{
		ArtifactIdentity: billing.ArtifactIdentity,
		CapturedAt:       billing.CapturedAt,
	})
}

func aggregateTelemetry(records []TelemetryRecord) (ObservedOperations, error) {
	seen := make(map[string]TelemetryRecord, len(records))
	result := ObservedOperations{RecordsSeen: len(records)}
	for _, record := range records {
		if existing, exists := seen[record.RecordID]; exists {
			if !reflect.DeepEqual(existing, record) {
				return ObservedOperations{}, fmt.Errorf("duplicate telemetry record_id %q conflicts with prior data", record.RecordID)
			}
			result.DuplicateRecordsIgnored++
			continue
		}
		seen[record.RecordID] = record
		result.RecordsUsed++
		result.CommandCount += record.Commands
		result.MissCount += record.Misses
	}
	if !finiteNonNegative(result.CommandCount) || !finiteNonNegative(result.MissCount) {
		return ObservedOperations{}, errors.New("aggregated telemetry must remain finite and non-negative")
	}
	return result, nil
}

func normalizedLedgerRecords(fixture Fixture, observed ObservedOperations) (costv1.CostDriver, costv1.CostObservation) {
	driverID := "upstash-" + fixture.ScenarioID + "-polling-cadence"
	observationID := "upstash-" + fixture.ScenarioID + "-observed"
	dimensions := costv1.Dimensions{
		Provider: "upstash",
		Service:  "redis",
		Tier:     "synthetic",
		Workload: fixture.Workload,
	}
	driver := costv1.CostDriver{
		SchemaVersion: costv1.SchemaVersion,
		ID:            driverID,
		Name:          "Declared Redis polling commands",
		Kind:          costv1.DriverCadence,
		Quantity: costv1.Quantity{
			Value: float64(fixture.Polling.WorkerReplicas) * float64(fixture.Polling.CommandsPerPoll),
			Unit:  "command",
		},
		Per:        &costv1.Quantity{Value: float64(fixture.Polling.IntervalSeconds), Unit: "second"},
		UnitPrice:  fixture.Pricing.UnitPrice,
		Window:     fixture.Window,
		Dimensions: dimensions,
		Evidence: costv1.DriverEvidence{
			Quantity:  declaredPriceEvidence(fixture.Source, "polling", "Declared synthetic polling inputs; no runtime measurement claim."),
			UnitPrice: declaredPriceEvidence(fixture.Source, "pricing", "Synthetic declared command rate; not a current provider price."),
		},
	}
	observation := costv1.CostObservation{
		SchemaVersion: costv1.SchemaVersion,
		ID:            observationID,
		DriverIDs:     []string{driverID},
		Window:        fixture.Window,
		Quantity:      costv1.Quantity{Value: observed.CommandCount, Unit: "command"},
		UnitPrice:     fixture.Pricing.UnitPrice,
		TotalCost:     observed.DerivedCost,
		Dimensions:    dimensions,
		Evidence: costv1.ClaimEvidence{
			Quantity:  observedTelemetryEvidence(fixture.Source),
			UnitPrice: declaredPriceEvidence(fixture.Source, "pricing", "Synthetic declared command rate; not a current provider price."),
			TotalCost: costv1.Evidence{
				Kind:        costv1.EvidencePredicted,
				Measurement: costv1.MeasurementDerived,
				Source: costv1.SourceReference{
					Type:             costv1.SourceProfitCtlDerived,
					ArtifactIdentity: "profitctl://upstash/reconciliation/" + fixture.ScenarioID + "/observed-cost",
					CapturedAt:       fixture.Source.CapturedAt,
				},
				Confidence:          costv1.ConfidenceLow,
				ConfidenceRationale: "Derived from synthetic observed commands and a declared synthetic command rate.",
			},
		},
	}
	return driver, observation
}

func validateSyntheticSource(source FixtureSource) error {
	evidence := observedSyntheticEvidence(source, "Synthetic fixture provenance.")
	return evidence.Validate()
}

func validateDeclaredPriceEvidence(source FixtureSource) error {
	evidence := declaredPriceEvidence(source, "pricing", "Synthetic declared command rate; not a current provider price.")
	return evidence.Validate()
}

func observedTelemetryEvidence(source FixtureSource) costv1.Evidence {
	return observedSyntheticEvidence(FixtureSource{
		ArtifactIdentity: source.ArtifactIdentity + "#telemetry",
		CapturedAt:       source.CapturedAt,
	}, "Synthetic telemetry aggregate; no runtime measurement claim.")
}

func observedSyntheticEvidence(source FixtureSource, rationale string) costv1.Evidence {
	return costv1.Evidence{
		Kind:        costv1.EvidenceObserved,
		Measurement: costv1.MeasurementSynthetic,
		Source: costv1.SourceReference{
			Type:             costv1.SourceSyntheticFixture,
			ArtifactIdentity: source.ArtifactIdentity,
			CapturedAt:       source.CapturedAt,
		},
		Confidence:          costv1.ConfidenceLow,
		ConfidenceRationale: rationale,
	}
}

func declaredPriceEvidence(source FixtureSource, fragment, rationale string) costv1.Evidence {
	return costv1.Evidence{
		Kind:        costv1.EvidenceUserSupplied,
		Measurement: costv1.MeasurementDeclared,
		Source: costv1.SourceReference{
			Type:             costv1.SourceUserSupplied,
			ArtifactIdentity: source.ArtifactIdentity + "#" + fragment,
			CapturedAt:       source.CapturedAt,
		},
		Confidence:          costv1.ConfidenceLow,
		ConfidenceRationale: rationale,
	}
}

func costForCommands(commands float64, price costv1.UnitPrice) (costv1.Money, error) {
	amount := commands * price.Amount.Amount / price.Per.Value
	if !finiteNonNegative(amount) {
		return costv1.Money{}, errors.New("command cost must remain finite and non-negative")
	}
	return costv1.Money{Amount: amount, Currency: price.Amount.Currency}, nil
}

func finiteNonNegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}
