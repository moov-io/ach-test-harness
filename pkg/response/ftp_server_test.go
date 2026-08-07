package response

import (
	"net"
	"testing"

	ftp "goftp.io/server/v2"
	"goftp.io/server/v2/driver/file"
)

func fileBackedFtpServer(t *testing.T) (string, *ftp.Server) {
	t.Helper()

	dir := t.TempDir()
	t.Logf("Using %s for temporary FTP directory", dir)

	driver, err := file.NewDriver(dir)
	if err != nil {
		t.Fatal(err)
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

	// Close the probe listener; NewServer/ListenAndServe will bind the port.
	// Using the free port we discovered above.
	port := addr.Port
	ln.Close()

	opts := &ftp.Options{
		Driver:   driver,
		Port:     port,
		Hostname: "127.0.0.1",
		Perm:     ftp.NewSimplePerm("user", "group"),
		Logger:   &ftp.DiscardLogger{},
	}
	server, err := ftp.NewServer(opts)
	if err != nil {
		t.Fatal(err)
	}
	go server.ListenAndServe()

	return dir, server
}
