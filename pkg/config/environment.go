package config

import (
	"os"
)

// EnvOpt gets a configuration from the environment
// with a default fallback.
func EnvOpt(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
