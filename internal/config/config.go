package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	Host        string
	Port        int
	Debug       bool
	PublicURL   string

	// Database
	DBPath      string

	// Auth
	CookieSecret string

	// Environment
	Environment string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Set defaults and override with environment variables
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT value: %w", err)
	}

	cfg := &Config{
		Host:         getEnv("HOST", "127.0.0.1"),
		Port:         port,
		Debug:        getEnv("DEBUG", "false") == "true",
		PublicURL:    getEnv("PUBLIC_URL", ""),
		DBPath:       getEnv("DB_PATH", "./statusphere.db"),
		CookieSecret: getEnv("COOKIE_SECRET", ""),
		Environment:  getEnv("NODE_ENV", "development"),
	}

	// Validate required configuration
	if cfg.CookieSecret == "" {
		return nil, fmt.Errorf("COOKIE_SECRET environment variable is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}