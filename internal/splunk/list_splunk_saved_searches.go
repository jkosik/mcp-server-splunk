package splunk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Flattened response provided by the MCP
type SavedSearch struct {
	Name        string `json:"name"`
	Search      string `json:"search"`
	Description string `json:"description"`
	Actions     string `json:"actions"`
	Disabled    bool   `json:"disabled"`
}

// GetSavedSearches retrieves paginated saved searches from Splunk
func (c *Client) GetSavedSearches(ctx context.Context, count, offset int) ([]SavedSearch, int, error) {
	url := fmt.Sprintf("%s/services/saved/searches?output_mode=json&count=%d&offset=%d", c.BaseURL, count, offset)

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
				Description string `json:"description"`
				Search      string `json:"search"`
				Actions     string `json:"actions"`
				Disabled    bool   `json:"disabled"`
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

	searches := make([]SavedSearch, len(result.Entry))
	for i, entry := range result.Entry {
		searches[i] = SavedSearch{
			Name:        entry.Name,
			Search:      entry.Content.Search,
			Description: entry.Content.Description,
			Actions:     entry.Content.Actions,
			Disabled:    entry.Content.Disabled,
		}
	}

	return searches, result.Paging.Total, nil
}
