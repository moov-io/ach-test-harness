package entries

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/response/match"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type EntryRepository interface {
	Search(ctx context.Context, opts SearchOptions) ([]*ach.EntryDetail, error)
}

type ftpRepository struct {
	rootPath string
}

func NewFTPRepository(cfg *service.FTPConfig) *ftpRepository {
	return &ftpRepository{
		rootPath: cfg.RootPath,
	}
}

func (r *ftpRepository) Search(ctx context.Context, opts SearchOptions) ([]*ach.EntryDetail, error) {
	_, span := telemetry.StartSpan(ctx, "repo-entry-search")
	defer span.End()

	out := make([]*ach.EntryDetail, 0)

	var filesProcessed int

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
		filesProcessed += 1

		entries, err := filterEntries(path, opts)
		if err != nil {
			return err
		}

		if len(entries) > 0 {
			span.AddEvent("found-entries", trace.WithAttributes(
				attribute.Int("search.entries", len(entries)),
				attribute.String("search.filename", path),
			))

			out = append(out, entries...)
		}
		return nil
	}
	span.SetAttributes(
		attribute.Int("search.files_processed", filesProcessed),
	)

	var walkingPath = r.rootPath
	if opts.Path != "" {
		walkingPath = filepath.Join(r.rootPath, opts.Path)
	}

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
