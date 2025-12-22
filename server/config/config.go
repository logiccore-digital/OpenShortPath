package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port                  int      `yaml:"port"`
	PostgresURI           string   `yaml:"postgres_uri"`
	SQLitePath            string   `yaml:"sqlite_path"`
	AvailableShortDomains []string `yaml:"available_short_domains"`
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{
		Port:                  3000,                       // default port
		SQLitePath:            "db.sqlite",                // default SQLite path
		AvailableShortDomains: []string{"localhost:3000"}, // default short domains
	}

	if configPath == "" {
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not specified in config
	if config.Port == 0 {
		config.Port = 3000
	}
	if config.SQLitePath == "" && config.PostgresURI == "" {
		config.SQLitePath = "db.sqlite"
	}
	if len(config.AvailableShortDomains) == 0 {
		config.AvailableShortDomains = []string{"localhost:3000"}
	}

	return config, nil
}
