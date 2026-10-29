package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testIssuer   = "https://auth.example.test"
	testAudience = "notes-api"
	testSubject  = "owner-123"
)

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}

func mustECKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate EC key: %v", err)
	}
	return key
}

func encodeBig(v []byte) string {
	return base64.RawURLEncoding.EncodeToString(v)
}

func rsaJWK(pub *rsa.PublicKey, kid string) map[string]string {
	e := pub.E
	eb := []byte{byte(e >> 16), byte(e >> 8), byte(e)}
	for len(eb) > 1 && eb[0] == 0 {
		eb = eb[1:]
	}
	return map[string]string{
		"kty": "RSA",
		"kid": kid,
		"use": "sig",
		"n":   encodeBig(pub.N.Bytes()),
		"e":   encodeBig(eb),
	}
}

func ecJWK(t *testing.T, pub *ecdsa.PublicKey, kid string) map[string]string {
	t.Helper()
	// Uncompressed point encoding: 0x04 followed by the X and Y coordinates.
	raw, err := pub.Bytes()
	if err != nil {
		t.Fatalf("encode EC point: %v", err)
	}
	n := (len(raw) - 1) / 2
	return map[string]string{
		"kty": "EC",
		"kid": kid,
		"use": "sig",
		"crv": "P-256",
		"x":   encodeBig(bytes.TrimLeft(raw[1:1+n], "\x00")),
		"y":   encodeBig(bytes.TrimLeft(raw[1+n:], "\x00")),
	}
}

// jwksServer serves one JWKS document built from keys.
func jwksServer(t *testing.T, keys ...map[string]string) *httptest.Server {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"keys": keys})
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	}))
}

func testClaims() jwt.RegisteredClaims {
	now := time.Now()
	return jwt.RegisteredClaims{
		Subject:   testSubject,
		Issuer:    testIssuer,
		Audience:  jwt.ClaimStrings{testAudience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	}
}

func sign(t *testing.T, method jwt.SigningMethod, key any, kid string, claims jwt.Claims) string {
	t.Helper()
	issued := jwt.NewWithClaims(method, claims)
	issued.Header["kid"] = kid
	signed, err := issued.SignedString(key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func TestVerify_RS256(t *testing.T) {
	key := mustRSAKey(t)
	srv := jwksServer(t, rsaJWK(&key.PublicKey, "rsa-1"))
	defer srv.Close()

	v, err := NewVerifier(srv.URL, testIssuer, testAudience)
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}
	sub, err := v.Verify(sign(t, jwt.SigningMethodRS256, key, "rsa-1", testClaims()))
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if sub != testSubject {
		t.Errorf("expected subject %q, got %q", testSubject, sub)
	}
}

func TestVerify_ES256(t *testing.T) {
	key := mustECKey(t)
	srv := jwksServer(t, ecJWK(t, &key.PublicKey, "ec-1"))
	defer srv.Close()

	v, err := NewVerifier(srv.URL, testIssuer, testAudience)
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}
	sub, err := v.Verify(sign(t, jwt.SigningMethodES256, key, "ec-1", testClaims()))
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if sub != testSubject {
		t.Errorf("expected subject %q, got %q", testSubject, sub)
	}
}

func TestVerify_Rejects(t *testing.T) {
	rsaKey := mustRSAKey(t)
	otherRSA := mustRSAKey(t)
	ecKey := mustECKey(t)
	srv := jwksServer(t, rsaJWK(&rsaKey.PublicKey, "rsa-1"), ecJWK(t, &ecKey.PublicKey, "ec-1"))
	defer srv.Close()

	v, err := NewVerifier(srv.URL, testIssuer, testAudience)
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}

	expired := testClaims()
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	wrongAud := testClaims()
	wrongAud.Audience = jwt.ClaimStrings{"other-api"}
	wrongIss := testClaims()
	wrongIss.Issuer = "https://evil.example.test"
	noSub := testClaims()
	noSub.Subject = ""

	cases := map[string]string{
		"expired":        sign(t, jwt.SigningMethodRS256, rsaKey, "rsa-1", expired),
		"wrong audience": sign(t, jwt.SigningMethodRS256, rsaKey, "rsa-1", wrongAud),
		"wrong issuer":   sign(t, jwt.SigningMethodRS256, rsaKey, "rsa-1", wrongIss),
		"no subject":     sign(t, jwt.SigningMethodRS256, rsaKey, "rsa-1", noSub),
		"wrong key":      sign(t, jwt.SigningMethodRS256, otherRSA, "rsa-1", testClaims()),
		"unknown kid":    sign(t, jwt.SigningMethodRS256, rsaKey, "rsa-9", testClaims()),
		"malformed":      "not-a-credential",
		"empty":          "",
		// EC-signed but pointing at the RSA key: the algorithm check rejects it.
		"alg confusion": sign(t, jwt.SigningMethodES256, ecKey, "rsa-1", testClaims()),
	}
	for name, raw := range cases {
		if sub, err := v.Verify(raw); err == nil {
			t.Errorf("%s: expected an error, got subject %q", name, sub)
		}
	}
}

func TestVerify_RefreshesOnRotation(t *testing.T) {
	oldKey := mustRSAKey(t)
	newKey := mustRSAKey(t)
	oldDoc, _ := json.Marshal(map[string]any{"keys": []map[string]string{rsaJWK(&oldKey.PublicKey, "rsa-1")}})
	newDoc, _ := json.Marshal(map[string]any{"keys": []map[string]string{rsaJWK(&newKey.PublicKey, "rsa-2")}})

	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if calls.Add(1) == 1 {
			_, _ = w.Write(oldDoc)
			return
		}
		_, _ = w.Write(newDoc)
	}))
	defer srv.Close()

	v, err := NewVerifier(srv.URL, testIssuer, testAudience)
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}
	// The verifier only knows rsa-1, so this triggers one refresh to rsa-2.
	sub, err := v.Verify(sign(t, jwt.SigningMethodRS256, newKey, "rsa-2", testClaims()))
	if err != nil {
		t.Fatalf("Verify after rotation: %v", err)
	}
	if sub != testSubject {
		t.Errorf("expected subject %q, got %q", testSubject, sub)
	}
}

func TestNewVerifier_FetchFailures(t *testing.T) {
	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer broken.Close()

	empty := jwksServer(t)
	defer empty.Close()

	unusable := jwksServer(t,
		map[string]string{"kty": "RSA", "use": "sig", "n": "x", "e": "y"},               // no kid
		map[string]string{"kty": "RSA", "kid": "enc", "use": "enc", "n": "x", "e": "y"}, // not signing
		map[string]string{"kty": "oct", "kid": "sym", "use": "sig"},                     // unsupported type
		map[string]string{"kty": "EC", "kid": "p521", "use": "sig", "crv": "P-521", "x": "x", "y": "y"},
	)
	defer unusable.Close()

	closed := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	closedURL := closed.URL
	closed.Close()

	for name, url := range map[string]string{
		"unreachable": closedURL,
		"bad JSON":    broken.URL,
		"no keys":     empty.URL,
		"no usable":   unusable.URL,
	} {
		if _, err := NewVerifier(url, testIssuer, testAudience); err == nil {
			t.Errorf("%s: expected an error, got nil", name)
		}
	}
}
