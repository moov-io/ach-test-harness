package batches

import (
	"context"
	"fmt"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/base"
)

type BatchService interface {
	Search(ctx context.Context, ops SearchOptions) ([]ach.Batcher, error)
}

type batchService struct {
	repository BatchRepository
}

func NewBatchService(repository BatchRepository) *batchService {
	return &batchService{
		repository: repository,
	}
}

type SearchOptions struct {
	AccountNumber string
	Amount        int
	RoutingNumber string
	TraceNumber   string
	CreatedAfter  string
	Path          string
}

func (s *batchService) Search(ctx context.Context, opts SearchOptions) ([]ach.Batcher, error) {
	return s.repository.Search(ctx, opts)
}

func (opts SearchOptions) fileTooOld(file *ach.File) (bool, error) {
	if opts.CreatedAfter == "" {
		return false, nil
	}

	tt, err := parseTimestamp(opts.CreatedAfter)
	if err != nil {
		return false, err
	}

	fileCreated, err := time.Parse("0601021504", file.Header.FileCreationDate+file.Header.FileCreationTime)
	if err != nil {
		return false, err
	}

	return fileCreated.Before(tt), nil
}

func parseTimestamp(when string) (time.Time, error) {
	formats := []string{base.ISO8601Format, "2006-01-02", time.RFC3339}
	for i := range formats {
		tt, err := time.Parse(formats[i], when)
		if !tt.IsZero() && err == nil {
			return tt, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse '%s'", when)
}
