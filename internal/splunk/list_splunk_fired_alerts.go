package splunk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Flattened response for Splunk fired alerts
type FiredAlert struct {
	Name string `json:"name"`
}

// GetFiredAlerts retrieves paginated fired alerts from Splunk for the list_splunk_fired_alerts tool
func (c *Client) GetFiredAlerts(ctx context.Context, count, offset int) ([]FiredAlert, int, error) {
	url := fmt.Sprintf("%s/services/alerts/fired_alerts?output_mode=json&count=%d&offset=%d", c.BaseURL, count, offset)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Splunk API response
	var result struct {
		Entry []struct {
			Name string `json:"name"`
		} `json:"entry"`
		Paging struct {
			Total   int `json:"total"`
			PerPage int `json:"perPage"`
			Offset  int `json:"offset"`
		} `json:"paging"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	alerts := make([]FiredAlert, len(result.Entry))
	for i, entry := range result.Entry {
		alerts[i] = FiredAlert{
			Name: entry.Name,
		}
	}

	return alerts, result.Paging.Total, nil
}
