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
// Only Type, Text, and State are rendered by Discord clients for bot gateway presence;
// Rich Presence fields (images, party, timestamps) are not supported for bots.
type ActivityConfig struct {
	Type  string `yaml:"type"`  // playing | watching | listening | competing
	Text  string `yaml:"text"`  // main activity name (leave empty to disable)
	State string `yaml:"state"` // line shown below the activity name
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
