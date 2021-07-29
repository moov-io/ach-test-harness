package entries

import (
	"github.com/moov-io/ach"
)

type EntryService interface {
	List() ([]*ach.EntryDetail, error)
}

type entryService struct {
	repository EntryRepository
}

func NewEntryService(repository EntryRepository) *entryService {
	return &entryService{
		repository: repository,
	}
}

func (s *entryService) List() ([]*ach.EntryDetail, error) {
	return s.repository.List()
}
