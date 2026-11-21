package decision

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"unicode"
)

type Summary struct {
	ID       string `json:"id"`
	Revision int    `json:"revision"`
	Project  string `json:"project"`
	Question string `json:"question"`
}

type ListResult struct {
	Cases      []Summary `json:"cases"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

// List finds saved cases by ID, project or question, in stable ID order.
// A cursor resumes after the last case examined. Sparse matches may return
// an empty page with a cursor. Evidence and outcomes stay inside read results.
func List(workspace, query, after string, limit int) (ListResult, error) {
	if len(query) > 256 || (after != "" && !slug.MatchString(after)) || limit < 0 || limit > 50 {
		return ListResult{}, invalid("query must fit 256 bytes, after must be a case ID, and limit must be between 0 and 50")
	}
	if limit == 0 {
		limit = 20
	}
	query = strings.ToLower(strings.TrimSpace(query))
	result := ListResult{Cases: []Summary{}}
	root, err := os.OpenRoot(workspace)
	if err != nil {
		return ListResult{}, err
	}
	defer root.Close()
	store, err := storeDirectory(root, false)
	if errors.Is(err, fs.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return ListResult{}, err
	}
	defer store.Close()
	entries, err := fs.ReadDir(store.FS(), ".") // sorted by filename
	if err != nil {
		return ListResult{}, err
	}
	const maxExamined = 200
	examined := 0
	cursor := after
	for _, e := range entries {
		id := e.Name()
		if id <= after || !e.IsDir() || !slug.MatchString(id) {
			continue
		}
		if len(result.Cases) == limit || examined == maxExamined {
			result.NextCursor = cursor
			break
		}
		examined++
		cursor = id
		dir, err := openDir(store, id, false, true)
		if err != nil {
			return ListResult{}, err
		}
		r, err := selectedRevision(dir, id, 0)
		dir.Close()
		if errors.Is(err, fs.ErrNotExist) {
			continue // interrupted creation before its first revision was published
		}
		if err != nil {
			return ListResult{}, err
		}
		if strings.Contains(id, query) || strings.Contains(strings.ToLower(r.Case.Project), query) ||
			strings.Contains(strings.ToLower(r.Case.Question), query) {
			result.Cases = append(result.Cases, Summary{id, r.Revision, summaryText(r.Case.Project, 120), summaryText(r.Case.Question, 280)})
		}
	}
	return result, nil
}

func summaryText(s string, limit int) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return ' '
		}
		return r
	}, s)
	runes := []rune(strings.Join(strings.Fields(s), " "))
	if len(runes) > limit {
		return string(runes[:limit-1]) + "…"
	}
	return string(runes)
}
