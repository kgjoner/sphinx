package sphinx

import (
	"context"
	"net/http"
	"strings"

	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/common/errcode"
)

type ctxKey string

const (
	ActorCtxKey ctxKey = "sphinx_actor"
	TokenCtxKey ctxKey = "sphinx_token"
)

type Middlewares struct {
	sphinx *Service
}

func (s *Service) Middlewares() *Middlewares {
	return &Middlewares{
		sphinx: s,
	}
}

// Ensure authentication via bearer token
func (m Middlewares) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" || authHeaderParts[1] == "" {
			err := apperr.NewUnauthorizedError("missing bearer token", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		user, err := m.sphinx.Me(tokenStr)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ActorCtxKey, *user)
		ctx = context.WithValue(ctx, TokenCtxKey, tokenStr)
		ctx = context.WithValue(ctx, presenter.ActorLogKey, user.ID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// If authorization header is present, ensure authentication via bearer token. Otherwise, allow request forward.
func (m Middlewares) TryAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		m.Authenticate(next).ServeHTTP(w, r)
	})
}

// Ensure authenticated user has at least one of listed roles. Admin users are always allowed.
func (m Middlewares) Guard(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actorValue := r.Context().Value(ActorCtxKey)
			if actorValue == nil {
				err := apperr.NewUnauthorizedError("no actor found, user must be authenticated prior guard middleware", errcode.InvalidAccess)
				presenter.HTTPError(err, w, r)
				return
			}

			actor := actorValue.(User)
			if actor.IsAdmin() {
				next.ServeHTTP(w, r)
				return
			}

			for _, p := range roles {
				if actor.HasRole(p) {
					next.ServeHTTP(w, r)
					return
				}
			}

			err := apperr.NewForbiddenError("user does not have enough permission")
			presenter.HTTPError(err, w, r)
		})
	}
}
