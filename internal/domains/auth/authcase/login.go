package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type Login struct {
	AuthRepo      auth.Repo
	PwHasher      shared.PasswordHasher
	DataHasher    shared.DataHasher
	TokenProvider auth.TokenProvider
}

type LoginInput struct {
	Entry                      shared.Entry
	Password                   string
	auth.SessionCreationFields `json:"-"`
}

func (i Login) Execute(input LoginInput) (out LoginOutput, err error) {
	audID := uuid.MustParse(config.Env.ROOT_APP_ID)
	principal, err := i.AuthRepo.GetPrincipalByEntry(input.Entry, audID)
	if err != nil {
		return out, err
	} else if principal == nil {
		return out, shared.ErrInvalidCredentials
	}

	proof, err := shared.VerifyPassword(principal.Password, input.Password, i.PwHasher)
	if err != nil {
		return out, err
	}

	session, err := auth.NewSession(input.SessionCreationFields, *principal, proof)
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

	refreshHash, err := shared.NewHashedData(tokens.RefreshToken, i.DataHasher)
	if err != nil {
		return out, err
	}

	session.UpdateRefreshToken(*refreshHash)
	err = i.AuthRepo.InsertSession(session)
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

type LoginOutput struct {
	UserID       uuid.UUID `json:"userID"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int       `json:"expiresIn"`
}
