package splunk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Index represents a Splunk index with name and disabled fields.
type Index struct {
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
}

// GetIndexes retrieves paginated indexes from Splunk
func (c *Client) GetIndexes(ctx context.Context, count, offset int) ([]Index, int, error) {
	url := fmt.Sprintf("%s/services/data/indexes?output_mode=json&count=%d&offset=%d", c.BaseURL, count, offset)

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
			Name    string `json:"name"`
			Content struct {
				Disabled bool `json:"disabled"`
			} `json:"content"`
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

	indexes := make([]Index, len(result.Entry))
	for i, entry := range result.Entry {
		indexes[i] = Index{
			Name:     entry.Name,
			Disabled: entry.Content.Disabled,
		}
	}

	return indexes, result.Paging.Total, nil
}
