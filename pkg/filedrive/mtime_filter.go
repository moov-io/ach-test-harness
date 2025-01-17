package filedrive

import (
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/base/log"
	"goftp.io/server/core"
)

type MTimeFilter struct {
	core.Driver
}

func (mtf MTimeFilter) ListDir(path string, callback func(core.FileInfo) error) error {
	now := time.Now()

	return mtf.Driver.ListDir(path, func(info core.FileInfo) error {
		if info.ModTime().Before(now) {
			return callback(info)
		}
		return nil
	})
}

type Factory struct {
	DriverFactory core.DriverFactory

	Logger       log.Logger
	ValidateOpts *ach.ValidateOpts
}

func (f *Factory) NewDriver() (core.Driver, error) {
	dd, err := f.DriverFactory.NewDriver()
	if err != nil {
		return nil, err
	}

	achDriver := NewACHDriver(f.Logger, f.ValidateOpts, dd)
	return MTimeFilter{
		Driver: achDriver,
	}, nil
}
