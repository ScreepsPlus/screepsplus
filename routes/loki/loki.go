package loki

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/mux"

	"github.com/screepsplus/screepsplus/grafana"
)

// NewRouter creates a new auth router
func NewRouter() *mux.Router {
	r := mux.NewRouter()
	AddHandlers(r)
	return r
}

// AddHandlers attaches handlers to a mux Router
func AddHandlers(r *mux.Router) {
	url, _ := url.Parse(os.Getenv("LOKI_URL"))
	proxy := httputil.NewSingleHostReverseProxy(url)

	r.Handle("/", loki(proxy))
}

func loki(next http.Handler) http.Handler {
	ourFunc := func(w http.ResponseWriter, r *http.Request) {
		client := grafana.NewClient()
		client.Client().SetCookies(r.Cookies())
		user, err := client.GetUser()
		if err != nil {
			log.Printf("loki getUser failed: %v", err)
			w.WriteHeader(500)
			return
		}
		r.Header.Set("X-Scope-OrgID", user.Login)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(ourFunc)
}