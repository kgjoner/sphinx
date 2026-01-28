package usercase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type GetPrivateUser struct {
	IdentityRepo identity.Repo
}

type GetPrivateUserInput struct {
	TargetID uuid.UUID    `json:"-"`
	Actor    shared.Actor `json:"-"`
}

func (i GetPrivateUser) Execute(input GetPrivateUserInput) (out identity.UserView, err error) {
	if err = identity.CanReadUserSensitiveData(&input.Actor, input.TargetID); err != nil {
		return out, err
	}

	target, err := i.IdentityRepo.GetUserByID(input.TargetID)
	if err != nil {
		return out, err
	} else if target == nil {
		return out, identity.ErrUserNotFound
	}

	return target.View(), nil
}
