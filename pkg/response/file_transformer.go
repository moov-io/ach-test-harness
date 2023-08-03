package response

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/response/match"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
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

func (ft *FileTransfomer) Transform(file *ach.File) error {
	out := ach.NewFile()
	out.SetValidation(ft.ValidateOpts)
	if err := createOutHeader(out, file, ft.ValidateOpts); err != nil {
		return err
	}

	var entrySaved = false
	for i := range file.Batches {
		mirror := newBatchMirror(ft.Writer, file.Batches[i])
		batch, err := ach.NewBatch(file.Batches[i].GetHeader())
		if err != nil {
			return fmt.Errorf("transform batch[%d] problem creating Batch: %v", i, err)
		}
		entries := file.Batches[i].GetEntries()
		for j := range entries {
			// Check if there's a matching Action and perform it. There may also be a future-dated action to execute.
			copyAction, processAction := ft.Matcher.FindAction(entries[j])
			if copyAction != nil {
				logger := ft.Matcher.Logger.With(copyAction)
				logger.Log("Processing matched action")

				mirror.saveEntry(copyAction.Copy, entries[j])
				entrySaved = true
			}
			if processAction != nil {
				logger := ft.Matcher.Logger.With(processAction)
				logger.Log("Processing matched action")

				entry, err := ft.Entry.MorphEntry(file.Header, entries[j], processAction)
				if err != nil {
					return fmt.Errorf("transform batch[%d] morph entry[%d] error: %v", i, j, err)
				}

				if processAction.Delay != nil {
					// need to save off the future-dated entry
					futOut := ach.NewFile()
					futOut.SetValidation(ft.ValidateOpts)
					if futErr := createOutHeader(futOut, file, ft.ValidateOpts); futErr != nil {
						return futErr
					}

					futMirror := newBatchMirror(ft.Writer, file.Batches[i])
					var futBatch ach.Batcher
					var futErr error

					// When the entry is corrected we need to change the SEC code
					if entry.Category == ach.CategoryNOC {
						bh := *file.Batches[i].GetHeader()
						bh.StandardEntryClassCode = ach.COR
						futBatch, futErr = ach.NewBatch(&bh)
					} else {
						futBatch, futErr = ach.NewBatch(file.Batches[i].GetHeader())
					}
					if futErr != nil {
						return fmt.Errorf("transform batch[%d] problem creating Batch: %v", i, futErr)
					}

					// Add the transformed entry onto the batch
					if entry != nil {
						futBatch.AddEntry(entry)
					}

					// Save off the entries as requested
					if futErr = futMirror.saveFiles(); futErr != nil {
						return fmt.Errorf("problem saving entries: %v", futErr)
					}
					// Create our Batch's Control and other fields
					if futErr = futBatch.Create(); futErr != nil {
						return fmt.Errorf("transform batch[%d] create error: %v", i, futErr)
					}
					futOut.AddBatch(futBatch)

					if futErr = writeOutFile(futOut, ft, processAction.Delay); futErr != nil {
						return futErr
					}
				} else {
					// When the entry is corrected we need to change the SEC code
					if entry.Category == ach.CategoryNOC {
						bh := batch.GetHeader()
						bh.StandardEntryClassCode = ach.COR
						if b, err := ach.NewBatch(bh); b != nil {
							batch = b // replace entire Batch
						} else {
							return fmt.Errorf("transform batch[%d] NOC entry[%d] error: %v", i, j, err)
						}
					}

					// Add the transformed entry onto the batch
					if entry != nil {
						batch.AddEntry(entry)
					}
					entrySaved = true
				}
			}
		}

		if entrySaved {
			// Save off the entries as requested
			if err := mirror.saveFiles(); err != nil {
				return fmt.Errorf("problem saving entries: %v", err)
			}
			// Create our Batch's Control and other fields
			if entries := batch.GetEntries(); len(entries) > 0 {
				if err := batch.Create(); err != nil {
					return fmt.Errorf("transform batch[%d] create error: %v", i, err)
				}
				out.AddBatch(batch)
			}
		}
	}

	if err := writeOutFile(out, ft, nil); err != nil {
		return err
	}
	return nil
}

func createOutHeader(out *ach.File, file *ach.File, opts *ach.ValidateOpts) error {
	out.Header = ach.NewFileHeader()
	out.Header.SetValidation(opts)

	out.Header.ImmediateDestination = file.Header.ImmediateOrigin
	out.Header.ImmediateDestinationName = file.Header.ImmediateOriginName
	out.Header.ImmediateOrigin = file.Header.ImmediateDestination
	out.Header.ImmediateOriginName = file.Header.ImmediateDestinationName
	out.Header.FileCreationDate = time.Now().Format("060102")
	out.Header.FileCreationTime = time.Now().Format("1504")
	out.Header.FileIDModifier = "A"

	if err := out.Header.Validate(); err != nil {
		return fmt.Errorf("file transform: header validate: %v", err)
	}

	return nil
}

func writeOutFile(out *ach.File, ft *FileTransfomer, delay *time.Duration) error {
	if out != nil && len(out.Batches) > 0 {
		if err := out.Create(); err != nil {
			return fmt.Errorf("transform out create: %v", err)
		}
		if err := out.Validate(); err == nil {
			generatedFilePath := filepath.Join(ft.returnPath, generateFilename(out)) // TODO(adam): need to determine return path
			if err := ft.Writer.WriteFile(generatedFilePath, out, delay); err != nil {
				return fmt.Errorf("transform write %s: %v", generatedFilePath, err)
			}
		} else {
			return fmt.Errorf("transform validate out file: %v", err)
		}
	}

	return nil
}

var (
	randomFilenameSource = rand.NewSource(time.Now().Unix())
)

func generateFilename(file *ach.File) string {
	if file == nil {
		return fmt.Sprintf("MISSING_%d.ach", randomFilenameSource.Int63())
	}
	for i := range file.Batches {
		bh := file.Batches[i].GetHeader()
		if bh.StandardEntryClassCode == ach.COR {
			return fmt.Sprintf("CORRECTION_%d.ach", randomFilenameSource.Int63())
		}
	}
	return fmt.Sprintf("RETURN_%d.ach", randomFilenameSource.Int63())
}
