package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Login struct {
	AuthRepo AuthRepo
}

type LoginInput struct {
	Entry                      auth.Entry
	Password                   string
	auth.SessionCreationFields `json:"-"`
}

func (i Login) Execute(input LoginInput) (*LoginOutput, error) {
	acc, err := i.AuthRepo.GetAccountByEntry(input.Entry)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}

	err = acc.AuthenticateViaPassword(input.Password)
	if err != nil {
		return nil, err
	}

	app, err := i.AuthRepo.GetApplicationByID(uuid.MustParse(config.Env.ROOT_APP_ID))
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, normalizederr.NewRequestError("Root application not found", errcode.ApplicationNotFound)
	}
	input.Application = *app

	access, refresh, err := acc.InitSession(&input.SessionCreationFields)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertSessions(acc.SessionsToPersist()...)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		AccountID:    acc.ID,
		AccessToken:  access.String(),
		RefreshToken: refresh.String(),
		ExpiresIn:    config.Env.JWT.ACCESS_LIFETIME_IN_SEC,
	}, nil
}

type LoginOutput struct {
	AccountID    uuid.UUID `json:"accountID"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int       `json:"expiresIn"`
}
