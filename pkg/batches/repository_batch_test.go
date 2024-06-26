package batches

import (
	"testing"

	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	logger := log.NewDefaultLogger()

	repo := NewFTPRepository(logger, &service.FTPConfig{
		RootPath: "./testdata",
		Paths: service.Paths{
			Files:  "/outbound/",
			Return: "/returned/",
		},
	})

	// return all
	batches, err := repo.Search(SearchOptions{})
	require.NoError(t, err)
	require.Len(t, batches, 2)

	// search by account number
	batches, err = repo.Search(SearchOptions{
		AccountNumber: "987654321",
	})

	require.NoError(t, err)
	require.Len(t, batches, 1)

	// search by timestamp in our files:
	// returned/2.ach was created on 1908161055 and has 1 entry
	// outbound/1.ach was created on 1908161059 and has 2 batches
	batches, err = repo.Search(SearchOptions{
		CreatedAfter: "2019-08-16T10:56:00+00:00",
	})

	// expect to get batches from outbound/1.ach
	require.NoError(t, err)
	require.Len(t, batches, 1)

	// search by subdirectory in our files:
	// outbound/1.ach was created on 1908161059 and has 2 batches
	batches, err = repo.Search(SearchOptions{
		Path: "outbound",
	})

	// expect to get batches from outbound/1.ach
	require.NoError(t, err)
	require.Len(t, batches, 1)
}

func TestRepository__filterBatches(t *testing.T) {
	var opts SearchOptions

	batches, err := filterBatches("/tmp/noexist/foobar", opts)
	require.NoError(t, err)
	require.Len(t, batches, 0)
}
