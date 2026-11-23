// generated-from:83a5239f515ffe938c49859f9bff6dfa6111234af33ad7caa1fce0163239ea25 DO NOT REMOVE, DO UPDATE

package main

import (
	"flag"
	"fmt"
	"os"

	achtestharness "github.com/moov-io/ach-test-harness"
	"github.com/moov-io/ach-test-harness/pkg/batches"
	"github.com/moov-io/ach-test-harness/pkg/entries"
	"github.com/moov-io/ach-test-harness/pkg/response"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
)

var (
	flagVersion = flag.Bool("version", false, "Print the version and exit")
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "ach-test-harness %s\n\n", achtestharness.Version)
		fmt.Fprintf(out, "ach-test-harness is an FTP server that responds to uploaded ACH files with returns and corrections.\n\n")
		fmt.Fprintf(out, "USAGE\n  ach-test-harness [flags]\n\n")
		fmt.Fprintf(out, "ENVIRONMENT\n")
		fmt.Fprintf(out, "  APP_CONFIG          Path to a YAML config file that is merged over the default config\n")
		fmt.Fprintf(out, "  APP_CONFIG_SECRETS  Path to a second YAML config file, merged after APP_CONFIG\n\n")
		fmt.Fprintf(out, "FLAGS\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *flagVersion {
		fmt.Println(achtestharness.Version)
		return
	}

	logger := log.NewDefaultLogger().
		Set("app", log.String("ach-test-harness")).
		Set("version", log.String(achtestharness.Version))

	env, err := service.NewEnvironment(&service.Environment{
		Logger: logger,
	})
	if err != nil {
		// NewEnvironment can return a nil Environment with the error,
		// so use the logger from above.
		logger.Fatal().LogErrorf("Error loading up environment: %v", err)
		os.Exit(1)
	}
	defer env.Shutdown()

	termListener := service.NewTerminationListener()

	stopServers := env.RunServers(termListener)
	defer stopServers()

	// Initialize our responders
	entryRepository := entries.NewFTPRepository(env.Config.Servers.FTP)
	entryService := entries.NewEntryService(entryRepository)
	entryController := entries.NewEntryController(env.Logger, entryService)
	entryController.AppendRoutes(env.Router)

	batchRepository := batches.NewFTPRepository(env.Config.Servers.FTP)
	batchService := batches.NewBatchService(batchRepository)
	batchController := batches.NewBatchController(env.Logger, batchService)
	batchController.AppendRoutes(env.Router)

	fileWriter := response.NewFileWriter(env.Logger, env.Config.Servers, env.FTPServer)
	fileTransformer := response.NewFileTransformer(env.Logger, env.Config, env.Config.Responses, fileWriter)
	response.Register(env.Logger, env.Config.ValidateOpts, env.FTPServer, fileTransformer)

	// Block for a signal to shutdown
	service.AwaitTermination(env.Logger, termListener)
}
