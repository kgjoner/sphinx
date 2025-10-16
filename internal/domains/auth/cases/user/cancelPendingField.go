package usercase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type CancelPendingField struct {
	AuthRepo auth.Repo
}

type CancelPendingFieldInput struct {
	UserID uuid.UUID `json:"-"`
	Field  string    `json:"-"`
}

func (i CancelPendingField) Execute(input CancelPendingFieldInput) (out bool, err error) {
	user, err := i.AuthRepo.GetUserByID(input.UserID)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, apperr.NewRequestError("User does not exit", errcode.UserNotFound)
	}

	err = user.CancelPendingField(input.Field)
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.UpdateUser(*user)
	if err != nil {
		return out, err
	}

	return true, nil
}
