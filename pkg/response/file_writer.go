package response

import (
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

func (w *FTPFileWriter) Write(filename string, file *ach.File) error {
	return nil
}
