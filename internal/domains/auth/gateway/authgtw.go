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
	common.Repos
	common.Services
	mid common.Middlewares
}

func Raise(router chi.Router, repos common.Repos, services common.Services) {
	authgtw := &AuthGateway{
		repos,
		services,
		common.Middlewares(repos),
	}

	router.Route("/account", authgtw.accountHandler)
	router.Route("/application", authgtw.applicationHandler)

	router.Route("/auth", func(r chi.Router) {
		r.With(authgtw.mid.AppToken).Post("/login", authgtw.login)
		r.With(authgtw.mid.Authenticate).Post("/logout", authgtw.logout)
		r.With(authgtw.mid.Authenticate).Post("/refresh", authgtw.refresh)
	})
}

// Login godoc
//
//	@Summary		Log user in
//	@Description	Exchange user credentials for auth tokens
//	@Router			/auth/login [post]
//	@Tags			Auth
//	@Security		AppToken
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		authcase.LoginInput	true	"Credentials. Entry can be: email, phone, username or document"
//	@Success		200		{object}	presenter.Success[authcase.LoginOutput]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) login(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseBody("entry", "password").
		AddApplication().
		AddHeader("x-forwarded-for", "ip").
		AddHeader("user-agent", "device")

	var input authcase.LoginInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	i := authcase.Login{
		AuthRepo: g.AuthRepo,
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

	i := authcase.Logout{
		AuthRepo: g.AuthRepo,
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

	i := authcase.Refresh{
		AuthRepo: g.AuthRepo,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}
