package splunk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Macro represents a Splunk macro with relevant fields.
type Macro struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Disabled   bool   `json:"disabled"`
}

// GetMacros retrieves paginated macros from Splunk
func (c *Client) GetMacros(ctx context.Context, count, offset int) ([]Macro, int, error) {
	url := fmt.Sprintf("%s/services/data/macros?output_mode=json&count=%d&offset=%d", c.BaseURL, count, offset)

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
				Definition string `json:"definition"`
				Disabled   bool   `json:"disabled"`
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

	macros := make([]Macro, len(result.Entry))
	for i, entry := range result.Entry {
		macros[i] = Macro{
			Name:       entry.Name,
			Definition: entry.Content.Definition,
			Disabled:   entry.Content.Disabled,
		}
	}

	return macros, result.Paging.Total, nil
}
