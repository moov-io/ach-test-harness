// generated-from:7d3fb3a8684b121784ee01a9a23f1582efc7e5f5074ade65f27ed9e3d4558222 DO NOT REMOVE, DO UPDATE

package test

import (
	"testing"

	"github.com/moov-io/base/log"
	"github.com/moov-io/base/stime"
	"github.com/stretchr/testify/require"

	"github.com/moov-io/ach-test-harness/pkg/service"
)

type TestEnvironment struct {
	Assert     *require.Assertions
	StaticTime stime.StaticTimeService

	service.Environment
}

func NewEnvironment(t *testing.T) *TestEnvironment {
	testEnv := &TestEnvironment{}

	testEnv.Assert = require.New(t)
	testEnv.Logger = log.NewDefaultLogger()
	testEnv.StaticTime = stime.NewStaticTimeService()
	testEnv.TimeService = testEnv.StaticTime

	cfg, err := service.LoadConfig(testEnv.Logger)
	if err != nil {
		t.Fatal(err)
	}
	testEnv.Config = cfg

	_, err = service.NewEnvironment(&testEnv.Environment)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(testEnv.Shutdown)

	return testEnv
}
