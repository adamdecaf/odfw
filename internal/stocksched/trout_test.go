package stocksched

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTroutLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("-short enabled")
	}

	counts, err := LoadTroutCounts(ODFWHttpClient, Willamette)
	require.NoError(t, err)

	if testing.Verbose() {
		t.Logf("%s", counts.Stockings[0])
	}
}

func TestTroutParse(t *testing.T) {
	fd, err := os.Open(filepath.Join("testdata", "willamette.html"))
	require.NoError(t, err)
	t.Cleanup(func() { fd.Close() })

	counts, err := parseTroutCounts(fd)
	require.NoError(t, err)
	require.Len(t, counts.Stockings, 50)

	if testing.Verbose() {
		for i := range counts.Stockings {
			t.Logf("%s", counts.Stockings[i])
		}
	}
}
