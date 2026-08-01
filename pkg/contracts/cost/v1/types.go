// Package v1 defines ProfitCtl's first versioned cost-driver and
// cost-observation contract.
package v1

const SchemaVersion = "profitctl.cost/v1"

type DriverKind string

const (
	DriverFixed       DriverKind = "fixed"
	DriverVariable    DriverKind = "variable"
	DriverCadence     DriverKind = "cadence"
	DriverConcurrency DriverKind = "concurrency"
	DriverUptime      DriverKind = "uptime"
)

type EvidenceKind string

const (
	EvidencePredicted    EvidenceKind = "predicted"
	EvidenceObserved     EvidenceKind = "observed"
	EvidenceBilled       EvidenceKind = "billed"
	EvidenceUserSupplied EvidenceKind = "user_supplied"
)

type MeasurementKind string

const (
	MeasurementMeasured  MeasurementKind = "measured"
	MeasurementSynthetic MeasurementKind = "synthetic"
	MeasurementDerived   MeasurementKind = "derived"
	MeasurementDeclared  MeasurementKind = "declared"
)

type SourceType string

const (
	SourceTemplate         SourceType = "template"
	SourceUserSupplied     SourceType = "user_supplied"
	SourceRepoDetected     SourceType = "repo_detected"
	SourceTelemetry        SourceType = "telemetry"
	SourceRuntimeLedger    SourceType = "runtime_ledger"
	SourceInvoice          SourceType = "invoice"
	SourceProviderCatalog  SourceType = "provider_catalog"
	SourceProfitCtlDerived SourceType = "profitctl_derived"
	SourceSyntheticFixture SourceType = "synthetic_fixture"
	SourceLegacyScenario   SourceType = "legacy_scenario"
)

type Confidence string

const (
	ConfidenceLow    Confidence = "low"
	ConfidenceMedium Confidence = "medium"
	ConfidenceHigh   Confidence = "high"
)

type Quantity struct {
	Value float64 `json:"value" yaml:"value"`
	Unit  string  `json:"unit" yaml:"unit"`
}

type Money struct {
	Amount   float64 `json:"amount" yaml:"amount"`
	Currency string  `json:"currency" yaml:"currency"`
}

type UnitPrice struct {
	Amount Money    `json:"amount" yaml:"amount"`
	Per    Quantity `json:"per" yaml:"per"`
}

type TimeWindow struct {
	Start string `json:"start" yaml:"start"`
	End   string `json:"end" yaml:"end"`
}

type Dimensions struct {
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`
	Service  string `json:"service,omitempty" yaml:"service,omitempty"`
	Region   string `json:"region,omitempty" yaml:"region,omitempty"`
	Tier     string `json:"tier,omitempty" yaml:"tier,omitempty"`
	Workload string `json:"workload" yaml:"workload"`
}

type SourceReference struct {
	Type             SourceType `json:"type" yaml:"type"`
	ArtifactIdentity string     `json:"artifact_identity,omitempty" yaml:"artifact_identity,omitempty"`
	URL              string     `json:"url,omitempty" yaml:"url,omitempty"`
	CapturedAt       string     `json:"captured_at" yaml:"captured_at"`
	RefreshOwner     string     `json:"refresh_owner,omitempty" yaml:"refresh_owner,omitempty"`
	RefreshCadence   string     `json:"refresh_cadence,omitempty" yaml:"refresh_cadence,omitempty"`
	StaleAfter       string     `json:"stale_after,omitempty" yaml:"stale_after,omitempty"`
}

// Evidence classifies a claim separately from its provenance. A source label
// alone never establishes that a value was measured.
type Evidence struct {
	Kind                EvidenceKind    `json:"kind" yaml:"kind"`
	Measurement         MeasurementKind `json:"measurement" yaml:"measurement"`
	Source              SourceReference `json:"source" yaml:"source"`
	Confidence          Confidence      `json:"confidence" yaml:"confidence"`
	ConfidenceRationale string          `json:"confidence_rationale" yaml:"confidence_rationale"`
}

type CostDriver struct {
	SchemaVersion string     `json:"schema_version" yaml:"schema_version"`
	ID            string     `json:"id" yaml:"id"`
	Name          string     `json:"name" yaml:"name"`
	Kind          DriverKind `json:"kind" yaml:"kind"`
	Quantity      Quantity   `json:"quantity" yaml:"quantity"`
	Per           *Quantity  `json:"per,omitempty" yaml:"per,omitempty"`
	UnitPrice     UnitPrice  `json:"unit_price" yaml:"unit_price"`
	Window        TimeWindow `json:"window" yaml:"window"`
	Dimensions    Dimensions `json:"dimensions" yaml:"dimensions"`
	Evidence      Evidence   `json:"evidence" yaml:"evidence"`
}

type ClaimEvidence struct {
	Quantity  Evidence `json:"quantity" yaml:"quantity"`
	UnitPrice Evidence `json:"unit_price" yaml:"unit_price"`
	TotalCost Evidence `json:"total_cost" yaml:"total_cost"`
}

type CostObservation struct {
	SchemaVersion string        `json:"schema_version" yaml:"schema_version"`
	ID            string        `json:"id" yaml:"id"`
	DriverIDs     []string      `json:"driver_ids" yaml:"driver_ids"`
	Window        TimeWindow    `json:"window" yaml:"window"`
	Quantity      Quantity      `json:"quantity" yaml:"quantity"`
	UnitPrice     UnitPrice     `json:"unit_price" yaml:"unit_price"`
	TotalCost     Money         `json:"total_cost" yaml:"total_cost"`
	Dimensions    Dimensions    `json:"dimensions" yaml:"dimensions"`
	Evidence      ClaimEvidence `json:"evidence" yaml:"evidence"`
}
