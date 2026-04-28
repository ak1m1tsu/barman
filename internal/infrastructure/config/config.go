package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultHandlerTimeout is used when timeouts.handler is not set in config.
const DefaultHandlerTimeout = 10 * time.Second

// Config holds all application configuration loaded from a YAML file.
type Config struct {
	Discord       DiscordConfig       `yaml:"discord"`
	Database      DatabaseConfig      `yaml:"database"`
	Timeouts      TimeoutsConfig      `yaml:"timeouts"`
	Notifications NotificationsConfig `yaml:"notifications"`
}

// NotificationsConfig holds the optional Discord webhook URL for bot notifications.
// Leave WebhookURL empty to disable all webhook delivery.
type NotificationsConfig struct {
	WebhookURL string `yaml:"webhook_url"`
}

// DiscordConfig holds the Discord bot credentials and guild-specific settings.
// GuildID may be empty, in which case commands are registered globally.
type DiscordConfig struct {
	Token    string         `yaml:"token"`
	AppID    string         `yaml:"app_id"`
	GuildID  string         `yaml:"guild_id"`
	Prefix   string         `yaml:"prefix"`
	OwnerIDs []string       `yaml:"owner_ids"`
	Activity ActivityConfig `yaml:"activity"`
}

// ActivityConfig holds Discord presence/activity settings.
// Image asset keys (LargeImage, SmallImage) must be uploaded in the Discord
// Developer Portal under Applications → Rich Presence → Art Assets first.
type ActivityConfig struct {
	Type           string           `yaml:"type"`             // playing | watching | listening | competing
	Text           string           `yaml:"text"`             // main activity name
	Details        string           `yaml:"details"`          // secondary line (may not show in all clients)
	State          string           `yaml:"state"`            // tertiary line
	LargeImage     string           `yaml:"large_image"`      // asset key from Discord Dev Portal
	LargeImageText string           `yaml:"large_image_text"` // tooltip for large image
	SmallImage     string           `yaml:"small_image"`      // asset key from Discord Dev Portal
	SmallImageText string           `yaml:"small_image_text"` // tooltip for small image
	Timestamps     TimestampsConfig `yaml:"timestamps"`
	Party          PartyConfig      `yaml:"party"`
}

// TimestampsConfig controls the rich-presence timestamp bar shown under the activity.
// StartNow causes the start timestamp to be set to the moment the bot connects,
// displaying elapsed time ("01:23 elapsed"). End is a fixed Unix timestamp in
// seconds; when set, Discord shows a countdown ("04:37 left").
type TimestampsConfig struct {
	StartNow bool  `yaml:"start_now"` // use bot startup time as start
	End      int64 `yaml:"end"`       // Unix seconds (0 = disabled)
}

// PartyConfig populates the Discord rich-presence party field.
// When MaxSize > 0 the client renders "CurrentSize of MaxSize" next to the activity.
// ID is an opaque string used by Discord to group party members; leave empty if unused.
type PartyConfig struct {
	ID          string `yaml:"id"`
	CurrentSize int    `yaml:"current_size"`
	MaxSize     int    `yaml:"max_size"`
}

// DatabaseConfig holds the file-system path of the SQLite database.
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// TimeoutsConfig holds per-handler context timeout durations.
// Use DefaultHandlerTimeout when Handler is zero.
type TimeoutsConfig struct {
	Handler time.Duration `yaml:"handler"`
}

// Load reads and parses the YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
