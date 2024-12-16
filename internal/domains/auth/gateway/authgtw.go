package authgtw

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/helpers/controller"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/common"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type AuthGateway struct {
	common.Pools
	common.Services
	mid common.Middlewares
}

func Raise(router chi.Router, pools common.Pools, services common.Services) {
	authgtw := &AuthGateway{
		pools,
		services,
		common.Middlewares{Pools: pools},
	}

	router.Route("/account", authgtw.accountHandler)
	router.Route("/application", authgtw.applicationHandler)

	router.Route("/auth", func(r chi.Router) {
		r.With(authgtw.mid.AppId).Post("/login", authgtw.login)
		r.With(authgtw.mid.Authenticate).Post("/logout", authgtw.logout)
		r.With(authgtw.mid.Authenticate).Post("/refresh", authgtw.refresh)

		r.With(authgtw.mid.AppId).Post("/open/init", authgtw.initOAuth)
		r.With(authgtw.mid.AppId).Post("/open/login", authgtw.loginViaOAuth)
	})
}

// Login godoc
//
//	@Summary		Log user in
//	@Description	Exchange user credentials for auth tokens
//	@Router			/auth/login [post]
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			x-app	header		string				true	"Application ID"
//	@Param			payload	body		authcase.LoginInput	true	"Credentials. Entry can be: email, phone, username or document"
//	@Success		200		{object}	presenter.Success[authcase.LoginOutput]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) login(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseBody("entry", "password").
		AddApplication().
		AddIp().
		AddHeader("user-agent", "device")

	var input authcase.LoginInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := authcase.Login{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
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
//	@Success		200	{object}	presenter.Success[bool]
//	@Failure		400	{object}	normalizederr.NormalizedError
//	@Failure		401	{object}	normalizederr.NormalizedError
//	@Failure		500	{object}	normalizederr.NormalizedError
func (g AuthGateway) logout(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddActor()

	var input authcase.LogoutInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := authcase.Logout{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// Init OAuth godoc
//
//	@Summary		Request an oauth code
//	@Description	Exchange user credentials for oauth code
//	@Router			/auth/open/init [post]
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			x-app	header		string					true	"Application ID"
//	@Param			payload	body		authcase.InitOAuthInput	true	"Credentials. Entry can be: email, phone, username or document"
//	@Success		200		{object}	presenter.Success[string]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) initOAuth(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseBody("entry", "password", "state").
		AddApplication()

	var input authcase.InitOAuthInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := authcase.InitOAuth{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// Login via OAuth godoc
//
//	@Summary		Request an oauth code
//	@Description	Exchange user credentials for oauth code
//	@Router			/auth/open/login [post]
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			x-app	header		string						true	"Application ID"
//	@Param			payload	body		authcase.LoginViaOAuthInput	true	"Credentials. Entry can be: email, phone, username or document"
//	@Success		200		{object}	presenter.Success[authcase.LoginViaOAuthOutput]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) loginViaOAuth(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseBody("code", "appSecret").
		AddApplication().
		AddIp().
		AddHeader("user-agent", "device")

	var input authcase.LoginViaOAuthInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := authcase.LoginViaOAuth{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
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
//	@Success		200	{object}	presenter.Success[authcase.LoginOutput]
//	@Failure		400	{object}	normalizederr.NormalizedError
//	@Failure		401	{object}	normalizederr.NormalizedError
//	@Failure		500	{object}	normalizederr.NormalizedError
func (g AuthGateway) refresh(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		AddActor()

	var input authcase.RefreshInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := authcase.Refresh{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}
