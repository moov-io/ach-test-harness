package response

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"

	"github.com/stretchr/testify/require"
)

// TestExampleConfig_Responses uploads the example files to a harness that uses
// examples/config.yml, and checks that each response rule writes its file.
func TestExampleConfig_Responses(t *testing.T) {
	t.Setenv("APP_CONFIG", filepath.Join("..", "..", "examples", "config.yml"))

	cfg, err := service.LoadConfig(log.NewTestLogger())
	require.NoError(t, err)

	fileTransformer, dir := testFileTransformer(t, cfg.Responses...)

	transform := func(t *testing.T, name string) {
		t.Helper()

		file, err := ach.ReadFile(filepath.Join("..", "..", "examples", name))
		require.NoError(t, err)
		require.NoError(t, fileTransformer.Transform(context.Background(), file))
	}

	t.Run("ppd-debit.ach is copied", func(t *testing.T) {
		transform(t, "ppd-debit.ach")

		fds, err := os.ReadDir(filepath.Join(dir, "reconciliation"))
		require.NoError(t, err)
		require.Len(t, fds, 1)
	})

	t.Run("utility-bill.ach gets a correction", func(t *testing.T) {
		transform(t, "utility-bill.ach")

		fds, err := os.ReadDir(filepath.Join(dir, "returned"))
		require.NoError(t, err)
		require.Len(t, fds, 1)

		file, err := ach.ReadFile(filepath.Join(dir, "returned", fds[0].Name()))
		require.NoError(t, err)
		require.Len(t, file.Batches, 1)

		entries := file.Batches[0].GetEntries()
		require.Len(t, entries, 1)
		require.NotNil(t, entries[0].Addenda98)
		require.Equal(t, "C01", entries[0].Addenda98.ChangeCode)
		require.Equal(t, "744567899", entries[0].Addenda98.CorrectedData)
	})
}
