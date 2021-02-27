package response

import (
	"fmt"
	"log"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type FileTransfomer struct {
	Matcher Matcher
	Entry   EntryTransformers
	Writer  FileWriter
}

func NewFileTransformer(cfg *service.Config, responses []service.Response, writer FileWriter) *FileTransfomer {
	return &FileTransfomer{
		Matcher: Matcher{
			Responses: cfg.Responses,
		},
		Entry: EntryTransformers([]EntryTransformer{
			&CorrectionTransformer{},
			&ReturnTransformer{},
		}),
		Writer: writer,
	}
}

func (ft *FileTransfomer) Transform(file *ach.File) error {
	out := ach.NewFile()
	out.Header = file.Header

	fmt.Printf("\n\nfile=%#v\n\n", file)

	for i := range file.Batches {
		batch, err := ach.NewBatch(file.Batches[i].GetHeader())
		if err != nil {
			log.Printf("ERROR0: %v\n", err)
		}
		entries := file.Batches[i].GetEntries()
		for j := range entries {
			// Check if there's a matching Action and perform it
			action := ft.Matcher.FindAction(entries[j])
			if action != nil {
				entry, err := ft.Entry.MorphEntry(entries[j], *action)
				if err != nil {
					return err
				}
				batch.AddEntry(entry)
			}
		}
		if err := batch.Create(); err != nil {
			log.Printf("ERROR1: %v\n", err)
		}
		out.AddBatch(batch)
	}
	if out != nil {
		if err := out.Create(); err != nil {
			log.Printf("ERROR2: %v\n", err)
		}
		if err := out.Validate(); err == nil {
			if err := ft.Writer.Write("RETURN_12345.ach", out); err != nil {
				// TODO(adam): do somethign
			}
		} else {
			// TODO(adam): do something
			fmt.Printf("\n\n\nERROR:\n%v\n\n\n", err)
		}
	}

	return nil
}
