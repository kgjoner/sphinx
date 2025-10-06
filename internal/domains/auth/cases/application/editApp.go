package appcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type EditApp struct {
	AuthRepo auth.Repo
}

type EditAppInput struct {
	ApplicationID uuid.UUID
	auth.ApplicationEditableFields
	Actor auth.User `json:"-"`
}

func (i EditApp) Execute(input EditAppInput) (*auth.Application, error) {
	app, err := i.AuthRepo.GetApplicationByID(input.ApplicationID)
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, apperr.NewRequestError("Application does not exist", errcode.ApplicationNotFound)
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
