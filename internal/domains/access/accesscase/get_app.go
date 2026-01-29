package accesscase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/shared"
)

type GetApp struct {
	AccessRepo access.Repo
}

type GetAppInput struct {
	ApplicationID uuid.UUID
	Actor         shared.Actor `json:"-"`
}

func (i GetApp) Execute(input GetAppInput) (out access.ApplicationView, err error) {
	if err := access.CanReadApplications(&input.Actor); err != nil {
		return out, err
	}

	app, err := i.AccessRepo.GetApplicationByID(input.ApplicationID)
	if err != nil {
		return out, err
	} else if app == nil {
		return out, access.ErrApplicationNotFound
	}

	return app.View(), nil
}
