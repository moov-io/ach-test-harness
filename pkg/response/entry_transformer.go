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
		out, err = et[i].MorphEntry(out, action)
		if err != nil {
			return out, fmt.Errorf("%T: %v", et, err)
		}
	}
	return out, nil
}

type CorrectionTransformer struct{}

func (t *CorrectionTransformer) MorphEntry(ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error) {
	if action.Correction == nil {
		return ed, nil
	}
	// fmt.Printf("  COR: %#v\n", ed)
	return ed, nil
}

type ReturnTransformer struct{}

func (t *ReturnTransformer) MorphEntry(ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error) {
	if action.Return == nil {
		return ed, nil
	}

	addenda99 := ach.NewAddenda99()
	addenda99.ReturnCode = action.Return.Code
	addenda99.OriginalTrace = ed.TraceNumber
	addenda99.OriginalDFI = "141142141" // TODO(adam):
	addenda99.TraceNumber = "34141"     // TODO(adam):

	ed.Addenda99 = addenda99
	ed.TraceNumber = "3414135154"
	ed.AddendaRecordIndicator = 1

	// fmt.Printf("  RET: %#v\n", ed)
	return ed, nil
}
