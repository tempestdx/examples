package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tempestdx/examples/deps/dashboards/server/models"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) CreateDashboard(ctx context.Context, dashboard models.Dashboard) (*models.Dashboard, error) {
	url := fmt.Sprintf("%s/dashboard/create", c.BaseURL)

	jsonData, err := json.Marshal(dashboard)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create dashboard, status: %s", resp.Status)
	}

	var createdDashboard models.Dashboard
	if err := json.NewDecoder(resp.Body).Decode(&createdDashboard); err != nil {
		return nil, err
	}

	return &createdDashboard, nil
}

func (c *Client) GetDashboard(ctx context.Context, id string) (*models.Dashboard, error) {
	url := fmt.Sprintf("%s/dashboard/get?id=%s", c.BaseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get dashboard, status: %s", resp.Status)
	}

	var dashboard models.Dashboard
	if err := json.NewDecoder(resp.Body).Decode(&dashboard); err != nil {
		return nil, err
	}

	return &dashboard, nil
}

func (c *Client) UpdateDashboard(ctx context.Context, id string, dashboard models.Dashboard) (*models.Dashboard, error) {
	url := fmt.Sprintf("%s/dashboard/update?id=%s", c.BaseURL, id)

	jsonData, err := json.Marshal(dashboard)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update dashboard, status: %s", resp.Status)
	}

	var updatedDashboard models.Dashboard
	if err := json.NewDecoder(resp.Body).Decode(&updatedDashboard); err != nil {
		return nil, err
	}

	return &updatedDashboard, nil
}

func (c *Client) DeleteDashboard(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/dashboard/delete?id=%s", c.BaseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete dashboard, status: %s", resp.Status)
	}

	return nil
}

func (c *Client) ListDashboards(ctx context.Context, next string) (*models.DashboardList, error) {
	url := fmt.Sprintf("%s/dashboard/list?next=%s", c.BaseURL, next)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list dashboards, status: %s", resp.Status)
	}

	var dashboards models.DashboardList
	if err := json.NewDecoder(resp.Body).Decode(&dashboards); err != nil {
		return nil, err
	}

	return &dashboards, nil
}

func (c *Client) Healthz(ctx context.Context) error {
	url := fmt.Sprintf("%s/healthz", c.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server is not healthy, status: %s", resp.Status)
	}

	return nil
}
