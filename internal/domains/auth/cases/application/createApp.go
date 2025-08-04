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
	Actor auth.User `json:"-"`
}

func (i CreateApplication) Execute(input CreateApplicationInput) (*CreateApplicationOutput, error) {
	app, secret, err := auth.NewApplication(&input.ApplicationCreationFields, input.Actor)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.InsertApplication(app)
	if err != nil {
		return nil, err
	}

	return &CreateApplicationOutput{
		Application: *app,
		Secret:      secret,
	}, nil
}

type CreateApplicationOutput struct {
	Application auth.Application `json:"application"`
	Secret      string           `json:"secret"`
}
