package entries

import (
	"io/fs"
	"path/filepath"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/response/match"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type EntryService interface {
	Search(opts SearchOptions) ([]*ach.EntryDetail, error)
}

type entryService struct {
	rootPath string
}

func NewEntryService(rootPath string) *entryService {
	return &entryService{
		rootPath: rootPath,
	}
}

type SearchOptions struct {
	AccountNumber string
	Amount        int
	RoutingNumber string
	TraceNumber   string
}

func (s *entryService) Search(opts SearchOptions) ([]*ach.EntryDetail, error) {
	var out []*ach.EntryDetail

	//nolint:gosimple
	var search filepath.WalkFunc
	search = func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return nil
		}
		entries, err := filterEntries(path, opts)
		if err != nil {
			return err
		}
		out = append(out, entries...)
		return nil
	}

	if err := filepath.Walk(s.rootPath, search); err != nil {
		return nil, err
	}

	return out, nil
}

func filterEntries(path string, opts SearchOptions) ([]*ach.EntryDetail, error) {
	file, err := ach.ReadFile(path)
	if file == nil || err != nil {
		return nil, nil
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
