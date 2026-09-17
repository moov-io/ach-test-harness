// Copyright 2026 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package response

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/internal/achx"
	"github.com/moov-io/ach-test-harness/pkg/service"

	"github.com/stretchr/testify/require"
)

func TestAcknowledgeCCDEntry(t *testing.T) {
	tests := []struct {
		name     string
		sec      string
		code     int
		data     string
		amount   int
		wantCode int
	}{
		{"checking credit", ach.CCD, ach.CheckingCredit, "AK", 1500, ach.CheckingZeroDollarRemittanceCredit},
		{"savings credit", ach.CCD, ach.SavingsCredit, "AK", 1500, ach.SavingsZeroDollarRemittanceCredit},
		{"whitespace", ach.CCD, ach.CheckingCredit, " \tAK\n ", 1500, ach.CheckingZeroDollarRemittanceCredit},
		{"checking debit", ach.CCD, ach.CheckingDebit, "AK", 1500, 0},
		{"savings debit", ach.CCD, ach.SavingsDebit, "AK", 1500, 0},
		{"checking credit prenote", ach.CCD, ach.CheckingPrenoteCredit, "AK", 0, 0},
		{"savings credit prenote", ach.CCD, ach.SavingsPrenoteCredit, "AK", 0, 0},
		{"checking debit prenote", ach.CCD, ach.CheckingPrenoteDebit, "AK", 0, 0},
		{"savings debit prenote", ach.CCD, ach.SavingsPrenoteDebit, "AK", 0, 0},
		{"checking zero-dollar remittance", ach.CCD, ach.CheckingZeroDollarRemittanceCredit, "AK", 0, 0},
		{"savings zero-dollar remittance", ach.CCD, ach.SavingsZeroDollarRemittanceCredit, "AK", 0, 0},
		{"checking credit with zero amount", ach.CCD, ach.CheckingCredit, "AK", 0, ach.CheckingZeroDollarRemittanceCredit},
		{"savings credit with zero amount", ach.CCD, ach.SavingsCredit, "AK", 0, ach.SavingsZeroDollarRemittanceCredit},
		{"GL credit", ach.CCD, ach.GLCredit, "AK", 1500, 0},
		{"loan credit", ach.CCD, ach.LoanCredit, "AK", 1500, 0},
		{"PPD", ach.PPD, ach.CheckingCredit, "AK", 1500, 0},
		{"WEB", ach.WEB, ach.CheckingCredit, "AK", 1500, 0},
		{"CTX", ach.CTX, ach.CheckingCredit, "AK", 1500, 0},
		{"ATX", ach.ATX, ach.CheckingCredit, "AK", 1500, 0},
		{"ACK", ach.ACK, ach.CheckingZeroDollarRemittanceCredit, "AK", 0, 0},
		{"no request", ach.CCD, ach.CheckingCredit, "", 1500, 0},
		{"different request", ach.CCD, ach.CheckingCredit, "AA", 1500, 0},
		{"lowercase", ach.CCD, ach.CheckingCredit, "ak", 1500, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := readACKSource(t)
			header := file.Batches[0].GetHeader()
			entry := file.Batches[0].GetEntries()[0]
			header.StandardEntryClassCode = tt.sec
			entry.TransactionCode = tt.code
			entry.DiscretionaryData = tt.data
			entry.Amount = tt.amount
			before := snapshotACKSource(t, file)

			out, err := acknowledgeCCDEntry(context.Background(), file.Header, header, entry)
			require.NoError(t, err)
			require.Equal(t, before, snapshotACKSource(t, file))
			if tt.wantCode == 0 {
				require.Nil(t, out)
				return
			}
			requireACKEntry(t, out, entry, file.Header, header, tt.wantCode)
		})
	}
}

func TestAcknowledgeCCDEntry_InvalidRouting(t *testing.T) {
	file := readACKSource(t)
	header := file.Batches[0].GetHeader()
	header.ODFIIdentification = "invalid"
	before := snapshotACKSource(t, file)

	out, err := acknowledgeCCDEntry(context.Background(), file.Header, header, file.Batches[0].GetEntries()[0])
	require.ErrorContains(t, err, "ACK entry detail validate")
	require.Nil(t, out)
	require.Equal(t, before, snapshotACKSource(t, file))
}

func TestFileTransformer_ACK(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		wantCode int
		copy     bool
	}{
		{"checking", ach.CheckingCredit, ach.CheckingZeroDollarRemittanceCredit, false},
		{"savings", ach.SavingsCredit, ach.SavingsZeroDollarRemittanceCredit, false},
		{"copy and ACK", ach.CheckingCredit, ach.CheckingZeroDollarRemittanceCredit, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := readACKSource(t)
			header := file.Batches[0].GetHeader()
			entry := file.Batches[0].GetEntries()[0]
			entry.TransactionCode = tt.code
			require.NoError(t, file.Batches[0].Create())
			require.NoError(t, file.Create())
			before := snapshotACKSource(t, file)

			var responses []service.Response
			if tt.copy {
				responses = append(responses, respCopyCredit)
			}
			transformer, dir := testFileTransformer(t, responses...)
			require.NoError(t, transformer.Transform(context.Background(), file))
			require.Equal(t, before, snapshotACKSource(t, file))

			files, err := os.ReadDir(filepath.Join(dir, "returned"))
			require.NoError(t, err)
			require.Len(t, files, 1)
			require.True(t, strings.HasPrefix(files[0].Name(), "ACK_"))
			out, err := ach.ReadFile(filepath.Join(dir, "returned", files[0].Name()))
			require.NoError(t, err)
			require.NoError(t, out.Validate())
			require.Equal(t, file.Header.ImmediateOrigin, out.Header.ImmediateDestination)
			require.Equal(t, file.Header.ImmediateDestination, out.Header.ImmediateOrigin)
			require.Len(t, out.Batches, 1)
			batch := out.Batches[0]
			require.IsType(t, &ach.BatchACK{}, batch)
			require.Equal(t, ach.ACK, batch.GetHeader().StandardEntryClassCode)
			require.Equal(t, ach.CreditsOnly, batch.GetHeader().ServiceClassCode)
			require.Equal(t, achx.ABA8(file.Header.ImmediateDestination), batch.GetHeader().ODFIIdentification)
			require.Len(t, batch.GetEntries(), 1)
			requireACKEntry(t, batch.GetEntries()[0], entry, file.Header, header, tt.wantCode)
			require.Zero(t, batch.GetControl().TotalDebitEntryDollarAmount)
			require.Zero(t, batch.GetControl().TotalCreditEntryDollarAmount)
			require.Equal(t, 1, batch.GetControl().EntryAddendaCount)
			info, err := files[0].Info()
			require.NoError(t, err)
			require.Less(t, info.ModTime(), time.Now())

			copies, err := os.ReadDir(filepath.Join(dir, "reconciliation"))
			if !tt.copy {
				require.ErrorIs(t, err, os.ErrNotExist)
				return
			}
			require.NoError(t, err)
			require.Len(t, copies, 1)
			copied, err := ach.ReadFile(filepath.Join(dir, "reconciliation", copies[0].Name()))
			// Reconciliation output intentionally contains batch records only.
			require.ErrorContains(t, err, "none or more than one file headers exists")
			require.Len(t, copied.Batches, 1)
			require.Equal(t, header.String(), copied.Batches[0].GetHeader().String())
			require.Len(t, copied.Batches[0].GetEntries(), 1)
			copiedEntry := copied.Batches[0].GetEntries()[0]
			require.Equal(t, entry.String(), copiedEntry.String())
			require.Len(t, copiedEntry.Addenda05, 1)
			require.Equal(t, entry.Addenda05[0].String(), copiedEntry.Addenda05[0].String())
		})
	}
}

func TestFileTransformer_ACKProcessPrecedence(t *testing.T) {
	tests := []struct {
		name   string
		action service.Action
		sec    string
		prefix string
	}{
		{"return", actionReturn, ach.CCD, "RETURN_"},
		{"correction", actionCorrection, ach.COR, "CORRECTION_"},
		{"delayed return", actionDelayReturn, ach.CCD, "RETURN_"},
		{"delayed correction", actionDelayCorrection, ach.COR, "CORRECTION_"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := readACKSource(t)
			before := snapshotACKSource(t, file)
			transformer, dir := testFileTransformer(t, service.Response{
				Match:  service.Match{EntryType: service.EntryTypeCredit},
				Action: tt.action,
			})
			require.NoError(t, transformer.Transform(context.Background(), file))
			require.Equal(t, before, snapshotACKSource(t, file))
			files, err := os.ReadDir(filepath.Join(dir, "returned"))
			require.NoError(t, err)
			require.Len(t, files, 1)
			require.True(t, strings.HasPrefix(files[0].Name(), tt.prefix))
			out, err := ach.ReadFile(filepath.Join(dir, "returned", files[0].Name()))
			require.NoError(t, err)
			require.NoError(t, out.Validate())
			require.Len(t, out.Batches, 1)
			require.Equal(t, tt.sec, out.Batches[0].GetHeader().StandardEntryClassCode)
			require.Len(t, out.Batches[0].GetEntries(), 1)
			entry := out.Batches[0].GetEntries()[0]
			if tt.action.Return != nil {
				require.NotNil(t, entry.Addenda99)
				require.Equal(t, tt.action.Return.Code, entry.Addenda99.ReturnCode)
			} else {
				require.NotNil(t, entry.Addenda98)
				require.Equal(t, tt.action.Correction.Code, entry.Addenda98.ChangeCode)
			}
			info, err := files[0].Info()
			require.NoError(t, err)
			if tt.action.Delay != nil {
				require.Greater(t, info.ModTime(), time.Now())
			} else {
				require.Less(t, info.ModTime(), time.Now())
			}
		})
	}
}

func TestFileTransformer_ACKMixedResponses(t *testing.T) {
	tests := []struct {
		name       string
		correction bool
		delay      *time.Duration
		prefixes   []string
	}{
		{"ACK and return", false, nil, []string{"RETURN"}},
		{"ACK return and COR", true, nil, []string{"CORRECTION"}},
		{"ACK and delayed return", false, &delay, []string{"ACK", "RETURN"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := readACKSource(t)
			batch := file.Batches[0]
			source := batch.GetEntries()[0]
			codes := []int{ach.SavingsCredit, ach.CheckingCredit}
			if tt.correction {
				codes = append(codes, ach.SavingsCredit)
			}
			for i, code := range codes {
				entry := *source
				entry.TransactionCode = code
				entry.Addenda05 = nil
				entry.AddendaRecordIndicator = 0
				entry.SetTraceNumber(batch.GetHeader().ODFIIdentification, 110542+i)
				batch.AddEntry(&entry)
			}
			require.NoError(t, batch.Create())
			require.NoError(t, file.Create())
			entries := batch.GetEntries()
			responses := []service.Response{{
				Match:  service.Match{TraceNumber: entries[2].TraceNumber},
				Action: service.Action{Return: actionReturn.Return, Delay: tt.delay},
			}}
			if tt.correction {
				responses = append(responses, service.Response{
					Match:  service.Match{TraceNumber: entries[3].TraceNumber},
					Action: actionCorrection,
				})
			}
			before := snapshotACKSource(t, file)
			transformer, dir := testFileTransformer(t, responses...)
			require.NoError(t, transformer.Transform(context.Background(), file))
			require.Equal(t, before, snapshotACKSource(t, file))
			files, err := os.ReadDir(filepath.Join(dir, "returned"))
			require.NoError(t, err)
			require.Len(t, files, len(tt.prefixes))
			var prefixes []string
			counts := make(map[string]int)
			for _, fd := range files {
				prefixes = append(prefixes, strings.SplitN(fd.Name(), "_", 2)[0])
				out, err := ach.ReadFile(filepath.Join(dir, "returned", fd.Name()))
				require.NoError(t, err)
				require.NoError(t, out.Validate())
				for _, outBatch := range out.Batches {
					sec := outBatch.GetHeader().StandardEntryClassCode
					counts[sec]++
					if sec != ach.ACK {
						require.Len(t, outBatch.GetEntries(), 1)
						continue
					}
					require.Equal(t, ach.CreditsOnly, outBatch.GetHeader().ServiceClassCode)
					require.Len(t, outBatch.GetEntries(), 2)
					var originalTraces, traces []string
					for _, ack := range outBatch.GetEntries() {
						originalTraces = append(originalTraces, ack.OriginalTraceNumberField())
						traces = append(traces, ack.TraceNumber)
						require.Zero(t, ack.Amount)
						require.Empty(t, ack.Addenda05)
						require.Nil(t, ack.Addenda98)
						require.Nil(t, ack.Addenda99)
					}
					require.ElementsMatch(t, []string{entries[0].TraceNumber, entries[1].TraceNumber}, originalTraces)
					require.True(t, sort.StringsAreSorted(traces))
					require.NotEqual(t, traces[0], traces[1])
				}
				info, err := fd.Info()
				require.NoError(t, err)
				if tt.delay != nil && strings.HasPrefix(fd.Name(), "RETURN_") {
					require.Greater(t, info.ModTime(), time.Now())
				} else {
					require.Less(t, info.ModTime(), time.Now())
				}
			}
			require.ElementsMatch(t, tt.prefixes, prefixes)
			expectedCounts := map[string]int{ach.ACK: 1, ach.CCD: 1}
			if tt.correction {
				expectedCounts[ach.COR] = 1
			}
			require.Equal(t, expectedCounts, counts)
		})
	}
}

func TestGenerateFilename_ACK(t *testing.T) {
	tests := []struct {
		name   string
		secs   []string
		prefix string
	}{
		{"empty", nil, "RETURN_"},
		{"ACK", []string{ach.ACK}, "ACK_"},
		{"multiple ACK batches", []string{ach.ACK, ach.ACK}, "ACK_"},
		{"return", []string{ach.CCD}, "RETURN_"},
		{"ACK then return", []string{ach.ACK, ach.CCD}, "RETURN_"},
		{"return then ACK", []string{ach.CCD, ach.ACK}, "RETURN_"},
		{"correction", []string{ach.COR}, "CORRECTION_"},
		{"ACK then correction", []string{ach.ACK, ach.COR}, "CORRECTION_"},
		{"correction then ACK", []string{ach.COR, ach.ACK}, "CORRECTION_"},
		{"return and correction", []string{ach.CCD, ach.COR}, "CORRECTION_"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := ach.NewFile()
			for _, sec := range tt.secs {
				header := ach.NewBatchHeader()
				header.StandardEntryClassCode = sec
				batch, err := ach.NewBatch(header)
				require.NoError(t, err)
				file.AddBatch(batch)
			}
			require.True(t, strings.HasPrefix(generateFilename(file), tt.prefix))
		})
	}
	require.True(t, strings.HasPrefix(generateFilename(nil), "MISSING_"))
}

func readACKSource(t *testing.T) *ach.File {
	t.Helper()
	file, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "with-addenda.ach"))
	require.NoError(t, err)
	entry := file.Batches[0].GetEntries()[0]
	entry.TransactionCode = ach.CheckingCredit
	entry.DiscretionaryData = "AK"
	require.NoError(t, file.Batches[0].Create())
	require.NoError(t, file.Create())
	require.NoError(t, file.Validate())
	return file
}

func snapshotACKSource(t *testing.T, file *ach.File) []byte {
	t.Helper()
	data, err := json.Marshal(file)
	require.NoError(t, err)
	return data
}

func requireACKEntry(t *testing.T, out, original *ach.EntryDetail, fh ach.FileHeader, bh *ach.BatchHeader, code int) {
	t.Helper()
	require.NotNil(t, out)
	require.NotSame(t, original, out)
	require.NoError(t, out.Validate())
	require.Equal(t, code, out.TransactionCode)
	require.Zero(t, out.Amount)
	require.Equal(t, ach.CategoryForward, out.Category)
	require.Equal(t, original.TraceNumber, out.OriginalTraceNumberField())
	require.NotEqual(t, original.TraceNumber, out.TraceNumber)
	require.Len(t, out.TraceNumber, 15)
	hasPrefix(t, out.TraceNumber, achx.ABA8(fh.ImmediateDestination))
	require.Equal(t, achx.ABA8(bh.ODFIIdentification), out.RDFIIdentification)
	require.Equal(t, achx.ABACheckDigit(bh.ODFIIdentification), out.CheckDigit)
	require.Equal(t, original.DFIAccountNumber, out.DFIAccountNumber)
	require.Equal(t, original.IndividualName, out.IndividualName)
	require.Zero(t, out.AddendaRecordIndicator)
	require.Nil(t, out.Addenda02)
	require.Empty(t, out.Addenda05)
	require.Nil(t, out.Addenda98)
	require.Nil(t, out.Addenda98Refused)
	require.Nil(t, out.Addenda99)
	require.Nil(t, out.Addenda99Dishonored)
	require.Nil(t, out.Addenda99Contested)
}
