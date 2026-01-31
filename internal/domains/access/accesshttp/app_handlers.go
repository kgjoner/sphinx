package accesshttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/v2/helpers/controller"
	"github.com/kgjoner/cornucopia/v2/helpers/presenter"
	"github.com/kgjoner/sphinx/internal/domains/access/accesscase"
	"github.com/kgjoner/sphinx/internal/shared/sharedhttp"
)

func (g gateway) applicationHandler(r chi.Router) {
	r.Get("/{id}", g.getApplication)

	r.With(g.Authenticate).Post("/", g.createApplication)
	r.With(g.Authenticate).Patch("/{id}", g.editApplication)
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
//	@Param			payload	body		access.ApplicationCreationFields	true	"Name and possible grantings."
//	@Success		200		{object}	presenter.Success[access.ApplicationView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) createApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input accesscase.CreateApplicationInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	repo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := accesscase.CreateApplication{
		AccessRepo: repo,
		PwHasher:   g.PwHasher,
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
//	@Param			id		path		string								true	"ID of target application"
//	@Param			payload	body		access.ApplicationEditableFields	true	"If grantings are passed, the new ones (even if empty array) will overwrite old ones."
//	@Success		200		{object}	presenter.Success[access.ApplicationView]
//	@Failure		400		{object}	apperr.AppError
//	@Failure		401		{object}	apperr.AppError
//	@Failure		403		{object}	apperr.AppError
//	@Failure		500		{object}	apperr.AppError
func (g gateway) editApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JSONBody().
		ParseURLParam("id", "applicationID").
		AddFromContext(sharedhttp.ActorCtxKey, "actor")

	var input accesscase.EditAppInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	repo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := accesscase.EditApp{
		AccessRepo: repo,
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
//	@Success		200	{object}	presenter.Success[access.ApplicationView]
//	@Failure		400	{object}	apperr.AppError
//	@Failure		500	{object}	apperr.AppError
func (g gateway) getApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseURLParam("id", "applicationID")

	var input accesscase.GetAppInput
	err := c.Write(&input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	repo := g.AccessFactory.NewDAO(r.Context(), g.PGPool.Connection())
	i := accesscase.GetApp{
		AccessRepo: repo,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HTTPError(err, w, r)
		return
	}

	presenter.HTTPSuccess(output, w, r)
}
