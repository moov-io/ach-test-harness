// generated-from:2c8b760c1363b7498fda493d7748a9433b06572be6ab5f2accb18b144acc9e94 DO NOT REMOVE, DO UPDATE

package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	ftp "goftp.io/server/core"
	"goftp.io/server/driver/file"

	"github.com/moov-io/base/admin"
	"github.com/moov-io/base/log"

	_ "github.com/moov-io/ach-test-harness"
)

// RunServers - Boots up all the servers and awaits till they are stopped.
func (env *Environment) RunServers(terminationListener chan error) func() {
	adminServer := bootAdminServer(terminationListener, env.Logger, env.Config.Servers.Admin)
	env.serveConfig(adminServer)

	env.Router = adminServer.Subrouter("/api")

	var shutdownFTPServer func()
	if env.Config.Servers.FTP != nil {
		ftpServer, shutdown := bootFTPServer(terminationListener, env.Logger, env.Config.Servers.FTP)
		env.FTPServer = ftpServer
		shutdownFTPServer = shutdown
	}

	return func() {
		adminServer.Shutdown()
		shutdownFTPServer()
	}
}

func bootFTPServer(errs chan<- error, logger log.Logger, cfg *FTPConfig) (*ftp.Server, func()) {
	// Setup data directory
	createDataDirectories(errs, logger, cfg)

	// Start the FTP server
	fileDriver := &file.DriverFactory{
		RootPath: cfg.RootPath,
		Perm:     ftp.NewSimplePerm("user", "group"),
	}
	opts := &ftp.ServerOpts{
		Factory:  fileDriver,
		Port:     cfg.Port,
		Hostname: cfg.Hostname,
		Auth: &ftp.SimpleAuth{
			Name:     cfg.Auth.Username,
			Password: cfg.Auth.Password,
		},
		PassivePorts: cfg.PassivePorts,
	}
	server := ftp.NewServer(opts)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			errs <- logger.Fatal().LogErrorf("problem running FTP server: %v", err).Err()
		}
	}()

	shutdown := func() {
		server.Shutdown()
	}

	return server, shutdown
}

func createDataDirectories(errs chan<- error, logger log.Logger, cfg *FTPConfig) {
	// Create the root path
	if err := os.MkdirAll(cfg.RootPath, 0777); err != nil {
		errs <- logger.Fatal().LogErrorf("problem creating root directory '%s': %v", cfg.RootPath, err).Err()
	}

	// Create sub-paths
	path := filepath.Join(cfg.RootPath, cfg.Paths.Files)
	if err := os.MkdirAll(path, 0777); err != nil {
		errs <- logger.Fatal().LogErrorf("problem creating files directory: %v", err).Err()
	}
	path = filepath.Join(cfg.RootPath, cfg.Paths.Return)
	if err := os.MkdirAll(path, 0777); err != nil {
		errs <- logger.Fatal().LogErrorf("problem creating return directory: %v", err).Err()
	}
}

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

func (env *Environment) serveConfig(svc *admin.Server) {
	svc.AddHandler("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(env.Config)
	})
}
