package response

import (
	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type Matcher struct {
	Responses []service.Response
}

func (m Matcher) FindAction(ed *ach.EntryDetail) *service.Action {
	for i := range m.Responses {
		if ed.TraceNumber == m.Responses[i].Match.TraceNumber {
			return &m.Responses[i].Action
		}
	}
	return nil
}
