package response

import (
	"bytes"
	"fmt"
	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"path/filepath"
	"sort"
	"time"
)

// batchMirror is an object that will save batches
type batchMirror struct {
	batches map[batchMirrorKey]map[string]*batchMirrorBatch // path+companyID -> (batch ID -> header+entries+control)

	writer FileWriter
}

type batchMirrorKey struct {
	path      string
	companyID string
}

func (key *batchMirrorKey) getFilePathName() string {
	timestamp := time.Now().Format("20060102-150405.00000")
	filename := fmt.Sprintf("%s_%s.ach", key.companyID, timestamp)
	return filepath.Join(key.path, filename)
}

type batchMirrorBatch struct {
	header  *ach.BatchHeader
	entries []*ach.EntryDetail
	control *ach.BatchControl
}

func (batch *batchMirrorBatch) write(buf *bytes.Buffer) error {
	buf.WriteString(batch.header.String() + "\n")
	for _, entry := range batch.entries {
		buf.WriteString(entry.String() + "\n")
	}
	control, err := calculateControl(batch.header, batch.entries)
	if err != nil {
		return fmt.Errorf("problem computing control: %v", err)
	}
	buf.WriteString(control + "\n")

	return nil
}

func newBatchMirror(w FileWriter) *batchMirror {
	return &batchMirror{
		batches: make(map[batchMirrorKey]map[string]*batchMirrorBatch),
		writer:  w,
	}
}

func (bm *batchMirror) saveEntry(b *ach.Batcher, copy *service.Copy, ed *ach.EntryDetail) {
	if b == nil || copy == nil || ed == nil {
		return
	}

	batcher := *b
	// Get the batchMirrorKey
	key := batchMirrorKey{
		path:      copy.Path,
		companyID: batcher.GetHeader().CompanyIdentification,
	}
	// Create a new batchMirrorBatch map if this key does not exist
	if _, exists := bm.batches[key]; !exists {
		bm.batches[key] = make(map[string]*batchMirrorBatch)
	}
	// Create an array of batchMirrorBatch if this batch ID does not exist
	if _, exists := bm.batches[key][batcher.GetHeader().BatchNumberField()]; !exists {
		bm.batches[key][batcher.GetHeader().BatchNumberField()] = &batchMirrorBatch{
			header:  batcher.GetHeader(),
			entries: make([]*ach.EntryDetail, 0),
			control: batcher.GetControl(),
		}
	}
	// Append this EntryDetail to the batchMirrorBatch's EntryDetails slice for the derived key
	bm.batches[key][batcher.GetHeader().BatchNumberField()].entries = append(bm.batches[key][batcher.GetHeader().BatchNumberField()].entries, ed)
}

func (bm *batchMirror) saveFiles() error {
	if len(bm.batches) == 0 {
		return nil
	}

	// Write files by Path/CompanyID
	for key, mirror := range bm.batches {
		var buf bytes.Buffer

		// sort the keys so that the batches appear in the correct order
		keys := make([]string, 0, len(mirror))
		for number, _ := range mirror {
			keys = append(keys, number)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, val := range keys {
			batch := mirror[val]
			if err := batch.write(&buf); err != nil {
				return err
			}
		}

		// Write the file out
		err := bm.writer.Write(key.getFilePathName(), &buf, nil)
		if err != nil {
			return fmt.Errorf("problem writing file: %v", err)
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
