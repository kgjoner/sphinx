package sphinx

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/pkg/tokens"
)

type middlewares struct {
	appID         uuid.UUID
	tokenProvider auth.TokenProvider
	authorizer    Authorizer
}

func NewMiddlewares(appID uuid.UUID, tokenSecret string, authorizer Authorizer) *middlewares {
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
			presenter.HTTPError(ErrInvalidAccess, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		sub, intent, err := m.tokenProvider.Validate(tokenStr)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		} else if intent != auth.IntentAccess {
			presenter.HTTPError(ErrInvalidAccess, w, r)
			return
		}

		r, err = m.authorizer.AuthorizeSubject(Subject(*sub), r)
		if err != nil {
			presenter.HTTPError(err, w, r)
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
