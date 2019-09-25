package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/screepsplus/screepsplus/models"
	"github.com/screepsplus/screepsplus/routes/api/auth"
	"github.com/volatiletech/authboss"
)

var jwtSecret = os.Getenv("JWT_SECRET")

type apiResp struct {
	Status string                 `json:"status"`
	Error  string                 `json:"error,omitempty"`
	Extra  map[string]interface{} `json:"inline"`
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
	sr := r.PathPrefix("/stats").Subrouter()
	sr.Use(jwtMiddleware.Handler)
	if _, ok := os.LookupEnv("GRAPHITE_URL"); ok {
		sr.Handle("/submit", statsSubmit())
	} else {
		sr.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
			serverError(w)
		})
	}
	sr = r.PathPrefix("/user").Subrouter()
	sr.Use(authboss.Middleware2(auth.AB(), authboss.RequireNone, authboss.RespondUnauthorized))
	sr.HandleFunc("/me", GetUserMe).Methods("GET")
	sr.HandleFunc("/email", PostUserEmail).Methods("POST")
	sr.HandleFunc("/stats", DeleteUserStats).Methods("DELETE")
}

// GetUserMe GET /api/user/me
func GetUserMe(w http.ResponseWriter, r *http.Request) {
	abuser, err := auth.AB().CurrentUser(r)
	if err != nil {
		serverError(w)
		return
	}
	user := abuser.(*models.User)
	bytes, err := json.Marshal(map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

// PostUserEmail POST /api/user/email
func PostUserEmail(w http.ResponseWriter, r *http.Request) {
	abuser, err := auth.AB().CurrentUser(r)
	if err != nil {
		serverError(w)
		return
	}
	user := abuser.(*models.User)
	if r.Header.Get("content-type") != "application/json" {
		badRequest(w)
		return
	}
	dec := json.NewDecoder(r.Body)
	data := struct {
		Email string `json:"email"`
	}{}
	err = dec.Decode(&data)
	if err != nil {
		badRequest(w)
		return
	}
	user.Email = &data.Email

	auth.AB().Config.Storage.Server.Save(r.Context(), user)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(apiResp{
		Status: "success",
	})
	w.Write(b)
}

// DeleteUserStats POST /api/user/email
func DeleteUserStats(w http.ResponseWriter, r *http.Request) {
	abuser, err := auth.AB().CurrentUser(r)
	if err != nil {
		serverError(w)
		return
	}
	user := abuser.(*models.User)

	_ = os.RemoveAll(fmt.Sprintf("/data/screeps/%s", user.Username))

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(apiResp{
		Status: "success",
	})
	w.Write(b)
}

func jsonErr(err string) []byte {
	b, _ := json.Marshal(apiResp{
		Status: "error",
		Error:  err,
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
