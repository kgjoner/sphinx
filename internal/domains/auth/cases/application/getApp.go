package appcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type GetApp struct {
	AuthRepo auth.Repo
}

type GetAppInput struct {
	ApplicationID uuid.UUID
}

func (i GetApp) Execute(input GetAppInput) (out auth.ApplicationView, err error) {
	app, err := i.AuthRepo.GetApplicationByID(input.ApplicationID)
	if err != nil {
		return out, err
	} else if app == nil {
		return out, apperr.NewRequestError("Application does not exist", errcode.ApplicationNotFound)
	}

	return app.View(), nil
}
