package splunk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Alert represents an alert definition from Splunk
// Only includes alerts with actions, and supports title filtering
// Fields: title, search, alert_type, actions, disabled, description
type Alert struct {
	Title       string `json:"title"`
	Search      string `json:"search"`
	AlertType   string `json:"alert_type"`
	Actions     string `json:"actions"`
	Disabled    bool   `json:"disabled"`
	Description string `json:"description"`
}

// GetAlerts retrieves paginated alerts from Splunk using SPL, with optional case-insensitive title filter
func (c *Client) GetAlerts(ctx context.Context, count, offset int, title string) ([]Alert, int, error) {
	// Build SPL
	spl := "| rest /services/saved/searches | search actions!=\"\" "
	if title != "" {
		title = strings.ToLower(title)
		spl += fmt.Sprintf("| where like(lower(title), \"%%%s%%\") ", title)
	}
	spl += "| table title search alert_type actions disabled description"

	// Prepare request to /services/search/jobs/export
	endpoint := fmt.Sprintf("%s/services/search/jobs/export", c.BaseURL)
	form := url.Values{}
	form.Set("search", spl)
	form.Set("output_mode", "json")

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse streaming JSON results
	dec := json.NewDecoder(resp.Body)
	var alerts []Alert
	for {
		var row map[string]interface{}
		if err := dec.Decode(&row); err != nil {
			if err == io.EOF {
				break
			}
			return nil, 0, fmt.Errorf("failed to decode response: %w", err)
		}
		// Each row is a result with fields under the 'result' key
		if result, ok := row["result"].(map[string]interface{}); ok {
			alert := Alert{
				Title:       getString(result, "title"),
				Search:      getString(result, "search"),
				AlertType:   getString(result, "alert_type"),
				Actions:     getString(result, "actions"),
				Description: getString(result, "description"),
			}
			if disabled, ok := result["disabled"].(string); ok {
				alert.Disabled = (disabled == "1" || strings.ToLower(disabled) == "true")
			}
			alerts = append(alerts, alert)
		}
	}

	// Mimicking Pagination and total counting, since SPL queries returns everything.
	total := len(alerts)
	start := offset
	if start > total {
		start = total
	}
	end := start + count
	if end > total {
		end = total
	}
	// But we provide cursor only the "count" of results. The real total is however known to the MCP backend and provided properly.
	return alerts[start:end], total, nil
}
