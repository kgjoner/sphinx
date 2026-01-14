package authgtw

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/common"
	appcase "github.com/kgjoner/sphinx/internal/domains/auth/cases/application"
)

func (g AuthGateway) applicationHandler(r chi.Router) {
	r.Get("/{id}", g.getApplication)

	r.With(g.mid.Authenticate).Post("/", g.createApplication)
	r.With(g.mid.Authenticate).Patch("/{id}", g.editApplication)
}

// CreateApplication godoc
//
//	@Summary		Create an application
//	@Description	Register a new application that share auth details with this server.
//	@Router			/application [post]
//	@Tags			Application
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		auth.ApplicationCreationFields	true	"Name and possible grantings."
//	@Success		200		{object}	presenter.Success[auth.ApplicationView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g AuthGateway) createApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(common.ActorCtxKey, "actor")

	var input appcase.CreateApplicationInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := appcase.CreateApplication{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// EditApplication godoc
//
//	@Summary		Edit an application
//	@Description	Update name or grantings of target application.
//	@Router			/application/{id} [patch]
//	@Tags			Application
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"ID of target application"
//	@Param			payload	body		auth.ApplicationEditableFields	true	"If grantings are passed, the new ones (even if empty array) will overwrite old ones."
//	@Success		200		{object}	presenter.Success[auth.ApplicationView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g AuthGateway) editApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		ParseURLParam("id", "applicationID").
		AddFromContext(common.ActorCtxKey, "actor")

	var input appcase.EditAppInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := appcase.EditApp{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}

// GetApplication godoc
//
//	@Summary		Get application data
//	@Description	Retrieve data from application referred by id
//	@Router			/application/{id} [get]
//	@Tags			Application
//	@Produce		json
//	@Param			id	path		string	true	"ID of target application"
//	@Success		200	{object}	presenter.Success[auth.ApplicationView]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g AuthGateway) getApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("id", "applicationID")

	var input appcase.GetAppInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.BasePool.NewDAO(r.Context())
	i := appcase.GetApp{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}
