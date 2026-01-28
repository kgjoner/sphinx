package accesshttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	appcase "github.com/kgjoner/sphinx/internal/domains/access/cases/application"
	"github.com/kgjoner/sphinx/internal/shared/api/sharedhttp"
)

func (g Gateway) applicationHandler(r chi.Router) {
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
func (g Gateway) createApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input appcase.CreateApplicationInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.AccessPool.NewDAO(r.Context())
	i := appcase.CreateApplication{
		AccessRepo: queries,
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
func (g Gateway) editApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		ParseURLParam("id", "applicationID").
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input appcase.EditAppInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.AccessPool.NewDAO(r.Context())
	i := appcase.EditApp{
		AccessRepo: queries,
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
func (g Gateway) getApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("id", "applicationID")

	var input appcase.GetAppInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	queries := g.AccessPool.NewDAO(r.Context())
	i := appcase.GetApp{
		AccessRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}
