package entries

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/response/match"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
)

type EntryRepository interface {
	Search(opts SearchOptions) ([]*ach.EntryDetail, error)
}

type ftpRepository struct {
	logger   log.Logger
	rootPath string
}

func NewFTPRepository(logger log.Logger, cfg *service.FTPConfig) *ftpRepository {
	return &ftpRepository{
		logger:   logger,
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

		r.logger.Logf("reading %s", path)
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

	var walkingPath = r.rootPath
	if opts.SubDirectory != "" {
		walkingPath = filepath.Join(r.rootPath, opts.SubDirectory)
	}

	r.logger.Logf("Waling directory %s", walkingPath)
	if err := filepath.WalkDir(walkingPath, search); err != nil {
		return nil, fmt.Errorf("failed reading directory content %s: %v", walkingPath, err)
	}

	return out, nil
}

func filterEntries(path string, opts SearchOptions) ([]*ach.EntryDetail, error) {
	file, _ := ach.ReadFile(path)
	if file == nil {
		return nil, nil
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
