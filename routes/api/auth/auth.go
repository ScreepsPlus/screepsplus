package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/screepsplus/screepsplus/auth"
	"github.com/volatiletech/authboss"
	"github.com/volatiletech/authboss-clientstate"
	_ "github.com/volatiletech/authboss/auth" // authboss modules
	"github.com/volatiletech/authboss/defaults"
	_ "github.com/volatiletech/authboss/logout"   // authboss modules
	_ "github.com/volatiletech/authboss/recover"  // authboss modules
	_ "github.com/volatiletech/authboss/register" // authboss modules
)

const (
	sessionCookieName = "screepsplus"
)

var (
	ab *authboss.Authboss
)

// AB returns the current authboss instance
func AB() *authboss.Authboss {
	return ab
}

// Init initializes authboss
func Init() *authboss.Authboss {
	rootURL := "http://localhost"
	if v, ok := os.LookupEnv("ROOT_URL"); ok {
		rootURL = v
	}
	mailFrom := "screepsplus@screepspl.us"
	if v, ok := os.LookupEnv("MAIL_FROM"); ok {
		mailFrom = v
	}
	mailFromName := "ScreepsPlus"
	if v, ok := os.LookupEnv("MAIL_FROM_NAME"); ok {
		mailFromName = v
	}
	domain := ".localhost"
	if v, ok := os.LookupEnv("DOMAIN"); ok {
		domain = v
	}

	ab = authboss.New()
	// TODO: These need to be ENV keys and _random_!
	cookieStoreKey, _ := base64.StdEncoding.DecodeString(`NpEPi8pEjKVjLGJ6kYCS+VTCzi6BUuDzU0wrwXyf5uDPArtlofn2AG6aTMiPmN3C909rsEWMNqJqhIVPGP3Exg==`)
	sessionStoreKey, _ := base64.StdEncoding.DecodeString(`AbfYwmmt8UCwUuhd9qvfNA9UCuN1cVcKJN1ofbiky6xCyyBj20whe40rJa3Su0WOWLWcPpO1taqJdsEI/65+JA==`)

	serverStore := ServerStorer{}
	cookieStore := abclientstate.NewCookieStorer(cookieStoreKey, nil)
	https := strings.HasPrefix(rootURL, "https")
	cookieStore.Secure = https
	cookieStore.Domain = domain
	
	sessionStore := abclientstate.NewSessionStorer(sessionCookieName, sessionStoreKey, nil)
	store := sessionStore.Store.(*sessions.CookieStore)
	store.Options.Secure = https
	store.Options.Domain = domain

	ab.Config.Storage.Server = serverStore        // dbImpl
	ab.Config.Storage.SessionState = sessionStore // sessImpl
	ab.Config.Storage.CookieState = cookieStore   // cookieImpl

	ab.Config.Paths.Mount = "/api/auth"
	ab.Config.Paths.RootURL = rootURL

	ab.Config.Mail.RootURL = fmt.Sprintf("%s%s", rootURL, "/auth")
	ab.Config.Mail.SubjectPrefix = fmt.Sprintf("[%s] ", mailFromName)
	ab.Config.Mail.From = mailFrom
	ab.Config.Mail.FromName = mailFromName

	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
	ab.Config.Core.MailRenderer = NewEmailRenderer()
	ab.Config.Modules.LogoutMethod = "POST"

	ab.Config.Modules.RegisterPreserveFields = []string{"email"}

	defaults.SetCore(&ab.Config, true, true)

	redir := ab.Config.Core.Redirector.(*defaults.Redirector)
	redir.CorceRedirectTo200 = true

	emailRule := defaults.Rules{
		FieldName:  defaults.FormValueEmail,
		Required:   true,
		MatchError: "Must be a valid email address",
		MustMatch:  regexp.MustCompile(`.*@.*\.[a-z]{1,}`),
	}

	ab.Config.Core.Mailer = NewSESMailer()

	r := ab.Config.Core.BodyReader.(*defaults.HTTPBodyReader)
	r.ReadJSON = true
	r.Rulesets["register"] = append(r.Rulesets["register"], emailRule)
	r.Whitelist["register"] = []string{
		defaults.FormValueEmail,
		defaults.FormValueUsername,
		defaults.FormValuePassword,
	}

	if err := ab.Init(); err != nil {
		panic(err)
	}
	return ab
}

// NewRouter creates a new auth router
func NewRouter() *mux.Router {
	r := mux.NewRouter()
	AddHandlers(r)
	return r
}

// AddHandlers attaches handlers to a mux Router
func AddHandlers(r *mux.Router) {
	r.Use(authboss.ModuleListMiddleware(ab))
	r.Use(authMigrate)
	r.PathPrefix("/").Handler(http.StripPrefix("/api/auth", ab.Config.Core.Router))
}

func authMigrate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.URL.Path)
		if r.URL.Path == "login" {
			var bodyBytes []byte
			bodyBytes, _ = ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			data := struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}{}
			json.Unmarshal(bodyBytes, &data)
			abuser, err := ab.Config.Storage.Server.Load(r.Context(), data.Username)
			if err == nil {
				user := abuser.(authboss.AuthableUser)
				pass := user.GetPassword()
				if !strings.HasPrefix(pass, "$") {
					log.Printf("Migrating password for user %s", data.Username)
					valid, err := auth.VerifyPassword(pass, data.Password)
					if err == nil && valid {
						ab.UpdatePassword(r.Context(), user, data.Password)
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func serverError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}
