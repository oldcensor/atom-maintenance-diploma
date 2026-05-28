package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TelemetryItem struct {
	EquipmentID  int64   `json:"equipment_id"`
	Name         string  `json:"name"`
	MeterType    string  `json:"meter_type"`
	CurrentValue float64 `json:"current_value"`
	Unit         string  `json:"unit"`
	Timestamp    string  `json:"timestamp"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) GetAll(ctx context.Context) ([]TelemetryItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/telemetry", nil)
	if err != nil {
		return nil, fmt.Errorf("simulator: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("simulator: get all: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("simulator: unexpected status %d", resp.StatusCode)
	}

	var items []TelemetryItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("simulator: decode: %w", err)
	}
	return items, nil
}

func (c *Client) GetByEquipmentID(ctx context.Context, id int64) (*TelemetryItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/v1/telemetry/%d", c.baseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("simulator: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("simulator: get by id %d: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("simulator: unexpected status %d", resp.StatusCode)
	}

	var item TelemetryItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("simulator: decode: %w", err)
	}
	return &item, nil
}
