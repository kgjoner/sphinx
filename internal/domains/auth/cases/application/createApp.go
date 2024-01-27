package appcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type CreateApplication struct {
	AuthRepo authcase.AuthRepo
}

type CreateApplicationInput struct {
	auth.ApplicationCreationFields
	Actor auth.Account
}

func (i CreateApplication) Execute(input CreateApplicationInput) (*auth.Application, error) {
	app, err := auth.NewApplication(&input.ApplicationCreationFields, input.Actor)
	if err != nil {
		return nil, err
	}

	_, err = i.AuthRepo.InsertApplication(*app)
	if err != nil {
		return nil, err
	}

	return app, nil
}
