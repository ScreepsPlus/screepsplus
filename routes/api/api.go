package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var jwtSecret = os.Getenv("JWT_SECRET")

type apiResp struct {
	OK    int                    `json:"ok"`
	Error string                 `json:"error,omitempty"`
	Extra map[string]interface{} `json:"inline"`
}

// NewRouter creates a new auth router
func NewRouter() *mux.Router {
	r := mux.NewRouter()
	AddHandlers(r)
	return r
}

// AddHandlers attaches handlers to a mux Router
func AddHandlers(r *mux.Router) {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		},
		Extractor: jwtmiddleware.FromFirst(
			jwtmiddleware.FromAuthHeader,
			func(r *http.Request) (string, error) {
				if user, pass, ok := r.BasicAuth(); ok && user == "token" {
					return pass, nil
				}
				return "", nil
			},
		),
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			log.Printf("API Auth Error: %v", err)
			unauthorized(w)
		},
		CredentialsOptional: false,
		SigningMethod:       jwt.SigningMethodHS256,
	})
	r.Use(jwtMiddleware.Handler)
	if _, ok := os.LookupEnv("GRAPHITE_URL"); ok {
		r.Handle("/stats/submit", statsSubmit())
	} else {
		r.HandleFunc("/stats/submit", func(w http.ResponseWriter, r *http.Request) {
			serverError(w)
		})
	}
}

func jsonErr(err string) []byte {
	b, _ := json.Marshal(apiResp{
		OK:    0,
		Error: err,
	})
	return b
}

func unauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(jsonErr("Unauthorized"))
}

func badRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write(jsonErr("Bad Request"))
}

func serverError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsonErr("Internal Server Error"))
}
