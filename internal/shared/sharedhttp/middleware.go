package sharedhttp

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v3/apperr"
	"github.com/kgjoner/cornucopia/v3/httpserver"
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

// Authenticate middleware verifies Bearer token from Authorization header,
// authorizes the user, and injects the actor and token into the request context.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" {
			httpserver.HTTPError(shared.ErrMissingCredentials, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		isRefreshRoute := strings.Contains(r.URL.Path, "/refresh")
		actor, err := m.authorizer.AuthorizeToken(tokenStr, isRefreshRoute)
		if err != nil {
			httpserver.HTTPError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ActorCtxKey, actor)
		ctx = context.WithValue(ctx, TokenCtxKey, tokenStr)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// AuthenticateApp middleware verifies Basic auth from Authorization header,
// authorizes the application, and injects the actor into the request context.
func (m *Middleware) AuthenticateApp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) < 2 || authHeaderParts[0] != "Basic" {
			httpserver.HTTPError(shared.ErrMissingCredentials, w, r)
			return
		}

		tokenStr := authHeaderParts[1]
		decodedBytes, err := base64.StdEncoding.DecodeString(tokenStr)
		if err != nil {
			httpserver.HTTPError(err, w, r)
			return
		}

		credentials := strings.Split(string(decodedBytes), ":")
		if len(credentials) != 2 {
			httpserver.HTTPError(shared.ErrInvalidCredentials, w, r)
			return
		}

		appID, _ := uuid.Parse(credentials[0])
		appSecret := credentials[1]

		actor, err := m.authorizer.AuthorizeApp(r.Context(), appID, appSecret)
		if err != nil {
			httpserver.HTTPError(err, w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ActorCtxKey, actor)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// AuthenticateAny middleware verifies Bearer or Basic auth from Authorization header,
// authorizes the user or application, and injects the actor into the request context.
func (m *Middleware) AuthenticateAny(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			m.Authenticate(next).ServeHTTP(w, r)
			return
		} else if strings.HasPrefix(authHeader, "Basic ") {
			m.AuthenticateApp(next).ServeHTTP(w, r)
			return
		} else {
			httpserver.HTTPError(shared.ErrMissingCredentials, w, r)
			return
		}
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
				httpserver.HTTPError(err, w, r)
				return
			}

			actor := untypedActor.(shared.Actor)
			if actor.Kind != shared.KindUser {
				err := apperr.NewForbiddenError("actor must be a user if no userID is provided")
				httpserver.HTTPError(err, w, r)
				return
			}

			targetUserID = actor.ID
		} else {
			parsedID, err := uuid.Parse(targetParam)
			if err != nil {
				err := apperr.Wrap(err, apperr.Request, apperr.BadRequest, "invalid user ID format")
				httpserver.HTTPError(err, w, r)
				return
			}
			targetUserID = parsedID
		}

		ctx = context.WithValue(ctx, TargetIDCtxKey, targetUserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
