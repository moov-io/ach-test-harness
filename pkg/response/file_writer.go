package response

import (
	"bytes"
	"fmt"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"

	ftp "goftp.io/server/core"
)

type FileWriter interface {
	Write(filename string, file *ach.File) error
}

func NewFileWriter(cfg service.ServerConfig, ftpServer *ftp.Server) FileWriter {
	if cfg.FTP != nil {
		return &FTPFileWriter{
			cfg:    cfg.FTP.Paths,
			server: ftpServer,
		}
	}
	return nil
}

type FTPFileWriter struct {
	cfg    service.Paths
	server *ftp.Server
}

func (w *FTPFileWriter) Write(filepath string, file *ach.File) error {
	var buf bytes.Buffer
	if err := ach.NewWriter(&buf).Write(file); err != nil {
		return fmt.Errorf("write %s: %v", filepath, err)
	}

	driver, err := w.server.Factory.NewDriver()
	if err != nil {
		return fmt.Errorf("get driver to write %s: %v", filepath, err)
	}

	if _, err := driver.PutFile(filepath, &buf, false); err != nil {
		return fmt.Errorf("PUT %s: %v", filepath, err)
	}

	return nil
}
