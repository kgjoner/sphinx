package sphinx

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/pkg/tokens"
)

type middlewares struct {
	appID         uuid.UUID
	tokenProvider auth.TokenProvider
	authorizer    Authorizer
}

// NewMiddlewares creates middlewares with JWKS-based token validation.
// This is the recommended approach for RS256 token validation.
func NewMiddlewares(appID uuid.UUID, sphinxBaseURL string, authorizer Authorizer) AuthMiddlewares {
	tokenProvider := NewJWKSProvider(sphinxBaseURL, "") // No HS256 fallback

	return &middlewares{
		appID:         appID,
		tokenProvider: tokenProvider,
		authorizer:    authorizer,
	}
}

// NewMiddlewaresWithHS256Fallback creates middlewares with JWKS-based RS256 validation and HS256 fallback.
// Use this during migration period when both RS256 and HS256 tokens may be in use.
func NewMiddlewaresWithHS256Fallback(appID uuid.UUID, sphinxBaseURL string, tokenSecret string, authorizer Authorizer) AuthMiddlewares {
	tokenProvider := NewJWKSProvider(sphinxBaseURL, tokenSecret) // With HS256 fallback

	return &middlewares{
		appID:         appID,
		tokenProvider: tokenProvider,
		authorizer:    authorizer,
	}
}

// NewMiddlewaresWithSecret creates middlewares with a shared secret for HS256 validation.
// Deprecated: Use NewMiddlewares with JWKS for RS256 support. This is kept for backward compatibility.
func NewMiddlewaresWithSecret(appID uuid.UUID, tokenSecret string, authorizer Authorizer) AuthMiddlewares {
	tokenProvider := tokens.NewJWTProvider(tokenSecret, 0, 0) // lifetime is not used in validation

	return &middlewares{
		appID:         appID,
		tokenProvider: tokenProvider,
		authorizer:    authorizer,
	}
}

// Ensure authentication via bearer token
func (m middlewares) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" || authHeaderParts[1] == "" {
			httpserver.Error(ErrInvalidAccess, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		sub, intent, err := m.tokenProvider.Validate(tokenStr)
		if err != nil {
			httpserver.Error(err, w, r)
			return
		} else if intent != auth.IntentAccess {
			httpserver.Error(ErrInvalidAccess, w, r)
			return
		}

		r, err = m.authorizer.AuthorizeSubject(Subject(*sub), r)
		if err != nil {
			httpserver.Error(err, w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// If authorization header is present, ensure authentication via bearer token. Otherwise, allow request forward.
func (m middlewares) TryAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		m.Authenticate(next).ServeHTTP(w, r)
	})
}
