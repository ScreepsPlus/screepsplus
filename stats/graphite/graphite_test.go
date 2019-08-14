package graphite

import (
	"testing"

	"github.com/screepsplus/screepsplus/stats"
)

func TestSend(t *testing.T) {
	type args struct {
		stats []stats.Stat
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Send(tt.args.stats)
		})
	}
}

func Test_worker(t *testing.T) {
	type args struct {
		host string
		port int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker(tt.args.host, tt.args.port)
		})
	}
}
