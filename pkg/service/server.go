// generated-from:2c8b760c1363b7498fda493d7748a9433b06572be6ab5f2accb18b144acc9e94 DO NOT REMOVE, DO UPDATE

package service

import (
	"fmt"

	ftp "goftp.io/server/core"
	"goftp.io/server/driver/file"

	"github.com/moov-io/base/admin"
	"github.com/moov-io/base/log"

	_ "github.com/moov-io/ach-test-harness"
)

// RunServers - Boots up all the servers and awaits till they are stopped.
func (env *Environment) RunServers(terminationListener chan error) func() {
	adminServer := bootAdminServer(terminationListener, env.Logger, env.Config.Servers.Admin)

	var shutdownFTPServer func()
	if env.Config.Servers.FTP != nil {
		ftpServer, shutdown := bootFTPServer(env.Config.Servers.FTP)
		env.FTPServer = ftpServer
		shutdownFTPServer = shutdown
	}

	return func() {
		adminServer.Shutdown()
		shutdownFTPServer()
	}
}

func bootFTPServer(cfg *FTPConfig) (*ftp.Server, func()) {
	fileDriver := &file.DriverFactory{
		RootPath: cfg.RootPath,
		Perm:     ftp.NewSimplePerm("user", "group"),
	}
	opts := &ftp.ServerOpts{
		Factory:      fileDriver,
		Port:         cfg.Port,
		Hostname:     cfg.Hostname,
		Auth:         &ftp.SimpleAuth{Name: "admin", Password: "admin"},
		PassivePorts: "30000-30009",
	}
	server := ftp.NewServer(opts)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			// errs <- err logger.Fatal().LogErrorf("...").Err()
		}
	}()

	shutdown := func() {
		server.Shutdown()
	}

	return server, shutdown
}

// func bootHTTPServer(name string, routes *mux.Router, errs chan<- error, logger log.Logger, config HTTPConfig)
// 	shutdownServer := func() {
// 		if err := serve.Shutdown(context.Background()); err != nil {
// 			logger.Fatal().LogErrorf("shutting down: %v", err)
// 		}
// 	}

func bootAdminServer(errs chan<- error, logger log.Logger, config HTTPConfig) *admin.Server {
	adminServer := admin.NewServer(config.Bind.Address)

	go func() {
		logger.Info().Log(fmt.Sprintf("listening on %s", adminServer.BindAddr()))
		if err := adminServer.Listen(); err != nil {
			errs <- logger.Fatal().LogErrorf("problem starting admin http: %w", err).Err()
		}
	}()

	return adminServer
}
