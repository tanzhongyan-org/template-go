package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	if err := os.Unsetenv("APP_NAME"); err != nil {
		t.Fatalf("failed to unset APP_NAME: %v", err)
	}
	if err := os.Unsetenv("APP_ENV"); err != nil {
		t.Fatalf("failed to unset APP_ENV: %v", err)
	}
	if err := os.Unsetenv("PORT"); err != nil {
		t.Fatalf("failed to unset PORT: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppName != "template-go" {
		t.Errorf("expected template-go, got %s", cfg.AppName)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected development, got %s", cfg.AppEnv)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected 8080, got %d", cfg.Port)
	}
}

func TestLoad_CustomEnv(t *testing.T) {
	t.Setenv("APP_NAME", "custom-service")
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "9000")
	t.Setenv("TRUST_IDENTITY_HEADER", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppName != "custom-service" || cfg.Port != 9000 || cfg.AppEnv != "production" {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("PORT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid PORT, got nil")
	}
}

func TestLoad_TrustIdentityHeader(t *testing.T) {
	t.Setenv("TRUST_IDENTITY_HEADER", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TrustIdentityHeader {
		t.Error("expected TrustIdentityHeader to default to false")
	}

	t.Setenv("TRUST_IDENTITY_HEADER", "true")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.TrustIdentityHeader {
		t.Error("expected TrustIdentityHeader to be true")
	}

	t.Setenv("TRUST_IDENTITY_HEADER", "maybe")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid TRUST_IDENTITY_HEADER")
	}
}

func TestLoad_DevIdentityRefusedOutsideDevelopment(t *testing.T) {
	for _, env := range []string{"production", "preview"} {
		t.Setenv("APP_ENV", env)
		t.Setenv("TRUST_IDENTITY_HEADER", "true")
		if _, err := Load(); err == nil {
			t.Errorf("expected boot to fail with TRUST_IDENTITY_HEADER=true and APP_ENV=%q", env)
		}
	}

	t.Setenv("APP_ENV", "production")
	t.Setenv("TRUST_IDENTITY_HEADER", "false")
	if _, err := Load(); err != nil {
		t.Errorf("expected boot to pass without the dev shortcut in production: %v", err)
	}
}

func TestLoad_JWTSettings(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("TRUST_IDENTITY_HEADER", "false")
	t.Setenv("AUTH_JWKS_URL", "https://auth.example.test/.well-known/jwks.json")
	t.Setenv("AUTH_ISSUER", "https://auth.example.test")
	t.Setenv("AUTH_AUDIENCE", "notes-api")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.JWTEnabled() {
		t.Error("expected JWTEnabled to be true")
	}

	t.Setenv("AUTH_AUDIENCE", "")
	if _, err := Load(); err == nil {
		t.Error("expected error when only some AUTH_ variables are set")
	}
}

func TestLoad_JWTDefaultsToOff(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.JWTEnabled() {
		t.Error("expected JWTEnabled to default to false")
	}
}
