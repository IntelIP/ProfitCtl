// Package v1 provides a local, file-backed normalized cost ledger.
package v1

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
)

// SchemaVersion identifies the persisted local-ledger document. Ledger records
// embed the additive profitctl.cost/v1 driver and observation contracts.
const SchemaVersion = "profitctl.ledger/v1"

var sha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// RawSourceReference preserves a stable pointer and content digest for the
// input artifact that produced a normalized observation.
type RawSourceReference struct {
	ArtifactIdentity string `json:"artifact_identity"`
	SHA256           string `json:"sha256"`
}

// ObservationRecord is one normalized observation plus its immutable input
// reference. Claim-level evidence remains on the embedded cost observation.
type ObservationRecord struct {
	Observation costv1.CostObservation `json:"observation"`
	RawSource   RawSourceReference     `json:"raw_source"`
}

// Ledger is a deterministic local document. Drivers and observations are
// ordered by ID before persistence and query output.
type Ledger struct {
	SchemaVersion string              `json:"schema_version"`
	Drivers       []costv1.CostDriver `json:"drivers"`
	Observations  []ObservationRecord `json:"observations"`
}

// QueryOptions narrows a query without changing the underlying ledger.
type QueryOptions struct {
	ObservationID string
	Workload      string
}

// QueryResult is the stable machine-readable ledger query envelope.
type QueryResult struct {
	SchemaVersion string              `json:"schema_version"`
	Drivers       []costv1.CostDriver `json:"drivers"`
	Observations  []ObservationRecord `json:"observations"`
}

// IngestStatus describes whether an observation changed the ledger.
type IngestStatus string

const (
	Ingested       IngestStatus = "ingested"
	AlreadyPresent IngestStatus = "already_present"
)

// IngestResult is the stable machine-readable ingest outcome.
type IngestResult struct {
	SchemaVersion string             `json:"schema_version"`
	Status        IngestStatus       `json:"status"`
	ObservationID string             `json:"observation_id"`
	RawSource     RawSourceReference `json:"raw_source"`
}

// New returns an empty valid ledger.
func New() Ledger {
	return Ledger{
		SchemaVersion: SchemaVersion,
		Drivers:       []costv1.CostDriver{},
		Observations:  []ObservationRecord{},
	}
}

func (r RawSourceReference) Validate() error {
	if strings.TrimSpace(r.ArtifactIdentity) == "" {
		return fmt.Errorf("artifact_identity is required")
	}
	if !sha256Pattern.MatchString(r.SHA256) {
		return fmt.Errorf("sha256 must be a lowercase SHA-256 digest")
	}
	return nil
}

// Validate proves that the persisted document is internally consistent and
// that every normalized value continues to satisfy the cost-v1 contract.
func (l Ledger) Validate() error {
	if l.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version must be %q", SchemaVersion)
	}

	drivers := make(map[string]struct{}, len(l.Drivers))
	for i, driver := range l.Drivers {
		if err := driver.Validate(); err != nil {
			return fmt.Errorf("drivers[%d]: %w", i, err)
		}
		if _, exists := drivers[driver.ID]; exists {
			return fmt.Errorf("drivers[%d]: duplicate driver id %q", i, driver.ID)
		}
		drivers[driver.ID] = struct{}{}
	}

	observations := make(map[string]struct{}, len(l.Observations))
	for i, record := range l.Observations {
		if err := record.Observation.Validate(); err != nil {
			return fmt.Errorf("observations[%d].observation: %w", i, err)
		}
		if err := record.RawSource.Validate(); err != nil {
			return fmt.Errorf("observations[%d].raw_source: %w", i, err)
		}
		if _, exists := observations[record.Observation.ID]; exists {
			return fmt.Errorf("observations[%d]: duplicate observation id %q", i, record.Observation.ID)
		}
		observations[record.Observation.ID] = struct{}{}
		for _, driverID := range record.Observation.DriverIDs {
			if _, exists := drivers[driverID]; !exists {
				return fmt.Errorf("observations[%d]: driver id %q is not present in the ledger", i, driverID)
			}
		}
	}
	return nil
}

func (l *Ledger) canonicalize() {
	sort.Slice(l.Drivers, func(i, j int) bool {
		return l.Drivers[i].ID < l.Drivers[j].ID
	})
	sort.Slice(l.Observations, func(i, j int) bool {
		return l.Observations[i].Observation.ID < l.Observations[j].Observation.ID
	})
}

func clone(l Ledger) Ledger {
	result := Ledger{
		SchemaVersion: l.SchemaVersion,
		Drivers:       append([]costv1.CostDriver(nil), l.Drivers...),
		Observations:  append([]ObservationRecord(nil), l.Observations...),
	}
	if result.Drivers == nil {
		result.Drivers = []costv1.CostDriver{}
	}
	if result.Observations == nil {
		result.Observations = []ObservationRecord{}
	}
	return result
}
