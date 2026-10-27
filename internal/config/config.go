package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds runtime configuration values.
type Config struct {
	AppName string
	AppEnv  string
	Port    int
}

// Load validates and returns configuration from environment variables.
func Load() (*Config, error) {
	appName := getEnv("APP_NAME", "template-go")
	appEnv := getEnv("APP_ENV", "development")
	portStr := getEnv("PORT", "8080")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT %q: %w", portStr, err)
	}

	return &Config{
		AppName: appName,
		AppEnv:  appEnv,
		Port:    port,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
