package discord_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ak1m1tsu/barman/internal/infrastructure/discord"
)

func TestWebhookHook_Levels(t *testing.T) {
	h := discord.NewErrorWebhookHook("http://example.com")
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

	h := discord.NewErrorWebhookHook(srv.URL)

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
	h := discord.NewErrorWebhookHook("http://127.0.0.1:1") // nothing listening here

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

func TestActivityWebhookHook_Levels(t *testing.T) {
	h := discord.NewActivityWebhookHook("http://example.com")
	levels := h.Levels()
	assert.Contains(t, levels, logrus.InfoLevel)
	assert.Len(t, levels, 1)
}

func TestActivityWebhookHook_Fire_SkipsEntryWithoutNotifyField(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	h := discord.NewActivityWebhookHook(srv.URL)

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.InfoLevel,
		Message: "some infrastructure log without notify",
		Time:    time.Now(),
		Data:    logrus.Fields{"op": "app.New"},
	}

	err := h.Fire(entry)
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, int32(0), callCount.Load(), "no HTTP call expected for entries without notify field")
}

func TestActivityWebhookHook_Fire_SendsEmbedWhenNotifySet(t *testing.T) {
	var received []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	h := discord.NewActivityWebhookHook(srv.URL)

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.InfoLevel,
		Message: "command invoked",
		Time:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		Data: logrus.Fields{
			"command":  "/react hug",
			"guild_id": "999",
			"user_id":  "42",
			"notify":   true,
		},
	}

	err := h.Fire(entry)
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	require.NotEmpty(t, received)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(received, &payload))

	embeds, ok := payload["embeds"].([]any)
	require.True(t, ok)
	require.Len(t, embeds, 1)

	embed := embeds[0].(map[string]any)
	assert.Equal(t, "🟢 ACTION", embed["title"])
	assert.Equal(t, "command invoked", embed["description"])
	assert.Equal(t, float64(0x57F287), embed["color"])

	// notify field must be stripped from embed fields
	fields, _ := embed["fields"].([]any)
	for _, f := range fields {
		field := f.(map[string]any)
		assert.NotEqual(t, "notify", field["name"], "notify metadata field must not appear in embed")
	}

	// The three content fields should be present: command, guild_id, user_id
	assert.Len(t, fields, 3)
}

func TestActivityWebhookHook_Fire_ToleratesUnreachableURL(t *testing.T) {
	h := discord.NewActivityWebhookHook("http://127.0.0.1:1")

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.InfoLevel,
		Message: "command invoked",
		Time:    time.Now(),
		Data:    logrus.Fields{"notify": true},
	}

	err := h.Fire(entry)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
}
