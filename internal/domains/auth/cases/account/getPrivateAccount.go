package accountcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type GetPrivateAccount struct {
	AuthRepo authcase.AuthRepo
}

type GetPrivateAccountInput struct {
	Id    uuid.UUID
	Actor auth.Account
}

func (i GetPrivateAccount) Execute(input GetPrivateAccountInput) (*auth.AccountPrivateView, error) {
	target := input.Actor
	if input.Id != uuid.Nil && input.Id != input.Actor.Id {
		if !input.Actor.HasRoleOnAuth(auth.RoleValues.ADMIN) {
			return nil, normalizederr.NewForbiddenError("Does not have permission to access this content.")
		}

		target, err := i.AuthRepo.GetAccountById(input.Id)
		if err != nil {
			return nil, err
		} else if target == nil {
			return nil, normalizederr.NewRequestError("Account does not exist", "")
		}
	}

	return target.PrivateView(input.Actor)
}
