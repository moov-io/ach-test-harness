package response

import (
	"fmt"

	ftp "goftp.io/server/core"
)

type FTPNotifier struct {
	ftp.NullNotifier
}

func (notify *FTPNotifier) AfterFilePut(conn *ftp.Conn, dstPath string, size int64, err error) {
	fmt.Printf("PUT %s (bytes:%d)\n", dstPath, size)
}
