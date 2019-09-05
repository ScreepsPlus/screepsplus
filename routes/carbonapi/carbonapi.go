package carbonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	m "github.com/grafana/grafana/pkg/models"
	"github.com/screepsplus/screepsplus/grafana"
)

var backendURL = os.Getenv("BACKEND_URL")
var grafanaURL = os.Getenv("GRAFANA_URL")

// NewRouter creates a new auth router
func NewRouter() *mux.Router {
	r := mux.NewRouter()
	AddHandlers(r)
	return r
}

// AddHandlers attaches handlers to a mux Router
func AddHandlers(r *mux.Router) {
	url, _ := url.Parse(backendURL)
	proxy := httputil.NewSingleHostReverseProxy(url)

	r.Path("/metrics/find").Handler(find(proxy))
	r.Path("/render").Handler(render(proxy))
}

func find(next http.Handler) http.Handler {
	ourFunc := func(w http.ResponseWriter, r *http.Request) {
		client := grafana.NewClient()
		client.Client().SetCookies(r.Cookies())
		r.ParseForm()
		query := r.Form.Get("query")
		if query == "*" || query == "screeps" {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(200)
			b := metricMap([]string{"screeps"}, "")
			w.Write(b)
			return
		}
		orgs, err := client.GetUserOrgs()
		if err != nil {
			log.Printf("find getOrgs %+v\n", err)
			w.WriteHeader(500)
			return
		}
		valid := false
		for _, org := range orgs {
			val := fmt.Sprintf("screeps.%s", org.Name)
			if strings.HasPrefix(query, val) {
				valid = true
				break
			}
		}
		if query == "screeps.*" || strings.HasPrefix(query, "screeps.*") {
			acl := getACL(orgs)
			query = strings.Replace(query, "*", acl, 1)

			switch r.Method {
			case http.MethodPost:
				r.PostForm.Set("query", query)
			case http.MethodGet:
				vals := r.URL.Query()
				vals.Set("query", query)
				r.URL.RawQuery = vals.Encode()
			}
			valid = true
		}
		if valid {
			if r.Method == http.MethodPost {
				str := r.PostForm.Encode()
				r.Body = ioutil.NopCloser(strings.NewReader(str))
				r.Header.Set("Content-Length", strconv.Itoa(len(str)))
				r.ContentLength = int64(len(str))
			}
			next.ServeHTTP(w, r)
		} else {
			log.Printf("403: %s, %v", query, orgs)
			w.WriteHeader(403)
		}
	}
	return http.HandlerFunc(ourFunc)
}

func render(next http.Handler) http.Handler {
	ourFunc := func(w http.ResponseWriter, r *http.Request) {
		client := grafana.NewClient()
		client.Client().SetCookies(r.Cookies())
		orgs, err := client.GetUserOrgs()
		if err != nil {
			log.Printf("render %+v\n", err)
			w.WriteHeader(500)
			return
		}

		acl := getACL(orgs)
		r.ParseForm()
		targets := r.Form["target"]
		validTargets := make([]string, 0)
		for _, target := range targets {
			valid := false
			for _, org := range orgs {
				val := fmt.Sprintf("screeps.%s", org.Name)
				if strings.Contains(target, val) {
					valid = true
					break
				}
			}
			if !valid && strings.Contains(target, "screeps.*") {
				target = strings.Replace(target, "screeps.*", fmt.Sprintf("screeps.%s", acl), -1)
				valid = true
			}
			if valid {
				validTargets = append(validTargets, target)
			}
		}
		r.Form["target"] = validTargets
		body := r.Form.Encode()
		r.Body = ioutil.NopCloser(bytes.NewBufferString(body))
		r.ContentLength = int64(len(body))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(ourFunc)
}

func getACL(orgs []m.UserOrgDTO) string {
	var list []string
	for _, org := range orgs {
		list = append(list, org.Name)
	}
	s := strings.Join(list, ",")
	s = fmt.Sprintf("{%s}", s)
	return s
}

func metricMap(list []string, base string) []byte {
	ret := make([]*graphiteMetric, 0, 0)
	for _, v := range list {
		id := v
		if base != "" {
			id = fmt.Sprintf("%s.%s", base, v)
		}
		obj := &graphiteMetric{}
		obj.AllowChildren = 1
		obj.Expandable = 1
		obj.ID = id
		obj.Leaf = 0
		obj.Text = v
		ret = append(ret, obj)
	}
	resp, _ := json.Marshal(ret)
	return resp
}

type graphiteMetric struct {
	AllowChildren int    `json:"allowChildren"`
	Expandable    int    `json:"expandable"`
	ID            string `json:"id"`
	Leaf          int    `json:"leaf"`
	Text          string `json:"text"`
}
