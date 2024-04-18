package appcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/config/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type EditApp struct {
	AuthRepo authcase.AuthRepo
}

type EditAppInput struct {
	ApplicationId uuid.UUID
	auth.ApplicationEditableFields
	Actor auth.Account
}

func (i EditApp) Execute(input EditAppInput) (*auth.Application, error) {
	app, err := i.AuthRepo.GetApplicationById(input.ApplicationId)
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, normalizederr.NewRequestError("Application does not exist", errcode.ApplicationNotFound)
	}

	err = app.Edit(&input.ApplicationEditableFields, input.Actor)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpdateApplication(*app)
	if err != nil {
		return nil, err
	}

	return app, nil
}
