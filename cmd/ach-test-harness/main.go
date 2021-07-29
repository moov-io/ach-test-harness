// generated-from:83a5239f515ffe938c49859f9bff6dfa6111234af33ad7caa1fce0163239ea25 DO NOT REMOVE, DO UPDATE

package main

import (
	"os"

	achtestharness "github.com/moov-io/ach-test-harness"
	"github.com/moov-io/ach-test-harness/pkg/response"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
)

func main() {
	env := &service.Environment{
		Logger: log.NewDefaultLogger().
			Set("app", log.String("ach-test-harness")).
			Set("version", log.String(achtestharness.Version)),
	}

	env, err := service.NewEnvironment(env)
	if err != nil {
		env.Logger.Fatal().LogErrorf("Error loading up environment: %v", err)
		os.Exit(1)
	}
	defer env.Shutdown()

	termListener := service.NewTerminationListener()

	stopServers := env.RunServers(termListener)
	defer stopServers()

	// Initialize our responders
	fileWriter := response.NewFileWriter(env.Logger, env.Config.Servers, env.FTPServer)
	fileTransformer := response.NewFileTransformer(env.Logger, env.Config, env.Config.Responses, fileWriter)
	response.Register(env.Logger, env.Config.ValidateOpts, env.FTPServer, fileTransformer)

	// Block for a signal to shutdown
	service.AwaitTermination(env.Logger, termListener)
}
