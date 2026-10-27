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
