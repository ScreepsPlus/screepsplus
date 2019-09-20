package routes

import (
	"log"
	"net/http"
	"os"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/screepsplus/screepsplus/routes/api"
	"github.com/screepsplus/screepsplus/routes/api/auth"
	"github.com/screepsplus/screepsplus/routes/carbonapi"
	"github.com/screepsplus/screepsplus/routes/consent"
	"github.com/screepsplus/screepsplus/routes/loki"
	"github.com/volatiletech/authboss"
)

// NewRouter creates a new mux Router
func NewRouter() http.Handler {
	r := mux.NewRouter()
	ab := auth.Init()

	r.Use(authboss.ModuleListMiddleware(ab))
	r.Use(ab.LoadClientStateMiddleware)
	if checkEnv("carbonapi", []string{"CARBONAPI_URL"}) {
		carbonapi.AddHandlers(r.PathPrefix("/grafana/carbonapi").Subrouter())
		carbonapi.AddHandlers(r.PathPrefix("/carbonapi").Subrouter())
	}
	if checkEnv("consent", []string{"HYDRA_URL"}) {
		consent.AddHandlers(r.PathPrefix("/consent").Subrouter())
	}
	if checkEnv("loki", []string{"LOKI_URL"}) {
		loki.AddHandlers(r.PathPrefix("/grafana/loki").Subrouter())
		loki.AddHandlers(r.PathPrefix("/loki").Subrouter())
	}
	auth.AddHandlers(r.PathPrefix("/api/auth").Subrouter())
	if checkEnv("api", []string{"GRAPHITE_URL"}) {
		api.AddHandlers(r.PathPrefix("/api").Subrouter())
	}

	static := packr.New("static", "../static")
	r.PathPrefix("/").Handler(http.FileServer(static))

	return handlers.LoggingHandler(os.Stdout, r)
}

func checkEnv(name string, keys []string) bool {
	for _, k := range keys {
		if _, ok := os.LookupEnv(k); !ok {
			log.Printf("Skipping module %s: Env key '%s' missing", name, k)
			return false
		}
	}
	log.Printf("Loading module %s", name)
	return true
}
