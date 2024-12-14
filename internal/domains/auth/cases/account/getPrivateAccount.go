package accountcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type GetPrivateAccount struct {
	AuthRepo authcase.AuthRepo
}

type GetPrivateAccountInput struct {
	Target auth.Account
	Actor auth.Account
}

func (i GetPrivateAccount) Execute(input GetPrivateAccountInput) (*auth.AccountPrivateView, error) {
	return input.Target.PrivateView(input.Actor)
}
