// generated-from:7d3fb3a8684b121784ee01a9a23f1582efc7e5f5074ade65f27ed9e3d4558222 DO NOT REMOVE, DO UPDATE

package test

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
	"github.com/moov-io/base/stime"
	"github.com/moovfinancial/go-zero-trust/pkg/middleware"
	"github.com/moovfinancial/go-zero-trust/pkg/middleware/middlewaretest"
	"github.com/stretchr/testify/require"

	"github.com/moovfinancial/ach-test-harness/pkg/service"
)

type TestEnvironment struct {
	Assert     *require.Assertions
	StaticTime stime.StaticTimeService
	Claims     middleware.TrustedClaims

	service.Environment
}

func NewEnvironment(t *testing.T, router *mux.Router) *TestEnvironment {
	testEnv := &TestEnvironment{}

	testEnv.PublicRouter = router
	testEnv.Assert = require.New(t)
	testEnv.Logger = log.NewDefaultLogger()
	testEnv.StaticTime = stime.NewStaticTimeService()
	testEnv.TimeService = testEnv.StaticTime
	testEnv.Claims = middlewaretest.NewRandomClaims()

	cfg, err := service.LoadConfig(testEnv.Logger)
	if err != nil {
		t.Fatal(err)
	}
	testEnv.Config = cfg

	cfg.Database = CreateTestDatabase(t, TestDatabaseConfig())

	claimsFunc := func() middleware.TrustedClaims { return testEnv.Claims }
	mw := middlewaretest.NewTestMiddlewareLazy(testEnv.StaticTime, claimsFunc, "ach-test-harness")
	testEnv.ZeroTrustMiddleware = mw.Handler

	_, err = service.NewEnvironment(&testEnv.Environment)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(testEnv.Shutdown)

	return testEnv
}
