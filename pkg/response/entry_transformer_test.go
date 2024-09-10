package response

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"

	"github.com/stretchr/testify/require"
)

func TestMorphEntry__Correction(t *testing.T) {
	file, err := ach.ReadFile(filepath.Join("..", "..", "examples", "utility-bill.ach"))
	require.NoError(t, err)

	file.Header.ImmediateDestination = "123456780"

	xform := &CorrectionTransformer{}
	action := service.Action{
		Correction: &service.Correction{
			Code: "C01",
			Data: "45111616",
		},
	}
	bh := file.Batches[0].GetHeader()
	ed := file.Batches[0].GetEntries()[0]
	out, err := xform.MorphEntry(context.Background(), file.Header, bh, ed, &action)
	require.NoError(t, err)

	if out.Addenda98 == nil {
		t.Fatal("exected Addenda98 record")
	}
	require.NotEqual(t, ed.TraceNumber, out.TraceNumber)
	require.Equal(t, ach.CheckingReturnNOCDebit, out.TransactionCode)
	require.Equal(t, ed.TraceNumber, out.Addenda98.OriginalTrace)
	require.Equal(t, "C01", out.Addenda98.ChangeCode)
	require.Equal(t, "45111616", out.Addenda98.CorrectedData)
	require.Equal(t, "23138010", out.Addenda98.OriginalDFI)

	require.Equal(t, "12104288", out.RDFIIdentification)
	require.Equal(t, "2", out.CheckDigit)

	if out.Addenda99 != nil {
		t.Fatal("unexpected Addenda99")
	}
}

func TestMorphEntry__Return(t *testing.T) {
	file, err := ach.ReadFile(filepath.Join("..", "..", "examples", "ppd-debit.ach"))
	require.NoError(t, err)

	file.Header.ImmediateDestination = "123456780"

	xform := &ReturnTransformer{}
	action := service.Action{
		Return: &service.Return{
			Code: "R01",
		},
	}
	bh := file.Batches[0].GetHeader()
	ed := file.Batches[0].GetEntries()[0]
	out, err := xform.MorphEntry(context.Background(), file.Header, bh, ed, &action)
	require.NoError(t, err)

	if out.Addenda98 != nil {
		t.Fatal("unexpected Addenda98")
	}
	if out.Addenda99 == nil {
		t.Fatal("exected Addenda99 record")
	}
	require.NotEqual(t, ed.TraceNumber, out.TraceNumber)
	require.Equal(t, ach.CheckingReturnNOCDebit, out.TransactionCode)
	require.Equal(t, "12104288", out.RDFIIdentification)
	require.Equal(t, "2", out.CheckDigit)
	require.Equal(t, ed.TraceNumber, out.Addenda99.OriginalTrace)
	require.Equal(t, "R01", out.Addenda99.ReturnCode)
	require.Equal(t, "23138010", out.Addenda99.OriginalDFI)
}

func TestMorphEntry_Return_GL(t *testing.T) {
	file, err := ach.ReadFile(filepath.Join("..", "..", "examples", "gl-debit.ach"))
	require.NoError(t, err)

	xform := &ReturnTransformer{}
	action := service.Action{
		Return: &service.Return{
			Code: "R03",
		},
	}
	bh := file.Batches[0].GetHeader()
	ed := file.Batches[0].GetEntries()[0]
	out, err := xform.MorphEntry(context.Background(), file.Header, bh, ed, &action)
	require.NoError(t, err)

	if out.Addenda98 != nil {
		t.Fatal("unexpected Addenda98")
	}
	if out.Addenda99 == nil {
		t.Fatal("exected Addenda99 record")
	}
	require.NotEqual(t, ed.TraceNumber, out.TraceNumber)
	require.Equal(t, ach.GLReturnNOCDebit, out.TransactionCode)
	require.Equal(t, "12104288", out.RDFIIdentification)
	require.Equal(t, "2", out.CheckDigit)
	require.Equal(t, ed.TraceNumber, out.Addenda99.OriginalTrace)
	require.Equal(t, "R03", out.Addenda99.ReturnCode)
	require.Equal(t, "23138010", out.Addenda99.OriginalDFI)

	// Try the reversal
	err = file.Reversal(time.Now())
	require.NoError(t, err)

	bh = file.Batches[0].GetHeader()
	ed = file.Batches[0].GetEntries()[0]

	out, err = xform.MorphEntry(context.Background(), file.Header, bh, ed, &action)
	require.NoError(t, err)
	require.Equal(t, ach.GLReturnNOCCredit, out.TransactionCode)
}

func TestMorphEntry_Return_Loan(t *testing.T) {
	file, err := ach.ReadFile(filepath.Join("..", "..", "examples", "loan-credit.ach"))
	require.NoError(t, err)

	xform := &ReturnTransformer{}
	action := service.Action{
		Return: &service.Return{
			Code: "R03",
		},
	}
	bh := file.Batches[0].GetHeader()
	ed := file.Batches[0].GetEntries()[0]
	out, err := xform.MorphEntry(context.Background(), file.Header, bh, ed, &action)
	require.NoError(t, err)

	if out.Addenda98 != nil {
		t.Fatal("unexpected Addenda98")
	}
	if out.Addenda99 == nil {
		t.Fatal("exected Addenda99 record")
	}
	require.NotEqual(t, ed.TraceNumber, out.TraceNumber)
	require.Equal(t, ach.LoanReturnNOCCredit, out.TransactionCode)
	require.Equal(t, "12104288", out.RDFIIdentification)
	require.Equal(t, "2", out.CheckDigit)
	require.Equal(t, ed.TraceNumber, out.Addenda99.OriginalTrace)
	require.Equal(t, "R03", out.Addenda99.ReturnCode)
	require.Equal(t, "23138010", out.Addenda99.OriginalDFI)

	// Try the reversal
	err = file.Reversal(time.Now())
	require.NoError(t, err)

	bh = file.Batches[0].GetHeader()
	ed = file.Batches[0].GetEntries()[0]

	out, err = xform.MorphEntry(context.Background(), file.Header, bh, ed, &action)
	require.NoError(t, err)
	require.Equal(t, ach.LoanReturnNOCDebit, out.TransactionCode)
}

func TestMorphEntry__Prenote(t *testing.T) {
	file, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "prenote.ach"))
	require.NoError(t, err)

	transactionCodes := []int{
		ach.CheckingPrenoteCredit, ach.CheckingPrenoteDebit,
		ach.SavingsPrenoteCredit, ach.SavingsPrenoteDebit,
	}

	t.Run("correction", func(t *testing.T) {
		action := service.Action{
			Correction: &service.Correction{
				Code: "C01",
				Data: "45111616",
			},
		}

		for _, txnCode := range transactionCodes {
			msg := fmt.Sprintf("input TransactionCode=%d", txnCode)

			bh := file.Batches[0].GetHeader()
			ed := file.Batches[0].GetEntries()[0]
			ed.TransactionCode = ach.CheckingPrenoteCredit

			xform := &CorrectionTransformer{}
			out, err := xform.MorphEntry(context.Background(), file.Header, bh, ed, &action)
			require.NoError(t, err, msg)

			require.Equal(t, ach.CheckingReturnNOCCredit, out.TransactionCode, msg)
			require.NotEqual(t, ed.TraceNumber, out.TraceNumber, msg)
			require.Equal(t, ed.TraceNumber, out.Addenda98.OriginalTrace, msg)

			require.NotNil(t, out.Addenda98, msg)
			require.Equal(t, "C01", out.Addenda98.ChangeCode, msg)
			require.Equal(t, "45111616", out.Addenda98.CorrectedData, msg)
		}
	})

	t.Run("return", func(t *testing.T) {
		action := service.Action{
			Return: &service.Return{
				Code: "R01",
			},
		}

		for _, txnCode := range transactionCodes {
			msg := fmt.Sprintf("input TransactionCode=%d", txnCode)

			bh := file.Batches[0].GetHeader()
			ed := file.Batches[0].GetEntries()[0]
			ed.TransactionCode = ach.CheckingPrenoteCredit

			xform := &ReturnTransformer{}
			out, err := xform.MorphEntry(context.Background(), file.Header, bh, ed, &action)
			require.NoError(t, err, msg)

			require.Equal(t, ach.CheckingReturnNOCCredit, out.TransactionCode, msg)
			require.NotEqual(t, ed.TraceNumber, out.TraceNumber, msg)

			require.NotNil(t, out.Addenda99, msg)
			require.Equal(t, ed.TraceNumber, out.Addenda99.OriginalTrace, msg)
			require.Equal(t, "R01", out.Addenda99.ReturnCode, msg)
		}
	})

	t.Run("loan", func(t *testing.T) {
		// TODO(adam):
	})

	t.Run("general-ledger", func(t *testing.T) {
		// TODO(adam):
	})
}
