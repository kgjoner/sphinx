package authhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
//	@Success		200		{object}	httpserver.SuccessResponse[authcase.LoginOutput]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) login(w http.ResponseWriter, r *http.Request) {
	var input authcase.LoginInput
	c := httpserver.Bind(r).
		JSONBody(&input).
		IP(&input.IP).
		Header("user-agent", &input.Device)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
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
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
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
	var input authcase.LogoutInput
	c := httpserver.Bind(r).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
		return
	}

	repo := g.AuthFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := authcase.Logout{
		AuthRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r, http.StatusNoContent)
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
//	@Success		200	{object}	httpserver.SuccessResponse[authcase.LoginOutput]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		401	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) refresh(w http.ResponseWriter, r *http.Request) {
	var input authcase.RefreshInput
	c := httpserver.Bind(r).
		FromContext(sharedhttp.ActorCtxKey, &input.Actor).
		FromContext(sharedhttp.TokenCtxKey, &input.Token)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
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
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
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
//	@Success		200			{object}	httpserver.SuccessResponse[authcase.LoginOutput]
//	@Failure		400			{object}	apperr.AppError
//	@Failure		401			{object}	apperr.AppError
//	@Failure		403			{object}	apperr.AppError
//	@Failure		500			{object}	apperr.AppError
func (g gateway) externalLogin(w http.ResponseWriter, r *http.Request) {
	var input authcase.ExternalLoginInput
	c := httpserver.Bind(r).
		PathParam("provider", &input.ProviderName).
		JSONBody(&input).
		IP(&input.IP).
		Header("user-agent", &input.Device).
		Languages(&input.Languages).
		// Legacy support for deprecated endpoint
		Header("authorization", &input.Authorization)

	if c.Err() != nil {
		httpserver.Error(c.Err(), w, r)
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
		httpserver.Error(err, w, r)
		return
	}

	httpserver.Success(output, w, r)
}
