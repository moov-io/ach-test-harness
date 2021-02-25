// generated-from:88728c8ab4eae05b171faaa88a77bce1900a7cbab5eeffb38231b49016675277 DO NOT REMOVE, DO UPDATE

package service

import (
	ftp "goftp.io/server/core"

	"github.com/moov-io/base/config"
	"github.com/moov-io/base/log"
	"github.com/moov-io/base/stime"

	_ "github.com/moov-io/ach-test-harness"
)

// Environment - Contains everything thats been instantiated for this service.
type Environment struct {
	Logger      log.Logger
	Config      *Config
	TimeService stime.TimeService

	// ftp or sftp server
	FTPServer *ftp.Server
	Shutdown  func()
}

// NewEnvironment - Generates a new default environment. Overrides can be specified via configs.
func NewEnvironment(env *Environment) (*Environment, error) {
	if env == nil {
		env = &Environment{}
	}

	env.Shutdown = func() {}

	if env.Logger == nil {
		env.Logger = log.NewDefaultLogger()
	}

	if env.Config == nil {
		cfg, err := LoadConfig(env.Logger)
		if err != nil {
			return nil, err
		}

		env.Config = cfg
	}

	if env.TimeService == nil {
		env.TimeService = stime.NewSystemTimeService()
	}

	return env, nil
}

func LoadConfig(logger log.Logger) (*Config, error) {
	configService := config.NewService(logger)

	global := &GlobalConfig{}
	if err := configService.Load(global); err != nil {
		return nil, err
	}

	cfg := &global.ACHTestHarness

	return cfg, nil
}
