package response

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

// batchMirror is an object that will save batches
type batchMirror struct {
	header  *ach.BatchHeader
	entries map[string][]*ach.EntryDetail // filepath -> entries
	control *ach.BatchControl

	writer FileWriter
}

func newBatchMirror(w FileWriter, b ach.Batcher) *batchMirror {
	return &batchMirror{
		header:  b.GetHeader(),
		entries: make(map[string][]*ach.EntryDetail),
		control: b.GetControl(),
		writer:  w,
	}
}

func (bm *batchMirror) saveEntry(copy *service.Copy, ed *ach.EntryDetail) {
	if copy == nil {
		return
	}
	bm.entries[copy.Path] = append(bm.entries[copy.Path], ed)
}

func (bm *batchMirror) saveFiles() error {
	if bm.header == nil || len(bm.entries) == 0 {
		return fmt.Errorf("missing BatchHeader or entries (found %d)", len(bm.entries))
	}
	for path, entries := range bm.entries {
		// Accumulate file contents
		var buf bytes.Buffer
		buf.WriteString(bm.header.String() + "\n")
		for i := range entries {
			buf.WriteString(entries[i].String() + "\n")
		}
		control, err := calculateControl(bm.header, entries)
		if err != nil {
			return fmt.Errorf("problem computing control: %v", err)
		}
		buf.WriteString(control)

		// Write the file out
		if filename, err := bm.filename(); err != nil {
			return fmt.Errorf("unable to get filename: %v", err)
		} else {
			bm.writer.Write(filepath.Join(path, filename), &buf)
		}
	}
	return nil
}

func calculateControl(bh *ach.BatchHeader, entries []*ach.EntryDetail) (string, error) {
	batch, _ := ach.NewBatch(bh)
	for i := range entries {
		batch.AddEntry(entries[i])
	}
	if err := batch.Create(); err != nil {
		return "", fmt.Errorf("error creating batch: %v", err)
	}
	return batch.GetControl().String(), nil
}

func (bm *batchMirror) filename() (string, error) {
	if bm.header == nil {
		return "", errors.New("missing BatchHeader")
	}
	timestamp := time.Now().Format("20060102-150405.00000")
	return fmt.Sprintf("%s_%s.ach", bm.header.CompanyIdentification, timestamp), nil
}
