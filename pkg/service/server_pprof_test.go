package service

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moov-io/base/admin"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestBootAdminServer_PprofExposure(t *testing.T) {
	tests := []struct {
		name   string
		pprof  *admin.Pprof
		status int
	}{
		{name: "omitted config keeps historical pprof on", pprof: nil, status: http.StatusOK},
		{name: "explicitly disabled", pprof: &admin.Pprof{Enabled: false}, status: http.StatusNotFound},
		{name: "explicitly enabled", pprof: &admin.Pprof{Enabled: true}, status: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := make(chan error, 1)
			server := bootAdminServer(errs, log.NewNopLogger(), HTTPConfig{
				Bind:  BindAddress{Address: "127.0.0.1:0"},
				Pprof: tc.pprof,
			})
			require.NotNil(t, server)
			t.Cleanup(server.Shutdown)

			for _, path := range []string{"/debug/pprof/", "/debug/pprof/cmdline"} {
				require.Equal(t, tc.status, adminHTTPStatus(t, server.BindAddr(), path), path)
			}

			require.Equal(t, http.StatusOK, adminHTTPStatus(t, server.BindAddr(), "/metrics"))
			require.Empty(t, errs)
		})
	}
}

func adminHTTPStatus(t *testing.T, addr, path string) int {
	t.Helper()

	var last error
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			last = err
			time.Sleep(20 * time.Millisecond)
			continue
		}

		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		_, err = fmt.Fprintf(conn, "GET %s HTTP/1.0\r\nHost: localhost\r\n\r\n", path)
		if err != nil {
			conn.Close()
			last = err
			time.Sleep(20 * time.Millisecond)
			continue
		}

		line, err := bufio.NewReader(conn).ReadString('\n')
		conn.Close()
		if err != nil {
			last = err
			time.Sleep(20 * time.Millisecond)
			continue
		}

		parts := strings.SplitN(strings.TrimSpace(line), " ", 3)
		require.GreaterOrEqual(t, len(parts), 2, line)
		code, err := strconv.Atoi(parts[1])
		require.NoError(t, err, line)
		return code
	}

	require.NoError(t, last)
	return 0
}
