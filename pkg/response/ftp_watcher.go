package response

import (
	"fmt"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/entries"
	"github.com/moov-io/base/log"

	ftp "goftp.io/server/core"
)

func Register(
	logger log.Logger,
	validateOpts *ach.ValidateOpts,
	ftpServer *ftp.Server,
	transformer *FileTransfomer,
	entryService entries.EntryService,
) {
	if ftpServer != nil && transformer != nil {
		ftpServer.RegisterNotifer(&FTPWatcher{
			logger:       logger,
			validateOpts: validateOpts,
			transformer:  transformer,
			entryService: entryService,
		})
	} else {
		logger.Info().Log("unable to register transformer")
	}
}

type FTPWatcher struct {
	ftp.NullNotifier

	logger       log.Logger
	validateOpts *ach.ValidateOpts
	transformer  *FileTransfomer
	entryService entries.EntryService
}

func (notify *FTPWatcher) AfterFilePut(conn *ftp.Conn, dstPath string, size int64, err error) {
	notify.logger.Info().Log(fmt.Sprintf("accepting file at %s", dstPath))

	if err != nil {
		notify.logger.Error().Log(fmt.Sprintf("error with file %s: %v", dstPath, err))
	}

	// Grab a file descriptor
	driver, err := conn.ServerOpts().Factory.NewDriver()
	if err != nil {
		notify.logger.Info().Log(fmt.Sprintf("ftp: error getting driver for file %s: %v", dstPath, err))
	}
	_, fd, err := driver.GetFile(dstPath, 0)
	if err != nil {
		notify.logger.Info().Log(fmt.Sprintf("ftp: error reading file %s: %v", dstPath, err))
	}
	// Read the file that was uploaded
	reader := ach.NewReader(fd)
	reader.SetValidation(notify.validateOpts)

	file, err := reader.Read()
	if err != nil {
		notify.logger.Info().Log(fmt.Sprintf("ftp: error reading ACH file %s: %v", dstPath, err))
	}
	if err := file.Create(); err != nil {
		notify.logger.Info().Log(fmt.Sprintf("ftp: error creating file %s: %v", dstPath, err))
	}

	if err := notify.entryService.AddFile(&file); err != nil {
		notify.logger.Info().Log(fmt.Sprintf("entry service: error adding file %s: %v", dstPath, err))
	}

	if err := notify.transformer.Transform(&file); err != nil {
		notify.logger.Info().Log(fmt.Sprintf("ftp: error transforming file %s: %v", dstPath, err))
	}

}
