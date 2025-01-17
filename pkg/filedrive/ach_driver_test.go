package filedrive

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
	"goftp.io/server/core"
)

// MockDriver is a simple mock implementation of core.Driver for testing purposes.
type MockDriver struct {
	core.Driver
}

func (m *MockDriver) PutFile(path string, r io.Reader, appendData bool) (int64, error) {
	// Mock implementation, just return success
	return 0, nil
}

func TestACHDriver_PutFile_InvalidACH(t *testing.T) {
	mockDriver := &MockDriver{}
	customDriver := NewACHDriver(log.NewDefaultLogger(), nil, mockDriver)

	// Create an invalid ACH file (e.g., missing required fields)
	var invalidACH bytes.Buffer
	invalidACH.WriteString("invalid ACH content")

	// Attempt to upload the invalid ACH file
	_, err := customDriver.PutFile("invalid.ach", &invalidACH, false)

	// Verify that an error is returned
	require.Error(t, err)
}

func TestACHDriver_PutFile(t *testing.T) {
	mockDriver := &MockDriver{}
	customDriver := NewACHDriver(log.NewDefaultLogger(), nil, mockDriver)

	achFile, err := os.Open(filepath.Join("..", "..", "testdata", "20230809-144155-102000021C.ach"))
	require.NoError(t, err)
	defer achFile.Close()

	// Attempt to upload the valid ACH file
	_, err = customDriver.PutFile("valid.ach", achFile, false)

	// Verify that no error is returned
	require.NoError(t, err)
}
