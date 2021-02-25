// generated-from:83a5239f515ffe938c49859f9bff6dfa6111234af33ad7caa1fce0163239ea25 DO NOT REMOVE, DO UPDATE

package main

import (
	"os"

	"github.com/moov-io/base/log"

	"github.com/moovfinancial/ach-test-harness"
	"github.com/moovfinancial/ach-test-harness/pkg/service"
)

func main() {
	env := &service.Environment{
		Logger: log.NewDefaultLogger().Set("app", log.String("ach-test-harness")).Set("version", log.String(ach-test-harness.Version)),
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

	service.AwaitTermination(env.Logger, termListener)
}
