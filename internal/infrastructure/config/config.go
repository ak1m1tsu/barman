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
	Discord  DiscordConfig  `yaml:"discord"`
	Database DatabaseConfig `yaml:"database"`
	Timeouts TimeoutsConfig `yaml:"timeouts"`
}

type DiscordConfig struct {
	Token    string   `yaml:"token"`
	AppID    string   `yaml:"app_id"`
	GuildID  string   `yaml:"guild_id"`
	Prefix   string   `yaml:"prefix"`
	OwnerIDs []string `yaml:"owner_ids"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

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
