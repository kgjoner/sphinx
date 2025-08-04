package usercase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type CancelPendingField struct {
	AuthRepo authcase.AuthRepo
}

type CancelPendingFieldInput struct {
	UserID uuid.UUID `json:"-"`
	Field  string    `json:"-"`
}

func (i CancelPendingField) Execute(input CancelPendingFieldInput) (bool, error) {
	acc, err := i.AuthRepo.GetUserByID(input.UserID)
	if err != nil {
		return false, err
	} else if acc == nil {
		return false, normalizederr.NewRequestError("User does not exit", errcode.UserNotFound)
	}

	err = acc.CancelPendingField(input.Field)
	if err != nil {
		return false, err
	}

	err = i.AuthRepo.UpdateUser(*acc)
	if err != nil {
		return false, err
	}

	return true, nil
}
