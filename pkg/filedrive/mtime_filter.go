package filedrive

import (
	"os"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/base/log"
	"goftp.io/server/v2"
	"goftp.io/server/v2/driver/file"
)

type MTimeFilter struct {
	server.Driver
}

func (mtf MTimeFilter) ListDir(ctx *server.Context, path string, callback func(os.FileInfo) error) error {
	now := time.Now()

	return mtf.Driver.ListDir(ctx, path, func(info os.FileInfo) error {
		if info.ModTime().Before(now) {
			return callback(info)
		}
		return nil
	})
}

// NewDriver builds the FTP filesystem driver used by the test harness:
// a file-backed store wrapped with ACH validation and future-mtime filtering.
func NewDriver(logger log.Logger, validateOpts *ach.ValidateOpts, rootPath string) (server.Driver, error) {
	base, err := file.NewDriver(rootPath)
	if err != nil {
		return nil, err
	}
	return MTimeFilter{
		Driver: NewACHDriver(logger, validateOpts, base),
	}, nil
}

var _ server.Driver = MTimeFilter{}
