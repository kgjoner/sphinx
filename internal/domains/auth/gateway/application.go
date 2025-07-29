package authgtw

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/helpers/controller"
	"github.com/kgjoner/cornucopia/helpers/presenter"
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
//	@Success		200		{object}	presenter.Success[auth.Application]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		403		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) createApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JsonBody().
		AddActor()

	var input appcase.CreateApplicationInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := appcase.CreateApplication{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
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
//	@Param			id		path		string							true	"Id of target application"
//	@Param			payload	body		auth.ApplicationEditableFields	true	"If grantings are passed, the new ones (even if empty array) will overwrite old ones."
//	@Success		200		{object}	presenter.Success[auth.Application]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		403		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) editApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		JsonBody().
		ParseUrlParam("id", "applicationId").
		AddActor()

	var input appcase.EditAppInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := appcase.EditApp{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

// GetApplication godoc
//
//	@Summary		Get application data
//	@Description	Retrieve data from application referred by id
//	@Router			/application/{id} [get]
//	@Tags			Application
//	@Produce		json
//	@Param			id	path		string	true	"Id of target application"
//	@Success		200	{object}	presenter.Success[auth.Application]
//	@Failure		400	{object}	normalizederr.NormalizedError
//	@Failure		500	{object}	normalizederr.NormalizedError
func (g AuthGateway) getApplication(w http.ResponseWriter, r *http.Request) {
	c := controller.New(r).
		ParseUrlParam("id", "applicationId")

	var input appcase.GetAppInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	queries := g.BasePool.NewQueries(r.Context())
	i := appcase.GetApp{
		AuthRepo: queries,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}
