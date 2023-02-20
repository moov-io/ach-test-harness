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

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ln.Close()
	})

	opts := &ftp.ServerOpts{
		Factory:  factory,
		Port:     ln.Addr().(*net.TCPAddr).Port,
		Hostname: "127.0.0.1",
	}
	server := ftp.NewServer(opts)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			t.Fatal(err)
		}
	}()

	return dir, server
}
