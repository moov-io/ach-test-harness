package filedrive

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"goftp.io/server/core"
	ftp "goftp.io/server/core"
	"goftp.io/server/driver/file"

	"github.com/stretchr/testify/require"
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

func listFiles(t *testing.T, driver core.Driver, path string) []core.FileInfo {
	t.Helper()

	var found []core.FileInfo
	err := driver.ListDir(path, func(info core.FileInfo) error {
		found = append(found, info)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func setupDriver(t *testing.T, basePath string) core.Driver {
	t.Helper()

	fileDriverFactory := &file.DriverFactory{
		RootPath: basePath,
		Perm:     ftp.NewSimplePerm("user", "group"),
	}
	filteringDriver := &Factory{
		DriverFactory: fileDriverFactory,
	}

	driver, err := filteringDriver.NewDriver()
	if err != nil {
		t.Fatal(err)
	}

	return driver
}
