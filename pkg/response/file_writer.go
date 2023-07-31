package response

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"

	ftp "goftp.io/server/core"
)

type FileWriter interface {
	Write(filepath string, r io.Reader, delay *time.Duration) error
	WriteFile(filename string, file *ach.File, delay *time.Duration) error
}

func NewFileWriter(logger log.Logger, cfg service.ServerConfig, ftpServer *ftp.Server) FileWriter {
	if cfg.FTP != nil {
		return &FTPFileWriter{
			cfg:      cfg.FTP.Paths,
			rootPath: cfg.FTP.RootPath,
			logger:   logger,
			server:   ftpServer,
		}
	}
	return nil
}

type FTPFileWriter struct {
	cfg      service.Paths
	rootPath string
	logger   log.Logger
	server   *ftp.Server
}

func (w *FTPFileWriter) WriteFile(filepath string, file *ach.File, delay *time.Duration) error {
	var buf bytes.Buffer
	if err := ach.NewWriter(&buf).Write(file); err != nil {
		return fmt.Errorf("write %s: %v", filepath, err)
	}
	w.logger.Info().Log(fmt.Sprintf("writing %s (%d bytes)", filepath, buf.Len()))
	return w.Write(filepath, &buf, delay)
}

func (w *FTPFileWriter) Write(path string, r io.Reader, delay *time.Duration) error {
	driver, err := w.server.Factory.NewDriver()
	if err != nil {
		return fmt.Errorf("get driver to write %s: %v", path, err)
	}

	if err := mkdir(driver, path); err != nil {
		return fmt.Errorf("mkdir: %s: %v", path, err)
	}

	if _, err := driver.PutFile(path, r, false); err != nil {
		return fmt.Errorf("STOR %s: %v", path, err)
	}

	if delay != nil {
		if err := os.Chtimes(filepath.Join(w.rootPath, path), time.Now(), time.Now().Add(*delay)); err != nil {
			return fmt.Errorf("chtimes: %s: %v", path, err)
		}
	}

	return nil
}

func mkdir(driver ftp.Driver, path string) error {
	dir, _ := filepath.Split(path)
	return driver.MakeDir(dir)
}
