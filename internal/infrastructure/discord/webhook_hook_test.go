package discord_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ak1m1tsu/barman/internal/infrastructure/discord"
)

func TestWebhookHook_Levels(t *testing.T) {
	h := discord.NewWebhookHook("http://example.com")
	levels := h.Levels()
	assert.Contains(t, levels, logrus.ErrorLevel)
	assert.Contains(t, levels, logrus.FatalLevel)
	assert.Contains(t, levels, logrus.PanicLevel)
	assert.Len(t, levels, 3)
}

func TestWebhookHook_Fire_SendsEmbed(t *testing.T) {
	var received []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	h := discord.NewWebhookHook(srv.URL)

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.ErrorLevel,
		Message: "something went wrong",
		Time:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		Data: logrus.Fields{
			"guild_id": "123",
			"error":    errors.New("db connection lost"),
		},
	}

	err := h.Fire(entry)
	require.NoError(t, err)

	// Fire is async — give it a moment to complete.
	time.Sleep(200 * time.Millisecond)

	require.NotEmpty(t, received)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(received, &payload))

	embeds, ok := payload["embeds"].([]any)
	require.True(t, ok)
	require.Len(t, embeds, 1)

	embed := embeds[0].(map[string]any)
	assert.Equal(t, "🔴 ERROR", embed["title"])
	assert.Equal(t, "something went wrong", embed["description"])
	assert.Equal(t, float64(0xED4245), embed["color"])

	fields := embed["fields"].([]any)
	require.Len(t, fields, 2)

	// Fields are sorted alphabetically: error, guild_id.
	f0 := fields[0].(map[string]any)
	assert.Equal(t, "error", f0["name"])
	assert.Equal(t, "db connection lost", f0["value"])

	f1 := fields[1].(map[string]any)
	assert.Equal(t, "guild_id", f1["name"])
	assert.Equal(t, "123", f1["value"])
}

func TestWebhookHook_Fire_ToleratesUnreachableURL(t *testing.T) {
	h := discord.NewWebhookHook("http://127.0.0.1:1") // nothing listening here

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.ErrorLevel,
		Message: "oops",
		Time:    time.Now(),
		Data:    logrus.Fields{},
	}

	// Must not panic or block.
	err := h.Fire(entry)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
}
