package accesscase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/access"
	"github.com/kgjoner/sphinx/internal/shared"
)

type EditApp struct {
	AccessRepo access.Repo
}

type EditAppInput struct {
	ApplicationID uuid.UUID
	access.ApplicationEditableFields
	Actor shared.Actor `json:"-"`
}

func (i EditApp) Execute(input EditAppInput) (out access.ApplicationView, err error) {
	if err := access.CanEditApplication(&input.Actor, input.ApplicationID); err != nil {
		return out, err
	}

	app, err := i.AccessRepo.GetApplicationByID(input.ApplicationID)
	if err != nil {
		return out, err
	} else if app == nil {
		return out, access.ErrApplicationNotFound
	}

	err = app.Edit(&input.ApplicationEditableFields)
	if err != nil {
		return out, err
	}

	err = i.AccessRepo.UpdateApplication(*app)
	if err != nil {
		return out, err
	}

	return app.View(), nil
}
