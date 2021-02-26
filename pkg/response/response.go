package response

import (
	"github.com/moov-io/base/log"

	ftp "goftp.io/server/core"
)

func Register(
	logger log.Logger,
	ftpServer *ftp.Server,
	transformer *FileTransfomer,
) {
	if ftpServer != nil && transformer != nil {
		// TODO(adam): log, or something
		ftpServer.RegisterNotifer(&FTPWatcher{
			transformer: transformer,
		})
	}
}
