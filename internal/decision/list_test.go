package decision

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindSavedDecisions(t *testing.T) {
	root := t.TempDir()
	page, err := List(root, "Condere", "", 0)
	require.NoError(t, err)
	require.Equal(t, []Summary{}, page.Cases)
	require.NoDirExists(t, filepath.Join(root, ".profitctl"))

	for _, project := range []string{"Jumpseat", "IntelIP", "Condere"} {
		c := example()
		c.Project = project
		c.Question = "Is a cache useful?"
		c.Evidence[0].Finding = "internal evidence detail"
		_, err := Save(root, strings.ToLower(project), 0, c)
		require.NoError(t, err)
		if project == "Condere" {
			c.Question = "Should research\n  run on every assessment?"
			_, err = Save(root, "condere", 1, c)
			require.NoError(t, err)
		}
	}
	page, err = List(root, " coNdeRe ", "", 0)
	require.NoError(t, err)
	require.Equal(t, []Summary{{"condere", 2, "Condere", "Should research run on every assessment?"}}, page.Cases)
	encoded, err := json.Marshal(page)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "internal evidence detail")
	page, err = List(root, "ANYTHING ELSE", "", 0)
	require.NoError(t, err)
	require.Empty(t, page.Cases)
	require.Empty(t, page.NextCursor)

	first, err := List(root, "CACHE", "", 1)
	require.NoError(t, err)
	require.Len(t, first.Cases, 1)
	require.Equal(t, "intelip", first.Cases[0].ID)
	require.Equal(t, "intelip", first.NextCursor)
	second, err := List(root, "CACHE", first.NextCursor, 1)
	require.NoError(t, err)
	require.Len(t, second.Cases, 1)
	require.Equal(t, "jumpseat", second.Cases[0].ID)
	require.Empty(t, second.NextCursor)
	_, err = List(root, "", "../", 2)
	require.ErrorIs(t, err, ErrInvalid)
	_, err = List(root, "", "", 51)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestFindResumesAfterBoundedScan(t *testing.T) {
	root := t.TempDir()
	_, err := Save(root, "z-case", 0, example())
	require.NoError(t, err)
	for i := range 200 {
		// Simulate case creation interrupted before publication. These still
		// consume a scan slot, so a sparse search can resume without timeout.
		dir := filepath.Join(root, ".profitctl", "decisions", fmt.Sprintf("a-%03d", i))
		require.NoError(t, os.Mkdir(dir, 0700))
	}
	first, err := List(root, "z-case", "", 0)
	require.NoError(t, err)
	require.Empty(t, first.Cases)
	require.Equal(t, "a-199", first.NextCursor)
	second, err := List(root, "z-case", first.NextCursor, 0)
	require.NoError(t, err)
	require.Len(t, second.Cases, 1)
	require.Equal(t, "z-case", second.Cases[0].ID)
	require.Empty(t, second.NextCursor)
}
