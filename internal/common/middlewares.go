package common

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type ctxKey string

const (
	ActorCtxKey       ctxKey = "sphinx_actor"
	TokenCtxKey       ctxKey = "sphinx_token"
	ApplicationCtxKey ctxKey = "sphinx_application"
	TargetCtxKey      ctxKey = "sphinx_target"
)

type Middlewares struct {
	Pools
}

func (m Middlewares) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" {
			err := apperr.NewUnauthorizedError("missing bearer token", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		token, err := auth.ParseAuthToken(tokenStr)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		}

		if token.IsRefresh() && !strings.Contains(r.URL.Path, "refresh") {
			err := apperr.NewUnauthorizedError("must provide an access token", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		authRepo := m.BasePool.NewDAO(r.Context())
		user, err := authRepo.GetUserByID(token.Claims.Sub)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		} else if user == nil {
			err := apperr.Fatal(apperr.NewUnauthorizedError("not existing user", errcode.InvalidAccess))
			presenter.HTTPError(err, w, r)
			return
		}

		err = user.Authenticate(token)
		authRepo.UpsertSessions(user.SessionsToPersist()...)
		if err != nil {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ActorCtxKey, *user)
			r = r.WithContext(ctx)
			presenter.HTTPError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ActorCtxKey, *user)
		ctx = context.WithValue(ctx, TokenCtxKey, tokenStr)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m Middlewares) AuthenticateApp(next http.Handler) http.Handler {
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

		authRepo := m.BasePool.NewDAO(r.Context())
		application, err := authRepo.GetApplicationByID(appID)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		} else if application == nil {
			err := apperr.NewUnauthorizedError("not existing app", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		err = application.Authenticate(appSecret)
		ctx := r.Context()
		ctx = context.WithValue(ctx, ApplicationCtxKey, *application)
		r = r.WithContext(ctx)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Target middleware resolves the target user from "id" URL parameter, or
// defaults to the actor if no parameter is given.
// Only admin users or authenticated applications can specify a target other than the actor.
//
// Must be used after Authenticate or AuthenticateApp middlewares.
func (m Middlewares) Target(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		actor := ctx.Value(ActorCtxKey)

		targetId := chi.URLParam(r, "id")
		if targetId == "" {
			if actor == nil {
				err := apperr.NewInternalError("target middleware must have a target ID or an actor", errcode.InvalidAccess)
				presenter.HTTPError(err, w, r)
				return
			}

			ctx = context.WithValue(ctx, TargetCtxKey, actor)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		app := ctx.Value(ApplicationCtxKey)
		isAdmin := actor != nil && actor.(auth.User).HasRoleOnAuth(auth.RoleAdmin)
		isAuthedApp := app != nil && app.(auth.Application).IsAuthenticated()
		if !isAdmin && !isAuthedApp {
			err := apperr.NewForbiddenError("does not have permission to execute this action")
			presenter.HTTPError(err, w, r)
			return
		}

		authRepo := m.BasePool.NewDAO(r.Context())
		id, err := uuid.Parse(targetId)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		}

		target, err := authRepo.GetUserByID(id)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		} else if target == nil {
			err := apperr.NewRequestError("target user does not exist", errcode.UserNotFound)
			presenter.HTTPError(err, w, r)
			return
		}

		ctx = context.WithValue(ctx, TargetCtxKey, *target)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
