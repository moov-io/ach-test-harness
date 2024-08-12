package entries

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestEntryController(t *testing.T) {
	router := mux.NewRouter()
	logger := log.NewDefaultLogger()

	t.Run("/entries returns list of entries", func(t *testing.T) {
		repo := NewFTPRepository(&service.FTPConfig{
			RootPath: "./testdata",
		})
		service := NewEntryService(repo)
		controller := NewEntryController(logger, service)
		controller.AppendRoutes(router)

		req, _ := http.NewRequest("GET", "/entries", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		wantJSON := []byte(`
			[
			  {
			    "id":"",
			    "transactionCode":27,
			    "RDFIIdentification":"23138010",
			    "checkDigit":"4",
			    "DFIAccountNumber":"744-5678-99",
			    "amount":500000,
			    "identificationNumber":"location1234567",
			    "individualName":"Best Co. #123456789012",
			    "discretionaryData":"S ",
			    "traceNumber":"031300010000001",
			    "category":"Forward"
			  },
			  {
			    "id":"",
			    "transactionCode":27,
			    "RDFIIdentification":"23138010",
			    "checkDigit":"4",
			    "DFIAccountNumber":"744-5678-99",
			    "amount":125,
			    "identificationNumber":"Fee123456789012",
			    "individualName":"Best Co. #123456789012",
			    "discretionaryData":"S ",
			    "traceNumber":"031300010000002",
			    "category":"Forward"
			  },
			  {
			    "id":"",
			    "transactionCode":22,
			    "RDFIIdentification":"23138010",
			    "checkDigit":"4",
			    "DFIAccountNumber":"987654321",
			    "amount":100000000,
			    "identificationNumber":"               ",
			    "individualName":"Credit Account 1      ",
			    "discretionaryData":"  ",
			    "traceNumber":"121042880000002",
			    "category":"Forward"
			  }
			]
		`)

		gotJSON := rr.Body.Bytes()

		fmt.Printf("\n\n%s\n\n", string(gotJSON))

		require.Truef(t, jsonpatch.Equal(wantJSON, gotJSON), "received JSON does not match expected json")
	})
}
