package entries

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type EntryRepository interface {
	List() ([]*ach.EntryDetail, error)
}

type ftpRepository struct {
	dataPath   string
	filesPath  string
	returnPath string
}

func NewFTPRepository(cfg service.FTPConfig) *ftpRepository {
	return &ftpRepository{
		dataPath:   cfg.RootPath,
		filesPath:  cfg.Paths.Files,
		returnPath: cfg.Paths.Return,
	}
}

func (r *ftpRepository) List() ([]*ach.EntryDetail, error) {
	var files []string

	for _, path := range []string{r.filesPath, r.returnPath} {
		err := filepath.Walk(r.dataPath+path, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			files = append(files, path)
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("reading directory %s: %v", r.dataPath+path, err)
		}
	}

	entries := make([]*ach.EntryDetail, 0)

	for _, filePath := range files {
		achFile, err := ach.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading ACH file %s: %v", filePath, err)
		}

		for _, batch := range achFile.Batches {
			entries = append(entries, batch.GetEntries()...)
		}
	}

	return entries, nil
}
