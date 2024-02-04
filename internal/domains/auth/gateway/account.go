package authgtw

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kgjoner/cornucopia/helpers/controller"
	"github.com/kgjoner/cornucopia/helpers/presenter"
	"github.com/kgjoner/cornucopia/utils/structop"
	accountcase "github.com/kgjoner/sphinx/internal/domains/auth/cases/account"
)

func (g AuthGateway) accountHandler(r chi.Router) {
	r.With(g.mid.AppToken).Post("/", g.createAccount)	
}

func (g AuthGateway) createAccount(w http.ResponseWriter, r *http.Request) {
	bodyKeys := structop.New(accountcase.CreateAccountInput{}.AccountCreationFields).Keys()
	c := controller.New(r).
		ParseBody(bodyKeys...).
		AddApplication()

	var input accountcase.CreateAccountInput
	err := c.Write(&input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	i := accountcase.CreateAccount{
		AuthRepo: g.AuthRepo,
	}

	output, err := i.Execute(input)
	if err != nil {
		presenter.HttpError(err, w, r)
		return
	}

	presenter.HttpSuccess(output, w, r, http.StatusCreated)
}
