package identcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
)

type UpdateExtraData struct {
	IdentityRepo identity.Repo
}

type UpdateExtraDataInput struct {
	identity.UserExtraFields
	TargetID uuid.UUID    `json:"-"`
	Actor    shared.Actor `json:"-"`
}

func (i UpdateExtraData) Execute(input UpdateExtraDataInput) (out identity.UserView, err error) {
	if err := identity.CanUpdateUser(&input.Actor, input.TargetID); err != nil {
		return out, err
	}

	target, err := i.IdentityRepo.GetUserByID(input.TargetID)
	if err != nil {
		return out, err
	} else if target == nil {
		return out, identity.ErrUserNotFound
	}

	err = target.UpdateExtraData(input.UserExtraFields)
	if err != nil {
		return out, err
	}

	err = i.IdentityRepo.UpdateUser(*target)
	if err != nil {
		return out, err
	}

	return target.View(), nil
}
