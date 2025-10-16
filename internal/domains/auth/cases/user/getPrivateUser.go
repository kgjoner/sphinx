package usercase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type GetPrivateUser struct {
	AuthRepo auth.Repo
}

type GetPrivateUserInput struct {
	Target auth.User `json:"-"`
	Actor  auth.User `json:"-"`
}

func (i GetPrivateUser) Execute(input GetPrivateUserInput) (auth.UserPrivateView, error) {
	return input.Target.PrivateView(input.Actor)
}
