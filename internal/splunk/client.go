package splunk

import (
	"net/http"
	"time"
)

// Client represents a Splunk client for all tools
// Credentials are passed directly, not read from environment variables
// Add HTTP client for reuse and timeouts
type Client struct {
	BaseURL   string
	AuthToken string
	HTTP      *http.Client
}

// NewClient creates a new Splunk client using provided credentials
func NewClient(baseURL, authToken string) *Client {
	return &Client{
		BaseURL:   baseURL,
		AuthToken: authToken,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
