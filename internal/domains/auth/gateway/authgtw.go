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
	mid common.Middlewares
}

func Raise(router chi.Router, repos common.Repos) {
	authgtw := &AuthGateway{
		repos,
		common.Middlewares{
			AuthRepo: repos.AuthRepo,
		},
	}

	router.Route("/account", authgtw.accountHandler)
	router.Route("/application", authgtw.applicationHandler)

	router.Route("/auth", func(r chi.Router) {
		r.With(authgtw.mid.AppToken).Post("/login", authgtw.login)
		r.With(authgtw.mid.Authenticate).Post("/logout", authgtw.logout)
		r.With(authgtw.mid.Authenticate).Post("/refresh", authgtw.refresh)
	})
}

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
