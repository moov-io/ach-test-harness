package response

import (
	"fmt"

	"github.com/moov-io/ach"
	ftp "goftp.io/server/core"
)

type FTPWatcher struct {
	ftp.NullNotifier

	transformer FileTransfomer
}

func (notify *FTPWatcher) AfterFilePut(conn *ftp.Conn, dstPath string, size int64, err error) {
	fmt.Printf("PUT %s (bytes:%d)\n", dstPath, size)

	file, err := ach.ReadFile(dstPath) // TODO(adam): needs path relative to FTP.RootPath
	if err != nil {
		// TODO(adam): log, or something
	}
	if err := notify.Transform(file); err != nil {
		// TODO(adam): log, or something
	}
}

// TODO(adam): steps
// 1. match the incoming file, call all Responders.Process
// 2. have responders write output files
