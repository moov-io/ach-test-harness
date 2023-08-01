package filedrive

import (
	"time"

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
}

func (f *Factory) NewDriver() (core.Driver, error) {
	dd, err := f.DriverFactory.NewDriver()
	if err != nil {
		return nil, err
	}
	return MTimeFilter{
		Driver: dd,
	}, nil
}
