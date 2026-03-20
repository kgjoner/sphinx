package authcase

import (
	"github.com/kgjoner/cornucopia/v3/repositories/cache"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	"github.com/kgjoner/sphinx/internal/shared"
)

type ExchangeGrant struct {
	AuthRepo      auth.Repo
	PwHasher      shared.PasswordHasher
	DataHasher    shared.DataHasher
	Challenger    auth.CodeChallenger
	CacheRepo     cache.DAO
	TokenProvider auth.TokenProvider
}

type ExchangeGrantInput struct {
	auth.GrantCredentials
	auth.SessionCreationFields `json:"-"`
}

func (i ExchangeGrant) Execute(input ExchangeGrantInput) (out LoginOutput, err error) {
	var grant *auth.Grant
	err = i.CacheRepo.GetJSON("grant:"+input.Code, &grant)
	if err != nil {
		if err == cache.ErrNil {
			return out, shared.ErrInvalidCredentials
		}
		return out, err
	}

	// Clear grant from cache independent of outcome
	defer func() { i.CacheRepo.Clear("grant:" + grant.Code) }()

	// Resolve Principal and Client
	principal, err := i.AuthRepo.GetPrincipal(grant.SubID, grant.AudID)
	if err != nil {
		return out, err
	} else if principal == nil {
		return out, shared.ErrInvalidCredentials
	}

	client, err := i.AuthRepo.GetClient(grant.ClientID)
	if err != nil {
		return out, err
	} else if client == nil {
		return out, auth.ErrInvalidClient
	}

	// Consume Grant
	var proof *auth.GrantProof
	if input.GrantCredentials.IsConfidentialClient() {
		proof, err = auth.VerifyGrant(*grant, *client, input.GrantCredentials, i.PwHasher)
	} else {
		proof, err = auth.VerifyGrant(*grant, *client, input.GrantCredentials, i.Challenger)
	}

	if err != nil {
		return out, err
	}

	// Create session
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
