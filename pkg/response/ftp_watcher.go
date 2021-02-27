package response

import (
	"github.com/moov-io/ach"
	ftp "goftp.io/server/core"
)

type FTPWatcher struct {
	ftp.NullNotifier

	transformer *FileTransfomer
}

func (notify *FTPWatcher) AfterFilePut(conn *ftp.Conn, dstPath string, size int64, err error) {
	// fmt.Printf("PUT %s (bytes:%d)\n", dstPath, size)
	// fmt.Printf("  conn: %#v\n", conn)

	// Grab a file descriptor
	driver, err := conn.ServerOpts().Factory.NewDriver()
	if err != nil {
		return // TODO(adam): log, or something
	}
	_, fd, err := driver.GetFile(dstPath, 0)
	if err != nil {
		return // TODO(adam): log, or something
	}
	// Read the file that was uploaded
	file, err := ach.NewReader(fd).Read()
	if err != nil {
		// TODO(adam): log, or something
	}
	if err := file.Create(); err != nil {
		// TODO(adam): log, or something
	}
	if err := notify.transformer.Transform(&file); err != nil {
		// TODO(adam): log, or something
	}
}

// TODO(adam): steps
// 1. match the incoming file, call all Responders.Process
// 2. have responders write output files
