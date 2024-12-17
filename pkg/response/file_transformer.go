package response

import (
	"cmp"
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"slices"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/response/match"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
	"github.com/moov-io/base/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type FileTransfomer struct {
	Matcher      match.Matcher
	Entry        EntryTransformers
	Writer       FileWriter
	ValidateOpts *ach.ValidateOpts

	returnPath string
}

func NewFileTransformer(logger log.Logger, cfg *service.Config, responses []service.Response, writer FileWriter) *FileTransfomer {
	xform := &FileTransfomer{
		Matcher: match.New(logger, cfg.Matching, responses),
		Entry: EntryTransformers([]EntryTransformer{
			&CorrectionTransformer{},
			&ReturnTransformer{},
		}),
		Writer:       writer,
		ValidateOpts: cfg.ValidateOpts,
	}
	if cfg.Servers.FTP != nil {
		xform.returnPath = cfg.Servers.FTP.Paths.Return
	}
	return xform
}

func (ft *FileTransfomer) Transform(ctx context.Context, file *ach.File) error {
	ctx, span := telemetry.StartSpan(ctx, "file-transform", trace.WithAttributes(
		attribute.Int("file.batches", len(file.Batches)),
	))
	defer span.End()

	// Track ach.File objects to write based on different delay durations, including a default of "0s"
	var outFiles = outFiles{}

	// batchMirror is used for copying entires to the reconciliation file (if needed)
	mirror := newBatchMirror(ft.Writer)

	for i := range file.Batches {

		// Track ach.Batcher to write based on different delay durations and whether the batch is for NOC
		var outBatches = outBatches{}

		bh := file.Batches[i].GetHeader()
		entries := file.Batches[i].GetEntries()
		for j := range entries {
			// Check if there's a matching Action and perform it. There may also be a future-dated action to execute.
			copyAction, processAction := ft.Matcher.FindAction(bh, entries[j])
			if copyAction != nil {
				// Save this Entry
				mirror.saveEntry(&file.Batches[i], copyAction.Copy, entries[j])
			}
			if processAction != nil {
				entry, err := ft.Entry.MorphEntry(ctx, file.Header, bh, entries[j], processAction)
				if err != nil {
					return fmt.Errorf("transform batch[%d] morph entry[%d] error: %v", i, j, err)
				}

				// Get the appropriate ach.Batch object to update
				batch, err := outBatches.getOutBatch(processAction.Delay, entry.Category, file.Header, *file.Batches[i].GetHeader(), i)
				if err != nil {
					return err
				}

				// Add the transformed entry onto the batch
				if entry != nil {
					(*batch).AddEntry(entry)
				}

				if processAction.Delay != nil {
					// Get the ach.File object corresponding to this delay to write to.
					// We don't use this ach.File yet, but it needs to be initialized for later.
					if _, err = outFiles.getOutFile(processAction.Delay, file, ft.ValidateOpts); err != nil {
						return err
					}
				}
			}
		}

		// Create our Batch's Control and other fields
		for delay, batchesByCategory := range outBatches {
			for _, batch := range batchesByCategory {
				if entries = (*batch).GetEntries(); len(entries) > 0 {
					// Sort the entries before the final build
					slices.SortFunc(entries, func(e1, e2 *ach.EntryDetail) int {
						return cmp.Compare(e1.TraceNumber, e2.TraceNumber)
					})

					if err := (*batch).Create(); err != nil {
						return fmt.Errorf("transform batch[%d] create error: %v", i, err)
					}
					out, err := outFiles.getOutFile(delay, file, ft.ValidateOpts)
					if err != nil {
						return err
					}
					out.AddBatch(*batch)
				}
			}
		}
	}

	// Save off the entries as requested
	if err := mirror.saveFiles(ctx); err != nil {
		return fmt.Errorf("problem saving entries: %v", err)
	}

	var outFileCount int
	for delay, out := range outFiles {
		outFileCount += 1

		if out != nil && len(out.Batches) > 0 {
			if err := out.Create(); err != nil {
				return fmt.Errorf("transform out create: %v", err)
			}
			if err := out.Validate(); err == nil {
				generatedFilePath := filepath.Join(ft.returnPath, generateFilename(out)) // TODO(adam): need to determine return path

				span.SetAttributes(
					attribute.String(fmt.Sprintf("outfile.%d.path", outFileCount), generatedFilePath),
				)

				if err := ft.Writer.WriteFile(generatedFilePath, out, delay); err != nil {
					return fmt.Errorf("transform write %s: %v", generatedFilePath, err)
				}
			} else {
				return fmt.Errorf("transform validate out file: %v", err)
			}
		}
	}
	return nil
}

type outFiles map[*time.Duration]*ach.File

func (outFiles outFiles) getOutFile(delay *time.Duration, file *ach.File, opts *ach.ValidateOpts) (*ach.File, error) {
	var outFile = outFiles[delay]
	if outFile == nil {
		outFile = ach.NewFile()
		outFile.SetValidation(opts)
		outFile.Header = ach.NewFileHeader()
		outFile.Header.SetValidation(opts)

		outFile.Header.ImmediateDestination = file.Header.ImmediateOrigin
		outFile.Header.ImmediateDestinationName = file.Header.ImmediateOriginName
		outFile.Header.ImmediateOrigin = file.Header.ImmediateDestination
		outFile.Header.ImmediateOriginName = file.Header.ImmediateDestinationName
		outFile.Header.FileCreationDate = time.Now().Format("060102")
		outFile.Header.FileCreationTime = time.Now().Format("1504")
		outFile.Header.FileIDModifier = "A"

		if err := outFile.Header.Validate(); err != nil {
			return nil, fmt.Errorf("file transform: header validate: %v", err)
		}
		outFiles[delay] = outFile
	}

	return outFile, nil
}

type outBatches map[*time.Duration]map[bool]*ach.Batcher

func (outBatches outBatches) getOutBatch(delay *time.Duration, category string, fh ach.FileHeader, bh ach.BatchHeader, i int) (*ach.Batcher, error) {
	var batchesByCategory = outBatches[delay]
	if batchesByCategory == nil {
		batchesByCategory = make(map[bool]*ach.Batcher)
		outBatches[delay] = batchesByCategory
	}

	var outBatch = batchesByCategory[category == ach.CategoryNOC]
	if outBatch == nil {
		// When the entry is corrected we need to change the SEC code
		if category == ach.CategoryNOC {
			bh.StandardEntryClassCode = ach.COR
		}

		// We need to flip the Origin / Destination values when setting up the out batch
		bh.ODFIIdentification = fh.ImmediateDestination

		batch, err := ach.NewBatch(&bh)
		if err != nil {
			return nil, fmt.Errorf("transform batch[%d] problem creating Batch: %v", i, err)
		}
		outBatch = &batch
		batchesByCategory[category == ach.CategoryNOC] = outBatch
	}

	return outBatch, nil
}

var (
	randomFilenameSource = rand.NewSource(time.Now().Unix())
)

func generateFilename(file *ach.File) string {
	timestamp := time.Now().Format("20060102-150405.00000")
	if file == nil {
		return fmt.Sprintf("MISSING_%s_%d.ach", timestamp, randomFilenameSource.Int63())
	}
	for i := range file.Batches {
		bh := file.Batches[i].GetHeader()
		if bh.StandardEntryClassCode == ach.COR {
			return fmt.Sprintf("CORRECTION_%s_%d.ach", timestamp, randomFilenameSource.Int63())
		}
	}
	return fmt.Sprintf("RETURN_%s_%d.ach", timestamp, randomFilenameSource.Int63())
}
