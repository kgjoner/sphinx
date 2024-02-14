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
	Entry string
	Actor auth.Account
}

func (i GetPrivateAccount) Execute(input GetPrivateAccountInput) (*auth.AccountPrivateView, error) {
	target := &input.Actor
	if input.Entry != "" {
		if !input.Actor.HasRoleOnAuth(auth.RoleValues.ADMIN) {
			return nil, normalizederr.NewForbiddenError("Does not have permission to access this content.")
		}

		var err error 
		if id, err := uuid.Parse(input.Entry); err != nil {
			target, err = i.AuthRepo.GetAccountById(id)
		} else {
			target, err = i.AuthRepo.GetAccountByEntry(input.Entry)
		}
		
		if err != nil {
			return nil, err
		} else if target == nil {
			return nil, normalizederr.NewRequestError("Account does not exist", "")
		}
	}

	return target.PrivateView(input.Actor)
}
