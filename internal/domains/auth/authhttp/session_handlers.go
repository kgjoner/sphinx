package authhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/cornucopia/v3/httpserver"
	"github.com/kgjoner/sphinx/internal/domains/auth/authcase"
	"github.com/kgjoner/sphinx/internal/shared/sharedhttp"
)

func (g gateway) sessionHandlers(r chi.Router) {
	r.Post("/login", g.login)
	r.Post("/external/{provider}", g.externalLogin)
	r.Post("/external", g.externalLogin) // deprecated endpoint

	authedR := r.With(g.Authenticate)
	authedR.Post("/logout", g.logout)
	authedR.Post("/refresh", g.refresh)
}

// Login godoc
//
//	@Summary		Log user in
//	@Description	Exchange user credentials for root app auth tokens
//	@Router			/auth/login [post]
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		authcase.LoginInput	true	"Credentials. Entry can be: email, phone, username or document"
//	@Success		200		{object}	httpserver.Success[authcase.LoginOutput]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) login(w http.ResponseWriter, r *http.Request) {
	c := httpserver.New(r).
		JSONBody().
		AddIP().
		AddHeader("user-agent", "device")

	var input authcase.LoginInput
	err := c.Write(&input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	repo := g.AuthFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := authcase.Login{
		AuthRepo:      repo,
		PwHasher:      g.PwHasher,
		DataHasher:    g.DataHasher,
		TokenProvider: g.TokenProvider,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	httpserver.HTTPSuccess(output, w, r)
}

// Logout godoc
//
//	@Summary		Log user out
//	@Description	Invalidate current session and their tokens
//	@Router			/auth/logout [post]
//	@Tags			Auth
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) logout(w http.ResponseWriter, r *http.Request) {
	c := httpserver.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input authcase.LogoutInput
	err := c.Write(&input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	repo := g.AuthFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := authcase.Logout{
		AuthRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	httpserver.HTTPSuccess(output, w, r, http.StatusNoContent)
}

// Refresh godoc
//
//	@Summary		Issue new auth tokens
//	@Description	Get new auth tokens from a refresh one. This is invalidated.
//	@Router			/auth/refresh [post]
//	@Tags			Auth
//	@Security		BearerRefresh
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httpserver.Success[authcase.LoginOutput]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) refresh(w http.ResponseWriter, r *http.Request) {
	c := httpserver.New(r).
		AddFromContext(sharedhttp.ActorCtxKey, "actor").
		AddFromContext(sharedhttp.TokenCtxKey, "token")

	var input authcase.RefreshInput
	err := c.Write(&input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	repo := g.AuthFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := authcase.Refresh{
		AuthRepo:      repo,
		DataHasher:    g.DataHasher,
		TokenProvider: g.TokenProvider,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	httpserver.HTTPSuccess(output, w, r)
}

// External Login godoc
//
//	@Summary		Log user in via external provider
//	@Description	Use an external provider to authenticate user and generate root app auth tokens.
//	@Router			/auth/external/{provider} [post]
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string						true	"Name of the external identity provider (e.g., google, facebook)."
//	@Param			payload		body		authcase.ExternalLoginInput	true	"You may inform email if creation is expected and provider does not provide it."
//	@Success		200			{object}	httpserver.Success[authcase.LoginOutput]
//	@Failure		400			{object}	apperr.AppError
//	@Failure		401			{object}	apperr.AppError
//	@Failure		403			{object}	apperr.AppError
//	@Failure		500			{object}	apperr.AppError
func (g gateway) externalLogin(w http.ResponseWriter, r *http.Request) {
	c := httpserver.New(r).
		ParseURLParam("provider", "providerName").
		JSONBody().
		AddIP().
		AddHeader("user-agent", "device").
		AddLanguages().
		// Legacy support for deprecated endpoint
		AddHeader("authorization")

	var input authcase.ExternalLoginInput
	err := c.Write(&input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	repo := g.AuthFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := authcase.ExternalLogin{
		AuthRepo:         repo,
		IdentityProvider: g.IdentityProvider,
		DataHasher:       g.DataHasher,
		TokenProvider:    g.TokenProvider,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.HTTPError(err, w, r)
		return
	}

	httpserver.HTTPSuccess(output, w, r)
}
