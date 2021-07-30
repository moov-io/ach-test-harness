package entries

import (
	"github.com/moov-io/ach"
)

type EntryService interface {
	Search(ops SearchOptions) ([]*ach.EntryDetail, error)
}

type entryService struct {
	repository EntryRepository
}

func NewEntryService(repository EntryRepository) *entryService {
	return &entryService{
		repository: repository,
	}
}

type SearchOptions struct {
	AccountNumber string
	Amount        int
	RoutingNumber string
	TraceNumber   string
}

func (s *entryService) Search(opts SearchOptions) ([]*ach.EntryDetail, error) {
	return s.repository.Search(opts)
}
