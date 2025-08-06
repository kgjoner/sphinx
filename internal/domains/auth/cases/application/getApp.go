package appcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type GetApp struct {
	AuthRepo auth.Repo
}

type GetAppInput struct {
	ApplicationID uuid.UUID
}

func (i GetApp) Execute(input GetAppInput) (*auth.Application, error) {
	app, err := i.AuthRepo.GetApplicationByID(input.ApplicationID)
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, normalizederr.NewRequestError("Application does not exist", errcode.ApplicationNotFound)
	}

	return app, nil
}
