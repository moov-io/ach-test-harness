package response

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"goftp.io/server/driver/file"
)

func TestFileWriter__mkdir(t *testing.T) {
	dir := t.TempDir()
	factory := &file.DriverFactory{
		RootPath: dir,
	}
	driver, err := factory.NewDriver()
	require.NoError(t, err)

	path := filepath.Join("foo", "Bar", "baz", "example.ach")

	err = mkdir(FTPFileDriver{Driver: driver}, path)
	require.NoError(t, err)

	// Check that the directory exists
	fd, err := os.Stat(filepath.Join(dir, filepath.Dir(path)))
	require.NoError(t, err)
	require.True(t, fd.IsDir())

	// Check the file does not exist
	fd, err = os.Stat(filepath.Join(dir, path))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
	require.Nil(t, fd)
}
