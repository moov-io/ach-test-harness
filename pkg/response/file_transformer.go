package response

import (
	"fmt"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
)

type FileTransfomer struct {
	Matcher Matcher
	Entry   EntryTransformers
	Writer  FileWriter
}

func NewFileTransformer(cfg service.Config, responses []service.Response, writer FileWriter) *FileTransfomer {
	return &FileTransfomer{
		Matcher: &Matcher{
			Responses: cfg.Responses,
		},
		Entry: []EntryTransformers{
			CorrectionTransformer{},
			ReturnTransformer{},
		},
		Writer: writer,
	}
}

func (ft *FileTransfomer) Transform(file *ach.File) (*ach.File, error) {
	out := ach.NewFile()
	out.Header = file.Header

	for i := range file.Batches {
		entries := file.Batches[i].GetEntries()
		for j := range entries {
			action := ft.Matcher.FindAction(entries[j])
			entry, err := ft.Entry.MorphEntry(entries[j], action)
			if err != nil {
				return out, err
			}
			// TODO(adam): append back to file
			fmt.Printf("entry[%d]=%#v\n", i, entry)
		}
	}
	if out != nil {
		if err := out.Validate(); err == nil {
			if err := ft.Writer.Write("RETURN_12345.ach", out); err != nil {
				// TODO(adam): do somethign
			}
		} else {
			// TODO(adam): do something
		}
	}

	return file, nil
}
