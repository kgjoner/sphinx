package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/config/errcode"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type LoginViaOAuth struct {
	AuthRepo AuthRepo
}

type LoginViaOAuthInput struct {
	Code                       string
	AppSecret                  string
	auth.SessionCreationFields `json:"-"`
}

func (i LoginViaOAuth) Execute(input LoginViaOAuthInput) (*LoginViaOAuthOutput, error) {
	acc, err := i.AuthRepo.GetAccountByOAuthCode(input.Code)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}

	err = acc.AuthenticateViaOAuth(input.Code, input.Application.Id, input.AppSecret)
	if err != nil {
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

	newLinks := acc.LinksToPersist()
	if len(newLinks) > 0 {
		err = i.AuthRepo.UpsertLinks(newLinks...)
		if err != nil {
			return nil, err
		}
	}

	return &LoginViaOAuthOutput{
		AccountId:    acc.Id,
		AccessToken:  access.String(),
		RefreshToken: refresh.String(),
	}, nil
}

type LoginViaOAuthOutput struct {
	AccountId    uuid.UUID `json:"accountId"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
}
