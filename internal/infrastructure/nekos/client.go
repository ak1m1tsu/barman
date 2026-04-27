package nekos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://nekos.best/api/v2"

// Client fetches reaction GIFs from nekos.best.
type Client struct {
	http *http.Client
}

// NewClient returns a Client with a 5-second HTTP timeout.
func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 5 * time.Second}}
}

// Fetch returns a GIF URL for the given reaction type.
func (c *Client) Fetch(ctx context.Context, reaction string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", baseURL, reaction), nil)
	if err != nil {
		return "", fmt.Errorf("nekos: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("nekos: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nekos: status %d", resp.StatusCode)
	}

	var body struct {
		Results []struct {
			URL string `json:"url"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("nekos: decode: %w", err)
	}

	if len(body.Results) == 0 {
		return "", fmt.Errorf("nekos: empty results for %q", reaction)
	}

	return body.Results[0].URL, nil
}
