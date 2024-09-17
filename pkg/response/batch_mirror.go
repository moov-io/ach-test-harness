package response

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

func (key *batchMirrorKey) getFilepath(data []byte) string {
	hash := fmt.Sprintf("%X", sha256.Sum256(data))
	if len(hash) > 8 {
		hash = hash[:8]
	}
	timestamp := time.Now().Format("20060102-150405.00000")
	filename := fmt.Sprintf("%s_%s_%s.ach", key.companyID, timestamp, hash)
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

func (bm *batchMirror) saveFiles(ctx context.Context) error {
	if len(bm.batches) == 0 {
		return nil
	}

	_, span := telemetry.StartSpan(ctx, "batch-mirror-save-files", trace.WithAttributes(
		attribute.Int("mirror.batches", len(bm.batches)),
	))
	defer span.End()

	// Write files by Path/CompanyID
	var batchCount int
	for key, mirror := range bm.batches {
		batchCount += 1

		var buf bytes.Buffer

		// sort the keys so that the batches appear in the correct order
		keys := make([]string, 0, len(mirror))
		for number := range mirror {
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
		path := key.getFilepath(buf.Bytes())
		span.SetAttributes(
			attribute.String(fmt.Sprintf("saved-files.%d.path", batchCount), path),
		)

		err := bm.writer.Write(path, &buf, nil)
		if err != nil {
			return fmt.Errorf("writing %s failed: %v", path, err)
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
