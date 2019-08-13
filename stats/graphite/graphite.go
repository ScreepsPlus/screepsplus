package graphite

import (
	"fmt"
	"github.com/screepsplus/screepsplus/pkg/stats"
	"os"
)

var queue chan []stats.Stat 
func Send(stats []stats.Stat) {
	queue <- stats

}

func Init() {
	queue := make(chan []stats.Stat, 100)
	host := "localhost"
	port := 2003
	if v, ok := os.LookupEnv("GRAPHITE_HOST"); ok {
		host = v
	}
	if v, ok := os.LookupEnv("GRAPHITE_PORT"); ok {
		port = v
	}
	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	
}
