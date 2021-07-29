package entries

import (
	"os"
	"path/filepath"

	"github.com/moov-io/ach"
)

func mockACHFile() (*ach.File, error) {
	f, err := os.Open(filepath.Join("testdata", "ccd-debit.ach"))
	if err != nil {
		return nil, err
	}

	r := ach.NewReader(f)
	achFile, err := r.Read()
	if err != nil {
		return nil, err
	}

	return &achFile, nil
}
