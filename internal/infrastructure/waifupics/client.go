package waifupics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.waifu.pics/nsfw/"

// nsfwEndpoints maps reaction names to waifu.pics NSFW endpoint slugs.
var nsfwEndpoints = map[string]string{
	"myatniy": "blowjob",
}

// Client fetches NSFW reaction GIFs from api.waifu.pics.
type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 5 * time.Second}}
}

// Fetch returns a GIF URL for the given reaction type.
// The reaction name is mapped to the corresponding waifu.pics endpoint.
func (c *Client) Fetch(ctx context.Context, reaction string) (string, error) {
	endpoint, ok := nsfwEndpoints[reaction]
	if !ok {
		return "", fmt.Errorf("waifupics: unsupported reaction %q", reaction)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("waifupics: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("waifupics: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("waifupics: status %d", resp.StatusCode)
	}

	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("waifupics: decode: %w", err)
	}

	if body.URL == "" {
		return "", fmt.Errorf("waifupics: empty url for %q", reaction)
	}

	return body.URL, nil
}
