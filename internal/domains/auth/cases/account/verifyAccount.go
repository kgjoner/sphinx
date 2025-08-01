package accountcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type VerifyAccount struct {
	AuthRepo authcase.AuthRepo
}

type VerifyAccountInput struct {
	AccountID        uuid.UUID             `json:"-"`
	VerificationKind auth.VerificationKind `json:"kind" validate:"required,oneof=email phone"`
	VerificationCode string                `json:"code" validate:"required"`
}

func (i VerifyAccount) Execute(input VerifyAccountInput) (bool, error) {
	acc, err := i.AuthRepo.GetAccountByID(input.AccountID)
	if err != nil {
		return false, err
	} else if acc == nil {
		return false, normalizederr.NewRequestError("Account does not exit", errcode.AccountNotFound)
	}

	err = acc.VerifyAccount(input.VerificationKind, input.VerificationCode)
	if err != nil {
		return false, err
	}

	err = i.AuthRepo.UpdateAccount(*acc)
	if err != nil {
		return false, err
	}

	return true, nil
}
