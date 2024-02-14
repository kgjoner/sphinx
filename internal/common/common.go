package common

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type Repos struct {
	AuthRepo authcase.AuthRepo
}

type Middlewares struct {
	AuthRepo authcase.AuthRepo
}

func (m Middlewares) AppToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appToken := r.Header.Get("authorization")
		if appToken == "" {
			err := normalizederr.NewUnauthorizedError("Missing app token.")
			presenter.HttpError(err, w, r)
			return
		}

		appId, err := uuid.Parse(appToken)
		if err != nil {
			err := normalizederr.NewUnauthorizedError("Bad formatted app token.")
			presenter.HttpError(err, w, r)
			return
		}

		application, err := m.AuthRepo.GetApplicationById(appId)
		if err != nil {
			presenter.HttpError(err, w, r)
			return
		} else if application == nil {
			err := normalizederr.NewUnauthorizedError("Invalid app token.")
			presenter.HttpError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "application", *application)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m Middlewares) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" {
			err := normalizederr.NewUnauthorizedError("Missing bearer token.")
			presenter.HttpError(err, w, r)
			return
		}
		
		tokenStr := authHeaderParts[1]
		token, err := auth.ParseAuthToken(tokenStr)
		if err != nil {
			presenter.HttpError(err, w, r)
			return
		}
		
		if token.IsRefresh() && !strings.Contains(r.URL.Path, "refresh") {
			err := normalizederr.NewUnauthorizedError("Must provide an access token.")
			presenter.HttpError(err, w, r)
			return
		}

		account, err := m.AuthRepo.GetAccountById(token.Claims.Sub)
		if err != nil {
			presenter.HttpError(err, w, r)
			return
		} else if account == nil {
			err := normalizederr.NewFatalUnauthorizedError("Not existing user.")
			presenter.HttpError(err, w, r)
			return
		}

		err = account.Authenticate(token)
		m.AuthRepo.UpsertSessions(account.SessionsToPersist()...)
		if err != nil {
			ctx := r.Context()
			ctx = context.WithValue(ctx, "actor", *account)
			r = r.WithContext(ctx)
			presenter.HttpError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "actor", *account)
		ctx = context.WithValue(ctx, "token", tokenStr)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
