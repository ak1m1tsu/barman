package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

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

var levelMeta = map[logrus.Level]struct {
	title string
	color int
}{
	logrus.ErrorLevel: {"🔴 ERROR", 0xED4245},
	logrus.FatalLevel: {"💀 FATAL", 0x992D22},
	logrus.PanicLevel: {"☠️ PANIC", 0x992D22},
}

// WebhookHook is a logrus hook that sends Error/Fatal/Panic entries to a
// Discord webhook as an embed. Each Fire call dispatches asynchronously so it
// never blocks the logging caller.
type WebhookHook struct {
	url    string
	client *http.Client
}

// NewWebhookHook creates a WebhookHook that POSTs to the given Discord webhook URL.
func NewWebhookHook(url string) *WebhookHook {
	return &WebhookHook{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (h *WebhookHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

func (h *WebhookHook) Fire(entry *logrus.Entry) error {
	go h.send(entry)
	return nil
}

func (h *WebhookHook) send(entry *logrus.Entry) {
	meta := levelMeta[entry.Level]

	fields := make([]webhookEmbedField, 0, len(entry.Data))
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
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

	resp, err := h.client.Post(h.url, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil || resp == nil {
		return
	}
	resp.Body.Close() //nolint:errcheck
}
