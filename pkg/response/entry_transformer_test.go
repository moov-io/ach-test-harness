package response

import (
	"path/filepath"
	"testing"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"

	"github.com/stretchr/testify/require"
)

func TestMorphEntry__Correction(t *testing.T) {
	file, err := ach.ReadFile(filepath.Join("..", "..", "examples", "utility-bill.ach"))
	require.NoError(t, err)

	xform := &CorrectionTransformer{}
	action := service.Action{
		Correction: &service.Correction{
			Code: "C01",
			Data: "45111616",
		},
	}
	ed := file.Batches[0].GetEntries()[0]
	out, err := xform.MorphEntry(file.Header, ed, &action)
	require.NoError(t, err)

	if out.Addenda98 == nil {
		t.Fatal("exected Addenda98 record")
	}
	require.NotEqual(t, ed.TraceNumber, out.TraceNumber)
	require.Equal(t, ed.TraceNumber, out.Addenda98.OriginalTrace)
	require.Equal(t, out.Addenda98.ChangeCode, "C01")
	require.Equal(t, out.Addenda98.CorrectedData, "45111616")

	if out.Addenda99 != nil {
		t.Fatal("unexpected Addenda99")
	}
}

func TestMorphEntry__Return(t *testing.T) {
	file, err := ach.ReadFile(filepath.Join("..", "..", "examples", "ppd-debit.ach"))
	require.NoError(t, err)

	xform := &ReturnTransformer{}
	action := service.Action{
		Return: &service.Return{
			Code: "R01",
		},
	}
	ed := file.Batches[0].GetEntries()[0]
	out, err := xform.MorphEntry(file.Header, ed, &action)
	require.NoError(t, err)

	if out.Addenda98 != nil {
		t.Fatal("unexpected Addenda98")
	}
	if out.Addenda99 == nil {
		t.Fatal("exected Addenda99 record")
	}
	require.NotEqual(t, ed.TraceNumber, out.TraceNumber)
	require.Equal(t, ed.TraceNumber, out.Addenda99.OriginalTrace)
	require.Equal(t, out.Addenda99.ReturnCode, "R01")
}
