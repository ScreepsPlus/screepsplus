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

const datasourceVersion = 5

// Sync syncs configuration data with grafana
type Sync struct {
	client      *Client
	grafanaURL  string
	influxdbURL string
	graphiteURL string
	lokiURL     string
	ctx         context.Context
}

// NewSync creates a new Sync
func NewSync() *Sync {
	d := &Sync{
		client:      NewClient(),
		grafanaURL:  os.Getenv("GRAFANA_URL"),
		influxdbURL: os.Getenv("INFLUXDB_URL"),
		graphiteURL: os.Getenv("GRAPHITE_URL"),
		lokiURL:     os.Getenv("LOKI_URL"),
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
				d.fixupOrg(&org)
				user = org.Name
				dataSources := []m.DataSource{
					{
						Name:     "ScreepsPlus-Graphite",
						Type:     "graphite",
						Access:   "direct",
						Url:      d.grafanaURL,
						JsonData: &simplejson.Json{},
						Version:  datasourceVersion,
					},
					{
						Name:     "ScreepsPlus-InfluxDB",
						Type:     "influxdb",
						Access:   "direct",
						Url:      d.influxdbURL,
						JsonData: &simplejson.Json{},
					},
					{
						Name:     "ScreepsPlus-Loki",
						Type:     "loki",
						Access:   "proxy",
						Url:      d.lokiURL,
						JsonData: &simplejson.Json{},
					},
				}
				dataSources[0].JsonData.Set("graphiteVersion", "1.1.x")
				dataSources[0].JsonData.Set("oauthPassThru", true)
				dataSources[1].JsonData.Set("oauthPassThru", true)
				dataSources[2].JsonData.Set("oauthPassThru", true)

				d.syncDatasources(user, dataSources)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (d *Sync) fixupOrg(org *m.OrgDTO) {
	if org.Name != strings.ToLower(org.Name) {
		newName := strings.ToLower(org.Name)
		log.Printf("Updating name for org %s to %s", org.Name, newName)
		if err := d.client.UpdateOrg(org.Id, m.UpdateOrgCommand{Name: newName}); err != nil {
			log.Printf("Error while updating org name: %v", err)
		}
		org.Name = newName
	}
}

func (d *Sync) syncDatasources(user string, dataSources []m.DataSource) {
	for _, base := range dataSources {
		if ds, _ := d.client.GetDataSourceByName(base.Name); ds == nil {
			data := m.AddDataSourceCommand{
				Name:            base.Name,
				Type:            base.Type,
				Access:          base.Access,
				Url:             base.Url,
				ReadOnly:        true,
				WithCredentials: true,
			}
			if ds, err := d.client.AddDataSource(data); err != nil {
				log.Printf("DataSource add failed for %s %v\n", user, err)
			} else {
				log.Printf("DataSource %s added for %s (v%d)\n", ds.Name, user, ds.Version)
			}
		} else if ds.Version < base.Version {
			data := m.UpdateDataSourceCommand{
				Name:            base.Name,
				Type:            base.Type,
				Access:          base.Access,
				Url:             base.Url,
				ReadOnly:        true,
				WithCredentials: true,
			}
			if ds, err := d.client.UpdateDataSource(data); err != nil {
				log.Printf("DataSource update failed for %s %v\n", user, err)
			} else {
				log.Printf("DataSource %s updated for %s (%d => %d)\n", ds.Name, user, ds.Version, base.Version)
			}
		}
	}
}
