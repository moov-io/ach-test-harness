package response

import (
	"fmt"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/internal/achx"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type EntryTransformer interface {
	MorphEntry(fh ach.FileHeader, ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error)
}

type EntryTransformers []EntryTransformer

func (et EntryTransformers) MorphEntry(fh ach.FileHeader, ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error) {
	var err error
	for i := range et {
		ed, err = et[i].MorphEntry(fh, ed, action)
		if err != nil {
			return ed, fmt.Errorf("%T: %v", et, err)
		}
	}
	return ed, nil
}

type CorrectionTransformer struct{}

func (t *CorrectionTransformer) MorphEntry(fh ach.FileHeader, ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error) {
	if action.Correction == nil {
		return ed, nil
	}

	// TODO(adam): COR transform
	// fmt.Printf("  COR: %#v\n", ed)
	return ed, nil
}

type ReturnTransformer struct{}

func (t *ReturnTransformer) MorphEntry(fh ach.FileHeader, ed *ach.EntryDetail, action service.Action) (*ach.EntryDetail, error) {
	if action.Return == nil {
		return ed, nil
	}

	out := ach.NewEntryDetail()

	// Set the TransactionCode from the EntryDetail
	switch ed.TransactionCode {
	case ach.CheckingCredit, ach.CheckingDebit, ach.SavingsCredit, ach.SavingsDebit:
		out.TransactionCode = ed.TransactionCode - 1
	}

	// Set the fields from the original EntryDetail
	out.RDFIIdentification = achx.ABA8(fh.ImmediateDestination)
	out.CheckDigit = achx.ABACheckDigit(fh.ImmediateDestination)
	out.DFIAccountNumber = ed.DFIAccountNumber
	out.Amount = ed.Amount
	out.IdentificationNumber = ed.IdentificationNumber
	out.IndividualName = ed.IndividualName
	out.DiscretionaryData = ed.DiscretionaryData
	out.AddendaRecordIndicator = 1
	out.TraceNumber = achx.TraceNumber(fh.ImmediateDestination)
	out.Category = ach.CategoryReturn

	// Create the Return addenda
	addenda99 := ach.NewAddenda99()
	addenda99.ReturnCode = action.Return.Code
	addenda99.OriginalTrace = ed.TraceNumber
	addenda99.OriginalDFI = fh.ImmediateDestination
	addenda99.TraceNumber = out.TraceNumber

	if err := addenda99.Validate(); err != nil {
		return out, fmt.Errorf("Addenda99 validate: %#v\n", addenda99)
	}

	// Add the Addenda99 on the return EntryDetail
	out.Addenda99 = addenda99

	if err := out.Validate(); err != nil {
		return out, fmt.Errorf("OUT EntryDetail validate: %v\n", out)
	}

	return out, nil
}
