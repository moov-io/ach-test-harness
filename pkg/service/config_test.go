// generated-from:b497f41560f9ad3b3f3fe17fb797f500908285ac22937d10d675391af26ee4ff DO NOT REMOVE, DO UPDATE

package service_test

import (
	"testing"

	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/config"
	"github.com/moov-io/base/log"

	"github.com/stretchr/testify/require"
)

func Test_ConfigLoading(t *testing.T) {
	logger := log.NewNopLogger()

	ConfigService := config.NewService(logger)

	gc := &service.GlobalConfig{}
	err := ConfigService.Load(gc)
	require.Nil(t, err)
}

// TODO JB: tests
