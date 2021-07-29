package entries

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntryService(t *testing.T) {
	service := NewEntryService("testdata")

	t.Run("AddFile adds entries from the file", func(t *testing.T) {
		var opts SearchOptions
		entries, err := service.Search(opts)
		require.NoError(t, err)
		require.Len(t, entries, 2)

		require.Equal(t, 500000, entries[0].Amount)
		require.Equal(t, 125, entries[1].Amount)
	})
}
