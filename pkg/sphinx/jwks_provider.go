package sphinx

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

var (
	ErrJWKSFetchFailed  = errors.New("failed to fetch JWKS")
	ErrNoMatchingKey    = errors.New("no matching key found in JWKS")
	ErrInvalidKeyFormat = errors.New("invalid key format in JWKS")
)

// JWKSProvider fetches and caches public keys from a JWKS endpoint.
// Optionally supports HS256 fallback for backward compatibility during migration.
type JWKSProvider struct {
	jwksURL              string
	httpClient           *http.Client
	cacheTTL             time.Duration
	hs256Secret          string // Optional: for HS256 fallback during migration

	// Cache
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey // kid -> public key
	lastFetch time.Time
}

// NewJWKSProvider creates a new JWKS-based token provider.
// If hs256Secret is provided, it will fall back to HS256 validation for tokens that can't be validated with RS256.
func NewJWKSProvider(sphinxBaseURL string, hs256Secret string) *JWKSProvider {
	jwksURL := sphinxBaseURL + "/.well-known/jwks.json"

	return &JWKSProvider{
		jwksURL:              jwksURL,
		httpClient:           &http.Client{Timeout: 10 * time.Second},
		cacheTTL:             60 * time.Minute,
		hs256Secret:          hs256Secret,
		keys:                 make(map[string]*rsa.PublicKey),
	}
}

// Validate validates a JWT token using public keys from JWKS.
func (p *JWKSProvider) Validate(signedToken string) (*auth.Subject, auth.Intent, error) {
	// Parse token to extract kid from header
	var kid string
	token, err := jwt.Parse(signedToken, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() == "HS256" && p.hs256Secret != "" {
			// HS256 fallback
			return []byte(p.hs256Secret), nil
		}

		// Extract kid from header
		if kidHeader, ok := t.Header["kid"].(string); ok {
			kid = kidHeader
		} else {
			return nil, errors.New("missing kid in token header")
		}

		// Validate algorithm
		if t.Method.Alg() != "RS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}

		// Get public key for this kid
		publicKey, err := p.getPublicKey(kid)
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	if token == nil || !token.Valid {
		return nil, "", auth.ErrInvalidAccess
	}

	// Parse claims
	var claims jwtClaims
	ms, _ := json.Marshal(token.Claims)
	merr := json.Unmarshal(ms, &claims)
	if merr != nil {
		return nil, "", auth.ErrInvalidAccess
	}

	if err != nil {
		// Check if token is expired
		msg := err.Error()
		if containsString(msg, "token is expired") || containsString(msg, "exp") {
			if claims.Intent == auth.IntentRefresh {
				return nil, "", auth.ErrExpiredSession
			}
			return nil, "", auth.ErrExpiredAccess
		}

		return nil, "", auth.ErrInvalidAccess
	}

	// Build subject
	var email htypes.Email
	if claims.Email != "" {
		email, err = htypes.ParseEmail(claims.Email)
		if err != nil {
			return nil, "", auth.ErrInvalidAccess
		}
	}

	subject := &auth.Subject{
		ID:         claims.Sub,
		Kind:       shared.KindUser,
		Email:      email,
		Name:       claims.Name,
		AudienceID: claims.Aud,
		Roles:      claims.Roles,
		SessionID:  claims.SessionID,
	}

	return subject, claims.Intent, nil
}

// Generate is not implemented for JWKS provider (only used for validation in SDK)
func (p *JWKSProvider) Generate(sub auth.Subject) (*auth.Tokens, error) {
	return nil, errors.New("token generation not supported in JWKS provider")
}

// getPublicKey retrieves a public key by kid, fetching from JWKS if needed.
func (p *JWKSProvider) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache first
	p.mu.RLock()
	if time.Since(p.lastFetch) < p.cacheTTL {
		if key, exists := p.keys[kid]; exists {
			p.mu.RUnlock()
			return key, nil
		}
	}
	p.mu.RUnlock()

	// Fetch from JWKS endpoint
	err := p.fetchKeys()
	if err != nil {
		return nil, err
	}

	// Try again after fetch
	p.mu.RLock()
	defer p.mu.RUnlock()
	if key, exists := p.keys[kid]; exists {
		return key, nil
	}

	return nil, ErrNoMatchingKey
}

// fetchKeys fetches public keys from the JWKS endpoint.
func (p *JWKSProvider) fetchKeys() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJWKSFetchFailed, err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJWKSFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrJWKSFetchFailed, resp.StatusCode)
	}

	var jwks jwksResponse
	err = json.NewDecoder(resp.Body).Decode(&jwks)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJWKSFetchFailed, err)
	}

	// Parse and cache keys
	newKeys := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Alg != "RS256" {
			continue
		}

		publicKey, err := parseJWK(key)
		if err != nil {
			continue // Skip invalid keys
		}

		newKeys[key.Kid] = publicKey
	}

	// Update cache
	p.mu.Lock()
	p.keys = newKeys
	p.lastFetch = time.Now()
	p.mu.Unlock()

	return nil
}

// parseJWK converts a JWK to an RSA public key.
func parseJWK(key jwk) (*rsa.PublicKey, error) {
	// Decode modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, ErrInvalidKeyFormat
	}

	// Decode exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, ErrInvalidKeyFormat
	}

	// Convert to big.Int
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}

// Helper types

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwtClaims struct {
	Sub       uuid.UUID   `json:"sub"`
	Aud       uuid.UUID   `json:"aud"`
	Iat       int64       `json:"iat"`
	Exp       int64       `json:"exp"`
	Intent    auth.Intent `json:"itn"`
	Email     string      `json:"email"`
	Name      string      `json:"name"`
	Roles     []string    `json:"roles"`
	SessionID uuid.UUID   `json:"sid"`
}

func (c jwtClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{c.Aud.String()}, nil
}
func (c jwtClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: time.Unix(c.Exp, 0)}, nil
}
func (c jwtClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: time.Unix(c.Iat, 0)}, nil
}
func (c jwtClaims) GetIssuer() (string, error) {
	return "", nil
}
func (c jwtClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{}, nil
}
func (c jwtClaims) GetSubject() (string, error) {
	return c.Sub.String(), nil
}

// containsString checks if a string contains a substring (case-insensitive).
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
