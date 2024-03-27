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

func (g AuthGateway) createApplication(w http.ResponseWriter, r *http.Request) {
	bodyKeys := structop.New(appcase.CreateApplicationInput{}.ApplicationCreationFields).Keys()
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
		AuthRepo: g.AuthRepo,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}

func (g AuthGateway) editApplication(w http.ResponseWriter, r *http.Request) {
	bodyKeys := structop.New(appcase.EditAppInput{}.ApplicationEditableFields).Keys()
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
		AuthRepo: g.AuthRepo,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r)
}
