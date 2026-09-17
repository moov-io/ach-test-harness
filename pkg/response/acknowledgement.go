// Copyright 2026 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package response

import (
	"context"
	"fmt"
	"strings"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/internal/achx"
	"github.com/moov-io/base/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// acknowledgeCCDEntry returns nil when the entry does not request an eligible ACK.
func acknowledgeCCDEntry(ctx context.Context, fh ach.FileHeader, bh *ach.BatchHeader, ed *ach.EntryDetail) (*ach.EntryDetail, error) {
	if bh.StandardEntryClassCode != ach.CCD || strings.TrimSpace(ed.DiscretionaryData) != "AK" {
		return nil, nil
	}

	var transactionCode int
	switch ed.TransactionCode {
	case ach.CheckingCredit:
		transactionCode = ach.CheckingZeroDollarRemittanceCredit
	case ach.SavingsCredit:
		transactionCode = ach.SavingsZeroDollarRemittanceCredit
	default:
		return nil, nil
	}

	_, span := telemetry.StartSpan(ctx, "entry-acknowledgement", trace.WithAttributes(
		attribute.String("entry.trace_number", ed.TraceNumber),
	))
	defer span.End()

	out := ach.NewEntryDetail()
	out.TransactionCode = transactionCode
	out.RDFIIdentification = achx.ABA8(bh.ODFIIdentification)
	out.CheckDigit = achx.ABACheckDigit(bh.ODFIIdentification)
	out.DFIAccountNumber = ed.DFIAccountNumber
	out.Amount = 0
	out.SetOriginalTraceNumber(ed.TraceNumber)
	out.IndividualName = ed.IndividualName
	out.Category = ach.CategoryForward

	responseTrace, err := achx.TraceNumber(fh.ImmediateDestination)
	if err != nil {
		return nil, fmt.Errorf("generating ACK trace number: %w", err)
	}
	out.TraceNumber = responseTrace

	if err := out.Validate(); err != nil {
		return nil, fmt.Errorf("ACK entry detail validate: %w", err)
	}
	return out, nil
}
