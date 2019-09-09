package graphite

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/screepsplus/screepsplus/stats"

	"github.com/go-resty/resty/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	queueLen = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "stats",
		Subsystem: "graphite",
		Name:      "queue_len",
	})
	queueTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "stats",
		Subsystem: "graphite",
		Name:      "queue_total",
		Help:      "Total number of stats queued",
	})
	successCnt = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "stats",
		Subsystem: "graphite",
		Name:      "send_success",
	})
	failureCnt = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "stats",
		Subsystem: "graphite",
		Name:      "send_failures",
	})
)

// Graphite manages graphite connection
type Graphite struct {
	url    string
	queue  chan []stats.Stat
	client *resty.Client
}

// New creates a new Graphite instance
func New(url string, workers int) *Graphite {
	g := &Graphite{
		queue: make(chan []stats.Stat, 100),
		url:   url,
	}
	g.client = resty.New().
		SetHostURL(url).
		SetRetryCount(5).
		SetRetryWaitTime(10 * time.Second).
		SetRetryMaxWaitTime(30 * time.Second)
	go g.worker()
	return g
}

// Send queues stats for sending
func (g *Graphite) Send(stats []stats.Stat) error {
	g.queue <- stats
	queueLen.Inc()
	queueTotal.Add(float64(len(stats)))
	return nil
}

func (g *Graphite) worker() {
	for list := range g.queue {
		queueLen.Dec()
		ts := time.Now().Unix()
		bodyArr := make([]string, len(list))
		for i, stat := range list {
			bodyArr[i] = fmt.Sprintf("%s %.3f %d\n", stat.Key, stat.Value, ts)
		}
		body := strings.Join(bodyArr, "")
		_, err := g.client.R().
			SetHeader("Content-Type", "text/plain").
			SetBody(body).
			Post("/")
		if err != nil {
			log.Printf("Error sending stats %v", err)
			failureCnt.Inc()
		} else {
			successCnt.Inc()
		}
	}
}

// ParseStat parses a Graphite plaintext format stat
// NOTE: This strips the timestamp
func ParseStat(s string) (*stats.Stat, error) {
	parts := strings.Split(s, " ")
	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}
	return &stats.Stat{
		Key:   parts[0],
		Value: value,
	}, nil
}
