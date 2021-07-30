package entries

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/response/match"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type EntryRepository interface {
	Search(opts SearchOptions) ([]*ach.EntryDetail, error)
}

type ftpRepository struct {
	rootPath string
}

func NewFTPRepository(cfg *service.FTPConfig) *ftpRepository {
	return &ftpRepository{
		rootPath: cfg.RootPath,
	}
}

func (r *ftpRepository) Search(opts SearchOptions) ([]*ach.EntryDetail, error) {
	out := make([]*ach.EntryDetail, 0)

	//nolint:gosimple
	var search fs.WalkDirFunc
	search = func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if err != nil {
			return nil
		}

		// read only *.ach files
		if strings.ToLower(filepath.Ext(path)) != ".ach" {
			return nil
		}

		entries, err := filterEntries(path, opts)
		if err != nil {
			return err
		}
		out = append(out, entries...)
		return nil
	}

	if err := filepath.WalkDir(r.rootPath, search); err != nil {
		return nil, fmt.Errorf("failed reading directory content %s: %v", r.rootPath, err)
	}

	return out, nil
}

func filterEntries(path string, opts SearchOptions) ([]*ach.EntryDetail, error) {
	file, err := ach.ReadFile(path)
	if file == nil || err != nil {
		return nil, fmt.Errorf("reading ACH file %s: %v", path, err)
	}

	tooOld, err := opts.fileTooOld(file)
	if tooOld || err != nil {
		return nil, err
	}

	mm := service.Match{
		AccountNumber: opts.AccountNumber,
		Amount: &service.Amount{
			Value: opts.Amount,
		},
		RoutingNumber: opts.RoutingNumber,
		TraceNumber:   opts.TraceNumber,
	}

	var out []*ach.EntryDetail
	for i := range file.Batches {
		entries := file.Batches[i].GetEntries()
		if mm.Empty() {
			out = append(out, entries...)
			continue
		}
		for j := range entries {
			if match.TraceNumber(mm, entries[j]) || match.AccountNumber(mm, entries[j]) ||
				match.RoutingNumber(mm, entries[j]) || match.Amount(mm, entries[j]) {
				// accumulate entry
				out = append(out, entries[j])
				continue
			}
		}
	}
	return out, nil
}
