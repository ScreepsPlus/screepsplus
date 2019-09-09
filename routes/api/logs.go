package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-resty/resty/v2"
	"github.com/grafana/loki/pkg/logproto"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
)

type rawScreepsLogData struct {
	Messages struct {
		Log     []string `json:"log"`
		Results []string `json:"results"`
	} `json:"messages"`
	Error  string            `json:"error"`
	Shard  string            `json:"shard"`
	Server string            `json:"server"`
	Labels map[string]string `json:"labels"`
}

func rawScreepsLogs() http.Handler {
	url := os.Getenv("LOKI_URL")
	client := resty.New().
		SetHostURL(url).
		SetRetryCount(5).
		SetRetryWaitTime(10 * time.Second).
		SetRetryMaxWaitTime(30 * time.Second)
	ourFunc := func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*jwt.Token)
		raw := rawScreepsLogData{}
		json.NewDecoder(r.Body).Decode(&raw)
		ts := time.Now()
		req := logproto.PushRequest{
			Streams: make([]*logproto.Stream, 3),
		}
		labels := map[string]string{}
		claims := user.Claims.(jwt.MapClaims)
		labels["user"] = claims["sub"].(string)
		if raw.Server != "" {
			labels["server"] = raw.Server
		}
		if raw.Shard != "" {
			labels["shard"] = raw.Shard
		}
		for k, v := range raw.Labels {
			labels[k] = v
		}
		ts, req.Streams[0] = streamFromArr(raw.Messages.Log, ts)
		labels["type"] = "log"
		req.Streams[0].Labels = formatLabels(labels)

		ts, req.Streams[1] = streamFromArr(raw.Messages.Results, ts)
		labels["type"] = "result"

		req.Streams[0].Labels = formatLabels(labels)
		req.Streams[2] = &logproto.Stream{
			Entries: make([]logproto.Entry, 0),
		}
		labels["type"] = "error"
		req.Streams[2].Labels = formatLabels(labels)

		if raw.Error != "" {
			entry := logproto.Entry{
				Timestamp: ts,
				Line:      raw.Error,
			}
			ts = ts.Add(1 * time.Nanosecond)
			req.Streams[2].Entries = append(req.Streams[2].Entries, entry)
		}

		buf, err := proto.Marshal(&req)
		if err != nil {
			serverError(w)
			return
		}
		buf = snappy.Encode(nil, buf)
		res, err := client.R().
			SetHeader("content-type", "application/x-protobuf").
			SetBody(buf).
			Post("/api/prom/push")
		if res.IsError() {
			serverError(w)
			return
		}
		resp := apiResp{
			OK: 1,
			Extra: map[string]interface{}{
				"ts": time.Now().Unix(),
			},
		}
		b, err := json.Marshal(resp)
		w.Write(b)
	}
	return http.HandlerFunc(ourFunc)
}

func formatLabels(labels map[string]string) string {
	pairs := make([]string, len(labels))
	i := 0
	for label, value := range labels {
		pairs[i] = fmt.Sprintf("%s=\"%s\"", label, value)
		i++
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

func streamFromArr(arr []string, ts time.Time) (time.Time, *logproto.Stream) {
	str := &logproto.Stream{
		Entries: make([]logproto.Entry, len(arr)),
	}
	for i, log := range arr {
		str.Entries[i] = logproto.Entry{
			Timestamp: ts,
			Line:      log,
		}
		ts = ts.Add(1 * time.Nanosecond)
	}
	return ts, str
}
