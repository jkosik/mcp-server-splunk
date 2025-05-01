package splunk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// FiredAlert represents a fired alert from Splunk audit logs
type FiredAlert struct {
	Time       string `json:"_time"`
	SearchName string `json:"ss_name"`
}

// GetFiredAlerts retrieves fired alerts from Splunk using search/jobs/export
func (c *Client) GetFiredAlerts(ctx context.Context, count, offset int, ssName, earliest string) ([]FiredAlert, int, error) {
	// First get total count
	totalSpl := fmt.Sprintf("search index=_audit action=alert_fired ss_name=\"%s\" earliest=%s | stats count",
		ssName, earliest)

	total, err := c.getCount(ctx, totalSpl)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Then get the actual results with pagination
	spl := fmt.Sprintf("search index=_audit action=alert_fired ss_name=\"%s\" earliest=%s | table _time ss_name | head %d | tail %d",
		ssName, earliest, offset+count, count)

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
	var alerts []FiredAlert

	for {
		var row struct {
			Offset  int  `json:"offset"`
			LastRow bool `json:"lastrow"`
			Result  struct {
				Time       string `json:"_time"`
				SearchName string `json:"ss_name"`
			} `json:"result"`
		}
		if err := dec.Decode(&row); err != nil {
			if err == io.EOF {
				break
			}
			return nil, 0, fmt.Errorf("failed to decode response: %w", err)
		}

		alerts = append(alerts, FiredAlert{
			Time:       row.Result.Time,
			SearchName: row.Result.SearchName,
		})
	}

	return alerts, total, nil
}

// getCount executes a count query and returns the result
func (c *Client) getCount(ctx context.Context, spl string) (int, error) {
	endpoint := fmt.Sprintf("%s/services/search/jobs/export", c.BaseURL)
	form := url.Values{}
	form.Set("search", spl)
	form.Set("output_mode", "json")

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the count result - we get two rows:
	// 1. Preview count (preview: true)
	// 2. Final count (preview: false)
	dec := json.NewDecoder(resp.Body)
	var finalCount int

	for {
		var result struct {
			Preview bool `json:"preview"`
			Result  struct {
				Count string `json:"count"`
			} `json:"result"`
		}
		if err := dec.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return 0, fmt.Errorf("failed to decode count response: %w", err)
		}

		// Skip preview results, only use the final count
		if !result.Preview {
			count, err := strconv.Atoi(result.Result.Count)
			if err != nil {
				return 0, fmt.Errorf("failed to parse count: %w", err)
			}
			finalCount = count
		}
	}

	return finalCount, nil
}
