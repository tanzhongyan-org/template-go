package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tanzhongyan-org/template-go/internal/core/config"
	"github.com/tanzhongyan-org/template-go/internal/domain/health"
)

func TestHealthHandler(t *testing.T) {
	cfg := &config.Config{
		AppName: "template-go",
		AppEnv:  "test",
		Port:    8080,
	}

	handler := NewHealthHandler(cfg)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var res health.Status
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Status != "healthy" {
		t.Errorf("expected healthy, got %s", res.Status)
	}
	if res.AppName != "template-go" {
		t.Errorf("expected template-go, got %s", res.AppName)
	}
}
