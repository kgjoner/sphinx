package common

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/controller"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Middlewares struct {
	Pools
}

func (m Middlewares) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" {
			err := normalizederr.NewUnauthorizedError("missing bearer token", errcode.InvalidAccess)
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
			err := normalizederr.NewUnauthorizedError("must provide an access token", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		authRepo := m.BasePool.NewDAO(r.Context())
		account, err := authRepo.GetAccountByID(token.Claims.Sub)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		} else if account == nil {
			err := normalizederr.NewFatalUnauthorizedError("not existing user", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		err = account.Authenticate(token)
		authRepo.UpsertSessions(account.SessionsToPersist()...)
		if err != nil {
			ctx := r.Context()
			ctx = context.WithValue(ctx, controller.ActorKey, *account)
			r = r.WithContext(ctx)
			presenter.HTTPError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, controller.ActorKey, *account)
		ctx = context.WithValue(ctx, controller.TokenKey, tokenStr)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m Middlewares) AuthenticateApp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Basic" {
			err := normalizederr.NewUnauthorizedError("missing app basic token", errcode.InvalidAccess)
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
			err := normalizederr.NewUnauthorizedError("invalid credentials", errcode.InvalidAccess)
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
			err := normalizederr.NewUnauthorizedError("not existing app", errcode.InvalidAccess)
			presenter.HTTPError(err, w, r)
			return
		}

		err = application.Authenticate(appSecret)
		ctx := r.Context()
		ctx = context.WithValue(ctx, controller.ApplicationKey, *application)
		r = r.WithContext(ctx)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m Middlewares) Target(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		actor := ctx.Value(controller.ActorKey)

		targetEntry := r.Header.Get("x-target")
		if targetEntry == "" {
			if actor == nil {
				err := normalizederr.NewRequestError("must provide a target header", errcode.InvalidAccess)
				presenter.HTTPError(err, w, r)
				return
			}

			ctx = context.WithValue(ctx, controller.TargetKey, actor)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		app := ctx.Value(controller.ApplicationKey)
		isAdmin := actor != nil && actor.(auth.Account).HasRoleOnAuth(auth.RoleAdmin)
		isAuthedApp := app != nil && app.(auth.Application).IsAuthenticated()
		if !isAdmin && !isAuthedApp {
			err := normalizederr.NewForbiddenError("does not have permission to execute this action")
			presenter.HTTPError(err, w, r)
			return
		}

		authRepo := m.BasePool.NewDAO(r.Context())
		var err error
		var target *auth.Account
		var entry auth.Entry
		if id, errif := uuid.Parse(targetEntry); errif == nil {
			target, err = authRepo.GetAccountByID(id)
		} else {
			entry, err = auth.ParseEntry(targetEntry)
			if err == nil {
				target, err = authRepo.GetAccountByEntry(entry)
			}
		}

		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		} else if target == nil {
			err := normalizederr.NewRequestError("target account does not exist", errcode.AccountNotFound)
			presenter.HTTPError(err, w, r)
			return
		}

		ctx = context.WithValue(ctx, controller.TargetKey, *target)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
