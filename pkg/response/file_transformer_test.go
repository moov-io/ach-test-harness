package response

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"

	"github.com/stretchr/testify/require"
)

func TestFileTransformer(t *testing.T) {
	var delay, err = time.ParseDuration("24h")
	require.NoError(t, err)

	var matchPrenote = service.Match{
		EntryType:     service.EntryTypePrenote,
		AccountNumber: "810044964044",
	}
	var matchEntry1 = service.Match{
		IndividualName: "Incorrect Name",
	}
	var actionCopy = service.Action{
		Copy: &service.Copy{
			Path: "/reconciliation/",
		},
	}
	var actionReturn = service.Action{
		Return: &service.Return{
			Code: "R03",
		},
	}
	var actionCorrection = service.Action{
		Correction: &service.Correction{
			Code: "C01",
			Data: "445566778",
		},
	}
	var actionDelayReturn = actionReturn
	actionDelayReturn.Delay = &delay
	var actionDelayCorrection = actionCorrection
	actionDelayCorrection.Delay = &delay

	t.Run("NoMatch", func(t *testing.T) {
		resp := service.Response{
			Match:  matchEntry1,
			Action: actionCopy,
		}
		fileTransformer, dir := testFileTransformer(t, resp)

		// read the file
		achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "prenote.ach"))
		require.NoError(t, err)
		require.NotNil(t, achIn)

		// transform the file
		err = fileTransformer.Transform(achIn)
		require.NoError(t, err)

		// verify no "returned" files created
		retdir := filepath.Join(dir, "returned")
		_, err = os.ReadDir(retdir)
		require.Error(t, err)

		// verify no "reconciliation" files created
		recondir := filepath.Join(dir, "reconciliation")
		_, err = os.ReadDir(recondir)
		require.Error(t, err)
	})

	t.Run("CopyOnly", func(t *testing.T) {
		resp := service.Response{
			Match:  matchEntry1,
			Action: actionCopy,
		}
		fileTransformer, dir := testFileTransformer(t, resp)

		achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20210308-1806-071000301.ach"))
		require.NoError(t, err)
		require.NotNil(t, achIn)

		// transform the file
		err = fileTransformer.Transform(achIn)
		require.NoError(t, err)

		// verify no "returned" files created
		retdir := filepath.Join(dir, "returned")
		_, err = os.ReadDir(retdir)
		require.Error(t, err)

		// verify the "reconciliation" file created
		recondir := filepath.Join(dir, "reconciliation")
		fds, err := os.ReadDir(recondir)
		require.NoError(t, err)
		require.Len(t, fds, 1)
		found, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name()))
		require.Equal(t, matchEntry1.IndividualName, strings.Trim(found.Batches[0].GetEntries()[0].IndividualName, " "))

		// verify the timestamp on the file is in the past
		fInfo, err := fds[0].Info()
		require.NoError(t, err)
		require.Less(t, fInfo.ModTime(), time.Now())
	})

	t.Run("ProcessOnly - Return", func(t *testing.T) {
		resp := service.Response{
			Match:  matchPrenote,
			Action: actionReturn,
		}
		fileTransformer, dir := testFileTransformer(t, resp)

		// read the file
		achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "prenote.ach"))
		require.NoError(t, err)
		require.NotNil(t, achIn)

		// transform the file
		err = fileTransformer.Transform(achIn)
		require.NoError(t, err)

		// verify the "returned" file created
		retdir := filepath.Join(dir, "returned")
		fds, err := os.ReadDir(retdir)
		require.NoError(t, err)
		require.Len(t, fds, 1)
		found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
		require.NoError(t, err)
		require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

		// verify the timestamp on the file is in the past
		fInfo, err := fds[0].Info()
		require.NoError(t, err)
		require.Less(t, fInfo.ModTime(), time.Now())

		// verify no "reconciliation" files created
		recondir := filepath.Join(dir, "reconciliation")
		_, err = os.ReadDir(recondir)
		require.Error(t, err)
	})

	t.Run("ProcessOnly - Correction", func(t *testing.T) {
		resp := service.Response{
			Match:  matchPrenote,
			Action: actionCorrection,
		}
		fileTransformer, dir := testFileTransformer(t, resp)

		// read the file
		achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "prenote.ach"))
		require.NoError(t, err)
		require.NotNil(t, achIn)

		// transform the file
		err = fileTransformer.Transform(achIn)
		require.NoError(t, err)

		// verify the "returned" file created
		retdir := filepath.Join(dir, "returned")
		fds, err := os.ReadDir(retdir)
		require.NoError(t, err)
		require.Len(t, fds, 1)
		found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
		require.NoError(t, err)
		require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)

		// verify the timestamp on the file is in the past
		fInfo, err := fds[0].Info()
		require.NoError(t, err)
		require.Less(t, fInfo.ModTime(), time.Now())

		// verify no "reconciliation" files created
		recondir := filepath.Join(dir, "reconciliation")
		_, err = os.ReadDir(recondir)
		require.Error(t, err)
	})

	t.Run("DelayProcessOnly - Return", func(t *testing.T) {
		resp := service.Response{
			Match:  matchEntry1,
			Action: actionDelayReturn,
		}
		fileTransformer, dir := testFileTransformer(t, resp)

		achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20210308-1806-071000301.ach"))
		require.NoError(t, err)
		require.NotNil(t, achIn)

		// transform the file
		err = fileTransformer.Transform(achIn)
		require.NoError(t, err)

		// verify the "returned" file created
		retdir := filepath.Join(dir, "returned")
		fds, err := os.ReadDir(retdir)
		require.NoError(t, err)
		require.Len(t, fds, 1)
		found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
		require.NoError(t, err)
		require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

		// verify the timestamp on the file is in the future
		fInfo, err := fds[0].Info()
		require.NoError(t, err)
		require.Greater(t, fInfo.ModTime(), time.Now())

		// verify no "reconciliation" files created
		recondir := filepath.Join(dir, "reconciliation")
		_, err = os.ReadDir(recondir)
		require.Error(t, err)
	})

	t.Run("DelayProcessOnly - Correction", func(t *testing.T) {
		resp := service.Response{
			Match:  matchEntry1,
			Action: actionDelayCorrection,
		}
		fileTransformer, dir := testFileTransformer(t, resp)

		achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20210308-1806-071000301.ach"))
		require.NoError(t, err)
		require.NotNil(t, achIn)

		// transform the file
		err = fileTransformer.Transform(achIn)
		require.NoError(t, err)

		// verify the "returned" file created
		retdir := filepath.Join(dir, "returned")
		fds, err := os.ReadDir(retdir)
		require.NoError(t, err)
		require.Len(t, fds, 1)
		found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
		require.NoError(t, err)
		require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)

		// verify the timestamp on the file is in the future
		fInfo, err := fds[0].Info()
		require.NoError(t, err)
		require.Greater(t, fInfo.ModTime(), time.Now())

		// verify no "reconciliation" files created
		recondir := filepath.Join(dir, "reconciliation")
		_, err = os.ReadDir(recondir)
		require.Error(t, err)
	})

	t.Run("CopyAndDelayProcess - Return", func(t *testing.T) {
		resp1 := service.Response{
			Match:  matchEntry1,
			Action: actionCopy,
		}
		resp2 := service.Response{
			Match:  matchEntry1,
			Action: actionDelayReturn,
		}
		fileTransformer, dir := testFileTransformer(t, resp1, resp2)

		achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20210308-1806-071000301.ach"))
		require.NoError(t, err)
		require.NotNil(t, achIn)

		// transform the file
		err = fileTransformer.Transform(achIn)
		require.NoError(t, err)

		// verify the "returned" file created
		retdir := filepath.Join(dir, "returned")
		fds, err := os.ReadDir(retdir)
		require.NoError(t, err)
		require.Len(t, fds, 1)
		found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
		require.NoError(t, err)
		require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

		// verify the timestamp on the file is in the future
		fInfo, err := fds[0].Info()
		require.NoError(t, err)
		require.Greater(t, fInfo.ModTime(), time.Now())

		// verify the "reconciliation" file created
		recondir := filepath.Join(dir, "reconciliation")
		fds, err = os.ReadDir(recondir)
		require.NoError(t, err)
		require.Len(t, fds, 1)
		found, _ = ach.ReadFile(filepath.Join(recondir, fds[0].Name()))
		require.Equal(t, matchEntry1.IndividualName, strings.Trim(found.Batches[0].GetEntries()[0].IndividualName, " "))

		// verify the timestamp on the file is in the past
		fInfo, err = fds[0].Info()
		require.NoError(t, err)
		require.Less(t, fInfo.ModTime(), time.Now())
	})

	t.Run("CopyAndDelayProcess - Correction", func(t *testing.T) {
		resp1 := service.Response{
			Match:  matchEntry1,
			Action: actionCopy,
		}
		resp2 := service.Response{
			Match:  matchEntry1,
			Action: actionDelayCorrection,
		}
		fileTransformer, dir := testFileTransformer(t, resp1, resp2)

		achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20210308-1806-071000301.ach"))
		require.NoError(t, err)
		require.NotNil(t, achIn)

		// transform the file
		err = fileTransformer.Transform(achIn)
		require.NoError(t, err)

		// verify the "returned" file created
		retdir := filepath.Join(dir, "returned")
		fds, err := os.ReadDir(retdir)
		require.NoError(t, err)
		require.Len(t, fds, 1)
		found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
		require.NoError(t, err)
		require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)

		// verify the timestamp on the file is in the future
		fInfo, err := fds[0].Info()
		require.NoError(t, err)
		require.Greater(t, fInfo.ModTime(), time.Now())

		// verify the "reconciliation" file created
		recondir := filepath.Join(dir, "reconciliation")
		fds, err = os.ReadDir(recondir)
		require.NoError(t, err)
		require.Len(t, fds, 1)
		found, _ = ach.ReadFile(filepath.Join(recondir, fds[0].Name()))
		require.Equal(t, matchEntry1.IndividualName, strings.Trim(found.Batches[0].GetEntries()[0].IndividualName, " "))

		// verify the timestamp on the file is in the past
		fInfo, err = fds[0].Info()
		require.NoError(t, err)
		require.Less(t, fInfo.ModTime(), time.Now())
	})
}

func testFileTransformer(t *testing.T, resp ...service.Response) (*FileTransfomer, string) {
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
	responses := resp

	logger := log.NewTestLogger()
	w := NewFileWriter(logger, cfg.Servers, ftpServer)

	return NewFileTransformer(logger, cfg, responses, w), dir
}
