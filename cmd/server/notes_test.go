package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanzhongyan-org/template-go/internal/config"
	"github.com/tanzhongyan-org/template-go/internal/db"
	"github.com/tanzhongyan-org/template-go/internal/handler"
	dbschema "github.com/tanzhongyan-org/template-go/sql"
)

func setupTestApp(t *testing.T) (*chi.Mux, *pgxpool.Pool) {
	t.Helper()
	t.Setenv("TRUST_IDENTITY_HEADER", "true")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if _, err := pool.Exec(ctx, dbschema.Schema); err != nil {
		t.Fatalf("failed to init test schema: %v", err)
	}

	queries := db.New(pool)
	router := SetupRouter(cfg, queries)
	return router, pool
}

func TestNotes_RoundtripAndTenantIsolation(t *testing.T) {
	router, pool := setupTestApp(t)
	defer pool.Close()

	// 1. Create a note for owner-A
	createPayload := handler.CreateNoteRequest{
		Title: "Owner A Note",
		Body:  "Confidential notes for owner A",
	}
	bodyBytes, _ := json.Marshal(createPayload)

	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", "owner-a")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var noteA handler.NoteResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &noteA); err != nil {
		t.Fatalf("failed to decode created note: %v", err)
	}
	if noteA.Title != "Owner A Note" {
		t.Errorf("expected title 'Owner A Note', got %q", noteA.Title)
	}
	if noteA.OwnerID != "owner-a" {
		t.Errorf("expected owner_id 'owner-a', got %q", noteA.OwnerID)
	}

	// 2. Owner A can fetch their note
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%s", noteA.ID), nil)
	req.Header.Set("x-user-id", "owner-a")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	// 3. Tenant Isolation: Owner B CANNOT read Owner A's note (404)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%s", noteA.ID), nil)
	req.Header.Set("x-user-id", "owner-b")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 for owner B reading owner A note, got %d", rr.Code)
	}

	// 4. Tenant Isolation: Owner B CANNOT delete Owner A's note (404)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%s", noteA.ID), nil)
	req.Header.Set("x-user-id", "owner-b")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 for owner B deleting owner A note, got %d", rr.Code)
	}

	// 5. Validation failure: empty title returns 400 Bad Request
	badPayload := handler.CreateNoteRequest{
		Title: "   ",
		Body:  "Valid body",
	}
	badBytes, _ := json.Marshal(badPayload)
	req = httptest.NewRequest(http.MethodPost, "/notes", bytes.NewReader(badBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", "owner-a")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty title, got %d", rr.Code)
	}

	// 6. Owner A deletes their note -> 204 No Content
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%s", noteA.ID), nil)
	req.Header.Set("x-user-id", "owner-a")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
}
