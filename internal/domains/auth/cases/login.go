package authcase

import (
	"github.com/google/uuid"
	"github.com/kgjoner/cornucopia/helpers/normalizederr"
	"github.com/kgjoner/sphinx/internal/domains/auth"
)

type Login struct {
	AuthRepo AuthRepo
}

type LoginInput struct {
	Entry    string
	Password string
	auth.SessionCreationFields `json:"-"`
}

func (i Login) Execute(input LoginInput) (*LoginOutput, error) {
	acc, err := i.AuthRepo.GetAccountByEntry(input.Entry)
	if err != nil {
		return nil, err
	} else if acc == nil {
		return nil, normalizederr.NewUnauthorizedError("Invalid credentials")
	}

	access, refresh, err := acc.InitSession(input.Password, &input.SessionCreationFields)
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

	return &LoginOutput{
		AccountId:    acc.Id,
		AccessToken:  access.String(),
		RefreshToken: refresh.String(),
	}, nil
}

type LoginOutput struct {
	AccountId    uuid.UUID `json:"accountId"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
}
