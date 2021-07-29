package entries

import (
	"sync"

	"github.com/moov-io/ach"
)

type EntryService interface {
	List() ([]*ach.EntryDetail, error)
	AddFile(file *ach.File) error
	Clean()
}

type entryService struct {
	entries []*ach.EntryDetail
	mutex   sync.RWMutex
}

func NewEntryService() *entryService {
	return &entryService{
		entries: make([]*ach.EntryDetail, 0),
	}
}

func (s *entryService) List() ([]*ach.EntryDetail, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.entries, nil
}

func (s *entryService) AddFile(file *ach.File) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, batch := range file.Batches {
		s.entries = append(s.entries, batch.GetEntries()...)
	}

	return nil
}

func (s *entryService) Clean() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.entries = make([]*ach.EntryDetail, 0)
}
