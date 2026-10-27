package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("APP_NAME")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("PORT")

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
