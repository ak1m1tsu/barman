package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

// NotifyField is the logrus field key that marks an Info/Warn entry for delivery
// to the activity webhook. Handlers add `.WithField("notify", true)` to opt in.
// The field is stripped from the embed fields before sending.
const NotifyField = "notify"

type webhookPayload struct {
	Embeds []webhookEmbed `json:"embeds"`
}

type webhookEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color"`
	Fields      []webhookEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp"`
}

type webhookEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type levelMeta struct {
	title string
	color int
}

var errorLevelMeta = map[logrus.Level]levelMeta{
	logrus.ErrorLevel: {"🔴 ERROR", 0xED4245},
	logrus.FatalLevel: {"💀 FATAL", 0x992D22},
	logrus.PanicLevel: {"☠️ PANIC", 0x992D22},
}

var activityLevelMeta = map[logrus.Level]levelMeta{
	logrus.InfoLevel: {"🟢 ACTION", 0x57F287},
}

// WebhookHook is a logrus hook that sends log entries to a Discord webhook as
// an embed. Each Fire call dispatches asynchronously so it never blocks the
// logging caller.
//
// Use NewErrorWebhookHook for Error/Fatal/Panic entries, or
// NewActivityWebhookHook for user-action Info entries (requires the "notify"
// field to be set).
type WebhookHook struct {
	url           string
	client        *http.Client
	meta          map[logrus.Level]levelMeta
	requireNotify bool // when true, Fire only sends if entry.Data["notify"] is set
}

// NewErrorWebhookHook creates a hook that forwards Error/Fatal/Panic entries
// to the given Discord webhook URL.
func NewErrorWebhookHook(url string) *WebhookHook {
	return &WebhookHook{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
		meta:   errorLevelMeta,
	}
}

// NewActivityWebhookHook creates a hook that forwards Info entries to the given
// Discord webhook URL, but only when the entry contains the "notify" field.
func NewActivityWebhookHook(url string) *WebhookHook {
	return &WebhookHook{
		url:           url,
		client:        &http.Client{Timeout: 5 * time.Second},
		meta:          activityLevelMeta,
		requireNotify: true,
	}
}

// NewWebhookHook is kept for backwards compatibility; it behaves like NewErrorWebhookHook.
func NewWebhookHook(url string) *WebhookHook {
	return NewErrorWebhookHook(url)
}

func (h *WebhookHook) Levels() []logrus.Level {
	levels := make([]logrus.Level, 0, len(h.meta))
	for l := range h.meta {
		levels = append(levels, l)
	}
	return levels
}

func (h *WebhookHook) Fire(entry *logrus.Entry) error {
	if h.requireNotify {
		if _, ok := entry.Data[NotifyField]; !ok {
			return nil
		}
	}
	go h.send(entry)
	return nil
}

func (h *WebhookHook) send(entry *logrus.Entry) {
	meta := h.meta[entry.Level]

	fields := make([]webhookEmbedField, 0, len(entry.Data))
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		if k == NotifyField {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := entry.Data[k]
		var val string
		if err, ok := v.(error); ok {
			val = err.Error()
		} else {
			val = fmt.Sprintf("%v", v)
		}
		fields = append(fields, webhookEmbedField{
			Name:   k,
			Value:  val,
			Inline: true,
		})
	}

	payload := webhookPayload{
		Embeds: []webhookEmbed{
			{
				Title:       meta.title,
				Description: entry.Message,
				Color:       meta.color,
				Fields:      fields,
				Timestamp:   entry.Time.UTC().Format(time.RFC3339),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, h.url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil || resp == nil {
		return
	}
	_ = resp.Body.Close()
}
