package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration loaded from a YAML file.
type Config struct {
	Discord  DiscordConfig  `yaml:"discord"`
	Database DatabaseConfig `yaml:"database"`
}

type DiscordConfig struct {
	Token   string `yaml:"token"`
	AppID   string `yaml:"app_id"`
	GuildID string `yaml:"guild_id"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
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
