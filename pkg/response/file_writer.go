package response

import (
	"bytes"
	"fmt"
	"io"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"

	ftp "goftp.io/server/core"
)

type FileWriter interface {
	Write(filepath string, r io.Reader) error
	WriteFile(filename string, file *ach.File) error
}

func NewFileWriter(logger log.Logger, cfg service.ServerConfig, ftpServer *ftp.Server) FileWriter {
	if cfg.FTP != nil {
		return &FTPFileWriter{
			cfg:    cfg.FTP.Paths,
			logger: logger,
			server: ftpServer,
		}
	}
	return nil
}

type FTPFileWriter struct {
	cfg    service.Paths
	logger log.Logger
	server *ftp.Server
}

func (w *FTPFileWriter) WriteFile(filepath string, file *ach.File) error {
	var buf bytes.Buffer
	if err := ach.NewWriter(&buf).Write(file); err != nil {
		return fmt.Errorf("write %s: %v", filepath, err)
	}
	w.logger.Info().Log(fmt.Sprintf("writing %s (%d bytes)", filepath, buf.Len()))
	return w.Write(filepath, &buf)
}

func (w *FTPFileWriter) Write(filepath string, r io.Reader) error {
	driver, err := w.server.Factory.NewDriver()
	if err != nil {
		return fmt.Errorf("get driver to write %s: %v", filepath, err)
	}
	if _, err := driver.PutFile(filepath, r, false); err != nil {
		return fmt.Errorf("PUT %s: %v", filepath, err)
	}
	return nil
}
