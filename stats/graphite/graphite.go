package graphite

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/screepsplus/screepsplus/stats"

	"github.com/go-resty/resty/v2"
)

var queue chan []stats.Stat

// Send queues stats for sending
func Send(stats []stats.Stat) {
	queue <- stats
}

func init() {
	queue = make(chan []stats.Stat, 100)
	host := "localhost"
	port := 2007
	if v, ok := os.LookupEnv("GRAPHITE_HOST"); ok {
		host = v
	}
	if v, ok := os.LookupEnv("GRAPHITE_PORT"); ok {
		port, _ = strconv.Atoi(v)
	}
	for i := 0; i < 4; i++ {
		go worker(host, port)
	}
}

func worker(host string, port int) {
	url := fmt.Sprintf("http://%s:%d", host, port)
	client := resty.New().
		SetHostURL(url).
		SetRetryCount(5).
		SetRetryWaitTime(10 * time.Second).
		SetRetryMaxWaitTime(30 * time.Second)
	for list := range queue {
		ts := time.Now().Unix()
		bodyArr := make([]string, len(list))
		for i, stat := range list {
			bodyArr[i] = fmt.Sprintf("%s %.3f %d\n", stat.Key, stat.Value, ts)
		}
		body := strings.Join(bodyArr, "")
		_, err := client.R().
			SetHeader("Content-Type", "text/plain").
			SetBody(body).
			Post("/")
		if err != nil {
			log.Printf("Error sending stats %v", err)
		}
	}
}
