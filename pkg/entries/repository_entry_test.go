package entries

import (
	"testing"

	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	repo := NewFTPRepository(service.FTPConfig{
		RootPath: "./testdata",
		Paths: service.Paths{
			Files:  "/outbound/",
			Return: "/returned/",
		},
	})

	entries, err := repo.List()
	require.NoError(t, err)
	require.Len(t, entries, 3)
}
