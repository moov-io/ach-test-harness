package filedrive

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
	"goftp.io/server/v2"
)

func TestMTimeFilter_ListDir(t *testing.T) {
	dir := t.TempDir()
	driver := setupDriver(t, dir)

	found := listFiles(t, driver, ".")
	require.Len(t, found, 0)

	err := os.WriteFile(filepath.Join(dir, "first.txt"), []byte("first file"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "second.txt"), []byte("second file"), 0600)
	require.NoError(t, err)

	found = listFiles(t, driver, ".")
	require.Len(t, found, 2)

	// Move one file into the future
	future := time.Now().Add(1 * time.Hour)
	err = os.Chtimes(filepath.Join(dir, "first.txt"), future, future)
	require.NoError(t, err)

	found = listFiles(t, driver, ".")
	require.Len(t, found, 1)
	require.Equal(t, "second.txt", found[0].Name())
}

func listFiles(t *testing.T, driver server.Driver, path string) []os.FileInfo {
	t.Helper()

	var found []os.FileInfo
	err := driver.ListDir(nil, path, func(info os.FileInfo) error {
		found = append(found, info)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func setupDriver(t *testing.T, basePath string) server.Driver {
	t.Helper()

	driver, err := NewDriver(log.NewDefaultLogger(), nil, basePath)
	if err != nil {
		t.Fatal(err)
	}
	return driver
}
