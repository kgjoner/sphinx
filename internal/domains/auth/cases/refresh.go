package authcase

import (
	"github.com/kgjoner/sphinx/internal/config"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Refresh struct {
	AuthRepo auth.Repo
}

type RefreshInput struct {
	Actor auth.User `json:"-"`
}

func (i Refresh) Execute(input RefreshInput) (out LoginOutput, err error) {
	accessToken, refreshToken, err := input.Actor.IssueNewTokens()
	if err != nil {
		return out, err
	}

	err = i.AuthRepo.UpsertSessions(input.Actor.SessionsToPersist()...)
	if err != nil {
		return out, err
	}

	return LoginOutput{
		UserID:       input.Actor.ID,
		AccessToken:  accessToken.String(),
		RefreshToken: refreshToken.String(),
		ExpiresIn:    config.Env.JWT.ACCESS_LIFETIME_IN_SEC,
	}, nil
}
