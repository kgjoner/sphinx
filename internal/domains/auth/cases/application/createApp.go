package appcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type CreateApplication struct {
	AuthRepo auth.Repo
}

type CreateApplicationInput struct {
	auth.ApplicationCreationFields
	Actor auth.User `json:"-"`
}

func (i CreateApplication) Execute(input CreateApplicationInput) (out CreateApplicationOutput, err error) {
	app, secret, err := auth.NewApplication(&input.ApplicationCreationFields, input.Actor)
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.InsertApplication(app)
	if err != nil {
		return out, err
	}

	return CreateApplicationOutput{
		Application: app.View(),
		Secret:      secret,
	}, nil
}

type CreateApplicationOutput struct {
	Application auth.ApplicationView `json:"application"`
	Secret      string               `json:"secret"`
}
