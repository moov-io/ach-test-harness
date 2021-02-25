// generated-from:49c22b0431e6d9826852cb04a312b6a5e9d1d3fc6168b36f0cb3040aae389631 DO NOT REMOVE, DO UPDATE

package service_test

import (
	"testing"

	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/assert"

	"github.com/moovfinancial/ach-test-harness/pkg/service"
)

func Test_Environment_Startup(t *testing.T) {
	a := assert.New(t)

	env := &service.Environment{
		Logger: log.NewDefaultLogger(),
	}

	env, err := service.NewEnvironment(env)
	a.Nil(err)

	t.Cleanup(env.Shutdown)
}
