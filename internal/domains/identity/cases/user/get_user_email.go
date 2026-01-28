package usercase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type GetUserEmail struct {
	IdentityRepo identity.Repo
}

type GetUserEmailInput struct {
	TargetID uuid.UUID    `json:"-"`
	Actor    shared.Actor `json:"-"`
}

func (i GetUserEmail) Execute(input GetUserEmailInput) (out htypes.Email, err error) {
	if err := identity.CanReadUserSensitiveData(&input.Actor, input.TargetID); err != nil {
		return out, err
	}

	target, err := i.IdentityRepo.GetUserByID(input.TargetID)
	if err != nil {
		return out, err
	} else if target == nil {
		return out, identity.ErrUserNotFound
	}

	return target.Email, nil
}
