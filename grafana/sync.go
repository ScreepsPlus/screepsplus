package grafana

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/grafana/grafana/pkg/components/simplejson"
	m "github.com/grafana/grafana/pkg/models"
)

// Sync syncs configuration data with grafana
type Sync struct {
	client      *Client
	grafanaURL  string
	influxdbURL string
	graphiteURL string
	ctx         context.Context
}

// NewSync creates a new Sync
func NewSync() *Sync {
	d := &Sync{
		client:      NewClient(),
		grafanaURL:  os.Getenv("GRAFANA_URL"),
		influxdbURL: os.Getenv("INFLUXDB_URL"),
		graphiteURL: os.Getenv("GRAPHITE_URL"),
	}
	return d
}

// Run starts the sync loop
func (d *Sync) Run(ctx context.Context) {
	d.ctx = ctx
	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()
	user := "admin"
	d.client.Client().OnBeforeRequest(func(client *resty.Client, req *resty.Request) error {
		req.SetHeader("X-Token-Sub", user)
		return nil
	})
	for {
		select {
		case <-ticker.C:
			user = "admin"
			orgs, err := d.client.GetOrgs()
			if err != nil {
				log.Printf("Sync Error: %v", err)
				continue
			}
			for _, org := range orgs {
				if org.Id == 1 {
					continue
				}
				if org.Name != strings.ToLower(org.Name) {
					newName := strings.ToLower(org.Name)
					log.Printf("Updating name for org %s to %s", org.Name, newName)
					if err := d.client.UpdateOrg(org.Id, m.UpdateOrgCommand{Name: newName}); err != nil {
						log.Printf("Error while updating org name: %v", err)
					}
					org.Name = newName
				}
				user = org.Name
				if ds, _ := d.client.GetDataSourceByName("ScreepsPlus-Graphite"); ds == nil {
					data := m.AddDataSourceCommand{
						Name:            "ScreepsPlus-Graphite",
						Type:            "graphite",
						Access:          "direct",
						Url:             d.grafanaURL,
						WithCredentials: true,
						JsonData:        &simplejson.Json{},
					}
					data.JsonData.Set("graphiteVersion", "1.1.x")
					if ds, err := d.client.AddDataSource(data); err != nil {
						log.Printf("DataSource add failed for %s %v\n", org.Name, err)
					} else {
						log.Printf("DataSource %s added for %s\n", ds.Name, org.Name)
					}
				}

				if ds, _ := d.client.GetDataSourceByName("ScreepsPlus-InfluxDB"); ds == nil {
					data := m.AddDataSourceCommand{
						Name:            "ScreepsPlus-InfluxDB",
						Type:            "influxdb",
						Access:          "proxy",
						Url:             d.influxdbURL,
						WithCredentials: true,
					}
					if ds, err := d.client.AddDataSource(data); err != nil {
						log.Printf("DataSource add failed for %s %v\n", org.Name, err)
					} else {
						log.Printf("DataSource %s added for %s\n", ds.Name, org.Name)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
