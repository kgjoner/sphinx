package usercase

import (
	"github.com/kgjoner/cornucopia/helpers/htypes"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type GetUserEmail struct {
	AuthRepo authcase.AuthRepo
}

type GetUserEmailInput struct {
	Target auth.User `json:"-"`
}

func (i GetUserEmail) Execute(input GetUserEmailInput) (htypes.Email, error) {
	return input.Target.Email, nil
}
