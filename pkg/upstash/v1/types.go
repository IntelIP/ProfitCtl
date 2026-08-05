// Package v1 adapts deterministic Upstash-shaped fixtures into a local cost
// reconciliation and normalized ledger input. It never calls a provider.
package v1

import (
	"fmt"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
	ledgerv1 "github.com/IntelIP/ProfitCtl/pkg/ledger/v1"
)

const (
	// FixtureSchemaVersion identifies the accepted synthetic fixture shape.
	FixtureSchemaVersion = "profitctl.upstash-fixture/v1"
	// ReconciliationSchemaVersion identifies the emitted reconciliation artifact.
	ReconciliationSchemaVersion = "profitctl.upstash-reconciliation/v1"
)

// Fixture is a deterministic, account-redacted input for the Upstash vertical
// slice. fixture_status is intentionally limited to synthetic in this slice.
type Fixture struct {
	SchemaVersion string            `json:"schema_version"`
	FixtureStatus string            `json:"fixture_status"`
	ScenarioID    string            `json:"scenario_id"`
	Source        FixtureSource     `json:"source"`
	Window        costv1.TimeWindow `json:"window"`
	Workload      string            `json:"workload"`
	Pricing       Pricing           `json:"pricing"`
	Polling       PollingWorkload   `json:"polling"`
	Telemetry     []TelemetryRecord `json:"telemetry"`
	Billing       *BillingExport    `json:"billing,omitempty"`
}

// FixtureSource gives the portable identity and capture time of a fixture or
// fixture-derived export. It contains no provider account identifier.
type FixtureSource struct {
	ArtifactIdentity string `json:"artifact_identity"`
	CapturedAt       string `json:"captured_at"`
}

// Pricing supplies a declared, synthetic command rate. It is not a current
// provider price claim.
type Pricing struct {
	UnitPrice costv1.UnitPrice `json:"unit_price"`
}

// PollingWorkload expresses the modeled polling inputs for one billing
// window. DeclaredDriver is a classification supplied by the fixture, not an
// inference from timestamps.
type PollingWorkload struct {
	IntervalSeconds int    `json:"interval_seconds"`
	WorkerReplicas  int    `json:"worker_replicas"`
	ActiveSeconds   int    `json:"active_seconds"`
	CommandsPerPoll int    `json:"commands_per_poll"`
	MissesPerPoll   int    `json:"misses_per_poll"`
	DeclaredDriver  string `json:"declared_driver"`
}

// TelemetryRecord is one synthetic Redis-operation export.
type TelemetryRecord struct {
	RecordID         string            `json:"record_id"`
	Window           costv1.TimeWindow `json:"window"`
	Commands         float64           `json:"commands"`
	Misses           float64           `json:"misses"`
	ArtifactIdentity string            `json:"artifact_identity"`
	CapturedAt       string            `json:"captured_at"`
}

// BillingExport is optional. An absent export is represented as unavailable,
// never as a zero-cost value.
type BillingExport struct {
	ExportID         string            `json:"export_id"`
	Window           costv1.TimeWindow `json:"window"`
	TotalCost        costv1.Money      `json:"total_cost"`
	ArtifactIdentity string            `json:"artifact_identity"`
	CapturedAt       string            `json:"captured_at"`
}

// Forecast contains the deterministic polling calculation for the window.
type Forecast struct {
	PollingIntervalSeconds int              `json:"polling_interval_seconds"`
	WorkerReplicas         int              `json:"worker_replicas"`
	ActiveSeconds          int              `json:"active_seconds"`
	CommandsPerPoll        int              `json:"commands_per_poll"`
	MissesPerPoll          int              `json:"misses_per_poll"`
	PollsPerWorker         float64          `json:"polls_per_worker"`
	PollingEvents          float64          `json:"polling_events"`
	PredictedCommandCount  float64          `json:"predicted_command_count"`
	PredictedMissCount     float64          `json:"predicted_miss_count"`
	UnitPrice              costv1.UnitPrice `json:"unit_price"`
	PredictedCost          costv1.Money     `json:"predicted_cost"`
}

// ObservedOperations is the de-duplicated telemetry aggregate.
type ObservedOperations struct {
	CommandCount            float64      `json:"command_count"`
	MissCount               float64      `json:"miss_count"`
	RecordsSeen             int          `json:"records_seen"`
	RecordsUsed             int          `json:"records_used"`
	DuplicateRecordsIgnored int          `json:"duplicate_records_ignored"`
	DerivedCost             costv1.Money `json:"derived_cost"`
}

// BillingEvidenceStatus distinguishes present fixture evidence from missing
// evidence without presenting missing evidence as a numeric zero.
type BillingEvidenceStatus string

const (
	BillingEvidenceAvailable   BillingEvidenceStatus = "available"
	BillingEvidenceUnavailable BillingEvidenceStatus = "unavailable"
)

// BillingSummary carries billing-window provenance when evidence is present.
type BillingSummary struct {
	EvidenceStatus    BillingEvidenceStatus `json:"evidence_status"`
	TotalCost         *costv1.Money         `json:"total_cost,omitempty"`
	Window            *costv1.TimeWindow    `json:"window,omitempty"`
	ArtifactIdentity  string                `json:"artifact_identity,omitempty"`
	CapturedAt        string                `json:"captured_at,omitempty"`
	UnavailableReason string                `json:"unavailable_reason,omitempty"`
}

// Variance compares modeled, observed, and (when supplied) fixture billing
// evidence. Observed cost is derived from observed commands and the declared
// unit price; it is not a billed-cost claim.
type Variance struct {
	ObservedMinusPredictedCommands float64       `json:"observed_minus_predicted_commands"`
	ObservedMinusPredictedCost     costv1.Money  `json:"observed_minus_predicted_cost"`
	BilledMinusObservedCost        *costv1.Money `json:"billed_minus_observed_cost,omitempty"`
}

// OperationalDriver identifies a fixture-declared driver and its modeled
// contribution. The explicit basis prevents temporal correlation from being
// represented as causal proof.
type OperationalDriver struct {
	Classification       string  `json:"classification"`
	Basis                string  `json:"basis"`
	ModeledPollingEvents float64 `json:"modeled_polling_events"`
	ModeledCommandCount  float64 `json:"modeled_command_count"`
}

// Reconciliation is the provider-free machine-readable result.
type Reconciliation struct {
	SchemaVersion             string             `json:"schema_version"`
	FixtureStatus             string             `json:"fixture_status"`
	ScenarioID                string             `json:"scenario_id"`
	Source                    FixtureSource      `json:"source"`
	Window                    costv1.TimeWindow  `json:"window"`
	Workload                  string             `json:"workload"`
	Forecast                  Forecast           `json:"forecast"`
	Observed                  ObservedOperations `json:"observed"`
	Billing                   BillingSummary     `json:"billing"`
	Variance                  Variance           `json:"variance"`
	DominantOperationalDriver OperationalDriver  `json:"dominant_operational_driver"`
	Limitations               []string           `json:"limitations"`
	LedgerObservationID       string             `json:"ledger_observation_id"`
}

// NormalizedLedgerInput is the PCTL-19 compatible, local-only ledger record
// derived from a reconciliation.
type NormalizedLedgerInput struct {
	Drivers     []costv1.CostDriver         `json:"-"`
	Observation costv1.CostObservation      `json:"-"`
	RawSource   ledgerv1.RawSourceReference `json:"-"`
}

// LedgerNormalization reports the local-ledger result embedded in the emitted
// artifact. The raw source includes the exact fixture digest.
type LedgerNormalization struct {
	Status        ledgerv1.IngestStatus       `json:"status"`
	ObservationID string                      `json:"observation_id"`
	RawSource     ledgerv1.RawSourceReference `json:"raw_source"`
}

// Artifact adds local-ledger status to the reconciliation result.
type Artifact struct {
	Reconciliation
	LedgerNormalization LedgerNormalization `json:"ledger_normalization"`
}

// NewArtifact makes a stable artifact after a reconciliation has been
// accepted by the local ledger.
func NewArtifact(reconciliation Reconciliation, ingest ledgerv1.IngestResult) (Artifact, error) {
	if reconciliation.LedgerObservationID != ingest.ObservationID {
		return Artifact{}, fmt.Errorf("ledger observation id %q does not match reconciliation %q", ingest.ObservationID, reconciliation.LedgerObservationID)
	}
	return Artifact{
		Reconciliation: reconciliation,
		LedgerNormalization: LedgerNormalization{
			Status:        ingest.Status,
			ObservationID: ingest.ObservationID,
			RawSource:     ingest.RawSource,
		},
	}, nil
}
