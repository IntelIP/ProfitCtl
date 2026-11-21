package decision

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/systemmodel"
	"gopkg.in/yaml.v3"
)

const privateIgnore = "*\n"

// Save appends a complete case snapshot if expectedRevision is current (zero
// creates a case). After an uncertain response the same base and input safely
// return the already published revision, even if later edits have landed.
func Save(workspace, id string, expectedRevision int, input Case) (Revision, error) {
	if !slug.MatchString(id) || expectedRevision < 0 || expectedRevision >= 1_000_000_000 {
		return Revision{}, invalid("supply a lowercase case ID of at most 64 letters, digits and hyphens, and a valid expected revision")
	}
	if err := input.validate(); err != nil {
		return Revision{}, err
	}
	root, err := os.OpenRoot(workspace)
	if err != nil {
		return Revision{}, err
	}
	defer root.Close()
	dir, err := caseDirectory(root, id, expectedRevision == 0)
	if expectedRevision > 0 && errors.Is(err, fs.ErrNotExist) {
		return Revision{}, ErrConflict
	}
	if err != nil {
		return Revision{}, err
	}
	defer dir.Close()
	next := expectedRevision + 1
	if existing, err := readRevision(dir, id, next); !errors.Is(err, fs.ErrNotExist) {
		return matchRetry(existing, input, err)
	}
	if expectedRevision > 0 {
		if _, err := readRevision(dir, id, expectedRevision); errors.Is(err, fs.ErrNotExist) {
			return Revision{}, ErrConflict
		} else if err != nil {
			return Revision{}, err
		}
	}
	if o := input.Outcome; o != nil {
		if o.ForecastRevision > expectedRevision {
			return Revision{}, invalid("outcome.forecast_revision must reference an earlier saved forecast")
		}
		prior, err := readRevision(dir, id, o.ForecastRevision)
		if errors.Is(err, fs.ErrNotExist) {
			return Revision{}, invalid("outcome.forecast_revision is missing")
		}
		if err != nil {
			return Revision{}, err
		}
		forecast := prior.Case
		if forecast.Project != input.Project || forecast.Workload != input.Workload ||
			forecast.Outcome != nil || blank(forecast.Recommendation) ||
			!slices.ContainsFunc(forecast.Alternatives, func(a Alternative) bool { return a.Name == input.Choice }) {
			return Revision{}, invalid("outcome requires an earlier forecast for the same project, workload and alternative")
		}
	}
	for _, calc := range input.Calculations {
		for _, input := range calc.Inputs {
			if err := verifyInput(root, input, calc); err != nil {
				return Revision{}, err
			}
		}
	}
	result := Revision{SchemaVersion: SchemaVersion, ID: id, Revision: next, RecordedAt: time.Now().UTC(), Case: input}
	result.RecordSHA256 = recordDigest(result)
	data, err := json.Marshal(result)
	if err != nil {
		return Revision{}, err
	}
	err = publish(dir, revisionFile(next), data)
	if errors.Is(err, fs.ErrExist) {
		existing, err := readRevision(dir, id, next)
		return matchRetry(existing, input, err)
	}
	if err != nil {
		return Revision{}, err
	}
	return result, nil
}

func matchRetry(existing Revision, input Case, err error) (Revision, error) {
	if err != nil {
		return Revision{}, err
	}
	// Compare the wire representation so omitted and empty optional lists are
	// equivalent before and after a JSON round trip.
	left, _ := json.Marshal(existing.Case)
	right, _ := json.Marshal(input)
	if !bytes.Equal(left, right) {
		return Revision{}, ErrConflict
	}
	return existing, nil
}

// Read returns an immutable revision; zero requests the latest one.
func Read(workspace, id string, revision int) (Revision, error) {
	if !slug.MatchString(id) || revision < 0 {
		return Revision{}, invalid("invalid decision ID or revision")
	}
	root, err := os.OpenRoot(workspace)
	if err != nil {
		return Revision{}, err
	}
	defer root.Close()
	dir, err := caseDirectory(root, id, false)
	if err != nil {
		return Revision{}, err
	}
	defer dir.Close()
	return selectedRevision(dir, id, revision)
}

func selectedRevision(dir *os.Root, id string, revision int) (Revision, error) {
	if revision == 0 {
		current, err := latest(dir)
		if err != nil {
			return Revision{}, err
		}
		revision = current
	}
	return readRevision(dir, id, revision)
}

func latest(dir *os.Root) (int, error) {
	file, err := dir.Open(".")
	if err != nil {
		return 0, err
	}
	defer file.Close()
	maximum := 0
	for {
		entries, err := file.ReadDir(128)
		for _, e := range entries {
			number, _ := strconv.Atoi(strings.TrimSuffix(e.Name(), ".json"))
			if number > maximum && e.Name() == revisionFile(number) {
				maximum = number
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	if maximum == 0 {
		return 0, fs.ErrNotExist
	}
	return maximum, nil
}

func revisionFile(n int) string { return strconv.Itoa(n) + ".json" }

// An unkeyed checksum detects accidental changes; workspace access controls
// establish the write boundary. Early local pilot revisions predate this field.
func recordDigest(r Revision) string {
	r.RecordSHA256 = ""
	data, _ := json.Marshal(r)
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func readRevision(dir *os.Root, id string, n int) (Revision, error) {
	f, err := openRegular(dir, revisionFile(n), true, MaxCaseBytes+1024)
	if err != nil {
		return Revision{}, err
	}
	defer f.Close()
	var r Revision
	if err := systemmodel.Decode(f, &r); err != nil || r.SchemaVersion != SchemaVersion || r.Revision != n ||
		r.ID != id || r.RecordedAt.IsZero() || r.Case.validate() != nil ||
		(r.RecordSHA256 != "" && r.RecordSHA256 != recordDigest(r)) {
		return Revision{}, errors.New("saved decision revision failed integrity validation")
	}
	return r, nil
}

func caseDirectory(root *os.Root, id string, create bool) (*os.Root, error) {
	store, err := storeDirectory(root, create)
	if err != nil {
		return nil, err
	}
	defer store.Close()
	if create {
		if err := ensureIgnored(store); err != nil {
			return nil, err
		}
	}
	return openDir(store, id, create, true)
}

func storeDirectory(root *os.Root, create bool) (*os.Root, error) {
	parent, err := openDir(root, ".profitctl", create, false)
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	return openDir(parent, "decisions", create, true)
}

// Work a component at a time. Operations are confined by the open
// directory handle; Lstat and SameFile also prevent redirection within it.
func openDir(parent *os.Root, name string, create, private bool) (*os.Root, error) {
	if create {
		err := parent.Mkdir(name, 0700)
		if err == nil {
			err = syncDirectory(parent)
		}
		if err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
	}
	info, err := parent.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || (private && exposed(info)) {
		return nil, errors.New("decision path requires a real directory with appropriate permissions")
	}
	dir, err := parent.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	opened, err := dir.Stat(".")
	if err != nil || !os.SameFile(info, opened) {
		dir.Close()
		return nil, errors.New("directory changed while opening")
	}
	return dir, nil
}

func openRegular(dir *os.Root, name string, private bool, maxBytes int64) (*os.File, error) {
	info, err := dir.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxBytes || (private && exposed(info)) {
		return nil, errors.New("expected a bounded regular file with appropriate permissions")
	}
	f, err := dir.Open(name)
	if err != nil {
		return nil, err
	}
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) {
		f.Close()
		return nil, errors.New("file changed while opening")
	}
	return f, nil
}

// Windows inherits the operator-selected workspace ACL. Go's POSIX mode
// bits do not express ACL grants there.
func exposed(info fs.FileInfo) bool {
	return runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0
}

func ensureIgnored(store *os.Root) error {
	f, err := openRegular(store, ".gitignore", false, 1024)
	if errors.Is(err, fs.ErrNotExist) {
		if err := publish(store, ".gitignore", []byte(privateIgnore)); !errors.Is(err, fs.ErrExist) {
			return err
		}
		f, err = openRegular(store, ".gitignore", false, 1024)
	}
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 1025))
	if err != nil || string(data) != privateIgnore {
		return errors.New("decision store .gitignore must contain only the wildcard rule")
	}
	return nil
}

// Write privately, flush, then link the complete file into its final name
// with create-if-absent semantics. Readers never see a partial revision and
// a second process cannot overwrite the winner. A crash may leave a harmless
// private .pending-* file, which readers ignore.
func publish(dir *os.Root, name string, data []byte) error {
	temp := ".pending-" + rand.Text()
	f, err := dir.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer dir.Remove(temp)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := dir.Link(temp, name); err != nil {
		return err
	}
	return syncDirectory(dir)
}

func syncDirectory(dir *os.Root) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	folder, err := dir.Open(".")
	if err != nil {
		return err
	}
	defer folder.Close()
	return folder.Sync()
}

func verifyInput(root *os.Root, input ArtifactInput, calc Calculation) error {
	dir := root
	parts := strings.Split(input.Path, "/")
	for _, part := range parts[:len(parts)-1] {
		child, err := openDir(dir, part, false, false)
		if dir != root {
			dir.Close()
		}
		if err != nil {
			return invalid("calculation input directory is unavailable or redirects")
		}
		dir = child
	}
	if dir != root {
		defer dir.Close()
	}
	const maxInputBytes = 2 << 20
	f, err := openRegular(dir, parts[len(parts)-1], false, maxInputBytes)
	if err != nil {
		return invalid("calculation input must be a readable workspace file of at most 2 MiB")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxInputBytes+1))
	digest := sha256.Sum256(data)
	if err != nil || len(data) > maxInputBytes || hex.EncodeToString(digest[:]) != input.SHA256 {
		return invalid("calculation input does not match its pinned sha256")
	}
	args := calc.Args
	if args[0] == "profitctl" {
		args = args[1:]
	}
	if slices.Contains([]string{"simulate", "validate", "compare"}, args[0]) && slices.Contains(args[1:], input.Path) {
		// The pricing CLI reads an optional calibration file relative to the
		// scenario. Pin that dependency as well so a replay uses the same data.
		var scenario struct {
			CalibrationFile string `yaml:"calibration_file"`
		}
		if err := yaml.Unmarshal(data, &scenario); err != nil {
			return invalid("calculation scenario could not be read as YAML")
		}
		if ref := strings.TrimPrefix(scenario.CalibrationFile, "./"); ref != "" {
			if !validInputPath(ref) || !slices.ContainsFunc(calc.Inputs, func(i ArtifactInput) bool {
				return i.Path == path.Join(path.Dir(input.Path), ref)
			}) {
				return invalid("the scenario calibration_file must also be an explicitly pinned workspace input")
			}
		}
	}
	return nil
}
