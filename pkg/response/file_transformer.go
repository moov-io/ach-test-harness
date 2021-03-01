package response

import (
	"fmt"
	"path/filepath"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type FileTransfomer struct {
	Matcher Matcher
	Entry   EntryTransformers
	Writer  FileWriter

	returnPath string
}

func NewFileTransformer(cfg *service.Config, responses []service.Response, writer FileWriter) *FileTransfomer {
	xform := &FileTransfomer{
		Matcher: Matcher{
			Responses: cfg.Responses,
		},
		Entry: EntryTransformers([]EntryTransformer{
			&CorrectionTransformer{},
			&ReturnTransformer{},
		}),
		Writer: writer,
	}
	if cfg.Servers.FTP != nil {
		xform.returnPath = cfg.Servers.FTP.Paths.Return
	}
	return xform
}

func (ft *FileTransfomer) Transform(file *ach.File) error {
	out := ach.NewFile()
	out.Header = file.Header

	for i := range file.Batches {
		batch, err := ach.NewBatch(file.Batches[i].GetHeader())
		if err != nil {
			return fmt.Errorf("transform batch[%d] problem creating Batch: %v", i, err)
		}
		entries := file.Batches[i].GetEntries()
		for j := range entries {
			// Check if there's a matching Action and perform it
			action := ft.Matcher.FindAction(entries[j])
			if action != nil {
				entry, err := ft.Entry.MorphEntry(entries[j], *action)
				if err != nil {
					return fmt.Errorf("transform batch[%d] morph entry[%d] error: %v", i, j, err)
				}
				batch.AddEntry(entry)
			}
		}
		if err := batch.Create(); err != nil {
			return fmt.Errorf("transform batch[%d] create error: %v", i, err)
		}
		out.AddBatch(batch)
	}
	if out != nil {
		if err := out.Create(); err != nil {
			return fmt.Errorf("transform out create: %v", err)
		}
		if err := out.Validate(); err == nil {
			filepath := filepath.Join(ft.returnPath, "RETURN_12345.ach")
			if err := ft.Writer.Write(filepath, out); err != nil {
				return fmt.Errorf("transform write %s: %v", filepath, err)
			}
		} else {
			return fmt.Errorf("transform validate out file: %v", err)
		}
	}

	return nil
}
