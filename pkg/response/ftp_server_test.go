package response

import (
	"net"
	"testing"

	ftp "goftp.io/server/core"
	"goftp.io/server/driver/file"
)

func fileBackedFtpServer(t *testing.T) (string, *ftp.Server) {
	t.Helper()

	dir := t.TempDir()
	t.Logf("Using %s for temporary FTP directory", dir)

	factory := &file.DriverFactory{
		RootPath: dir,
	}

	ln, err := net.Listen("tcp", ":0") //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ln.Close()
	})

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected listener address: %T", ln.Addr())
	}

	opts := &ftp.ServerOpts{
		Factory:  factory,
		Port:     addr.Port,
		Hostname: "127.0.0.1",
	}
	server := ftp.NewServer(opts)
	go server.ListenAndServe()

	return dir, server
}
