package oauthcase

import (
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/repositories/cache"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type ExchangeGrant struct {
	AuthRepo  authcase.AuthRepo
	CacheRepo cache.DAO
}

type ExchangeGrantInput struct {
	auth.GrantCredentials
	auth.SessionCreationFields `json:"-"`
}

func (i ExchangeGrant) Execute(input ExchangeGrantInput) (*authcase.LoginOutput, error) {
	var grant *auth.AuthorizationGrant
	err := i.CacheRepo.GetJSON("grant:"+input.Code, &grant)
	if err != nil {
		if err == cache.ErrNil {
			return nil, normalizederr.NewUnauthorizedError("Invalid code", errcode.InvalidCredentials)
		}
		return nil, err
	}

	// Clear grant from cache independent of outcome
	defer func() { i.CacheRepo.Clear("grant:" + grant.Code) }()

	// Get user by link ID
	acc, err := i.AuthRepo.GetUserByLink(grant.LinkID)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}

	// Authenticate via authorization grant
	err = acc.AuthenticateViaGrant(grant, &input.GrantCredentials)
	if err != nil {
		return nil, err
	}

	// Create session
	input.SessionCreationFields.Application.ID = input.ClientID
	access, refresh, err := acc.InitSession(&input.SessionCreationFields)
	if err != nil {
		return nil, err
	}

	// Persist sessions
	err = i.AuthRepo.UpsertSessions(acc.SessionsToPersist()...)
	if err != nil {
		return nil, err
	}

	return &authcase.LoginOutput{
		UserID:       acc.ID,
		AccessToken:  access.String(),
		RefreshToken: refresh.String(),
		ExpiresIn:    config.Env.JWT.ACCESS_LIFETIME_IN_SEC,
	}, nil
}
