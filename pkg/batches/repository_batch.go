package batches

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

type BatchRepository interface {
	Search(opts SearchOptions) ([]ach.Batcher, error)
}

type batchRepository struct {
	logger   log.Logger
	rootPath string
}

func NewFTPRepository(logger log.Logger, cfg *service.FTPConfig) *batchRepository {
	return &batchRepository{
		logger:   logger,
		rootPath: cfg.RootPath,
	}
}

func (r *batchRepository) Search(opts SearchOptions) ([]ach.Batcher, error) {
	out := make([]ach.Batcher, 0)

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

		batches, err := filterBatches(path, opts)
		if err != nil {
			return err
		}
		out = append(out, batches...)
		return nil
	}

	var walkingPath = r.rootPath
	if opts.Path != "" {
		walkingPath = filepath.Join(r.rootPath, opts.Path)
	}

	r.logger.Logf("Waling directory %s", walkingPath)
	if err := filepath.WalkDir(walkingPath, search); err != nil {
		return nil, fmt.Errorf("failed reading directory content %s: %v", walkingPath, err)
	}

	return out, nil
}

func filterBatches(path string, opts SearchOptions) ([]ach.Batcher, error) {
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

	var out []ach.Batcher
	for i := range file.Batches {
		entries := file.Batches[i].GetEntries()
		if mm.Empty() {
			out = append(out, file.Batches[i])
			continue
		}
		for j := range entries {
			if match.TraceNumber(mm, entries[j]) || match.AccountNumber(mm, entries[j]) ||
				match.RoutingNumber(mm, entries[j]) || match.Amount(mm, entries[j]) {
				// accumulate batch
				out = append(out, file.Batches[i])
				continue
			}
		}
	}
	return out, nil
}
