package otakugifs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ak1m1tsu/barman/internal/infrastructure/otakugifs"
)

func newTestClient(t *testing.T, handler http.Handler) (*otakugifs.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return otakugifs.NewClientWithBaseURL(srv.URL + "?reaction="), srv
}

func TestClient_Fetch(t *testing.T) {
	ctx := context.Background()

	t.Run("returns url on success", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "hug", r.URL.Query().Get("reaction"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"url":"https://cdn.otakugifs.xyz/gifs/hug/1.gif"}`)) //nolint:errcheck
		}))

		url, err := c.Fetch(ctx, "hug")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.otakugifs.xyz/gifs/hug/1.gif", url)
	})

	t.Run("returns error on non-200 status", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		_, err := c.Fetch(ctx, "unknown")
		assert.Error(t, err)
	})

	t.Run("returns error on empty url in response", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"url":""}`)) //nolint:errcheck
		}))

		_, err := c.Fetch(ctx, "hug")
		assert.Error(t, err)
	})

	t.Run("returns error on malformed json", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`not json`)) //nolint:errcheck
		}))

		_, err := c.Fetch(ctx, "hug")
		assert.Error(t, err)
	})
}
