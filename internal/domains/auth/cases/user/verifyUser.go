package usercase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type VerifyUser struct {
	AuthRepo auth.Repo
}

type VerifyUserInput struct {
	UserID           uuid.UUID             `json:"-"`
	VerificationKind auth.VerificationKind `json:"kind" validate:"required,oneof=email phone"`
	VerificationCode string                `json:"code" validate:"required"`
}

func (i VerifyUser) Execute(input VerifyUserInput) (bool, error) {
	user, err := i.AuthRepo.GetUserByID(input.UserID)
	if err != nil {
		return false, err
	} else if user == nil {
		return false, normalizederr.NewRequestError("User does not exit", errcode.UserNotFound)
	}

	err = user.VerifyUser(input.VerificationKind, input.VerificationCode)
	if err != nil {
		return false, err
	}

	err = i.AuthRepo.UpdateUser(*user)
	if err != nil {
		return false, err
	}

	return true, nil
}
