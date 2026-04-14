package otakugifs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.otakugifs.xyz/gif?reaction="

// Client fetches reaction GIFs from api.otakugifs.xyz.
type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 5 * time.Second}}
}

// Fetch returns a GIF URL for the given reaction type.
func (c *Client) Fetch(ctx context.Context, reaction string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+reaction, nil)
	if err != nil {
		return "", fmt.Errorf("otakugifs: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("otakugifs: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("otakugifs: status %d", resp.StatusCode)
	}

	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("otakugifs: decode: %w", err)
	}

	if body.URL == "" {
		return "", fmt.Errorf("otakugifs: empty url for %q", reaction)
	}

	return body.URL, nil
}
