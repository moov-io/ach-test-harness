package batches

import (
	"encoding/json"
	"net/http"

	"github.com/moov-io/base/log"

	"github.com/gorilla/mux"
)

func NewBatchController(logger log.Logger, service BatchService) *batchController {
	return &batchController{
		logger:  logger,
		service: service,
	}
}

type batchController struct {
	logger  log.Logger
	service BatchService
}

func (c *batchController) AppendRoutes(router *mux.Router) *mux.Router {
	router.
		Name("Batch.search").
		Methods("GET").
		Path("/batches").
		HandlerFunc(c.Search())

	return router
}

func (c *batchController) Search() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		batches, err := c.service.Search(readSearchOptions(r))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})

			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(batches)
	}
}

func readSearchOptions(r *http.Request) SearchOptions {
	query := r.URL.Query()
	opts := SearchOptions{
		AccountNumber: query.Get("accountNumber"),
		RoutingNumber: query.Get("routingNumber"),
		TraceNumber:   query.Get("traceNumber"),
		CreatedAfter:  query.Get("createdAfter"),
		Path:          query.Get("path"),
	}
	return opts
}
