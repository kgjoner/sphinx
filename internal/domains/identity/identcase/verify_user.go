package identcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/domains/identity"
)

type VerifyUser struct {
	IdentityRepo identity.Repo
}

type VerifyUserInput struct {
	UserID           uuid.UUID                 `json:"-"`
	VerificationKind identity.VerificationKind `json:"-" validate:"required,oneof=email phone"`
	VerificationCode string                    `json:"code" validate:"required"`
}

func (i VerifyUser) Execute(input VerifyUserInput) (out bool, err error) {
	user, err := i.IdentityRepo.GetUserByID(input.UserID)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, identity.ErrUserNotFound
	}

	err = user.VerifyUser(input.VerificationKind, input.VerificationCode)
	if err != nil {
		return out, err
	}

	err = i.IdentityRepo.UpdateUser(*user)
	if err != nil {
		return out, err
	}

	return true, nil
}
