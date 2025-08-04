package usercase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type GetPrivateUser struct {
	AuthRepo authcase.AuthRepo
}

type GetPrivateUserInput struct {
	Target auth.User `json:"-"`
	Actor  auth.User `json:"-"`
}

func (i GetPrivateUser) Execute(input GetPrivateUserInput) (*auth.UserPrivateView, error) {
	return input.Target.PrivateView(input.Actor)
}
