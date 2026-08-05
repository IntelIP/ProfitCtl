package v1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
)

// Load reads and validates an existing ledger document.
func Load(path string) (Ledger, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Ledger{}, err
	}

	var ledger Ledger
	if err := decodeStrictJSON(data, &ledger); err != nil {
		return Ledger{}, fmt.Errorf("decode ledger %q: %w", path, err)
	}
	ledger.canonicalize()
	if err := ledger.Validate(); err != nil {
		return Ledger{}, fmt.Errorf("validate ledger %q: %w", path, err)
	}
	return ledger, nil
}

// Open reads an existing ledger or returns a new local ledger when the target
// does not exist yet. It never creates a file by itself.
func Open(path string) (Ledger, error) {
	ledger, err := Load(path)
	if errors.Is(err, os.ErrNotExist) {
		return New(), nil
	}
	return ledger, err
}

// Save validates and atomically persists a canonical ledger document.
func Save(path string, ledger Ledger) error {
	ledger = clone(ledger)
	ledger.canonicalize()
	if err := ledger.Validate(); err != nil {
		return err
	}

	payload, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ledger: %w", err)
	}
	payload = append(payload, '\n')

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
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	return nil
}

// Ingest normalizes contract-v1 drivers and one observation into the ledger.
// The observation ID is its natural idempotency key: an identical retry is a
// no-op, while a changed payload with the same ID fails closed.
func (l *Ledger) Ingest(drivers []costv1.CostDriver, observation costv1.CostObservation, rawSource RawSourceReference) (IngestResult, error) {
	if l == nil {
		return IngestResult{}, errors.New("ledger is required")
	}
	if l.SchemaVersion == "" {
		*l = New()
	}
	if err := rawSource.Validate(); err != nil {
		return IngestResult{}, fmt.Errorf("raw source: %w", err)
	}
	for i, driver := range drivers {
		if err := driver.Validate(); err != nil {
			return IngestResult{}, fmt.Errorf("drivers[%d]: %w", i, err)
		}
	}
	if err := observation.Validate(); err != nil {
		return IngestResult{}, fmt.Errorf("observation: %w", err)
	}

	candidate := clone(*l)
	if err := candidate.Validate(); err != nil {
		return IngestResult{}, fmt.Errorf("existing ledger: %w", err)
	}

	driverByID := make(map[string]costv1.CostDriver, len(candidate.Drivers))
	for _, driver := range candidate.Drivers {
		driverByID[driver.ID] = driver
	}
	for _, driver := range drivers {
		if existing, exists := driverByID[driver.ID]; exists {
			if !reflect.DeepEqual(existing, driver) {
				return IngestResult{}, fmt.Errorf("driver id %q conflicts with existing ledger data", driver.ID)
			}
			continue
		}
		driverByID[driver.ID] = driver
		candidate.Drivers = append(candidate.Drivers, driver)
	}

	for _, driverID := range observation.DriverIDs {
		if _, exists := driverByID[driverID]; !exists {
			return IngestResult{}, fmt.Errorf("observation driver id %q is not present in ingest input or ledger", driverID)
		}
	}

	record := ObservationRecord{Observation: observation, RawSource: rawSource}
	for _, existing := range candidate.Observations {
		if existing.Observation.ID != observation.ID {
			continue
		}
		if !reflect.DeepEqual(existing, record) {
			return IngestResult{}, fmt.Errorf("observation id %q conflicts with existing ledger data", observation.ID)
		}
		return IngestResult{
			SchemaVersion: SchemaVersion,
			Status:        AlreadyPresent,
			ObservationID: observation.ID,
			RawSource:     rawSource,
		}, nil
	}

	candidate.Observations = append(candidate.Observations, record)
	candidate.canonicalize()
	if err := candidate.Validate(); err != nil {
		return IngestResult{}, err
	}
	*l = candidate
	return IngestResult{
		SchemaVersion: SchemaVersion,
		Status:        Ingested,
		ObservationID: observation.ID,
		RawSource:     rawSource,
	}, nil
}

// IngestFixture ingests the accepted PCTL-18 deterministic fixture shape.
// It records the fixture artifact identity and its SHA-256 digest, not a local
// absolute path, so the ledger remains portable and auditable.
func (l *Ledger) IngestFixture(path string) (IngestResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return IngestResult{}, err
	}

	var fixture contractFixture
	if err := decodeStrictJSON(data, &fixture); err != nil {
		return IngestResult{}, fmt.Errorf("decode fixture %q: %w", path, err)
	}
	if strings.TrimSpace(fixture.FixtureStatus) == "" {
		return IngestResult{}, fmt.Errorf("fixture %q: fixture_status is required", path)
	}
	if len(fixture.Drivers) == 0 {
		return IngestResult{}, fmt.Errorf("fixture %q: drivers is required", path)
	}

	identity, err := fixtureArtifactIdentity(fixture)
	if err != nil {
		return IngestResult{}, fmt.Errorf("fixture %q: %w", path, err)
	}
	digest := sha256.Sum256(data)
	return l.Ingest(fixture.Drivers, fixture.Observation, RawSourceReference{
		ArtifactIdentity: identity,
		SHA256:           hex.EncodeToString(digest[:]),
	})
}

// Query returns a deterministically ordered JSON-ready view of matching
// observations. Empty filters return every record.
func (l Ledger) Query(options QueryOptions) QueryResult {
	copy := clone(l)
	copy.canonicalize()

	result := QueryResult{
		SchemaVersion: SchemaVersion,
		Drivers:       []costv1.CostDriver{},
		Observations:  []ObservationRecord{},
	}
	referencedDrivers := make(map[string]struct{})
	for _, record := range copy.Observations {
		if options.ObservationID != "" && record.Observation.ID != options.ObservationID {
			continue
		}
		if options.Workload != "" && record.Observation.Dimensions.Workload != options.Workload {
			continue
		}
		result.Observations = append(result.Observations, record)
		for _, driverID := range record.Observation.DriverIDs {
			referencedDrivers[driverID] = struct{}{}
		}
	}
	for _, driver := range copy.Drivers {
		if _, exists := referencedDrivers[driver.ID]; exists {
			result.Drivers = append(result.Drivers, driver)
		}
	}
	return result
}

// MarshalQuery returns stable, indented JSON for a query result.
func MarshalQuery(result QueryResult) ([]byte, error) {
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

// MarshalIngest returns stable, indented JSON for an ingest result.
func MarshalIngest(result IngestResult) ([]byte, error) {
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

type contractFixture struct {
	FixtureStatus string                 `json:"fixture_status"`
	Note          string                 `json:"note"`
	Drivers       []costv1.CostDriver    `json:"drivers"`
	Observation   costv1.CostObservation `json:"observation"`
}

func fixtureArtifactIdentity(fixture contractFixture) (string, error) {
	evidence := []costv1.Evidence{
		fixture.Observation.Evidence.Quantity,
		fixture.Observation.Evidence.UnitPrice,
		fixture.Observation.Evidence.TotalCost,
	}
	for _, driver := range fixture.Drivers {
		evidence = append(evidence, driver.Evidence.Quantity, driver.Evidence.UnitPrice)
	}
	for _, claim := range evidence {
		identity := strings.TrimSpace(claim.Source.ArtifactIdentity)
		if identity != "" {
			if fragment := strings.Index(identity, "#"); fragment >= 0 {
				identity = identity[:fragment]
			}
			return identity, nil
		}
		if sourceURL := strings.TrimSpace(claim.Source.URL); sourceURL != "" {
			return sourceURL, nil
		}
	}
	return "", errors.New("fixture evidence requires artifact_identity or url")
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
