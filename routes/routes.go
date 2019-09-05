package routes

import (
	"github.com/gorilla/mux"
	"github.com/screepsplus/screepsplus/routes/carbonapi"
	"github.com/screepsplus/screepsplus/routes/consent"
)

// NewRouter creates a new mux Router
func NewRouter() *mux.Router {
	r := mux.NewRouter()
	carbonapi.AddHandlers(r.PathPrefix("/grafana/carbonapi").Subrouter())
	carbonapi.AddHandlers(r.PathPrefix("/carbonapi").Subrouter())
	consent.AddHandlers(r.PathPrefix("/consent").Subrouter())
	return r
}
