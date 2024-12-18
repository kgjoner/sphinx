package accountcase

import (
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type GetAccountEmail struct {
	AuthRepo authcase.AuthRepo
}

type GetAccountEmailInput struct {
	Target auth.Account
}

func (i GetAccountEmail) Execute(input GetAccountEmailInput) (htypes.Email, error) {
	return input.Target.Email, nil
}
