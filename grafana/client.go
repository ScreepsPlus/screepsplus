package grafana

import (
	"os"
	"strconv"

	"github.com/go-resty/resty/v2"
	m "github.com/grafana/grafana/pkg/models"
)

// Error is a grafana API error
type Error struct {
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

// NewClient returns a new grafana client
func NewClient() *Client {
	c := resty.New()
	c.
		SetHostURL(os.Getenv("GRAFANA_URL")).
		SetError(Error{})
	return &Client{
		client: c,
	}
}

// Client is a Grafana API client
type Client struct {
	client *resty.Client
}

// Client returns the resty client for this instance
func (c *Client) Client() *resty.Client {
	return c.client
}

// GetUser GET /api/user
func (c *Client) GetUser() (*m.UserProfileDTO, error) {
	resp, err := c.client.R().
		SetResult(&m.UserProfileDTO{}).
		Get("/api/user")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(Error)
	}
	return resp.Result().(*m.UserProfileDTO), nil
}

// GetOrgs GET /api/orgs
func (c *Client) GetOrgs() ([]m.OrgDTO, error) {
	resp, err := c.client.R().
		SetResult([]m.OrgDTO{}).
		Get("/api/orgs")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(Error)
	}
	return resp.Result().([]m.OrgDTO), nil
}

// GetUserOrgs GET /api/user/orgs
func (c *Client) GetUserOrgs() ([]m.UserOrgDTO, error) {
	resp, err := c.client.R().
		SetResult([]m.UserOrgDTO{}).
		Get("/api/user/orgs")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(Error)
	}
	return resp.Result().([]m.UserOrgDTO), nil
}

// UpdateOrg PUT /api/orgs/{orgID}
func (c *Client) UpdateOrg(orgID int64, data m.UpdateOrgCommand) error {
	resp, err := c.client.R().
		SetPathParams(map[string]string{
			"orgID": strconv.Itoa(int(orgID)),
		}).
		SetBody(data).
		SetContentLength(true).
		Put("/api/orgs/{orgID}")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.Error().(Error)
	}
	return nil
}

// GetDataSourceByName GET /api/datasources/name/{name}
func (c *Client) GetDataSourceByName(name string) (*m.DataSource, error) {
	resp, err := c.client.R().
		SetPathParams(map[string]string{
			"name": name,
		}).
		SetResult(&m.DataSource{}).
		Get("/api/datasources/name/{name}")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(Error)
	}
	return resp.Result().(*m.DataSource), nil
}

// AddDataSource POST /api/datasources
func (c *Client) AddDataSource(data m.AddDataSourceCommand) (*m.DataSource, error) {
	resp, err := c.client.R().
		SetResult(&m.DataSource{}).
		SetBody(data).
		Post("/api/datasources")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(Error)
	}
	return resp.Result().(*m.DataSource), nil
}

// UpdateDataSource PUT /api/datasources
func (c *Client) UpdateDataSource(data m.UpdateDataSourceCommand) (*m.DataSource, error) {
	resp, err := c.client.R().
		SetResult(&m.DataSource{}).
		SetBody(data).
		Put("/api/datasources")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(Error)
	}
	return resp.Result().(*m.DataSource), nil
}
