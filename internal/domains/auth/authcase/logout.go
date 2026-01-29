package authcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Logout struct {
	AuthRepo auth.Repo
}

type LogoutInput struct {
	Actor shared.Actor `json:"-"`
}

func (i Logout) Execute(input LogoutInput) (out bool, err error) {
	session, err := i.AuthRepo.GetSessionByID(input.Actor.SessionID)
	if err != nil {
		return out, err
	} else if session == nil {
		return out, auth.ErrSessionNotFound
	}

	err = session.Validate(&input.Actor)
	if err != nil {
		return out, err
	}

	err = session.Terminate()
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.UpdateSession(*session)
	if err != nil {
		return out, err
	}

	return true, nil
}
