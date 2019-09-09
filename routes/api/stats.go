package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/screepsplus/screepsplus/stats"
	"github.com/screepsplus/screepsplus/stats/graphite"
)

var formats = map[string]string{
	"application/json": "json",
	"text/grafana":     "grafana",
	"text/graphite":    "graphite",
}

func getStats(r *http.Request) ([]stats.Stat, string, error) {
	contentType := r.Header.Get("content-type")
	format, ok := formats[contentType]
	if !ok {
		return nil, "unsupported", fmt.Errorf("Unsupported format: %s", contentType)
	}
	switch format {
	case "json":
		format = "json"
		s, err := stats.FromJSON(r.Body)
		if err != nil {
			return nil, "json", err
		}
		return s, format, nil
	case "grafana":
		fallthrough
	case "graphite":
		scanner := bufio.NewScanner(r.Body)
		stats := make([]stats.Stat, 0)
		for scanner.Scan() {
			s, err := graphite.ParseStat(scanner.Text())
			if err != nil {
				continue
			}
			stats = append(stats, *s)
		}
		return stats, format, nil
	}
	return nil, "invalid", fmt.Errorf("Invalid stat type")
}

func statsSubmit() http.Handler {
	url := os.Getenv("GRAPHITE_URL")
	g := graphite.New(url, 4)
	ourFunc := func(w http.ResponseWriter, r *http.Request) {
		s, format, err := getStats(r)
		if err != nil {
			log.Printf("Error in stats submit: %v", err)
			badRequest(w)
			return
		}
		if err := g.Send(s); err != nil {
			log.Printf("Error in stats submit: %v", err)
			serverError(w)
			return
		}
		resp := apiResp{
			OK: 1,
			Extra: map[string]interface{}{
				"ts":     time.Now().Unix(),
				"format": format,
			},
		}
		if format == "grafana" {
			resp.Extra["message"] = "Warning: Format 'text/grafana' is deprecated, please use 'text/graphite' instead"
		}
		b, err := json.Marshal(resp)
		w.Write(b)
	}
	return http.HandlerFunc(ourFunc)
}
