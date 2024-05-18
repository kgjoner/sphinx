package authgtw

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/helpers/controller"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/utils/structop"
	appcase "github.com/kgjoner/sphinx/internal/domains/auth/cases/application"
)

func (g AuthGateway) applicationHandler(r chi.Router) {
	r.With(g.mid.Authenticate).Post("/", g.createApplication)
	r.With(g.mid.Authenticate).Patch("/{id}", g.editApplication)

}

// CreateApplication godoc
//
//	@Summary		Create an application
//	@Description	Register a new application that share auth details with this server.
//	@Router			/application [post]
//	@Tags			Application
//	@Security		AppToken
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		auth.ApplicationCreationFields	true	"Name and possible grantings."
//	@Success		200		{object}	presenter.Success[auth.Application]
//	@Failure		400		{object}	normalizederr.NormalizedError
//	@Failure		401		{object}	normalizederr.NormalizedError
//	@Failure		403		{object}	normalizederr.NormalizedError
//	@Failure		500		{object}	normalizederr.NormalizedError
func (g AuthGateway) createApplication(w http.ResponseWriter, r *http.Request) {
	bodyKeys := structop.New(appcase.CreateApplicationInput{}.ApplicationCreationFields).JsonKeys()
	c := controller.New(r).
		ParseBody(bodyKeys...).
		AddActor()

	var input appcase.CreateApplicationInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	i := appcase.CreateApplication{
		AuthRepo: g.AuthRepo.New(r.Context()),
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
//	@Security		AppToken
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
	bodyKeys := structop.New(appcase.EditAppInput{}.ApplicationEditableFields).JsonKeys()
	c := controller.New(r).
		ParseBody(bodyKeys...).
		ParseUrlParam("id", "applicationId").
		AddActor()

	var input appcase.EditAppInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	i := appcase.EditApp{
		AuthRepo: g.AuthRepo.New(r.Context()),
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}
