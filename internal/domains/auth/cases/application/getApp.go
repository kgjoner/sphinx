package appcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/config/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type GetApp struct {
	AuthRepo authcase.AuthRepo
}

type GetAppInput struct {
	ApplicationId uuid.UUID
}

func (i GetApp) Execute(input GetAppInput) (*auth.Application, error) {
	app, err := i.AuthRepo.GetApplicationById(input.ApplicationId)
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, normalizederr.NewRequestError("Application does not exist", errcode.ApplicationNotFound)
	}

	return app, nil
}
