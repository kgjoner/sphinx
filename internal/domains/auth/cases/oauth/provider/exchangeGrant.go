package oauthcase

import (
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/cornucopia/helpers/validator"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
	authcase "github.com/kgjoner/sphinx/internal/domains/auth/cases"
)

type ExchangeGrant struct {
	AuthRepo authcase.AuthRepo
}

type ExchangeGrantInput struct {
	auth.OAuthAuthenticateFields
	auth.SessionCreationFields `json:"-"`
}

func (i ExchangeGrant) Execute(input ExchangeGrantInput) (*authcase.LoginOutput, error) {
	err := validator.Validate(input)
	if err != nil {
		return nil, err
	}

	app, err := i.AuthRepo.GetApplicationById(input.AppId)
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}
	input.Application = *app

	acc, err := i.AuthRepo.GetAccountByOAuthCode(input.Code)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}

	err = acc.AuthenticateViaOAuth(&input.OAuthAuthenticateFields)
	if err != nil {
		if normerr, ok := err.(normalizederr.NormalizedError); ok && normerr.Code == errcode.InvalidCredentials {
			errRepo := i.AuthRepo.UpsertLinks(acc.LinksToPersist()...)
			if errRepo != nil {
				return nil, normalizederr.NewFatalUnauthorizedError(err.Error() + ", and links could not be persisted: " + errRepo.Error())
			}
		}

		return nil, err
	}

	access, refresh, err := acc.InitSession(&input.SessionCreationFields)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertSessions(acc.SessionsToPersist()...)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertLinks(acc.LinksToPersist()...)
	if err != nil {
		return nil, err
	}

	return &authcase.LoginOutput{
		AccountId:    acc.Id,
		AccessToken:  access.String(),
		RefreshToken: refresh.String(),
		ExpiresIn:    config.Env.JWT.ACCESS_LIFETIME_IN_SEC,
	}, nil
}
