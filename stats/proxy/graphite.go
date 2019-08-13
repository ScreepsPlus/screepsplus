package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"gopkg.in/resty.v1"
)

type graphiteProxy struct {
	resty  *resty.Client
	router *mux.Router
	proxy  *httputil.ReverseProxy
}

type graphiteMetric struct {
	AllowChildren int    `json:"allowChildren"`
	Expandable    int    `json:"expandable"`
	ID            string `json:"id"`
	Leaf          int    `json:"leaf"`
	Text          string `json:"text"`
}

// NewGraphiteProxy creates a new auth proxy for Graphite
func NewGraphiteProxy(backendURL string) http.Handler {
	url, _ := url.Parse(backendURL)
	gp := graphiteProxy{
		resty:  resty.New(),
		router: mux.NewRouter(),
		proxy:  httputil.NewSingleHostReverseProxy(url),
	}
	gp.resty.SetHostURL(backendURL)
	gp.router.HandleFunc("/render", gp.RenderHandler)
	gp.router.HandleFunc("/metrics/find", gp.MetricsFindHandler)
	return gp
}

func (g graphiteProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.router.ServeHTTP(w, r)
}

func (g graphiteProxy) RenderHandler(w http.ResponseWriter, r *http.Request) {
	orgs := r.TLS.PeerCertificates[0].DNSNames
	acl := g.getACL(orgs)
	r.ParseForm()
	targets := r.Form["target"]
	validTargets := make([]string, 0)
	for _, target := range targets {
		valid := false
		for _, org := range orgs {
			val := fmt.Sprintf("screeps.%s", org)
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
	g.proxy.ServeHTTP(w, r)
}

func (g graphiteProxy) MetricsFindHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "*" || query == "screeps" {
		w.Header().Add("Content-Type", "application/json")
		metrics := g.metricMap([]string{"screeps"}, "")
		b, err := json.Marshal(metrics)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.WriteHeader(200)
		w.Write(b)
		return
	}

	orgs := r.TLS.PeerCertificates[0].DNSNames

	valid := false
	for _, org := range orgs {
		val := fmt.Sprintf("screeps.%s", org)
		if strings.HasPrefix(query, val) {
			valid = true
			break
		}
	}
	if query == "screeps.*" || strings.HasPrefix(query, "screeps.*.") {
		acl := g.getACL(orgs)
		query = strings.Replace(query, "*", acl, 1)
		vals := r.URL.Query()
		vals.Set("query", query)
		r.URL.RawQuery = vals.Encode()
		valid = true
	}
	if valid {
		g.proxy.ServeHTTP(w, r)
	} else {
		w.WriteHeader(403)
	}
}

func (g *graphiteProxy) getACL(orgs []string) string {
	s := strings.Join(orgs, ",")
	s = fmt.Sprintf("{%s}", s)
	return s
}

func (g *graphiteProxy) metricMap(list []string, base string) []graphiteMetric {
	ret := make([]graphiteMetric, 0, 0)
	for _, v := range list {
		id := v
		if base != "" {
			id = fmt.Sprintf("%s.%s", base, v)
		}
		obj := graphiteMetric{
			AllowChildren: 1,
			Expandable:    1,
			ID:            id,
			Leaf:          0,
			Text:          v,
		}
		ret = append(ret, obj)
	}
	return ret
}
