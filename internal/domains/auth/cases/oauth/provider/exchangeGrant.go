package oauthcase

import (
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/cornucopia/v2/repositories/cache"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type ExchangeGrant struct {
	AuthRepo  auth.Repo
	CacheRepo cache.DAO
}

type ExchangeGrantInput struct {
	auth.GrantCredentials
	auth.SessionCreationFields `json:"-"`
}

func (i ExchangeGrant) Execute(input ExchangeGrantInput) (out authcase.LoginOutput, err error) {
	var grant *auth.AuthorizationGrant
	err = i.CacheRepo.GetJSON("grant:"+input.Code, &grant)
	if err != nil {
		if err == cache.ErrNil {
			return out, apperr.NewUnauthorizedError("Invalid code", errcode.InvalidCredentials)
		}
		return out, err
	}

	// Clear grant from cache independent of outcome
	defer func() { i.CacheRepo.Clear("grant:" + grant.Code) }()

	// Get user by link ID
	user, err := i.AuthRepo.GetUserByLink(grant.LinkID)
	if err != nil {
		return out, err
	} else if user == nil {
		return out, apperr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}

	// Authenticate via authorization grant
	err = user.AuthenticateViaGrant(grant, &input.GrantCredentials)
	if err != nil {
		return out, err
	}

	// Create session
	input.SessionCreationFields.Application.ID = input.ClientID
	access, refresh, err := user.InitSession(&input.SessionCreationFields)
	if err != nil {
		return out, err
	}

	// Persist sessions
	err = i.AuthRepo.UpsertSessions(user.SessionsToPersist()...)
	if err != nil {
		return out, err
	}

	return authcase.LoginOutput{
		UserID:       user.ID,
		AccessToken:  access.String(),
		RefreshToken: refresh.String(),
		ExpiresIn:    config.Env.JWT.ACCESS_LIFETIME_IN_SEC,
	}, nil
}
