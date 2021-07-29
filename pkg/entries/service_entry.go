package entries

import "github.com/moov-io/ach"

type EntryService interface {
	List() ([]*ach.EntryDetail, error)
	Add(entry *ach.EntryDetail)
}

type entryService struct {
}

func NewEntryService() *entryService {
	return &entryService{}
}

func (s *entryService) List() ([]*ach.EntryDetail, error) {
	return nil, nil
}

func (s *entryService) Add(entry *ach.EntryDetail) {
}
