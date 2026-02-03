package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type ExternalLogin struct {
	AuthRepo         auth.Repo
	IdentityProvider shared.IdentityProvider
	DataHasher       shared.DataHasher
	TokenProvider    auth.TokenProvider
}

type ExternalLoginInput struct {
	shared.IdentityProviderInput
	auth.SessionCreationFields `json:"-"`
	Languages                  []string `json:"-"`
}

func (i ExternalLogin) Execute(input ExternalLoginInput) (out LoginOutput, err error) {
	extSubject, err := i.IdentityProvider.Authenticate(input.IdentityProviderInput)
	if err != nil {
		return out, err
	} else if extSubject == nil || extSubject.ID == "" || extSubject.Email.IsZero() || extSubject.ProviderName == "" {
		return out, shared.ErrInvalidExternalSubject
	}

	audID := uuid.MustParse(config.Env.ROOT_APP_ID)
	principal, err := i.AuthRepo.GetPrincipalByExtSubID(extSubject.ProviderName, extSubject.ID, audID)
	if err != nil {
		return out, err
	} else if principal == nil {
		return out, auth.ErrNoUserForThisSubject
	}

	proof, err := auth.VerifyExternalLogin(*principal, *extSubject)
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
		ExpiresIn:    config.Env.JWT.ACCESS_LIFETIME_IN_SEC,
	}, nil
}
