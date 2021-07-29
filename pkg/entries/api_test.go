package entries

import (
	"net/http"
	"net/http/httptest"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestEntryController(t *testing.T) {
	router := mux.NewRouter()
	logger := log.NewDefaultLogger()

	achFile, err := mockACHFile()
	require.NoError(t, err)

	service := NewEntryService()
	err = service.AddFile(achFile)
	require.NoError(t, err)

	controller := NewEntryController(logger, service)
	controller.AppendRoutes(router)

	t.Run("/entries returns list of entries", func(t *testing.T) {
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
		    "DFIAccountNumber":"744-5678-99      ",
		    "amount":500000,
		    "identificationNumber":"location1234567",
		    "individualName":"Best Co. #123456789012",
		    "discretionaryData":"S ",
		    "traceNumber":"031300010000001"
		  },
		  {
		    "id":"",
		    "transactionCode":27,
		    "RDFIIdentification":"23138010",
		    "checkDigit":"4",
		    "DFIAccountNumber":"744-5678-99      ",
		    "amount":125,
		    "identificationNumber":"Fee123456789012",
		    "individualName":"Best Co. #123456789012",
		    "discretionaryData":"S ",
		    "traceNumber":"031300010000002"
		  }
		]
	`)
		gotJSON := rr.Body.Bytes()

		require.Truef(t, jsonpatch.Equal(wantJSON, gotJSON), "received JSON does not match expected json")
	})

	t.Run("DELETE /entries removes entries", func(t *testing.T) {
		// delete all entries
		req, _ := http.NewRequest("DELETE", "/entries", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)

		// check that there are no entries
		entries, err := service.List()

		require.NoError(t, err)
		require.Len(t, entries, 0)
	})
}
