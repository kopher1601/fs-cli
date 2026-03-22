package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration.
type Config struct {
	APIKey          string
	RefreshInterval int // seconds
}

// Load reads configuration from environment variables and flags.
func Load() (*Config, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	return &Config{
		APIKey:          apiKey,
		RefreshInterval: 10,
	}, nil
}
