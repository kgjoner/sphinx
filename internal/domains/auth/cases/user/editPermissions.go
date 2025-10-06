package usercase

import (
	"database/sql"

	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type EditUserPermissions struct {
	AuthRepo auth.Repo
}

type EditUserPermissionsInput struct {
	Roles        []auth.Role
	ShouldRemove sql.NullBool     `validate:"required"`
	Target       auth.User        `json:"-"`
	Application  auth.Application `json:"-"`
}

func (i EditUserPermissions) Execute(input EditUserPermissionsInput) (bool, error) {
	if !input.ShouldRemove.Valid {
		return false, apperr.NewRequestError("Must inform whether permissions should be added or removed.")
	}

	var err error
	targetAcc := &input.Target

	if input.Roles != nil {
		for _, r := range input.Roles {
			if input.ShouldRemove.Bool {
				err = targetAcc.RemoveRole(r, input.Application)
			} else {
				err = targetAcc.AddRole(r, input.Application)
			}

			if err != nil {
				return false, err
			}
		}
	}

	err = i.AuthRepo.UpsertLinks(targetAcc.LinksToPersist()...)
	if err != nil {
		return false, err
	}

	return true, nil
}
