package response

import (
	"context"
	"fmt"

	"github.com/moov-io/ach"
	"github.com/moov-io/base/log"
	"github.com/moov-io/base/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	ftp "goftp.io/server/v2"
)

func Register(
	logger log.Logger,
	validateOpts *ach.ValidateOpts,
	ftpServer *ftp.Server,
	transformer *FileTransfomer,
) {
	if ftpServer != nil && transformer != nil {
		ftpServer.RegisterNotifer(&FTPWatcher{
			logger:       logger,
			validateOpts: validateOpts,
			transformer:  transformer,
		})
	} else {
		logger.Error().Log("unable to register transformer")
	}
}

type FTPWatcher struct {
	ftp.NullNotifier

	logger       log.Logger
	validateOpts *ach.ValidateOpts
	transformer  *FileTransfomer
}

func (notify *FTPWatcher) AfterFilePut(ctx *ftp.Context, dstPath string, size int64, err error) {
	otelCtx, span := telemetry.StartSpan(context.Background(), "after-file-put", trace.WithAttributes(
		attribute.String("ftp.destination", dstPath),
		attribute.Int64("ftp.file_size_bytes", size),
	))
	defer span.End()

	notify.logger.Info().Log(fmt.Sprintf("accepting file at %s", dstPath))

	if err != nil {
		notify.logger.Error().Log(fmt.Sprintf("error with file %s: %v", dstPath, err))
	}

	// Grab a file descriptor from the server driver
	driver := ctx.Sess.Server().Driver
	if driver == nil {
		notify.logger.Error().Log(fmt.Sprintf("ftp: nil driver for file %s", dstPath))
		return
	}
	_, fd, err := driver.GetFile(ctx, dstPath, 0)
	if err != nil {
		notify.logger.Error().Log(fmt.Sprintf("ftp: error reading file %s: %v", dstPath, err))
		return
	}
	// Read the file that was uploaded
	reader := ach.NewReader(fd)
	reader.SetValidation(notify.validateOpts)

	// TODO(adam): ACH file Iterator

	file, err := reader.Read()
	if err != nil {
		span.RecordError(err)
		notify.logger.Error().Log(fmt.Sprintf("ftp: error reading ACH file %s: %v", dstPath, err))
	}
	if err := file.Create(); err != nil {
		notify.logger.Error().Log(fmt.Sprintf("ftp: error creating file %s: %v", dstPath, err))
	}

	if err := notify.transformer.Transform(otelCtx, &file); err != nil {
		notify.logger.Error().Log(fmt.Sprintf("ftp: error transforming file %s: %v", dstPath, err))
	}
}
