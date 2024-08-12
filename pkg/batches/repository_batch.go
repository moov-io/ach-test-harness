package batches

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

type BatchRepository interface {
	Search(ctx context.Context, opts SearchOptions) ([]ach.Batcher, error)
}

type batchRepository struct {
	rootPath string
}

func NewFTPRepository(cfg *service.FTPConfig) *batchRepository {
	return &batchRepository{
		rootPath: cfg.RootPath,
	}
}

func (r *batchRepository) Search(ctx context.Context, opts SearchOptions) ([]ach.Batcher, error) {
	_, span := telemetry.StartSpan(ctx, "repo-batch-search")
	defer span.End()

	out := make([]ach.Batcher, 0)

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

		batches, err := filterBatches(path, opts)
		if err != nil {
			return err
		}

		if len(batches) > 0 {
			span.AddEvent("found-batches", trace.WithAttributes(
				attribute.Int("search.batches", len(batches)),
				attribute.String("search.filename", path),
			))

			out = append(out, batches...)
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
				match.RoutingNumber(mm, entries[j]) {
				// accumulate batch
				out = append(out, file.Batches[i])
				continue
			}
		}
	}
	return out, nil
}
