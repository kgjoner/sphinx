package authcase

import (
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Refresh struct {
	AuthRepo      auth.Repo
	Hasher        shared.DataHasher
	TokenProvider auth.TokenProvider
}

type RefreshInput struct {
	Actor shared.Actor `json:"-"`
	Token string       `json:"-"`
}

func (i Refresh) Execute(input RefreshInput) (out LoginOutput, err error) {
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

	proof, err := shared.VerifyData(session.RefreshToken, input.Token, i.Hasher)
	if err != nil {
		err := i.AuthRepo.TerminateAllSubjectSessions(input.Actor.ID)
		if err != nil {
			return out, err
		}

		return out, err
	}

	err = session.Authenticate(*proof)
	if err != nil {
		return out, err
	}

	sub, err := session.ToSubject()
	if err != nil {
		return out, err
	}

	tokens, err := i.TokenProvider.Generate(*sub)
	if err != nil {
		return out, err
	}

	refreshHash, err := shared.NewHashedData(tokens.RefreshToken, i.Hasher)
	if err != nil {
		return out, err
	}

	session.UpdateRefreshToken(*refreshHash)
	err = i.AuthRepo.UpdateSession(*session)
	if err != nil {
		return out, err
	}

	return LoginOutput{
		UserID:       sub.ID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}
