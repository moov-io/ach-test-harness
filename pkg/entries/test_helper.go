package entries

import (
	"os"
	"path/filepath"

	"github.com/moov-io/ach"
)

func mockACHFile() (*ach.File, error) {
	return ach.ReadFile(filepath.Join("testdata", "ccd-debit.ach"))
}
