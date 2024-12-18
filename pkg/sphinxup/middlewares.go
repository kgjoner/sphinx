package sphinxup

import (
	"context"
	"net/http"
	"strings"

	"github.com/kgjoner/cornucopia/helpers/controller"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/config/errcode"
	sphinx "github.com/kgjoner/sphinx/pkg/service"
)

type middlewares struct {
	Sphinx sphinx.Service
}

func NewMiddlewares(svc sphinx.Service) *middlewares {
	return &middlewares{
		svc,
	}
}

// Ensure authentication via bearer token
func (m middlewares) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" || authHeaderParts[1] == "" {
			err := normalizederr.NewUnauthorizedError("missing bearer token", errcode.InvalidAccess)
			presenter.HttpError(err, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		acc, err := m.Sphinx.Account(tokenStr)
		if err != nil {
			presenter.HttpError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, controller.ActorKey, *acc)
		ctx = context.WithValue(ctx, controller.TokenKey, tokenStr)
		r = r.WithContext(ctx)

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

		m.Authenticate(next)
	})
}

// Ensure authenticated account has at least one of listed permissions. Admin accounts are always allowed.
func (m middlewares) Guard(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actorValue := r.Context().Value("actor")
			if actorValue == nil {
				err := normalizederr.NewUnauthorizedError("no actor found, user must be authenticated prior guard middleware", errcode.InvalidAccess)
				presenter.HttpError(err, w, r)
			}

			actor := actorValue.(sphinx.Account)
			if actor.IsAdmin() {
				next.ServeHTTP(w, r)
				return
			}

			for _, p := range permissions {
				if actor.HasGranting(p) {
					next.ServeHTTP(w, r)
					return
				}

				if actor.HasRole(p) {
					next.ServeHTTP(w, r)
					return
				}
			}

			err := normalizederr.NewForbiddenError("user does not have enough permission")
			presenter.HttpError(err, w, r)
		})
	}
}
