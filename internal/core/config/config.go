package config

import (
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
func Load() *Config {
	appName := getEnv("APP_NAME", "template-go")
	appEnv := getEnv("APP_ENV", "development")
	portStr := getEnv("PORT", "8080")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 8080
	}

	return &Config{
		AppName: appName,
		AppEnv:  appEnv,
		Port:    port,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
