package entries

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/moov-io/base/log"
	"github.com/moov-io/base/telemetry"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
)

func NewEntryController(logger log.Logger, service EntryService) *entryController {
	return &entryController{
		logger:  logger,
		service: service,
	}
}

type entryController struct {
	logger  log.Logger
	service EntryService
}

func (c *entryController) AppendRoutes(router *mux.Router) *mux.Router {
	router.
		Name("Entry.search").
		Methods("GET").
		Path("/entries").
		HandlerFunc(c.Search())

	return router
}

func (c *entryController) Search() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := telemetry.StartSpan(r.Context(), "api-entry-search")
		defer span.End()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		opts := readSearchOptions(r)
		span.SetAttributes(
			attribute.String("search.trace_number", opts.TraceNumber),
		)

		entries, err := c.service.Search(ctx, opts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})

			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(entries)
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
	if n, _ := strconv.ParseInt(query.Get("amount"), 10, 32); n > 0 {
		opts.Amount = int(n)
	}
	return opts
}
