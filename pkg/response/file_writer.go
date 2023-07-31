package response

import (
	"bytes"
	"errors"
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
		driver, err := ftpServer.Factory.NewDriver()
		if err != nil {
			return nil // fmt.Errorf("get driver to write: %v", err)
		}
		return &FTPFileWriter{
			cfg:      cfg.FTP.Paths,
			rootPath: cfg.FTP.RootPath,
			logger:   logger,
			server:   ftpServer,
			driver: FTPFileDriver{
				Driver: driver,
			},
		}
	}
	return nil
}

type FTPFileWriter struct {
	cfg      service.Paths
	rootPath string
	logger   log.Logger
	server   *ftp.Server
	driver   FTPFileDriver
}

type FTPFileDriver struct {
	ftp.Driver
}

func (d *FTPFileDriver) ListDir(path string, callback func(ftp.FileInfo) error) error {
	return d.Driver.ListDir(path, func(f ftp.FileInfo) error {
		if f.ModTime().After(time.Now()) {
			// TODO JB: test and see if this actually works
			return errors.New("file is in the future")
		}
		return nil
	})
}

func (w *FTPFileWriter) WriteFile(filepath string, file *ach.File, futureDated *time.Duration) error {
	var buf bytes.Buffer
	if err := ach.NewWriter(&buf).Write(file); err != nil {
		return fmt.Errorf("write %s: %v", filepath, err)
	}
	w.logger.Info().Log(fmt.Sprintf("writing %s (%d bytes)", filepath, buf.Len()))
	return w.Write(filepath, &buf, futureDated)
}

func (w *FTPFileWriter) Write(path string, r io.Reader, futureDated *time.Duration) error {
	driver := w.driver

	if err := mkdir(driver, path); err != nil {
		return fmt.Errorf("mkdir: %s: %v", path, err)
	}

	if _, err := driver.PutFile(path, r, false); err != nil {
		return fmt.Errorf("STOR %s: %v", path, err)
	}

	if futureDated != nil {
		if err := os.Chtimes(filepath.Join(w.rootPath, path), time.Now(), time.Now().Add(*futureDated)); err != nil {
			return fmt.Errorf("chtimes: %s: %v", path, err)
		}
	}

	return nil
}

func mkdir(driver FTPFileDriver, path string) error {
	dir, _ := filepath.Split(path)
	return driver.MakeDir(dir)
}
