// internal/config/atproto.go
package config

import (
	"os"

	"github.com/referendumApp/statusphere-example-app-go/internal/atproto"
)

// GetATProtoConfig returns the AT Protocol configuration from environment variables
func GetATProtoConfig() atproto.Config {
	return atproto.Config{
		PdsHost: getEnvWithDefault("ATPROTO_PDS_HOST", "https://bsky.social"),
	}
}

// getEnvWithDefault gets an environment variable or returns the default if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}