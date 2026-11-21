package decision

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func example() Case {
	return Case{
		Project: "example", Workload: "accepted reports", SourceRevision: "fixture-v1",
		Question: "Does caching pay?", Window: "one week",
		Alternatives: []Alternative{
			{Name: "baseline", Cost: "Unmeasured", Quality: "Refresh on every request"},
			{Name: "cache", Cost: "Unmeasured", Quality: "Bounded staleness"},
		},
		Evidence: []Evidence{{Kind: "assumed", Source: "case owner", Finding: "Repeat lookups are possible"}},
		Missing:  []string{"Count repeat requests"}, Recommendation: "Retain baseline pending measurement",
		ReversalCondition: "Adopt caching if savings cover upkeep within the freshness budget",
	}
}

func TestRevisionHistoryAndObservedOutcome(t *testing.T) {
	root := t.TempDir()
	c := example()
	first, err := Save(root, "case-1", 0, c)
	require.NoError(t, err)
	require.Equal(t, 1, first.Revision)
	c.Choice = "baseline"
	c.Outcome = &Outcome{ForecastRevision: 1, Observed: "No repeated requests in fixture week",
		EvidenceSource: "fixture usage extract", Conclusion: "No cache hits possible"}
	_, err = Save(root, "case-1", 1, c)
	require.ErrorIs(t, err, ErrInvalid)
	c.Evidence = append(c.Evidence, Evidence{Kind: "observed", Source: "fixture usage extract", Finding: "0 repeated requests"})
	second, err := Save(root, "case-1", 1, c)
	require.NoError(t, err)
	require.Equal(t, 2, second.Revision)
	result, err := Read(root, "case-1", 0)
	require.NoError(t, err)
	require.Equal(t, second, result)
	result, err = Read(root, "case-1", 1)
	require.NoError(t, err)
	require.Equal(t, first, result)
	result, err = Save(root, "case-1", 0, example())
	require.NoError(t, err)
	require.Equal(t, first, result)
	c.Outcome.ForecastRevision = 2 // points to an outcome, not a forecast
	_, err = Save(root, "case-1", 2, c)
	require.ErrorIs(t, err, ErrInvalid)
	c.Outcome.ForecastRevision = 3
	_, err = Save(root, "case-1", 2, c)
	require.ErrorIs(t, err, ErrInvalid)
	c.Outcome.ForecastRevision = 1
	c.Workload = "all page views"
	_, err = Save(root, "case-1", 2, c)
	require.ErrorIs(t, err, ErrInvalid)
	_, err = Save(root, "other-id", 0, c)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestConcurrentRevisionAndPrivatePublication(t *testing.T) {
	root := t.TempDir()
	_, err := Save(root, "case", 0, example())
	require.NoError(t, err)
	store := filepath.Join(root, ".profitctl", "decisions")
	require.FileExists(t, filepath.Join(store, ".gitignore"))
	caseDir := filepath.Join(store, "case")
	require.NoError(t, os.WriteFile(filepath.Join(caseDir, ".pending-interrupted"), []byte("partial"), 0600))
	variants := []string{"baseline", "cache"}
	errorsCh := make(chan error, len(variants))
	for _, choice := range variants {
		go func() {
			c := example()
			c.Choice = choice
			_, err := Save(root, "case", 1, c)
			errorsCh <- err
		}()
	}
	var success, conflict int
	for range variants {
		switch err := <-errorsCh; {
		case err == nil:
			success++
		case errors.Is(err, ErrConflict):
			conflict++
		default:
			t.Fatalf("unexpected save error: %v", err)
		}
	}
	require.Equal(t, 1, success)
	require.Equal(t, 1, conflict)
	got, err := Read(root, "case", 0)
	require.NoError(t, err)
	require.Equal(t, 2, got.Revision)
	retry, err := Save(root, "case", 1, got.Case)
	require.NoError(t, err)
	require.Equal(t, got, retry)
	info, err := os.Stat(filepath.Join(caseDir, "2.json"))
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		require.Equal(t, fs.FileMode(0600), info.Mode().Perm())
		dirInfo, err := os.Stat(caseDir)
		require.NoError(t, err)
		require.Equal(t, fs.FileMode(0700), dirInfo.Mode().Perm())
	}
}

func TestPinnedInputsAndRetryAfterInputChanges(t *testing.T) {
	root := t.TempDir()
	contents := []byte("{}\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, "sample.json"), contents, 0600))
	digest := sha256.Sum256(contents)
	c := example()
	c.Calculations = []Calculation{{Args: []string{"system", "outcomes", "sample.json"},
		Inputs: []ArtifactInput{{Path: "sample.json", SHA256: hex.EncodeToString(digest[:])}}}}
	_, err := Save(root, "case", 0, c)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "sample.json"), []byte("changed"), 0600))
	_, err = Save(root, "case", 0, c) // original save did complete
	require.NoError(t, err)
	_, err = Save(root, "case", 1, c) // new evidence must be current
	require.ErrorIs(t, err, ErrInvalid)
	_, err = Read(root, "case", 0)
	require.NoError(t, err) // historical records are readable without their original inputs
	for _, invalidPath := range []string{"../outside.json", ".profitctl/decisions/case/1.json", "dir/.private/x.yml", "/abs.csv", "C:\\file.json"} {
		c.Calculations[0].Args[2] = invalidPath
		c.Calculations[0].Inputs[0].Path = invalidPath
		_, err = Save(root, "case", 1, c)
		require.ErrorIs(t, err, ErrInvalid)
	}
	_, err = Save(root, "../case", 0, example())
	require.ErrorIs(t, err, ErrInvalid)
}

func TestCaseChecksumAndCalibrationDependency(t *testing.T) {
	root := t.TempDir()
	pin := func(name, data string) ArtifactInput {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(data), 0600))
		sum := sha256.Sum256([]byte(data))
		return ArtifactInput{name, hex.EncodeToString(sum[:])}
	}
	scenario := pin("project.yml", "calibration_file: usage.csv\n")
	usage := pin("usage.csv", "metric,value\nexample,1\n")
	c := example()
	c.Calculations = []Calculation{{Args: []string{"validate", "-f", scenario.Path}, Inputs: []ArtifactInput{scenario}}}
	_, err := Save(root, "case", 0, c)
	require.ErrorIs(t, err, ErrInvalid)
	c.Calculations[0].Inputs = append(c.Calculations[0].Inputs, usage)
	c.Calculations[0].Args = []string{"simulate", "--json", scenario.Path}
	_, err = Save(root, "case", 0, c)
	require.ErrorIs(t, err, ErrInvalid) // simulate would read the unpinned default profit.yml
	require.ErrorContains(t, err, "simulate --file SCENARIO")
	c.Calculations[0].Args = []string{"simulate", "--scenario", scenario.Path}
	_, err = Save(root, "case", 0, c)
	require.ErrorContains(t, err, "simulate --file SCENARIO")
	c.Calculations[0].Args = []string{"validate", "-f", scenario.Path}
	r, err := Save(root, "case", 0, c)
	require.NoError(t, err)
	require.Len(t, r.RecordSHA256, 64)
	r.Case.Recommendation = "different recommendation"
	tampered, err := json.Marshal(r)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, ".profitctl", "decisions", "case", "1.json"), tampered, 0600))
	_, err = Read(root, "case", 1)
	require.ErrorContains(t, err, "integrity")
}

func TestRedirectedDirectoriesAndRevisionDestinations(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires a Windows account with appropriate privilege")
	}
	root := t.TempDir()
	other := t.TempDir()
	require.NoError(t, os.Symlink(other, filepath.Join(root, ".profitctl")))
	_, err := Save(root, "case", 0, example())
	require.Error(t, err)
	entries, err := os.ReadDir(other)
	require.NoError(t, err)
	require.Empty(t, entries)
	require.NoError(t, os.Remove(filepath.Join(root, ".profitctl")))
	_, err = Save(root, "case", 0, example())
	require.NoError(t, err)
	sentinel := filepath.Join(other, "sentinel")
	require.NoError(t, os.WriteFile(sentinel, []byte("preserved"), 0600))
	target := filepath.Join(root, ".profitctl", "decisions", "case", "2.json")
	for _, link := range []func(string, string) error{os.Link, os.Symlink} {
		require.NoError(t, link(sentinel, target))
		_, err = Save(root, "case", 1, example())
		require.Error(t, err)
		unchanged, err := os.ReadFile(sentinel)
		require.NoError(t, err)
		require.Equal(t, "preserved", string(unchanged))
		require.NoError(t, os.Remove(target))
	}
}
