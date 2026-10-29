package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds runtime configuration values.
type Config struct {
	AppName     string
	AppEnv      string
	Port        int
	DatabaseURL string
	// TrustIdentityHeader enables the x-user-id header as the caller identity.
	// TEMPLATE-TODO(auth): replace header identity with real authentication
	TrustIdentityHeader bool
	// JWKSURL, JWTIssuer and JWTAudience enable product JWT checks. When
	// JWKSURL is set, the notes routes accept a signed Bearer credential
	// verified against the JWKS document instead of the x-user-id header.
	JWKSURL     string
	JWTIssuer   string
	JWTAudience string
}

// JWTEnabled reports whether the product JWT path is configured.
func (c *Config) JWTEnabled() bool {
	return c.JWKSURL != ""
}

// Load validates and returns configuration from environment variables.
func Load() (*Config, error) {
	appName := getEnv("APP_NAME", "template-go")
	appEnv := getEnv("APP_ENV", "development")
	portStr := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/template_go?sslmode=disable")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT %q: %w", portStr, err)
	}

	trustStr := getEnv("TRUST_IDENTITY_HEADER", "false")
	trust, err := strconv.ParseBool(trustStr)
	if err != nil {
		return nil, fmt.Errorf("invalid TRUST_IDENTITY_HEADER %q: %w", trustStr, err)
	}

	// The dev identity shortcut must never survive into a deployed
	// environment. Fail the boot here instead of serving with it on.
	if trust && (appEnv == "production" || appEnv == "preview") {
		return nil, fmt.Errorf("TRUST_IDENTITY_HEADER must be false when APP_ENV=%q", appEnv)
	}

	jwksURL := getEnv("AUTH_JWKS_URL", "")
	issuer := getEnv("AUTH_ISSUER", "")
	audience := getEnv("AUTH_AUDIENCE", "")
	if jwksURL == "" || issuer == "" || audience == "" {
		if jwksURL != "" || issuer != "" || audience != "" {
			return nil, fmt.Errorf("AUTH_JWKS_URL, AUTH_ISSUER and AUTH_AUDIENCE must be set together")
		}
	}

	return &Config{
		AppName:     appName,
		AppEnv:      appEnv,
		Port:        port,
		DatabaseURL: dbURL,

		TrustIdentityHeader: trust,
		JWKSURL:             jwksURL,
		JWTIssuer:           issuer,
		JWTAudience:         audience,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
