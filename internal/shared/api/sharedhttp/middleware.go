package sharedhttp

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Middleware struct {
	authorizer Authorizer
}

func NewMiddleware(authorizer Authorizer) *Middleware {
	return &Middleware{
		authorizer: authorizer,
	}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" {
			presenter.HTTPError(shared.ErrMissingCredentials, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		actor, intent, err := m.authorizer.AuthorizeUser(tokenStr)
		if err != nil {
			presenter.HTTPError(err, w, r)
			return
		}

		if intent != auth.IntentAccess && intent != auth.IntentRefresh {
			presenter.HTTPError(auth.ErrInvalidAccess, w, r)
			return
		}

		isRefreshRoute := strings.Contains(r.URL.Path, "/refresh")
		if (intent == auth.IntentRefresh && !isRefreshRoute) ||
			(intent == auth.IntentAccess && isRefreshRoute) {
			presenter.HTTPError(auth.ErrInvalidAccess, w, r)
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
			presenter.HTTPError(shared.ErrMissingCredentials, w, r)
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
			presenter.HTTPError(shared.ErrInvalidCredentials, w, r)
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

// TargetUser middleware resolves a target user ID from "userID" URL parameter, or
// defaults to the actor UserID if no parameter is given.
//
// It is a convenience to be used in both /user/me and /user/{userID} endpoints.
func (m *Middleware) TargetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var targetUserID uuid.UUID

		targetParam := chi.URLParam(r, "userID")
		if targetParam == "" {
			untypedActor := ctx.Value(ActorCtxKey)
			if untypedActor == nil {
				err := apperr.NewInternalError("target middleware must have a userID or an actor")
				presenter.HTTPError(err, w, r)
				return
			}

			actor := untypedActor.(shared.Actor)
			if actor.Kind != shared.KindUser {
				err := apperr.NewForbiddenError("actor must be a user if no userID is provided")
				presenter.HTTPError(err, w, r)
				return
			}

			targetUserID = actor.ID
		} else {
			parsedID, err := uuid.Parse(targetParam)
			if err != nil {
				err := apperr.Wrap(err, apperr.Request, apperr.BadRequest, "invalid user ID format")
				presenter.HTTPError(err, w, r)
				return
			}
			targetUserID = parsedID
		}

		ctx = context.WithValue(ctx, TargetIDCtxKey, targetUserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
