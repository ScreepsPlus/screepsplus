package auth

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"regexp"

	"github.com/gorilla/mux"
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

// Init initializes authboss
func Init() *authboss.Authboss {
	rootURL := "http://localhost"
	if v, ok := os.LookupEnv("ROOT_URL"); ok {
		rootURL = v
	}
	ab = authboss.New()
	// TODO: These need to be ENV keys and _random_!
	cookieStoreKey, _ := base64.StdEncoding.DecodeString(`NpEPi8pEjKVjLGJ6kYCS+VTCzi6BUuDzU0wrwXyf5uDPArtlofn2AG6aTMiPmN3C909rsEWMNqJqhIVPGP3Exg==`)
	sessionStoreKey, _ := base64.StdEncoding.DecodeString(`AbfYwmmt8UCwUuhd9qvfNA9UCuN1cVcKJN1ofbiky6xCyyBj20whe40rJa3Su0WOWLWcPpO1taqJdsEI/65+JA==`)

	serverStore := ServerStorer{}
	cookieStore := abclientstate.NewCookieStorer(cookieStoreKey, nil)
	cookieStore.Secure = strings.HasPrefix(rootURL, "https")
	sessionStore := abclientstate.NewSessionStorer(sessionCookieName, sessionStoreKey, nil)
	ab.Config.Storage.Server = serverStore        // dbImpl
	ab.Config.Storage.SessionState = sessionStore // sessImpl
	ab.Config.Storage.CookieState = cookieStore   // cookieImpl

	ab.Config.Paths.Mount = "/auth"
	ab.Config.Paths.RootURL = rootURL

	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
	ab.Config.Core.MailRenderer = defaults.JSONRenderer{}

	ab.Config.Modules.RegisterPreserveFields = []string{"email"}

	defaults.SetCore(&ab.Config, true, true)

	emailRule := defaults.Rules{
		FieldName:  defaults.FormValueEmail,
		Required:   true,
		MatchError: "Must be a valid email address",
		MustMatch:  regexp.MustCompile(`.*@.*\.[a-z]{1,}`),
	}
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
	r.PathPrefix("/").Handler(http.StripPrefix("/api/auth", ab.Config.Core.Router))
}
