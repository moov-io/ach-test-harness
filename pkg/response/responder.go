package response

import (
	"github.com/moov-io/ach-test-harness/pkg/service"

	ftp "goftp.io/server/core"
)

type Responder interface {
	// TODO(adam):
}

func Setup(cfg []service.Response, ftpServer *ftp.Server) {
	// for now just register a Notifier
	notifier := &FTPNotifier{}
	ftpServer.RegisterNotifer(notifier)
}
