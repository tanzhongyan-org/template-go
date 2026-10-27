package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tanzhongyan-org/template-go/internal/config"
	"github.com/tanzhongyan-org/template-go/internal/db"
	"github.com/tanzhongyan-org/template-go/internal/handler"
)

func TestSetupRouter_Root(t *testing.T) {
	cfg := &config.Config{
		AppName: "template-go",
		AppEnv:  "test",
		Port:    8080,
	}

	router := SetupRouter(cfg, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestSetupRouter_Health(t *testing.T) {
	cfg := &config.Config{
		AppName: "template-go",
		AppEnv:  "test",
		Port:    8080,
	}

	router := SetupRouter(cfg, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var res handler.HealthStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Status != "healthy" {
		t.Errorf("expected healthy, got %s", res.Status)
	}
}

func newIdentityRouter(t *testing.T, trust bool) http.Handler {
	t.Helper()
	cfg := &config.Config{AppName: "template-go", AppEnv: "test", Port: 8080, TrustIdentityHeader: trust}
	// The identity check runs before any query, so a nil pool is never touched.
	return SetupRouter(cfg, db.New(nil))
}

func assertJSONError(t *testing.T, rr *httptest.ResponseRecorder, status int) {
	t.Helper()
	if rr.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json content type, got %q", ct)
	}
	var body struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil || body.Error.Message == "" {
		t.Errorf("expected error envelope, got %q (%v)", rr.Body.String(), err)
	}
}

func TestNotes_MissingIdentityHeaderIs401(t *testing.T) {
	router := newIdentityRouter(t, true)
	for _, header := range []string{"", "   "} {
		req := httptest.NewRequest(http.MethodGet, "/notes", nil)
		req.Header.Set("x-user-id", header)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assertJSONError(t, rr, http.StatusUnauthorized)
	}
}

func TestNotes_HeaderIgnoredWhenIdentityNotTrusted(t *testing.T) {
	router := newIdentityRouter(t, false)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notes", nil)
	req.Header.Set("x-user-id", "owner-a")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assertJSONError(t, rr, http.StatusUnauthorized)
}

func TestNotes_OversizedBodyIs413(t *testing.T) {
	router := newIdentityRouter(t, true)
	payload := `{"title":"t","body":"` + strings.Repeat("x", 2<<20) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(payload))
	req.Header.Set("x-user-id", "owner-a")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assertJSONError(t, rr, http.StatusRequestEntityTooLarge)
}

func TestAPIRoot_IsJSON(t *testing.T) {
	router := newIdentityRouter(t, false)
	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected response: %d %q", rr.Code, rr.Header().Get("Content-Type"))
	}
}
