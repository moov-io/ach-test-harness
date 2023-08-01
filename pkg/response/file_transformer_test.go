package response

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/ach/cmd/achcli/describe"
	"github.com/moov-io/base/log"

	"github.com/stretchr/testify/require"
)

func TestFileTransformer__CorrectedPrenote(t *testing.T) {
	resp := service.Response{
		Match: service.Match{
			EntryType:     service.EntryTypePrenote,
			AccountNumber: "810044964044",
		},
		Action: service.Action{
			Correction: &service.Correction{
				Code: "C01",
				Data: "445566778",
			},
		},
	}
	fileTransformer, dir := testFileTransformer(t, resp)

	prenote, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "prenote.ach"))
	require.NoError(t, err)
	require.NotNil(t, prenote)

	err = fileTransformer.Transform(prenote)
	require.NoError(t, err)

	retdir := filepath.Join(dir, "returned")

	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)

	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)

	var out bytes.Buffer
	describe.File(&out, found, nil)
	require.Contains(t, out.String(), "26 (Checking Return NOC Debit)")
	require.Contains(t, out.String(), "C01")
}

func TestFileTransformer__ReturnedPrenote(t *testing.T) {
	resp := service.Response{
		Match: service.Match{
			EntryType:     service.EntryTypePrenote,
			AccountNumber: "810044964044",
		},
		Action: service.Action{
			Return: &service.Return{
				Code: "R03",
			},
		},
	}
	fileTransformer, dir := testFileTransformer(t, resp)

	prenote, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "prenote.ach"))
	require.NoError(t, err)
	require.NotNil(t, prenote)

	err = fileTransformer.Transform(prenote)
	require.NoError(t, err)

	retdir := filepath.Join(dir, "returned")

	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)

	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)

	var out bytes.Buffer
	describe.File(&out, found, nil)
	require.Contains(t, out.String(), "28 (Checking Prenote Debit)")
	require.Contains(t, out.String(), "R03")
}

func TestFileTransformer__DelayCorrectedPrenote(t *testing.T) {
	var delay, err = time.ParseDuration("12h")
	require.NoError(t, err)

	resp := service.Response{
		Match: service.Match{
			EntryType:     service.EntryTypePrenote,
			AccountNumber: "810044964044",
		},
		Action: service.Action{
			Delay: &delay,
			Correction: &service.Correction{
				Code: "C01",
				Data: "445566778",
			},
		},
	}
	fileTransformer, dir := testFileTransformer(t, resp)

	prenote, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "prenote.ach"))
	require.NoError(t, err)
	require.NotNil(t, prenote)

	err = fileTransformer.Transform(prenote)
	require.NoError(t, err)

	retdir := filepath.Join(dir, "returned")

	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	finfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Less(t, finfo.ModTime(), time.Now().Add(delay))

	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)

	var out bytes.Buffer
	describe.File(&out, found, nil)
	require.Contains(t, out.String(), "26 (Checking Return NOC Debit)")
	require.Contains(t, out.String(), "C01")
}

func TestFileTransformer__DelayReturnedPrenote(t *testing.T) {
	var delay, err = time.ParseDuration("24h")
	require.NoError(t, err)

	resp := service.Response{
		Match: service.Match{
			EntryType:     service.EntryTypePrenote,
			AccountNumber: "810044964044",
		},
		Action: service.Action{
			Delay: &delay,
			Return: &service.Return{
				Code: "R03",
			},
		},
	}
	fileTransformer, dir := testFileTransformer(t, resp)

	prenote, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "prenote.ach"))
	require.NoError(t, err)
	require.NotNil(t, prenote)

	err = fileTransformer.Transform(prenote)
	require.NoError(t, err)

	retdir := filepath.Join(dir, "returned")

	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	finfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Less(t, finfo.ModTime(), time.Now().Add(delay))

	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)

	var out bytes.Buffer
	describe.File(&out, found, nil)
	require.Contains(t, out.String(), "28 (Checking Prenote Debit)")
	require.Contains(t, out.String(), "R03")
}

func testFileTransformer(t *testing.T, resp service.Response) (*FileTransfomer, string) {
	t.Helper()

	dir, ftpServer := fileBackedFtpServer(t)

	cfg := &service.Config{
		Matching: service.Matching{
			Debug: true,
		},
		Servers: service.ServerConfig{
			FTP: &service.FTPConfig{
				RootPath: dir,
				Paths: service.Paths{
					Return: "./returned/",
				},
			},
		},
	}
	responses := []service.Response{resp}

	logger := log.NewTestLogger()
	w := NewFileWriter(logger, cfg.Servers, ftpServer)

	return NewFileTransformer(logger, cfg, responses, w), dir
}

// TODO JB: more tests
