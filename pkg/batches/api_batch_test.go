package batches

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

func TestBatchController(t *testing.T) {
	router := mux.NewRouter()
	logger := log.NewDefaultLogger()

	t.Run("/batches returns list of batches", func(t *testing.T) {
		repo := NewFTPRepository(logger, &service.FTPConfig{
			RootPath: "./testdata",
		})
		newBatchService := NewBatchService(repo)
		controller := NewBatchController(logger, newBatchService)
		controller.AppendRoutes(router)

		req, _ := http.NewRequest("GET", "/batches", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		wantJSON := []byte(`
			[
			  {
				"batchHeader": {
				  "id": "",
				  "serviceClassCode": 225,
				  "companyName": "Name on Account",
				  "companyIdentification": "231380104",
				  "standardEntryClassCode": "CCD",
				  "companyEntryDescription": "Vndr Pay",
				  "effectiveEntryDate": "190816",
				  "settlementDate": "   ",
				  "originatorStatusCode": 1,
				  "ODFIIdentification": "03130001",
				  "batchNumber": 1
				},
				"entryDetails": [
				  {
					"id": "",
					"transactionCode": 27,
					"RDFIIdentification": "23138010",
					"checkDigit": "4",
					"DFIAccountNumber": "744-5678-99",
					"amount": 500000,
					"identificationNumber": "location1234567",
					"individualName": "Best Co. #123456789012",
					"discretionaryData": "S ",
					"traceNumber": "031300010000001",
					"category": "Forward"
				  },
				  {
					"id": "",
					"transactionCode": 27,
					"RDFIIdentification": "23138010",
					"checkDigit": "4",
					"DFIAccountNumber": "744-5678-99",
					"amount": 125,
					"identificationNumber": "Fee123456789012",
					"individualName": "Best Co. #123456789012",
					"discretionaryData": "S ",
					"traceNumber": "031300010000002",
					"category": "Forward"
				  }
				],
				"batchControl": {
				  "id": "",
				  "serviceClassCode": 225,
				  "entryAddendaCount": 2,
				  "entryHash": 46276020,
				  "totalDebit": 500125,
				  "totalCredit": 0,
				  "companyIdentification": "231380104",
				  "ODFIIdentification": "03130001",
				  "batchNumber": 1
				},
				"offset": null
			  },
			  {
				"batchHeader": {
				  "id": "",
				  "serviceClassCode": 220,
				  "companyName": "Name on Account",
				  "companyIdentification": "231380104",
				  "standardEntryClassCode": "PPD",
				  "companyEntryDescription": "REG.SALARY",
				  "effectiveEntryDate": "190816",
				  "settlementDate": "   ",
				  "originatorStatusCode": 1,
				  "ODFIIdentification": "12104288",
				  "batchNumber": 1
				},
				"entryDetails": [
				  {
					"id": "",
					"transactionCode": 22,
					"RDFIIdentification": "23138010",
					"checkDigit": "4",
					"DFIAccountNumber": "987654321",
					"amount": 100000000,
					"identificationNumber": "               ",
					"individualName": "Credit Account 1      ",
					"discretionaryData": "  ",
					"traceNumber": "121042880000002",
					"category": "Forward"
				  }
				],
				"batchControl": {
				  "id": "",
				  "serviceClassCode": 220,
				  "entryAddendaCount": 1,
				  "entryHash": 23138010,
				  "totalDebit": 0,
				  "totalCredit": 100000000,
				  "companyIdentification": "231380104",
				  "ODFIIdentification": "12104288",
				  "batchNumber": 1
				},
				"offset": null
			  }
			]
		`)

		gotJSON := rr.Body.Bytes()

		fmt.Printf("\n\n%s\n\n", string(gotJSON))

		require.Truef(t, jsonpatch.Equal(wantJSON, gotJSON), "received JSON does not match expected json")
	})
}
