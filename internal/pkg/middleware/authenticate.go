package middleware

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/pkg/errcode"
	"github.com/kgjoner/sphinx/internal/pkg/types"
)

type ctxKey string

const (
	ActorCtxKey  ctxKey = "sphinx_actor"
	TokenCtxKey  ctxKey = "sphinx_token"
	TargetCtxKey ctxKey = "sphinx_target"
)

type Authorizer interface {
	AuthorizeUser(token string) (types.Actor, error)
	AuthorizeApp(appID uuid.UUID, appSecret string) (types.Actor, error)
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" {
			err := apperr.NewUnauthorizedError("missing bearer token", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		actor, err := m.authorizer.AuthorizeUser(tokenStr)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ActorCtxKey, actor)
		ctx = context.WithValue(ctx, TokenCtxKey, tokenStr)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) AuthenticateApp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Basic" {
			err := apperr.NewUnauthorizedError("missing app basic token", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		decodedBytes, err := base64.StdEncoding.DecodeString(tokenStr)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		}

		credentials := strings.Split(string(decodedBytes), ":")
		if len(credentials) != 2 {
			err := apperr.NewUnauthorizedError("invalid credentials", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		appID, _ := uuid.Parse(credentials[0])
		appSecret := credentials[1]

		actor, err := m.authorizer.AuthorizeApp(appID, appSecret)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ActorCtxKey, actor)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
