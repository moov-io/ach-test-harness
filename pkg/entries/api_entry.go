package entries

import (
	"encoding/json"
	"net/http"

	"github.com/moov-io/base/log"

	"github.com/gorilla/mux"
)

type EntryController interface {
	AppendRoutes(router *mux.Router) *mux.Router
	List()
}

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
		Name("Entry.list").
		Methods("GET").
		Path("/entries").
		HandlerFunc(c.List())

	router.
		Name("Entry.delete").
		Methods("DELETE").
		Path("/entries").
		HandlerFunc(c.Clean())

	return router
}

func (c *entryController) List() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		entries, err := c.service.List()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)

			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(entries)
	}
}

func (c *entryController) Clean() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c.service.Clean()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNoContent)
	}
}
