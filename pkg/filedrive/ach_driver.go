package filedrive

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/moov-io/ach"
	"github.com/moov-io/base/log"
	"github.com/moov-io/base/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"goftp.io/server/core"
)

// ACHDriver wraps the goftp driver to add additional logic and error checking.
type ACHDriver struct {
	core.Driver

	logger       log.Logger
	validateOpts *ach.ValidateOpts
}

func NewACHDriver(logger log.Logger, validateOpts *ach.ValidateOpts, driver core.Driver) *ACHDriver {
	return &ACHDriver{
		Driver:       driver,
		logger:       logger,
		validateOpts: validateOpts,
	}
}

// PutFile overrides the existing method to prevent erroneous ACH files from being uploaded.
func (d *ACHDriver) PutFile(path string, r io.Reader, appendData bool) (int64, error) {
	_, span := telemetry.StartSpan(context.Background(), "put-file", trace.WithAttributes(
		attribute.String("ftp.destination", path),
	))
	defer span.End()

	d.logger.Info().Log(fmt.Sprintf("receiving file for %s", path))

	// Read the file that was uploaded
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)

	reader := ach.NewReader(tee)
	reader.SetValidation(d.validateOpts)

	file, err := reader.Read()
	if err != nil {
		span.RecordError(err)
		d.logger.Error().Log(fmt.Sprintf("ftp: error reading ACH file %s: %v", path, err))
		return 0, err
	}

	if err := file.Create(); err != nil {
		d.logger.Error().Log(fmt.Sprintf("ftp: error creating file %s: %v", path, err))
		return 0, err
	}

	span.SetAttributes(attribute.Int("ftp.file_size_bytes", buf.Len()))
	d.logger.Info().Log(fmt.Sprintf("accepting file at %s", path))

	// Call the original PutFile method with a reset reader.
	return d.Driver.PutFile(path, &buf, appendData)
}
