package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanzhongyan-org/template-go/internal/config"
	"github.com/tanzhongyan-org/template-go/internal/db"
	"github.com/tanzhongyan-org/template-go/internal/handler"
	dbschema "github.com/tanzhongyan-org/template-go/sql"
)

const (
	jwtTestIssuer   = "https://auth.example.test"
	jwtTestAudience = "notes-api"
	jwtTestKid      = "test-rsa-1"
)

func mustJWTKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}

// jwtTestServer serves a JWKS document holding the public half of key.
func jwtTestServer(t *testing.T, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()
	e := key.E
	eb := []byte{byte(e >> 16), byte(e >> 8), byte(e)}
	for len(eb) > 1 && eb[0] == 0 {
		eb = eb[1:]
	}
	doc, err := json.Marshal(map[string]any{"keys": []map[string]string{{
		"kty": "RSA",
		"kid": jwtTestKid,
		"use": "sig",
		"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(eb),
	}}})
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(doc)
	}))
}

func signJWT(t *testing.T, key *rsa.PrivateKey, claims jwt.RegisteredClaims) string {
	t.Helper()
	issued := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	issued.Header["kid"] = jwtTestKid
	signed, err := issued.SignedString(key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func jwtClaims(subject string) jwt.RegisteredClaims {
	now := time.Now()
	return jwt.RegisteredClaims{
		Subject:   subject,
		Issuer:    jwtTestIssuer,
		Audience:  jwt.ClaimStrings{jwtTestAudience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	}
}

func newJWTRouter(t *testing.T, jwksURL string) http.Handler {
	t.Helper()
	cfg := &config.Config{
		AppName: "template-go", AppEnv: "test", Port: 8080,
		TrustIdentityHeader: false,
		JWKSURL:             jwksURL,
		JWTIssuer:           jwtTestIssuer,
		JWTAudience:         jwtTestAudience,
	}
	// The identity check runs before any query, so a nil pool is never touched.
	return SetupRouter(cfg, db.New(nil))
}

func TestNotes_JWTShapesAre401(t *testing.T) {
	key := mustJWTKey(t)
	srv := jwtTestServer(t, key)
	defer srv.Close()
	router := newJWTRouter(t, srv.URL)

	expired := jwtClaims("owner-a")
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))

	cases := map[string]string{
		"missing header":  "",
		"bare scheme":     "Bearer ",
		"wrong scheme":    "Token abc",
		"malformed":       "Bearer not-a-credential",
		"expired":         "Bearer " + signJWT(t, key, expired),
		"unknown subject": "Bearer " + signJWT(t, mustJWTKey(t), jwtClaims("owner-a")),
	}
	for name, header := range cases {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notes", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assertJSONError(t, rr, http.StatusUnauthorized)
		t.Logf("%s: rejected as expected", name)
	}
}

func TestNotes_JWTIgnoresTheDevHeader(t *testing.T) {
	key := mustJWTKey(t)
	srv := jwtTestServer(t, key)
	defer srv.Close()
	router := newJWTRouter(t, srv.URL)

	// The dev header alone must not authenticate once JWT is configured.
	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	req.Header.Set("x-user-id", "owner-a")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assertJSONError(t, rr, http.StatusUnauthorized)
}

func setupJWTApp(t *testing.T, jwksURL string) (*chi.Mux, *pgxpool.Pool) {
	t.Helper()
	t.Setenv("TRUST_IDENTITY_HEADER", "false")
	t.Setenv("AUTH_JWKS_URL", jwksURL)
	t.Setenv("AUTH_ISSUER", jwtTestIssuer)
	t.Setenv("AUTH_AUDIENCE", jwtTestAudience)
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
	return SetupRouter(cfg, queries), pool
}

func TestNotes_JWTRoundtrip(t *testing.T) {
	key := mustJWTKey(t)
	srv := jwtTestServer(t, key)
	defer srv.Close()
	router, pool := setupJWTApp(t, srv.URL)
	defer pool.Close()

	ownerA := "Bearer " + signJWT(t, key, jwtClaims("jwt-owner-a"))
	ownerB := "Bearer " + signJWT(t, key, jwtClaims("jwt-owner-b"))

	payload, _ := json.Marshal(handler.CreateNoteRequest{Title: "JWT note", Body: "Owned by the subject"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", ownerA)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var note handler.NoteResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &note); err != nil {
		t.Fatalf("failed to decode created note: %v", err)
	}
	if note.OwnerID != "jwt-owner-a" {
		t.Errorf("expected owner_id 'jwt-owner-a', got %q", note.OwnerID)
	}

	// Another subject cannot read the note.
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/notes/%s", note.ID), nil)
	req.Header.Set("Authorization", ownerB)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 for the other subject, got %d", rr.Code)
	}

	// The owning subject deletes it.
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/notes/%s", note.ID), nil)
	req.Header.Set("Authorization", ownerA)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
}
