package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/v2/helpers/apperr"
	"github.com/kgjoner/sphinx/internal/common/errcode"
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Login struct {
	AuthRepo auth.Repo
}

type LoginInput struct {
	Entry                      auth.Entry
	Password                   string
	auth.SessionCreationFields `json:"-"`
}

func (i Login) Execute(input LoginInput) (*LoginOutput, error) {
	user, err := i.AuthRepo.GetUserByEntry(input.Entry)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, apperr.NewUnauthorizedError("Invalid credentials", errcode.InvalidCredentials)
	}

	err = user.AuthenticateViaPassword(input.Password)
	if err != nil {
		return nil, err
	}

	app, err := i.AuthRepo.GetApplicationByID(uuid.MustParse(config.Env.ROOT_APP_ID))
	if err != nil {
		return nil, err
	} else if app == nil {
		return nil, apperr.NewRequestError("Root application not found", errcode.ApplicationNotFound)
	}
	input.Application = *app

	access, refresh, err := user.InitSession(&input.SessionCreationFields)
	if err != nil {
		return nil, err
	}

	err = i.AuthRepo.UpsertSessions(user.SessionsToPersist()...)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		UserID:       user.ID,
		AccessToken:  access.String(),
		RefreshToken: refresh.String(),
		ExpiresIn:    config.Env.JWT.ACCESS_LIFETIME_IN_SEC,
	}, nil
}

type LoginOutput struct {
	UserID       uuid.UUID `json:"userID"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int       `json:"expiresIn"`
}
