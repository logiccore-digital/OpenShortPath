package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type JWT struct {
	Algorithm  string `yaml:"algorithm"`   // "HS256" or "RS256"
	SecretKey  string `yaml:"secret_key"`  // Secret key for HS256 (optional if using RS256)
	PublicKey  string `yaml:"public_key"`  // Public key for RS256 in PEM format (optional if using HS256)
	PrivateKey string `yaml:"private_key"` // Private key for RS256 in PEM format (required for signing when using RS256)
}

type Config struct {
	Port                  int      `yaml:"port"`
	PostgresURI           string   `yaml:"postgres_uri"`
	SQLitePath            string   `yaml:"sqlite_path"`
	AvailableShortDomains []string `yaml:"available_short_domains"`
	AuthProvider          string   `yaml:"auth_provider"` // "external_jwt" or "local"
	JWT                   *JWT     `yaml:"jwt,omitempty"`
	AdminPassword         string   `yaml:"admin_password"`           // Super long password for administrative purposes
	DashboardDevServerURL string   `yaml:"dashboard_dev_server_url"` // URL for dashboard dev server (optional, for development)
	LandingDevServerURL   string   `yaml:"landing_dev_server_url"`  // URL for landing page dev server (optional, for development)
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

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Auth provider is required and must be either "local" or "external_jwt"
	if c.AuthProvider == "" {
		return fmt.Errorf("auth_provider is required (must be 'local' or 'external_jwt')")
	}
	if c.AuthProvider != "local" && c.AuthProvider != "external_jwt" {
		return fmt.Errorf("invalid auth_provider: %s (must be 'local' or 'external_jwt')", c.AuthProvider)
	}

	// If auth_provider is "local", JWT config must be provided
	if c.AuthProvider == "local" {
		if c.JWT == nil {
			return fmt.Errorf("JWT config is required when auth_provider is 'local'")
		}

		// Validate algorithm-specific requirements
		switch c.JWT.Algorithm {
		case "HS256":
			if c.JWT.SecretKey == "" {
				return fmt.Errorf("secret_key is required when using HS256 algorithm")
			}
		case "RS256":
			if c.JWT.PrivateKey == "" {
				return fmt.Errorf("private_key is required when using RS256 algorithm for local auth")
			}
			if c.JWT.PublicKey == "" {
				return fmt.Errorf("public_key is required when using RS256 algorithm")
			}
		default:
			if c.JWT.Algorithm != "" {
				return fmt.Errorf("unsupported JWT algorithm: %s (must be HS256 or RS256)", c.JWT.Algorithm)
			}
		}
	}

	// If JWT config is provided, validate algorithm requirements
	if c.JWT != nil && c.JWT.Algorithm != "" {
		switch c.JWT.Algorithm {
		case "HS256":
			if c.JWT.SecretKey == "" {
				return fmt.Errorf("secret_key is required when using HS256 algorithm")
			}
		case "RS256":
			if c.JWT.PublicKey == "" {
				return fmt.Errorf("public_key is required when using RS256 algorithm")
			}
			// Private key is only required for local auth (signing)
			if c.AuthProvider == "local" && c.JWT.PrivateKey == "" {
				return fmt.Errorf("private_key is required when using RS256 algorithm for local auth")
			}
		default:
			return fmt.Errorf("unsupported JWT algorithm: %s (must be HS256 or RS256)", c.JWT.Algorithm)
		}
	}

	return nil
}
