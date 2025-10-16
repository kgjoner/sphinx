package authcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Logout struct {
	AuthRepo auth.Repo
}

type LogoutInput struct {
	Actor auth.User `json:"-"`
}

func (i Logout) Execute(input LogoutInput) (out bool, err error) {
	_, err = input.Actor.TerminateAuthedSession()
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.UpsertSessions(input.Actor.SessionsToPersist()...)
	if err != nil {
		return out, err
	}

	return true, nil
}
