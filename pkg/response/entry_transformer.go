package response

import (
	"fmt"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type EntryTransformer interface {
	MorphEntry(ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error)
}

type EntryTransformers []EntryTransformer

func (et EntryTransformers) MorphEntry(ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error) {
	out := &(*ed) // make a copy
	var err error
	for i := range et {
		out, err = et.MorphEntry(out, action)
		if err != nil {
			return out, fmt.Errorf("%T: %v", et, err)
		}
	}
	return out, nil
}

type CorrectionTransformer struct{}

func (t *CorrectionTransformer) MorphEntry(ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error) {
	return ed, nil
}

type ReturnTransformer struct{}

func (t *ReturnTransformer) MorphEntry(ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error) {
	return ed, nil
}
