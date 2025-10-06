package usercase

import (
	"github.com/kgjoner/cornucopia/v2/helpers/htypes"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type GetUserEmail struct {
	AuthRepo auth.Repo
}

type GetUserEmailInput struct {
	Target auth.User `json:"-"`
}

func (i GetUserEmail) Execute(input GetUserEmailInput) (htypes.Email, error) {
	return input.Target.Email, nil
}
