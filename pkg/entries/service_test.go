package entries

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moov-io/ach"
	"github.com/stretchr/testify/require"
)

func TestEntryService(t *testing.T) {
	// load ACH file
	f, err := os.Open(filepath.Join("testdata", "ccd-debit.ach"))
	require.NoError(t, err)

	r := ach.NewReader(f)
	achFile, err := r.Read()
	require.NoError(t, err)

	service := NewEntryService()

	t.Run("AddFile adds entries from the file", func(t *testing.T) {
		err := service.AddFile(&achFile)

		require.NoError(t, err)

		entries, err := service.List()

		require.NoError(t, err)
		require.Len(t, entries, 2)

		require.Equal(t, 500000, entries[0].Amount)
		require.Equal(t, 125, entries[1].Amount)
	})

	t.Run("Clean removes entries from the service", func(t *testing.T) {
		entries, err := service.List()

		require.NoError(t, err)
		require.Len(t, entries, 2)

		service.Clean()

		entries, err = service.List()

		require.NoError(t, err)
		require.Len(t, entries, 0)
	})
}
