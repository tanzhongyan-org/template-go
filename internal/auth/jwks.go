// Package auth verifies product JWTs against the JWKS document served by the
// product's auth provider (baseline section 1: the auth plugin serves JWKS to
// other backends). It fails closed: any fetch, parse, signature, expiry,
// issuer, audience or subject problem rejects the credential.
package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwksTimeout bounds one JWKS fetch. jwksMaxBytes caps the document size.
const (
	jwksTimeout  = 10 * time.Second
	jwksMaxBytes = 1 << 20
)

// Verifier checks JWTs against a remote JWKS document.
type Verifier struct {
	jwksURL  string
	issuer   string
	audience string
	client   *http.Client

	mu   sync.Mutex
	keys map[string]crypto.PublicKey
}

// NewVerifier fetches the JWKS document and returns a verifier for credentials
// whose issuer and audience match. It fails when the document cannot be
// fetched or holds no usable signing key.
func NewVerifier(jwksURL, issuer, audience string) (*Verifier, error) {
	v := &Verifier{
		jwksURL:  jwksURL,
		issuer:   issuer,
		audience: audience,
		client:   &http.Client{Timeout: jwksTimeout},
		keys:     make(map[string]crypto.PublicKey),
	}
	if err := v.refresh(); err != nil {
		return nil, err
	}
	if len(v.keys) == 0 {
		return nil, fmt.Errorf("auth: JWKS document at %s holds no usable signing key", jwksURL)
	}
	return v, nil
}

// Verify parses raw, checks its signature against the JWKS keys, validates
// expiry, issuer and audience, and returns the subject as the caller identity.
// An unknown key ID triggers one JWKS refresh before the credential is
// rejected, so key rotations do not lock callers out.
func (v *Verifier) Verify(raw string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(raw, claims, v.keyFunc(false), jwt.WithExpirationRequired(),
		jwt.WithIssuer(v.issuer), jwt.WithAudience(v.audience),
		jwt.WithValidMethods([]string{"RS256", "ES256", "ES384"}))
	if err != nil {
		return "", fmt.Errorf("auth: invalid credential: %w", err)
	}
	if !parsed.Valid {
		return "", fmt.Errorf("auth: invalid credential")
	}
	if claims.Subject == "" {
		return "", fmt.Errorf("auth: credential carries no subject")
	}
	return claims.Subject, nil
}

// keyFunc returns the jwt key lookup. The lookup refreshes the JWKS document
// once on an unknown key ID unless refreshed is already true.
func (v *Verifier) keyFunc(refreshed bool) func(*jwt.Token) (any, error) {
	return func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		key, err := v.lookup(kid, refreshed)
		if err != nil {
			return nil, err
		}
		if err := checkAlg(t.Method.Alg(), key); err != nil {
			return nil, err
		}
		return key, nil
	}
}

// lookup returns the key for kid, refreshing the document once first when the
// ID is unknown and refreshed is false.
func (v *Verifier) lookup(kid string, refreshed bool) (crypto.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if key, ok := v.keys[kid]; ok {
		return key, nil
	}
	if refreshed {
		return nil, fmt.Errorf("auth: unknown key ID %q", kid)
	}
	keys, err := fetchKeys(v.client, v.jwksURL)
	if err != nil {
		return nil, err
	}
	v.keys = keys
	if key, ok := v.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("auth: unknown key ID %q", kid)
}

// checkAlg rejects algorithm confusion: RSA keys only speak RS256 and ECDSA
// keys only speak the ECDSA methods.
func checkAlg(alg string, key crypto.PublicKey) error {
	switch key.(type) {
	case *rsa.PublicKey:
		if alg != "RS256" {
			return fmt.Errorf("auth: method %q does not match an RSA key", alg)
		}
	case *ecdsa.PublicKey:
		if alg != "ES256" && alg != "ES384" {
			return fmt.Errorf("auth: method %q does not match an ECDSA key", alg)
		}
	default:
		return fmt.Errorf("auth: unsupported key type %T", key)
	}
	return nil
}

// refresh replaces the cached keys with a fresh fetch.
func (v *Verifier) refresh() error {
	keys, err := fetchKeys(v.client, v.jwksURL)
	if err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	v.keys = keys
	return nil
}

// jwksDocument is the subset of a JWKS document this package reads.
type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

// jwk is one JSON Web Key. Only public RSA and ECDSA signing keys apply;
// anything else is skipped.
type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	// RSA members.
	N string `json:"n"`
	E string `json:"e"`
	// ECDSA members.
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// fetchKeys downloads the JWKS document and builds the key map. Keys without
// an ID, with a non-signing use, or of an unsupported type are skipped.
func fetchKeys(client *http.Client, url string) (map[string]crypto.PublicKey, error) {
	resp, err := client.Get(url) //nolint:gosec,noctx // bounded client with a timeout; URL is operator config
	if err != nil {
		return nil, fmt.Errorf("auth: fetch JWKS: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth: fetch JWKS: unexpected status %s", resp.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, jwksMaxBytes))
	if err != nil {
		return nil, fmt.Errorf("auth: read JWKS: %w", err)
	}
	var doc jwksDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("auth: parse JWKS: %w", err)
	}
	keys := make(map[string]crypto.PublicKey, len(doc.Keys))
	for _, k := range doc.Keys {
		if k.Kid == "" || (k.Use != "" && k.Use != "sig") {
			continue
		}
		key, err := k.publicKey()
		if err != nil {
			continue
		}
		keys[k.Kid] = key
	}
	return keys, nil
}

// publicKey builds the Go public key for one JWK.
func (k jwk) publicKey() (crypto.PublicKey, error) {
	switch k.Kty {
	case "RSA":
		return rsaKey(k.N, k.E)
	case "EC":
		return ecdsaKey(k.Crv, k.X, k.Y)
	default:
		return nil, fmt.Errorf("unsupported key type %q", k.Kty)
	}
}

// rsaKey decodes the base64url modulus and exponent into an RSA public key.
func rsaKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("bad RSA modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("bad RSA exponent: %w", err)
	}
	if len(eBytes) == 0 || len(eBytes) > 4 {
		return nil, fmt.Errorf("bad RSA exponent")
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

// ecdsaKey decodes the base64url coordinates on the named curve.
func ecdsaKey(crv, xStr, yStr string) (*ecdsa.PublicKey, error) {
	var curve elliptic.Curve
	var size int
	switch crv {
	case "P-256":
		curve, size = elliptic.P256(), 32
	case "P-384":
		curve, size = elliptic.P384(), 48
	default:
		return nil, fmt.Errorf("unsupported curve %q", crv)
	}
	xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, fmt.Errorf("bad EC x coordinate: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, fmt.Errorf("bad EC y coordinate: %w", err)
	}
	if len(xBytes) > size || len(yBytes) > size {
		return nil, fmt.Errorf("EC coordinates exceed %s size", crv)
	}
	// Uncompressed point encoding: 0x04 followed by the padded coordinates.
	raw := make([]byte, 0, 1+2*size)
	raw = append(raw, 4)
	raw = append(raw, make([]byte, size-len(xBytes))...)
	raw = append(raw, xBytes...)
	raw = append(raw, make([]byte, size-len(yBytes))...)
	raw = append(raw, yBytes...)
	pub, err := ecdsa.ParseUncompressedPublicKey(curve, raw)
	if err != nil {
		return nil, fmt.Errorf("bad EC point on %s: %w", crv, err)
	}
	return pub, nil
}
