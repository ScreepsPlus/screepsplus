package auth

import (
	"bytes"
	"net/http"
	"net/url"

	"strings"

	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"github.com/ory/hydra/sdk/go/hydra/client"
	"github.com/ory/hydra/sdk/go/hydra/client/admin"
	hmodels "github.com/ory/hydra/sdk/go/hydra/models"
	"github.com/screepsplus/screepsplus/db"
	"github.com/screepsplus/screepsplus/models"

	"html/template"

	"log"
)

var loginTemplate *template.Template
var consentTemplate *template.Template
var adminClient *client.OryHydra

var hydraURL = "http://192.168.0.119:4445"

func init() {
	box := packr.NewBox("../../templates")
	loginHTML, _ := box.FindString("login.html.tmpl")
	loginTemplate = template.Must(template.New("login_view").Parse(loginHTML))
	adminURL, _ := url.Parse(hydraURL)
	adminClient = client.NewHTTPClientWithConfig(nil, &client.TransportConfig{Schemes: []string{adminURL.Scheme}, Host: adminURL.Host, BasePath: adminURL.Path})
}

// NewRouter creates a new auth router
func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/consent", consentHandler)
	return r
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		chal := r.URL.Query().Get("login_challenge")
		glr := admin.NewGetLoginRequestParams().WithLoginChallenge(chal)
		resp, err := adminClient.Admin.GetLoginRequest(glr)
		if err != nil {
			log.Printf("login err: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		pay := resp.Payload
		if pay.Skip {
			resp, err := adminClient.Admin.AcceptLoginRequest(admin.NewAcceptLoginRequestParams().WithLoginChallenge(chal).WithBody(&hmodels.HandledLoginRequest{
				Subject: &pay.Subject,
			}))
			if err != nil {
				log.Printf("login err: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, resp.Payload.RedirectTo, http.StatusFound)
		} else {
			data := map[string]interface{}{
				"LoginChallenge": chal,
			}
			render(w, r, loginTemplate, "login_view", data)
		}
	case http.MethodPost:
		r.ParseForm()
		chal := r.Form.Get("login_challenge")
		username := strings.ToLower(r.Form.Get("username"))
		password := r.Form.Get("password")
		remember := r.Form.Get("remember") == "yes"

		invalid := func() {
			data := map[string]interface{}{
				"LoginChallenge": chal,
				"Error":          "Invalid username or password",
			}
			render(w, r, loginTemplate, "login_view", data)
		}

		user := models.User{}
		if db.DB().Where(&models.User{Username: username}).First(&user).RecordNotFound() {
			invalid()
			return
		}
		if user.VerifyPassword(password) {
			resp, err := adminClient.Admin.AcceptLoginRequest(admin.NewAcceptLoginRequestParams().WithLoginChallenge(chal).WithBody(&hmodels.HandledLoginRequest{
				Subject:     &username,
				Remember:    remember,
				RememberFor: 0,
			}))
			if err != nil {
				log.Printf("login err: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, resp.Payload.RedirectTo, http.StatusFound)
		} else {
			invalid()
		}
	}
}

func consentHandler(w http.ResponseWriter, r *http.Request) {
	chal := r.URL.Query().Get("consent_challenge")
	resp, err := adminClient.Admin.GetConsentRequest(admin.NewGetConsentRequestParams().WithConsentChallenge(chal))
	if err != nil {
		log.Printf("consent err: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	pay := resp.Payload
	aresp, err := adminClient.Admin.AcceptConsentRequest(admin.NewAcceptConsentRequestParams().WithConsentChallenge(chal).WithBody(&hmodels.HandledConsentRequest{
		GrantedAudience: pay.RequestedAudience,
		GrantedScope:    pay.RequestedScope,
		Remember:        true,
		RememberFor:     0,
	}))

	if err != nil {
		log.Printf("consent err: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, aresp.Payload.RedirectTo, http.StatusFound)
}

func render(w http.ResponseWriter, r *http.Request, tpl *template.Template, name string, data interface{}) {
	buf := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(buf, name, data); err != nil {
		log.Printf("render error: %s %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}
