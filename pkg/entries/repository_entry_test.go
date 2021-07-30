package entries

import (
	"testing"

	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	repo := NewFTPRepository(&service.FTPConfig{
		RootPath: "./testdata",
		Paths: service.Paths{
			Files:  "/outbound/",
			Return: "/returned/",
		},
	})

	// return all
	entries, err := repo.Search(SearchOptions{})
	require.NoError(t, err)
	require.Len(t, entries, 3)

	// search by account number
	entries, err = repo.Search(SearchOptions{
		AccountNumber: "987654321",
	})

	require.NoError(t, err)
	require.Len(t, entries, 1)
}
